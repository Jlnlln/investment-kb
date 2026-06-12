package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
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

	// 5. 生成编号
	now := time.Now()
	ids, err := idgen.GenerateIDs(result, now)
	if err != nil {
		return fmt.Errorf("生成编号失败: %w", err)
	}

	// 6. 渲染 Markdown
	rawMD := markdown.RenderRawMaterial(ids, result, string(rawText), now)
	qaMD := markdown.RenderKnowledgeCard(ids, result, now)
	crMD := markdown.RenderCandidateRules(ids, result.CandidateRules)

	var caseMD string
	if result.ShouldGenerateCase && result.Case != nil {
		caseMD = markdown.RenderMarketCase(ids, *result.Case)
	}

	// 7. Dry-run 模式：只打印不写入
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
		Provider:   cfg.AI.Provider,
		Model:      cfg.AI.Model,
		BaseURL:    cfg.AI.BaseURL,
		APIKey:     apiKey,
		TimeoutSec: cfg.AI.TimeoutSec,
		MaxRetries: 3,
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

	return &result, nil
}