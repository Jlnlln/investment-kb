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
	"investment-kb/internal/config"
	"investment-kb/internal/idgen"
	"investment-kb/internal/markdown"
	"investment-kb/internal/model"
	"investment-kb/internal/obsidian"
	"investment-kb/internal/prompt"
)

// ExtractOptions 提取选项
type ExtractOptions struct {
	InputPath  string
	Source     string
	DryRun     bool
	Mock       bool
	ConfigPath string
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
		fmt.Printf("⚠️  检测到可能重复导入：原文哈希已存在 (%s)\n", result.RawHash)
		fmt.Printf("   V1 可以暂时继续导入，后续可支持 --allow-duplicate 控制是否允许重复导入。\n")
	}

	// 5. 生成编号
	now := time.Now()
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	// 6. 渲染 Markdown（在校验通过后才渲染）
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)
	qaMD := markdown.RenderKnowledgeCard(ids, result, now)
	crMD := markdown.RenderCandidateRules(cfg, ids, result, result.CandidateRules)

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

// checkAbsoluteClaims 检查绝对化收益表达
func checkAbsoluteClaims(result *model.ExtractionResult) error {
	claims := []string{
		result.Summary,
		result.CoreConclusion,
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

	claims = append(claims, result.MyUnderstanding)

	// 绝对化收益关键词列表
	absoluteClaims := []string{
		"保证盈利",
		"没有亏损风险",
		"必然上涨",
		"一定赚钱",
		"判断错了也不会亏",
		"只赚不亏",
		"无风险",
		"稳赚",
		"绝对安全",
		"必胜",
	}

	// 检查每个关键词
	for _, claim := range claims {
		for _, keyword := range absoluteClaims {
			if strings.Contains(claim, keyword) {
				return fmt.Errorf("AI 输出包含绝对化收益表达：%s。请调整 Prompt 或人工检查后重试。", keyword)
			}
		}
	}

	return nil
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
// 返回需要打印的 warning 信息列表
func WarnRuleTypeDomainMismatch(result *model.ExtractionResult) []string {
	var warnings []string

	for _, rule := range result.CandidateRules {
		// 只检查 rule_type 为 "买入规则" 的
		if rule.RuleType == "买入规则" {
			// 检查 domain_code 是否为 "BUY"
			if rule.DomainCode != "BUY" {
				warning := fmt.Sprintf("候选规则分类可能不一致：买入规则「%s」的 domain_code=%s，建议检查是否应为 BUY-%s。",
					rule.RuleName,
					rule.DomainCode,
					rule.TopicCode)
				warnings = append(warnings, warning)
			}
		}
	}

	return warnings
}
