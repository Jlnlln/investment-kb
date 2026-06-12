package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/model"
)

// RenderRawMaterial 生成原始材料 RAW Markdown
func RenderRawMaterial(ids *model.DocumentIDs, result *model.ExtractionResult, rawText string, now time.Time) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.RawID, result.Title))

	// 元数据
	sb.WriteString(fmt.Sprintf("来源：%s\n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s\n", formatTags(result.Tags)))
	sb.WriteString("整理状态：已整理\n")
	sb.WriteString(fmt.Sprintf("生成时间：%s\n", now.Format("2006-01-02")))
	if result.RawHash != "" {
		sb.WriteString(fmt.Sprintf("原文哈希：%s\n", result.RawHash))
	}
	sb.WriteString("\n")

	// 对应知识卡片
	sb.WriteString(fmt.Sprintf("对应知识卡片：%s｜%s\n\n", ids.QAID, result.Title))

	// 对应候选规则
	sb.WriteString("对应候选规则：\n\n")
	for i, crID := range ids.CandidateIDs {
		if i < len(result.CandidateRules) {
			sb.WriteString(fmt.Sprintf("- %s｜%s\n", crID, result.CandidateRules[i].RuleName))
		}
	}
	sb.WriteString("\n")

	// 对应案例
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("对应案例：%s｜%s\n\n", ids.CaseID, result.Case.CaseName))
	} else {
		sb.WriteString("对应案例：暂不单独生成市场案例\n\n")
	}

	// 案例说明
	if ids.CaseID == "" && result.CaseInsufficientReason != "" {
		sb.WriteString("案例说明：")
		sb.WriteString(result.CaseInsufficientReason)
		sb.WriteString("\n\n")
	}

	// 原文
	sb.WriteString("## 原文\n\n")
	sb.WriteString(rawText)
	sb.WriteString("\n\n")

	return sb.String()
}

// formatTags 格式化标签列表
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, " / ")
}