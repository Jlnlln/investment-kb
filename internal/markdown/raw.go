package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// RenderRawMaterial 生成原始材料 RAW Markdown
func RenderRawMaterial(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, rawText string, now time.Time) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.RawID, result.Title))

	// 元数据
	sb.WriteString(fmt.Sprintf("来源：%s  \n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s  \n", formatTags(result.Tags)))
	sb.WriteString("整理状态：已整理  \n")
	sb.WriteString(fmt.Sprintf("生成时间：%s  \n", now.Format("2006-01-02")))
	if result.RawHash != "" {
		sb.WriteString(fmt.Sprintf("原文哈希：%s  \n", result.RawHash))
	}
	sb.WriteString("\n")

	// 对应知识卡片/宏观理解卡/市场观察卡（根据 material_type 动态生成）
	materialType := string(result.MaterialType)
	if materialType == "" {
		materialType = "rule_candidate"
	}

	switch materialType {
	case "rule_candidate":
		// 规则型材料：链接到 QA 卡
		if ids.QAID != "" {
			sb.WriteString(fmt.Sprintf("对应知识卡片：%s\n\n", ObsidianHeadingLink(GetQaPath(cfg), JoinHeading(ids.QAID, result.Title), JoinHeading(ids.QAID, result.Title))))
		}
	case "macro_knowledge":
		// 宏观理解型材料：链接到 KNOW 卡
		if ids.KNOWID != "" {
			knowPath := "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/宏观理解卡库.md" // TODO: 从 config 读取
			sb.WriteString(fmt.Sprintf("对应宏观理解卡：%s\n\n", ObsidianHeadingLink(knowPath, JoinHeading(ids.KNOWID, result.Title), JoinHeading(ids.KNOWID, result.Title))))
		}
	case "market_observation":
		// 市场观察型材料：链接到 OBS 卡
		if ids.OBSID != "" {
			obsPath := "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/市场观察卡库.md" // TODO: 从 config 读取
			sb.WriteString(fmt.Sprintf("对应市场观察卡：%s\n\n", ObsidianHeadingLink(obsPath, JoinHeading(ids.OBSID, result.Title), JoinHeading(ids.OBSID, result.Title))))
		}
	}

	// 对应候选规则（仅 rule_candidate 类型显示）
	if materialType == "rule_candidate" {
		sb.WriteString("对应候选规则：\n\n")
		for i, crID := range ids.CandidateIDs {
			if i < len(result.CandidateRules) {
				rule := result.CandidateRules[i]
				heading := JoinCandidateRuleHeading(crID, rule.DomainCode, rule.TopicCode, rule.RuleName)
				sb.WriteString(fmt.Sprintf("- %s\n", ObsidianHeadingLink(GetCandidateRulePath(cfg), heading, heading)))
			}
		}
		sb.WriteString("\n")
	}

	// 对应案例
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("对应案例：%s\n\n", ObsidianHeadingLink(GetMarketCasePath(cfg), JoinHeading(ids.CaseID, result.Case.CaseName), JoinHeading(ids.CaseID, result.Case.CaseName))))
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
