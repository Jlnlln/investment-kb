package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"investment-kb/internal/ai"
	"investment-kb/internal/classify"
	"investment-kb/internal/config"
	"investment-kb/internal/dedup"
	"investment-kb/internal/idgen"
	"investment-kb/internal/markdown"
	"investment-kb/internal/model"
	"investment-kb/internal/obsidian"
	"investment-kb/internal/prompt"
)

// ExtractOptions 提取选项
type ExtractOptions struct {
	InputPath      string
	Source         string
	DryRun         bool
	Mock           bool
	AllowDuplicate bool
	ConfigPath     string
}

// 哈希管理（内联）
var (
	hashes      = make(map[string]bool)
	hashesMutex sync.RWMutex
	hashesPath  = "data/import_hashes.json"
)

func loadHashes() error {
	hashesMutex.Lock()
	defer hashesMutex.Unlock()

	data, err := os.ReadFile(hashesPath)
	if err != nil {
		if os.IsNotExist(err) {
			hashes = make(map[string]bool)
			return nil
		}
		return fmt.Errorf("读取哈希文件失败: %w", err)
	}

	if len(data) == 0 {
		hashes = make(map[string]bool)
		return nil
	}

	if err := json.Unmarshal(data, &hashes); err != nil {
		return fmt.Errorf("解析哈希文件失败: %w", err)
	}

	return nil
}

func checkHash(hash string) bool {
	hashesMutex.RLock()
	defer hashesMutex.RUnlock()

	return hashes[hash]
}

func saveHash(hash string) error {
	hashesMutex.Lock()
	defer hashesMutex.Unlock()

	hashes[hash] = true

	// saveAllHashesLocked 不再自行加锁，因为当前已持有锁
	return saveAllHashesLocked()
}

// saveAllHashesLocked 将哈希记录写入磁盘。
// 调用方必须已持有 hashesMutex（写锁）。
func saveAllHashesLocked() error {
	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(hashesPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	data, err := json.MarshalIndent(hashes, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化哈希记录失败: %w", err)
	}

	if err := os.WriteFile(hashesPath, data, 0644); err != nil {
		return fmt.Errorf("写入哈希文件失败: %w", err)
	}

	return nil
}

// Extract 执行提取流程
func Extract(opts *ExtractOptions) error {
	// 1. 读取输入文件
	rawText, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	// 2. 加载配置
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 3. 获取 ExtractionResult
	var result *model.ExtractionResult
	if opts.Mock {
		result = model.MockExtractionResult()
		fmt.Printf("🧪 使用 Mock 数据\n\n")
	} else {
		// 调用 AI 获取结果
		fmt.Printf("🤖 正在调用 AI...\n")
		result, err = callAI(cfg, string(rawText), opts.Source)
		if err != nil {
			return fmt.Errorf("AI 调用失败: %w", err)
		}
		fmt.Printf("✅ AI 返回结果\n\n")
	}

	// 4. 计算原文 sha256 hash（用于去重）
	hash := sha256.Sum256(rawText)
	result.RawHash = fmt.Sprintf("%x", hash)

	// 4.1 加载已导入的哈希记录
	if err := loadHashes(); err != nil {
		return fmt.Errorf("加载已导入哈希记录失败: %w", err)
	}

	// 4.2 检查是否重复导入
	if checkHash(result.RawHash) {
		if !opts.AllowDuplicate {
			return fmt.Errorf("检测到重复导入：原文哈希已存在 (%s)。如需强制导入，请加 --allow-duplicate", result.RawHash)
		}
		fmt.Printf("⚠️  检测到重复导入：原文哈希已存在 (%s)，因启用 --allow-duplicate，继续写入\n", result.RawHash)
	}

	// 4.5 候选规则去重：同一篇材料下，rule_name 相同则只保留第一条
	result.CandidateRules = deduplicateCandidateRules(result.CandidateRules)

	// 4.6 领域分类映射：AI domain_code 为"建议"，程序做二次分类
	fmt.Printf("🔍 领域分类映射...\n")
	for i := range result.CandidateRules {
		rule := &result.CandidateRules[i]
		// 先做基础映射（BUY→VALUATION 等）
		mappedDomain := idgen.MapCRDomain(rule.DomainCode)
		// 保存 AI 原始分类
		rule.OriginalDomainCode = rule.DomainCode
		// 程序二次分类
		finalDomain := classify.ClassifyDomainWithLog(*rule, mappedDomain)
		rule.DomainCode = finalDomain
	}
	// 同步映射顶层 domain_code
	topMappedDomain := idgen.MapCRDomain(result.DomainCode)
	topFinalDomain := classify.ClassifyDomainWithLog(model.CandidateRule{
		RuleName:          result.Title,
		RuleContent:       result.CoreConclusion,
		TriggerConditions: result.ApplicableScenarios,
		Actions:           result.RiskBoundaries,
		DomainCode:        result.DomainCode,
	}, topMappedDomain)
	result.DomainCode = topFinalDomain

	// 4.7 跨文章相似规则检查
	fmt.Printf("🔍 跨文章相似规则检查...\n")
	crLibraryPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.CandidateRule)
	existingCRs, err := dedup.ParseExistingCRs(crLibraryPath)
	if err != nil {
		fmt.Printf("⚠️  读取候选规则库失败，跳过相似检查: %v\n", err)
		existingCRs = nil
	}
	var similarResults []map[int][]dedup.SimilarRule // 每条新 CR 对应的相似规则
	if existingCRs != nil {
		similarResults = make([]map[int][]dedup.SimilarRule, 0)
		for i, rule := range result.CandidateRules {
			similarRules := dedup.CheckSimilarRules(
				rule.DomainCode, rule.TopicCode, rule.RuleName,
				rule.TriggerConditions, rule.Actions,
				existingCRs,
			)
			if len(similarRules) > 0 {
				for _, sr := range similarRules {
					fmt.Printf("   🔗 %s 与已有规则 %s 相似（%s）\n", rule.RuleName, sr.CRID, sr.Level)
				}
			}
			// 暂时用 slice 存储，后面渲染时使用
			similarResults = append(similarResults, map[int][]dedup.SimilarRule{i: similarRules})
		}
	}

	// 5. 生成编号
	now := time.Now()
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	// 6. 渲染 Markdown（在校验通过后才渲染）
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)
	qaMD := markdown.RenderKnowledgeCard(cfg, ids, result, now)
	// 准备每条 CR 的相似规则数据
	similarData := prepareSimilarData(similarResults, len(result.CandidateRules))
	crMD := markdown.RenderCandidateRules(cfg, ids, result, result.CandidateRules, similarData)

	var caseMD string
	if result.ShouldGenerateCase && result.Case != nil {
		caseMD = markdown.RenderMarketCase(cfg, ids, result, *result.Case)
	}

	// 7. Dry-run 模式：执行校验，通过后才打印
	if opts.DryRun {
		fmt.Printf("=== RAW ===\n\n%s\n", rawMD)
		fmt.Printf("=== QA ===\n\n%s\n", qaMD)
		fmt.Printf("=== CR ===\n\n%s\n", crMD)
		if caseMD != "" {
			fmt.Printf("=== CASE ===\n\n%s\n", caseMD)
		}
		return nil
	}

	// 8. 写入 Obsidian（全部成功后才保存编号状态）
	fmt.Printf("📝 正在写入 Obsidian...\n")

	if err := obsidian.AppendMarkdown(cfg.ObsidianVaultPath, cfg.Files.RawMaterial, rawMD); err != nil {
		return fmt.Errorf("写入原始材料失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.RawID)

	if err := obsidian.AppendMarkdown(cfg.ObsidianVaultPath, cfg.Files.QA, qaMD); err != nil {
		return fmt.Errorf("写入知识卡片失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.QAID)

	if err := obsidian.AppendMarkdown(cfg.ObsidianVaultPath, cfg.Files.CandidateRule, crMD); err != nil {
		return fmt.Errorf("写入候选规则失败: %w", err)
	}

	// 生成规则验证卡草稿
	for i, crID := range ids.CandidateIDs {
		if i >= len(result.CandidateRules) {
			break
		}
		rule := result.CandidateRules[i]
		var ruleSimilarRules []dedup.SimilarRule
		if i < len(similarData) {
			ruleSimilarRules = similarData[i]
		}
		vcContent, vcRelativePath := markdown.RenderValidationCard(cfg, crID, ids.QAID, ids.RawID, result, rule, ruleSimilarRules)
		vcFullPath := filepath.Join(cfg.ObsidianVaultPath, vcRelativePath)
		if err := os.MkdirAll(filepath.Dir(vcFullPath), 0755); err != nil {
			return fmt.Errorf("创建验证卡目录失败: %w", err)
		}
		if err := os.WriteFile(vcFullPath, []byte(vcContent), 0644); err != nil {
			return fmt.Errorf("写入验证卡失败 %s: %w", crID, err)
		}
		fmt.Printf("   ✅ %s 验证卡\n", crID)
	}

	for _, crID := range ids.CandidateIDs {
		fmt.Printf("   ✅ %s\n", crID)
	}

	if caseMD != "" {
		if err := obsidian.AppendMarkdown(cfg.ObsidianVaultPath, cfg.Files.MarketCase, caseMD); err != nil {
			return fmt.Errorf("写入市场案例失败: %w", err)
		}
		fmt.Printf("   ✅ %s\n", ids.CaseID)
	} else {
		fmt.Printf("   ⚠️  未生成市场案例（原因：%s）\n", result.CaseInsufficientReason)
	}

	// 9. Markdown 全部写入成功，保存编号状态
	if err := idgen.SaveState(); err != nil {
		return fmt.Errorf("保存编号状态失败: %w", err)
	}

	// 10. 保存哈希记录到 import_hashes.json
	if err := saveHash(result.RawHash); err != nil {
		return fmt.Errorf("保存哈希记录失败: %w", err)
	}

	fmt.Printf("\n✅ 完成\n")
	return nil
}

// deduplicateCandidateRules 对同一篇材料下的候选规则去重
// 去重规则：rule_name 相同则只保留第一条（保留顺序）
func deduplicateCandidateRules(rules []model.CandidateRule) []model.CandidateRule {
	if len(rules) <= 1 {
		return rules
	}
	seen := make(map[string]bool, len(rules))
	result := make([]model.CandidateRule, 0, len(rules))
	for _, rule := range rules {
		key := rule.RuleName
		if seen[key] {
			fmt.Printf("   ⚠️  去重：跳过重复候选规则「%s」\n", rule.RuleName)
			continue
		}
		seen[key] = true
		result = append(result, rule)
	}
	return result
}

// callAI 调用 AI 获取 ExtractionResult
func callAI(cfg *config.Config, rawText, source string) (*model.ExtractionResult, error) {
	// 创建 AI 客户端
	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("未设置 API Key（环境变量：%s）", cfg.AI.APIKeyEnv)
	}

	client, err := ai.NewClient(&ai.Config{
		Provider:    cfg.AI.Provider,
		Model:       cfg.AI.Model,
		BaseURL:     cfg.AI.BaseURL,
		APIKey:      apiKey,
		TimeoutSec:  cfg.AI.TimeoutSec,
		MaxRetries:  3,
		Temperature: cfg.AI.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 AI 客户端失败: %w", err)
	}

	// 加载 Prompt
	systemPrompt, err := prompt.Load("prompts/qa_extract_prompt.md")
	if err != nil {
		return nil, fmt.Errorf("加载 Prompt 失败: %w", err)
	}

	// 构造用户输入
	userPrompt := fmt.Sprintf("来源：%s\n\n%s", source, rawText)

	// 调用 AI 并解析结果
	var result model.ExtractionResult
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.AI.TimeoutSec)*time.Second)
	defer cancel()

	if err := client.CompleteJSON(ctx, systemPrompt, userPrompt, &result); err != nil {
		return nil, err
	}

	// 校验 AI 输出
	if err := validateExtractionResult(&result, rawText, source); err != nil {
		return nil, err
	}

	return &result, nil
}

// validateExtractionResult 校验 ExtractionResult
func validateExtractionResult(result *model.ExtractionResult, rawText, source string) error {
	// 1. CASE 校验
	if !result.ShouldGenerateCase {
		if result.Case != nil {
			return fmt.Errorf("ShouldGenerateCase=false 时，Case 不能为非 nil")
		}
		if result.CaseInsufficientReason == "" {
			return fmt.Errorf("ShouldGenerateCase=false 时，CaseInsufficientReason 不能为空")
		}
	} else {
		if result.Case == nil {
			return fmt.Errorf("ShouldGenerateCase=true 时，Case 不能为 nil")
		}
		case_ := *result.Case
		if case_.CaseName == "" {
			return fmt.Errorf("Case.CaseName 不能为空")
		}
		if case_.DomainCode == "" {
			return fmt.Errorf("Case.DomainCode 不能为空")
		}
		if case_.TopicCode == "" {
			return fmt.Errorf("Case.TopicCode 不能为空")
		}
		if case_.KeyDecisionQuestion == "" {
			return fmt.Errorf("Case.KeyDecisionQuestion 不能为空")
		}
		if case_.FinalInsight == "" {
			return fmt.Errorf("Case.FinalInsight 不能为空")
		}
	}

	// 2. 禁止表达检查（硬性校验）
	if err := ai.ContainsForbiddenPhrasesInResult(result); err != nil {
		return err
	}

	// 3. 禁止绝对化收益表达检查
	if err := checkAbsoluteClaims(result); err != nil {
		return err
	}

	// 4. candidate_rules 类型集中 warning
	warnOnConsistentRuleTypes(result.CandidateRules)

	// 5. 买入规则 domain_code 检查（warning，不终止）
	warnings := WarnRuleTypeDomainMismatch(result)
	for _, warning := range warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}

	// 6. my_understanding 空值检查（warning，不终止）
	if result.MyUnderstanding == "" {
		fmt.Printf("⚠️  my_understanding 为空，已在 Markdown 中使用「待补充。」\n")
	}

	return nil
}

// absoluteClaimKeywords 绝对化收益关键词列表
var absoluteClaimKeywords = []string{
	"保证盈利",
	"保证上涨",
	"没有亏损风险",
	"必然上涨",
	"一定上涨",
	"一定赚钱",
	"判断错了也不会亏",
	"只赚不亏",
	"无风险",
	"稳赚不赔",
	"稳赚",
	"绝对安全",
	"必胜",
	"直接满仓",
	"满仓买入",
	"应该满仓",
	"必须满仓",
	"可以满仓",
	"应直接满仓",
	"可直接满仓",
	"高确定性时可直接满仓",
	"梭哈",
	"全仓押注",
	"全仓买入",
	"一把梭哈",
}

// negationMarkers 否定标记（按长度降序，优先匹配长词，避免短词误截断）
var negationMarkers = []string{
	"不等于", "并非", "不代表", "不意味着", "不能保证", "不能", "不建议", "不要", "不可", "不会",
	"没有", "不存在", "不是", "无法", "无须", "无需", "避免", "杜绝", "禁止", "严禁", "而非", "勿", "别",
	"不",
}

// checkAbsoluteClaims 检查绝对化收益表达（支持否定语境放行）
func checkAbsoluteClaims(result *model.ExtractionResult) error {
	claims := []string{
		result.Summary,
		result.CoreConclusion,
		result.MyUnderstanding,
	}

	for _, logic := range result.CoreLogic {
		claims = append(claims, logic.Title, logic.Content)
	}

	for _, boundary := range result.RiskBoundaries {
		claims = append(claims, boundary)
	}

	for _, rule := range result.CandidateRules {
		claims = append(claims,
			rule.RuleContent,
			rule.RiskBoundary,
		)
		claims = append(claims, rule.Actions...)
		claims = append(claims, rule.TriggerConditions...)
		claims = append(claims, rule.NotApplicable...)
	}

	for _, claim := range claims {
		if err := checkAbsoluteClaimInText(claim); err != nil {
			return err
		}
	}

	return nil
}

// checkAbsoluteClaimInText 检查单条文本中的绝对化收益表达
// 对每个危险词查找所有出现位置，取前后 24 个 rune 的上下文（不含关键词本身），
// 若上下文中存在否定标记则放行，否则返回 hard error。
func checkAbsoluteClaimInText(text string) error {
	if text == "" {
		return nil
	}

	runes := []rune(text)
	for _, keyword := range absoluteClaimKeywords {
		keywordRunes := []rune(keyword)
		if len(keywordRunes) > len(runes) {
			continue
		}

		for i := 0; i <= len(runes)-len(keywordRunes); i++ {
			// 匹配关键词
			match := true
			for j := 0; j < len(keywordRunes); j++ {
				if runes[i+j] != keywordRunes[j] {
					match = false
					break
				}
			}
			if !match {
				continue
			}

			// 取关键词前后的上下文（不含关键词本身），按 rune 计数
			beforeStart := i - 24
			if beforeStart < 0 {
				beforeStart = 0
			}
			before := runes[beforeStart:i]

			afterEnd := i + len(keywordRunes) + 24
			if afterEnd > len(runes) {
				afterEnd = len(runes)
			}
			after := runes[i+len(keywordRunes):afterEnd]

			// 检查上下文是否包含否定标记
			context := string(before) + string(after)
			if !hasNegationMarker(context) {
				return fmt.Errorf("AI 输出包含绝对化收益表达：%s。请调整 Prompt 或人工检查后重试。", keyword)
			}
		}
	}
	return nil
}

// hasNegationMarker 检查文本中是否包含否定标记
func hasNegationMarker(context string) bool {
	for _, marker := range negationMarkers {
		if strings.Contains(context, marker) {
			return true
		}
	}
	return false
}

// warnOnConsistentRuleTypes 检查 candidate_rules 类型是否过于集中
func warnOnConsistentRuleTypes(rules []model.CandidateRule) {
	if len(rules) < 3 {
		return
	}

	// 收集所有类型
	typeCounts := make(map[string]int)
	for _, rule := range rules {
		typeCounts[rule.RuleType]++
	}

	// 找出出现次数最多的类型
	var maxCount int
	for _, count := range typeCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	// 如果某个类型占比超过 50%，打印 warning
	if maxCount >= len(rules)/2 {
		fmt.Printf("⚠️  候选规则全部为同一类型，请检查是否遗漏仓位规则、风控规则或账户适配规则。\n")
	}
}

// WarnRuleTypeDomainMismatch 检查买入规则的 domain_code 是否匹配
// 新系统中买入规则映射到 VALUATION，但旧分类 BUY 也兼容
// 返回需要打印的 warning 信息列表
func WarnRuleTypeDomainMismatch(result *model.ExtractionResult) []string {
	var warnings []string

	validBuyDomains := map[string]bool{
		"BUY":       true,
		"VALUATION": true,
	}

	for _, rule := range result.CandidateRules {
		// 只检查 rule_type 为 "买入规则" 的
		if rule.RuleType == "买入规则" {
			// 检查 domain_code 是否为有效买入领域
			if !validBuyDomains[rule.DomainCode] {
				warning := fmt.Sprintf("候选规则分类可能不一致：买入规则「%s」的 domain_code=%s，建议检查是否应为 VALUATION-%s 或 BUY-%s。",
					rule.RuleName,
					rule.DomainCode,
					rule.TopicCode,
					rule.TopicCode)
				warnings = append(warnings, warning)
			}
		}
	}

	return warnings
}

// prepareSimilarData 将 map 格式的相似结果转换为按索引排列的 slice
func prepareSimilarData(similarResults []map[int][]dedup.SimilarRule, ruleCount int) [][]dedup.SimilarRule {
	data := make([][]dedup.SimilarRule, ruleCount)
	for _, m := range similarResults {
		for idx, rules := range m {
			if idx < ruleCount {
				data[idx] = rules
			}
		}
	}
	return data
}
