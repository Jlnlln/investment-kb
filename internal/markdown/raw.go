package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// ValidateRawConsistency 校验 RAW 标题与原文内容的一致性
// 如果标题关键词在原文中完全不出现，返回警告信息
func ValidateRawConsistency(result *model.ExtractionResult, rawText string) []string {
	var warnings []string
	
	title := result.Title
	
	// 简化校验：检查标题中的连续子串（4-6个字符）是否出现在原文中
	// 对于中文，提取滑动窗口子串
	if len(title) >= 4 {
		found := false
		// 提取标题中的 4-6 字符子串（滑动窗口）
		for i := 0; i <= len(title)-4; i++ {
			substr := title[i:i+4]
			// 跳过纯标点或数字
			if isMeaningful(substr) && strings.Contains(rawText, substr) {
				found = true
				break
			}
		}
		if !found {
			warnings = append(warnings, fmt.Sprintf("⚠️  标题关键词与原文内容可能存在错配（标题：%s）", title))
			warnings = append(warnings, "   建议检查输入文件是否与内容匹配（mock 模式下可能出现此警告）")
		}
	}
	
	return warnings
}

// isMeaningful 检查字符串是否包含有意义的内容（非纯标点/数字）
func isMeaningful(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r >= 0x4e00 && r <= 0x9fff) {
			return true
		}
	}
	return false
}

// extractTitleKeywords 从标题中提取关键词
func extractTitleKeywords(title string) []string {
	// 按常见分隔符分割
	separators := []string{"｜", "|", "与", "和", "对", "的", "如何", "为什么", "怎么"}
	processed := title
	for _, sep := range separators {
		processed = strings.ReplaceAll(processed, sep, " ")
	}
	
	words := strings.Fields(processed)
	var keywords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) >= 2 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}



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
	// source_file 用于追溯原始输入文件，防止内容错配
	if ids.SourceFile != "" {
		sb.WriteString(fmt.Sprintf("来源文件：%s  \n", ids.SourceFile))
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
		sb.WriteString("对应候选规则：\n\n")
		for i, crID := range ids.CandidateIDs {
			if i < len(result.CandidateRules) {
				rule := result.CandidateRules[i]
				heading := JoinCandidateRuleHeading(crID, rule.DomainCode, rule.TopicCode, rule.RuleName)
				sb.WriteString(fmt.Sprintf("- %s\n", ObsidianHeadingLink(GetCandidateRulePath(cfg), heading, heading)))
			}
		}
		sb.WriteString("\n")
	case "macro_knowledge":
		// 宏观理解型材料：链接到 KNOW 卡（单文件模式，直接 WikiLink）
		if ids.KNOWID != "" {
			sb.WriteString(fmt.Sprintf("对应宏观理解卡：[[%s｜%s]]\n\n", ids.KNOWID, result.Title))
		}
		sb.WriteString("对应知识卡片：不生成\n")
		sb.WriteString("对应候选规则：不生成\n")
		sb.WriteString("对应规则验证卡：不生成\n\n")
		if result.NoRuleReason != "" {
			sb.WriteString(fmt.Sprintf("不生成规则原因：%s\n\n", result.NoRuleReason))
		}
	case "market_observation":
		// 市场观察型材料：链接到 OBS 卡（单文件模式，直接 WikiLink）
		if ids.OBSID != "" {
			sb.WriteString(fmt.Sprintf("对应市场观察卡：[[%s｜%s]]\n\n", ids.OBSID, result.Title))
		}
		sb.WriteString("对应知识卡片：不生成\n")
		sb.WriteString("对应候选规则：不生成\n")
		sb.WriteString("对应规则验证卡：不生成\n\n")
	case "archive_only":
		// 仅存档：标注全部不生成
		sb.WriteString("对应知识卡片：不生成\n")
		sb.WriteString("对应候选规则：不生成\n\n")
	}

	// 对应候选规则（仅 rule_candidate 类型显示——已在上面处理）

	// 对应案例（仅 rule_candidate 类型显示）
	if materialType == "rule_candidate" {
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
