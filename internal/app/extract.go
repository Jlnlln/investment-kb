package app

import (
	"bytes"
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
	MockIndex      int    // Mock 数据变体编号（默认 1）
	ForceType      string // 强制指定材料类型，跳过 AI 判断
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
		return fmt.Errorf("读取哈希文件失败 %s: %w", hashesPath, err)
	}

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(data)) == 0 {
		hashes = make(map[string]bool)
		return nil
	}

	loaded := make(map[string]bool)
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("解析哈希文件失败 %s: %w", hashesPath, err)
	}
	hashes = loaded

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
	rawBytes, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}
	originalText := string(rawBytes)
	rawHash := hashString(originalText)

	// 1.1 输入清洗：去掉 [!tip] 使用说明 后的内容
	cleanedText := cleanRawText(originalText)
	cleanedHash := hashString(cleanedText)
	rawText := []byte(cleanedText)

	// 2. 加载配置
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 2.1 同步编号状态（扫描实际目录，防止 id_state.json 不一致）
	idgen.SyncStateFromDisk(cfg.ObsidianVaultPath, cfg.Files.MacroKnowledgeDir)

	// 3. 获取 ExtractionResult
	var result *model.ExtractionResult
	if opts.Mock {
		// 根据 ForceType 和 MockIndex 选择对应的 Mock 数据
		if opts.ForceType == "macro_knowledge" {
			if opts.MockIndex == 2 {
				result = model.MockMacroKnowledgeResult2()
				fmt.Printf("🧪 使用 Mock 数据（macro_knowledge variant 2）\n\n")
			} else {
				result = model.MockMacroKnowledgeResult()
				fmt.Printf("🧪 使用 Mock 数据（macro_knowledge）\n\n")
			}
		} else {
			result = model.MockExtractionResult()
			fmt.Printf("🧪 使用 Mock 数据（rule_candidate）\n\n")
		}
	} else {
		// 调用 AI 获取结果
		fmt.Printf("🤖 正在调用 AI...\n")
		result, err = callAI(cfg, string(rawText), opts.Source)
		if err != nil {
			return fmt.Errorf("AI 调用失败: %w", err)
		}
		fmt.Printf("✅ AI 返回结果\n\n")
	}

	// 3.1 如果指定了 ForceType，覆盖 material_type（跳过 AI 判断）
	if opts.ForceType != "" {
		result.MaterialType = model.MaterialType(opts.ForceType)
		// 同步更新 Generate* 标志，确保提取函数行为正确
		switch model.MaterialType(opts.ForceType) {
		case model.MaterialTypeRuleCandidate:
			result.GenerateQA = true
			result.GenerateCandidateRules = true
			result.GenerateValidationCards = true
			result.GenerateKnowledgeCard = false
			result.GenerateObservationCard = false
		case model.MaterialTypeMacroKnowledge:
			result.GenerateQA = false
			result.GenerateCandidateRules = false
			result.GenerateValidationCards = false
			result.GenerateKnowledgeCard = true
			result.GenerateObservationCard = false
		case model.MaterialTypeMarketObservation:
			result.GenerateQA = false
			result.GenerateCandidateRules = false
			result.GenerateValidationCards = false
			result.GenerateKnowledgeCard = false
			result.GenerateObservationCard = true
		case model.MaterialTypeArchiveOnly:
			result.GenerateQA = false
			result.GenerateCandidateRules = false
			result.GenerateValidationCards = false
			result.GenerateKnowledgeCard = false
			result.GenerateObservationCard = false
		}
		fmt.Printf("⚠️  强制指定材料类型：%s（跳过 AI 判断）\n\n", opts.ForceType)
	}

	// 4. 计算清洗后文本 hash（用于去重），同时保留原始 hash 用于追溯
	result.RawHash = cleanedHash
	result.SourceMeta = model.SourceMeta{
		SourceFile:   opts.InputPath,
		RawHash:      rawHash,
		CleanedHash:  cleanedHash,
		MaterialType: result.MaterialType,
	}

	// 4.1 加载已导入的哈希记录
	if err := loadHashes(); err != nil {
		return fmt.Errorf("加载已导入哈希记录失败: %w", err)
	}

	// 4.2 检查是否重复导入
	if checkHash(result.RawHash) && !opts.AllowDuplicate {
		return fmt.Errorf("检测到重复导入：清洗后原文哈希已存在 (%s)。如需强制导入，请使用 -allow-duplicate。", result.RawHash)
	}

	// 4.5 根据 material_type 路由处理
	now := time.Now() // 定义时间戳
	var materialType string
	if result.MaterialType != "" {
		materialType = string(result.MaterialType)
	} else {
		// 兼容旧版本：如果没有 material_type 字段，默认为 rule_candidate
		materialType = "rule_candidate"
		result.MaterialType = model.MaterialTypeRuleCandidate
		result.GenerateQA = true
		result.GenerateCandidateRules = true
		result.GenerateValidationCards = true
	}

	result.SourceMeta.MaterialType = result.MaterialType

	if opts.Mock {
		if err := validateMockInputBinding(opts, result); err != nil {
			return err
		}
	}
	if err := enforceRawConsistency(result, string(rawText), opts); err != nil {
		return err
	}

	fmt.Printf("📋 材料类型：%s\n", materialType)

	// 路由处理
	switch materialType {
	case "rule_candidate":
		// 原有逻辑：生成 RAW + QA + CR + 验证卡
		return extractRuleCandidate(opts, cfg, result, rawText, now)
	case "macro_knowledge":
		// 生成 RAW + KNOW 卡
		return extractMacroKnowledge(opts, cfg, result, rawText, now)
	case "market_observation":
		// 生成 RAW + OBS 卡
		return extractMarketObservation(opts, cfg, result, rawText, now)
	case "archive_only":
		// 仅生成 RAW
		return extractArchiveOnly(opts, cfg, result, rawText, now)
	default:
		return fmt.Errorf("未知的材料类型：%s", materialType)
	}
}

// extractRuleCandidate 处理规则型材料（原有逻辑）
func extractRuleCandidate(opts *ExtractOptions, cfg *config.Config, result *model.ExtractionResult, rawText []byte, now time.Time) error {
	// 4.6 候选规则去重：同一篇材料下，rule_name 相同则只保留第一条
	result.CandidateRules = deduplicateCandidateRules(result.CandidateRules)

	// 4.7 领域分类映射：AI domain_code 为"建议"，程序做二次分类
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

	// 4.8 跨文章相似规则检查
	fmt.Printf("🔍 跨文章相似规则检查...\n")
	existingCRs, err := loadExistingCandidateRules(cfg)
	if err != nil {
		fmt.Printf("⚠️  读取候选规则失败，跳过相似检查: %v\n", err)
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
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	// 设置来源文件（用于追溯和一致性校验）
	ids.SourceFile = opts.InputPath
	result.SourceMeta.RawID = ids.RawID

	// 6. 渲染 Markdown
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)
	qaMD := markdown.RenderKnowledgeCard(cfg, ids, result, now)

	// 准备每条 CR 的相似规则数据
	similarData := prepareSimilarData(similarResults, len(result.CandidateRules))
	crMD := markdown.RenderCandidateRules(cfg, ids, result, result.CandidateRules, similarData)

	var caseMD string
	if result.ShouldGenerateCase && result.Case != nil {
		caseMD = markdown.RenderMarketCase(cfg, ids, result, *result.Case)
	}

	// 7. Dry-run 模式
	if opts.DryRun {
		fmt.Printf("=== RAW ===\n\n%s\n", rawMD)
		fmt.Printf("=== QA ===\n\n%s\n", qaMD)
		fmt.Printf("=== CR ===\n\n%s\n", crMD)
		if caseMD != "" {
			fmt.Printf("=== CASE ===\n\n%s\n", caseMD)
		}
		return nil
	}

	// 8. 写入 Obsidian
	fmt.Printf("📝 正在写入 Obsidian...\n")

	if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.RawMaterial, rawMD, ids.RawID); err != nil {
		return fmt.Errorf("写入原始材料失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.RawID)

	if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.QA, qaMD, ids.QAID); err != nil {
		return fmt.Errorf("写入知识卡片失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.QAID)

	if markdown.UseStandaloneCandidateRules(cfg) {
		if err := writeCandidateRuleFiles(cfg, ids, result, similarData); err != nil {
			return err
		}
		if err := markdown.UpdateCandidateRuleIndex(cfg); err != nil {
			return fmt.Errorf("更新候选规则索引失败: %w", err)
		}
	} else {
		if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.CandidateRule, crMD, firstID(ids.CandidateIDs)); err != nil {
			return fmt.Errorf("写入候选规则失败: %w", err)
		}
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
		if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.MarketCase, caseMD, ids.CaseID); err != nil {
			return fmt.Errorf("写入市场案例失败: %w", err)
		}
		fmt.Printf("   ✅ %s\n", ids.CaseID)
	} else {
		fmt.Printf("   ⚠️  未生成市场案例（原因：%s）\n", result.CaseInsufficientReason)
	}

	// 9. 保存状态
	if err := idgen.SaveState(); err != nil {
		return fmt.Errorf("保存编号状态失败: %w", err)
	}

	// 10. 清理孤立验证卡
	cleanOrphanValidationCards(cfg)

	// 11. 保存哈希记录
	if err := saveHash(result.RawHash); err != nil {
		return fmt.Errorf("保存哈希记录失败: %w", err)
	}

	fmt.Printf("\n✅ 完成（规则型材料）\n")
	return nil
}

// extractMacroKnowledge 处理宏观理解型材料
func extractMacroKnowledge(opts *ExtractOptions, cfg *config.Config, result *model.ExtractionResult, rawText []byte, now time.Time) error {
	// 生成编号（KNOW 卡使用特殊前缀）
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	// 设置来源文件（用于追溯和一致性校验）
	ids.SourceFile = opts.InputPath
	result.SourceMeta.RawID = ids.RawID

	// 渲染 Markdown
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)
	knowMD := markdown.RenderKnowCard(cfg, ids, result, now)

	// Dry-run 模式
	if opts.DryRun {
		fmt.Printf("=== RAW ===\n\n%s\n", rawMD)
		fmt.Printf("=== KNOW ===\n\n%s\n", knowMD)
		return nil
	}

	// 写入 Obsidian
	fmt.Printf("📝 正在写入 Obsidian...\n")

	if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.RawMaterial, rawMD, ids.RawID); err != nil {
		return fmt.Errorf("写入原始材料失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.RawID)

	// 写入 KNOW 卡（单文件模式：每张 KNOW 是独立的 .md 文件）
	knowRelativePath := markdown.GetKnowRelativePath(cfg, ids.KNOWID, result.Title)
	knowFullPath := filepath.Join(cfg.ObsidianVaultPath, knowRelativePath)
	if err := os.MkdirAll(filepath.Dir(knowFullPath), 0755); err != nil {
		return fmt.Errorf("创建 KNOW 卡目录失败: %w", err)
	}
	if err := os.WriteFile(knowFullPath, []byte(knowMD), 0644); err != nil {
		return fmt.Errorf("写入宏观理解卡失败: %w", err)
	}
	fmt.Printf("   ✅ %s（独立文件：%s）\n", ids.KNOWID, knowRelativePath)

	// KNOW 卡相似去重检查（轻量级）
	var layer, topic string
	if ids.KNOWID != "" {
		parts := strings.SplitN(ids.KNOWID, "-", 4)
		if len(parts) >= 3 {
			layer = parts[1]
			topic = parts[2]
		}
	}
	similarWarnings := markdown.CheckSimilarKnowCards(cfg.ObsidianVaultPath, cfg.Files.MacroKnowledgeDir, ids.KNOWID, result.Title, layer, topic)
	if len(similarWarnings) > 0 {
		fmt.Printf("   ⚠️  知似去重提示：\n")
		for _, w := range similarWarnings {
			fmt.Printf("      %s\n", w)
		}
		// 在 KNOW 卡文件中追加相似理解卡提示（不自动删除，由人工决定）
		similarSection := renderSimilarKnowSection(similarWarnings)
		if err := appendSimilarKnowSection(knowFullPath, similarSection); err != nil {
			fmt.Printf("      ⚠️  写入相似提示失败: %v\n", err)
		}
	}

	// 更新宏观理解卡索引
	if err := markdown.UpdateKnowIndex(cfg); err != nil {
		fmt.Printf("   ⚠️  更新索引失败: %v\n", err)
	}

	// 保存状态
	if err := idgen.SaveState(); err != nil {
		return fmt.Errorf("保存编号状态失败: %w", err)
	}

	// 清理孤立验证卡
	cleanOrphanValidationCards(cfg)

	// 保存哈希记录
	if err := saveHash(result.RawHash); err != nil {
		return fmt.Errorf("保存哈希记录失败: %w", err)
	}

	fmt.Printf("\n✅ 完成（宏观理解型材料）\n")
	return nil
}

// extractMarketObservation 处理市场状态观察型材料
func extractMarketObservation(opts *ExtractOptions, cfg *config.Config, result *model.ExtractionResult, rawText []byte, now time.Time) error {
	// 生成编号（OBS 卡使用特殊前缀）
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	ids.SourceFile = opts.InputPath
	result.SourceMeta.RawID = ids.RawID

	// 渲染 Markdown
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)

	// OBS 卡渲染（待完整实现，暂时只写 RAW）
	_ = ids.OBSID // 预留 OBS 卡 ID

	// Dry-run 模式
	if opts.DryRun {
		fmt.Printf("=== RAW ===\n\n%s\n", rawMD)
		fmt.Printf("⚠️  OBS 卡生成功能待完整实现\n")
		return nil
	}

	// 写入 Obsidian
	fmt.Printf("📝 正在写入 Obsidian...\n")

	if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.RawMaterial, rawMD, ids.RawID); err != nil {
		return fmt.Errorf("写入原始材料失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.RawID)

	fmt.Printf("   ⚠️  OBS 卡生成功能待完整实现（预留 ID：%s）\n", ids.OBSID)

	// 保存状态
	if err := idgen.SaveState(); err != nil {
		return fmt.Errorf("保存编号状态失败: %w", err)
	}

	// 保存哈希记录
	if err := saveHash(result.RawHash); err != nil {
		return fmt.Errorf("保存哈希记录失败: %w", err)
	}

	fmt.Printf("\n✅ 完成（市场状态观察型材料）\n")
	return nil
}

// extractArchiveOnly 处理仅存档材料
func extractArchiveOnly(opts *ExtractOptions, cfg *config.Config, result *model.ExtractionResult, rawText []byte, now time.Time) error {
	// 生成编号
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	ids.SourceFile = opts.InputPath
	result.SourceMeta.RawID = ids.RawID

	// 渲染 Markdown（仅 RAW）
	rawMD := markdown.RenderRawMaterial(cfg, ids, result, string(rawText), now)

	// Dry-run 模式
	if opts.DryRun {
		fmt.Printf("=== RAW ===\n\n%s\n", rawMD)
		return nil
	}

	// 写入 Obsidian（仅 RAW）
	fmt.Printf("📝 正在写入 Obsidian...\n")

	if _, err := obsidian.AppendMarkdownIfMissing(cfg.ObsidianVaultPath, cfg.Files.RawMaterial, rawMD, ids.RawID); err != nil {
		return fmt.Errorf("写入原始材料失败: %w", err)
	}
	fmt.Printf("   ✅ %s\n", ids.RawID)

	// 保存状态
	if err := idgen.SaveState(); err != nil {
		return fmt.Errorf("保存编号状态失败: %w", err)
	}

	// 保存哈希记录
	if err := saveHash(result.RawHash); err != nil {
		return fmt.Errorf("保存哈希记录失败: %w", err)
	}

	fmt.Printf("\n✅ 完成（仅存档材料）\n")
	return nil
}

func hashString(text string) string {
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash)
}

func expectedMockInputPath(opts *ExtractOptions) string {
	if opts.ForceType == "macro_knowledge" {
		if opts.MockIndex == 2 {
			return filepath.Join("testdata", "inputs", "know_revenue_income.md")
		}
		return filepath.Join("testdata", "inputs", "know_rate.md")
	}
	return filepath.Join("testdata", "inputs", "rule_safety_margin.md")
}

func validateMockInputBinding(opts *ExtractOptions, result *model.ExtractionResult) error {
	expected := filepath.Clean(expectedMockInputPath(opts))
	actual := filepath.Clean(opts.InputPath)
	if !strings.EqualFold(actual, expected) && !strings.EqualFold(filepath.Base(actual), filepath.Base(expected)) {
		return fmt.Errorf("mock 输入文件不匹配：mock-index=%d material_type=%s 需要使用 %s，实际为 %s", opts.MockIndex, result.MaterialType, expected, opts.InputPath)
	}
	return nil
}

func enforceRawConsistency(result *model.ExtractionResult, rawText string, opts *ExtractOptions) error {
	warnings := markdown.ValidateRawConsistency(result, rawText)
	if len(warnings) == 0 {
		return nil
	}
	fmt.Printf("\n⚠️  一致性校验警告：\n")
	for _, w := range warnings {
		fmt.Printf("   %s\n", w)
	}
	// 一致性校验仅警告，不阻断执行（校验逻辑尚未成熟，存在误报）
	return nil
}

// cleanRawText 清洗原始文本：去掉 YAML frontmatter 和模板使用说明
func cleanRawText(text string) string {
	// 1. 去掉模板的 YAML frontmatter（只保留正文部分）
	lines := strings.Split(text, "\n")
	var cleaned []string
	inFrontmatter := false
	frontmatterSeen := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 第一次遇到 --- 进入 frontmatter
		if !frontmatterSeen && trimmed == "---" {
			inFrontmatter = true
			frontmatterSeen = true
			continue
		}
		// 仍在 frontmatter 中，遇到第二个 --- 结束
		if inFrontmatter && trimmed == "---" {
			inFrontmatter = false
			continue
		}
		if inFrontmatter {
			continue // 跳过 frontmatter 里的内容
		}
		cleaned = append(cleaned, line)
	}

	result := strings.Join(cleaned, "\n")

	// 2. 截断 [!tip] 使用说明 后面的内容
	cutMarkers := []string{
		"> [!tip] 使用说明",
		">[!tip] 使用说明",
	}
	for _, marker := range cutMarkers {
		if idx := strings.Index(result, marker); idx >= 0 {
			result = result[:idx]
		}
	}

	return strings.TrimSpace(result)
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
// 对每个危险词查找所有出现位置，取前后 48 个 rune 的上下文（不含关键词本身），
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
			beforeStart := i - 48
			if beforeStart < 0 {
				beforeStart = 0
			}
			before := runes[beforeStart:i]

			afterEnd := i + len(keywordRunes) + 48
			if afterEnd > len(runes) {
				afterEnd = len(runes)
			}
			after := runes[i+len(keywordRunes) : afterEnd]

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

func firstID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func loadExistingCandidateRules(cfg *config.Config) ([]dedup.RuleFingerprint, error) {
	if markdown.UseStandaloneCandidateRules(cfg) {
		return dedup.ParseExistingCRDir(filepath.Join(cfg.ObsidianVaultPath, markdown.GetCandidateRuleDir(cfg)))
	}
	return dedup.ParseExistingCRs(filepath.Join(cfg.ObsidianVaultPath, cfg.Files.CandidateRule))
}

func writeCandidateRuleFiles(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, similarData [][]dedup.SimilarRule) error {
	for i, crID := range ids.CandidateIDs {
		if i >= len(result.CandidateRules) {
			break
		}
		rule := result.CandidateRules[i]
		var ruleSimilarRules []dedup.SimilarRule
		if i < len(similarData) {
			ruleSimilarRules = similarData[i]
		}
		content := markdown.RenderCandidateRuleFile(cfg, ids, result, rule, crID, ruleSimilarRules)
		relativePath := markdown.CandidateRuleRelativePath(cfg, crID, rule.DomainCode, rule.TopicCode, rule.RuleName)
		fullPath := filepath.Join(cfg.ObsidianVaultPath, relativePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("创建候选规则目录失败: %w", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("写入候选规则失败 %s: %w", crID, err)
		}
		fmt.Printf("   ✅ %s（独立文件：%s）\n", crID, relativePath)
	}
	return nil
}

// cleanOrphanValidationCards 清理孤立的验证卡（候选规则库中不存在的）
func cleanOrphanValidationCards(cfg *config.Config) {
	vcDir := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.ValidationCardDir)

	crIDs := make(map[string]bool)
	if markdown.UseStandaloneCandidateRules(cfg) {
		crDir := filepath.Join(cfg.ObsidianVaultPath, markdown.GetCandidateRuleDir(cfg))
		entries, err := os.ReadDir(crDir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || !strings.HasPrefix(entry.Name(), "CR-") {
				continue
			}
			id := strings.TrimSuffix(entry.Name(), ".md")
			if idx := strings.Index(id, "｜"); idx > 0 {
				id = id[:idx]
			}
			crIDs[id] = true
		}
	} else {
		crPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.CandidateRule)
		crData, err := os.ReadFile(crPath)
		if err != nil {
			return
		}
		lines := strings.Split(string(crData), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# CR-") {
				id := strings.TrimPrefix(line, "# ")
				if idx := strings.Index(id, "｜"); idx > 0 {
					id = id[:idx]
				}
				crIDs[id] = true
			}
		}
	}

	// 扫描验证卡目录，删除不在 CR 列表中的
	entries, err := os.ReadDir(vcDir)
	if err != nil {
		return // 目录不存在，跳过
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 提取文件名（去掉 .md 后缀）
		name := entry.Name()
		vcID := strings.TrimSuffix(name, ".md")

		// 检查是否在 CR 列表中
		if !crIDs[vcID] {
			vcPath := filepath.Join(vcDir, name)
			if err := os.Remove(vcPath); err == nil {
				fmt.Printf("   🧹 清理孤立验证卡：%s\n", vcID)
			}
		}
	}
}

// renderSimilarKnowSection 渲染相似理解卡提示区段
func renderSimilarKnowSection(warnings []string) string {
	var sb strings.Builder
	sb.WriteString("\n## 相似理解卡\n\n")
	for _, w := range warnings {
		sb.WriteString(fmt.Sprintf("- %s\n", w))
	}
	sb.WriteString("\n处理建议：\n")
	sb.WriteString("- [ ] 独立保留\n")
	sb.WriteString("- [ ] 合并到已有 KNOW\n")
	sb.WriteString("- [ ] 作为已有 KNOW 的补充材料\n")
	sb.WriteString("- [ ] 废弃\n")
	sb.WriteString("\n")
	return sb.String()
}

// appendSimilarKnowSection 将相似理解卡提示追加到 KNOW 文件末尾
func appendSimilarKnowSection(filePath string, section string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 KNOW 文件失败: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(section); err != nil {
		return fmt.Errorf("写入相似提示失败: %w", err)
	}
	return nil
}
