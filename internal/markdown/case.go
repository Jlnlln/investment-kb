package markdown

import (
	"fmt"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// RenderMarketCase 生成市场案例 CASE Markdown
func RenderMarketCase(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, c model.MarketCase) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.CaseID, c.CaseName))

	// 元数据
	sb.WriteString(fmt.Sprintf("来源材料：%s\n", RawMaterialLink(cfg, ids.RawID, result.Title, ids.RawID)))
	sb.WriteString(fmt.Sprintf("关联知识卡片：%s\n", QaLink(cfg, ids.QAID, result.Title, ids.QAID)))
	sb.WriteString(fmt.Sprintf("主题标签：%s\n\n", formatTags([]string{
		c.DomainCode, c.TopicCode,
	})))

	// 分隔线
	sb.WriteString("---\n\n")

	// 1. 时间 / 背景
	sb.WriteString("## 1. 时间 / 背景\n\n")
	sb.WriteString(c.TimeBackground)
	sb.WriteString("\n\n")

	// 2. 涉及资产
	sb.WriteString("---\n\n")
	sb.WriteString("## 2. 涉及资产\n\n")
	for _, asset := range c.Assets {
		sb.WriteString(fmt.Sprintf("- %s\n", asset))
	}
	sb.WriteString("\n")

	// 3. 当时市场状态
	sb.WriteString("---\n\n")
	sb.WriteString("## 3. 当时市场状态\n\n")
	sb.WriteString(c.MarketStatus)
	sb.WriteString("\n\n")

	// 4. 当时的关键决策问题
	sb.WriteString("---\n\n")
	sb.WriteString("## 4. 当时的关键决策问题\n\n")
	sb.WriteString(c.KeyDecisionQuestion)
	sb.WriteString("\n\n")

	// 5. 不同应对方案
	sb.WriteString("---\n\n")
	sb.WriteString("## 5. 不同应对方案\n\n")
	for i, solution := range c.AlternativeSolutions {
		sb.WriteString(fmt.Sprintf("### 5.%d 方案 %d\n\n", i+1, i+1))
		sb.WriteString(solution)
		sb.WriteString("\n\n")
	}

	// 6. 最终启发
	sb.WriteString("---\n\n")
	sb.WriteString("## 6. 最终启发\n\n")
	sb.WriteString(c.FinalInsight)
	sb.WriteString("\n\n")

	// 7. 可提炼规则
	sb.WriteString("---\n\n")
	sb.WriteString("## 7. 可提炼规则\n\n")
	for _, ruleID := range c.ExtractedRules {
		sb.WriteString(fmt.Sprintf("- %s\n", ruleID))
	}
	sb.WriteString("\n")

	return sb.String()
}
