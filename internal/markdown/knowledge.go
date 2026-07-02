package markdown

import (
	"fmt"
	"investment-kb/internal/config"
	"investment-kb/internal/model"
	"strings"
	"time"
)

// RenderKnowCard 渲染宏观理解卡（单文件模式）
// 每张 KNOW 卡是独立的 .md 文件，拥有完整 frontmatter
func RenderKnowCard(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, now time.Time) string {
	var sb strings.Builder

	// 解析 layer 和 topic
	var layer, topic string
	if ids.KNOWID != "" {
		parts := strings.SplitN(ids.KNOWID, "-", 4)
		if len(parts) >= 3 {
			layer = parts[1]
			topic = parts[2]
		}
	}

	// YAML frontmatter（每张 KNOW 卡独立拥有，Obsidian 可识别）
	fmt.Fprintf(&sb, "---\n")
	fmt.Fprintf(&sb, "uid: %s\n", ids.KNOWID)
	fmt.Fprintf(&sb, "title: %s\n", result.Title)
	fmt.Fprintf(&sb, "source: %s\n", result.Source)
	fmt.Fprintf(&sb, "material_type: macro_knowledge\n")
	fmt.Fprintf(&sb, "layer: %s\n", layer)
	fmt.Fprintf(&sb, "topic: %s\n", topic)
	// tags 用 YAML 列表格式，Obsidian 才能识别
	fmt.Fprintf(&sb, "tags:\n")
	for _, tag := range result.Tags {
		fmt.Fprintf(&sb, "  - %s\n", tag)
	}
	fmt.Fprintf(&sb, "created: %s\n", now.Format("2006-01-02"))
	fmt.Fprintf(&sb, "---\n\n")

	// 标题（用｜分隔，与 RAW/CR 格式一致）
	fmt.Fprintf(&sb, "# %s｜%s\n\n", ids.KNOWID, result.Title)

	// 原始材料链接（RAW 在聚合文件中，需要锚点链接）
	rawPath := GetRawMaterialPath(cfg)
	rawHeading := JoinHeading(ids.RawID, result.Title)
	fmt.Fprintf(&sb, "原始材料：%s\n\n", ObsidianHeadingLink(rawPath, rawHeading, ids.RawID+"｜"+result.Title))

	// 核心结论
	fmt.Fprintf(&sb, "## 核心结论\n\n")
	fmt.Fprintf(&sb, "%s\n\n", result.CoreConclusion)

	// 核心逻辑
	fmt.Fprintf(&sb, "## 核心逻辑\n\n")
	for i, logic := range result.CoreLogic {
		fmt.Fprintf(&sb, "### %d.%d %s\n\n", 1, i+1, logic.Title)
		fmt.Fprintf(&sb, "%s\n\n", logic.Content)
	}

	// 可复用理解（优先渲染 ReusableUnderstanding 列表，如果没有则渲染 Summary）
	if len(result.ReusableUnderstanding) > 0 {
		fmt.Fprintf(&sb, "## 可复用理解\n\n")
		for _, understanding := range result.ReusableUnderstanding {
			fmt.Fprintf(&sb, "- %s\n", understanding)
		}
		fmt.Fprintf(&sb, "\n")
	} else if result.Summary != "" {
		// 兼容旧版本：如果没有 ReusableUnderstanding，渲染 Summary
		fmt.Fprintf(&sb, "## 可复用理解\n\n")
		fmt.Fprintf(&sb, "%s\n\n", result.Summary)
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

	return sb.String()
}
