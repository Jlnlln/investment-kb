package markdown

import (
	"fmt"
	"investment-kb/internal/config"
	"investment-kb/internal/model"
	"strings"
	"time"
)

// RenderKnowCard 渲染宏观理解卡
func RenderKnowCard(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, now time.Time) string {
	var sb strings.Builder

	// YAML frontmatter
	fmt.Fprintf(&sb, "---\n")
	fmt.Fprintf(&sb, "uid: %s\n", ids.KNOWID)
	fmt.Fprintf(&sb, "title: %s\n", result.Title)
	fmt.Fprintf(&sb, "source: %s\n", result.Source)
	fmt.Fprintf(&sb, "material_type: macro_knowledge\n")
	// tags 用 YAML 列表格式，Obsidian 才能识别
	fmt.Fprintf(&sb, "tags:\n")
	for _, tag := range result.Tags {
		fmt.Fprintf(&sb, "  - %s\n", tag)
	}
	fmt.Fprintf(&sb, "created: %s\n", now.Format("2006-01-02"))
	fmt.Fprintf(&sb, "---\n\n")

	// 标题
	fmt.Fprintf(&sb, "# %s %s\n\n", ids.KNOWID, result.Title)

	// 核心结论
	fmt.Fprintf(&sb, "## 核心结论\n\n")
	fmt.Fprintf(&sb, "%s\n\n", result.CoreConclusion)

	// 核心逻辑
	fmt.Fprintf(&sb, "## 核心逻辑\n\n")
	for _, logic := range result.CoreLogic {
		fmt.Fprintf(&sb, "### %s\n\n", logic.Title)
		fmt.Fprintf(&sb, "%s\n\n", logic.Content)
	}

	// 适用场景
	if len(result.ApplicableScenarios) > 0 {
		fmt.Fprintf(&sb, "## 适用场景\n\n")
		for _, scene := range result.ApplicableScenarios {
			fmt.Fprintf(&sb, "- %s\n", scene)
		}
		fmt.Fprintf(&sb, "\n")
	}

	// 风险边界
	if len(result.RiskBoundaries) > 0 {
		fmt.Fprintf(&sb, "## 风险边界\n\n")
		for _, boundary := range result.RiskBoundaries {
			fmt.Fprintf(&sb, "- %s\n", boundary)
		}
		fmt.Fprintf(&sb, "\n")
	}

	// 不生成规则的原因
	if result.NoRuleReason != "" {
		fmt.Fprintf(&sb, "## 不生成规则的原因\n\n")
		fmt.Fprintf(&sb, "%s\n\n", result.NoRuleReason)
	}

	// 我的理解
	fmt.Fprintf(&sb, "## 我的理解\n\n")
	if result.MyUnderstanding != "" {
		fmt.Fprintf(&sb, "%s\n", result.MyUnderstanding)
	} else {
		fmt.Fprintf(&sb, "待补充。\n")
	}
	fmt.Fprintf(&sb, "\n")

	// 原始材料链接
	fmt.Fprintf(&sb, "---\n\n")
	fmt.Fprintf(&sb, "**原始材料**：[[%s]]\n", ids.RawID)

	return sb.String()
}
