package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// ValidateRawConsistency 校验 RAW 标题、材料类型和原文内容的一致性。
func ValidateRawConsistency(result *model.ExtractionResult, rawText string) []string {
	var warnings []string

	if strings.TrimSpace(result.Title) == "" || strings.TrimSpace(rawText) == "" {
		warnings = append(warnings, "标题或原文为空，无法完成一致性校验")
		return warnings
	}

	titleKeywords := coreKeywords(result.Title)
	if !hasEnoughKeywordMatches(rawText, titleKeywords) {
		warnings = append(warnings, fmt.Sprintf("标题核心关键词与原文内容不匹配（标题：%s，关键词：%s）", result.Title, strings.Join(titleKeywords, "/")))
	}

	if result.MaterialType == model.MaterialTypeMacroKnowledge {
		topicKeywords := macroTopicKeywords(result)
		if len(topicKeywords) > 0 && !containsAnyKeyword(rawText, topicKeywords) {
			warnings = append(warnings, fmt.Sprintf("macro_knowledge 核心主题关键词未在原文中出现（主题关键词：%s）", strings.Join(topicKeywords, "/")))
		}
	}

	return warnings
}

func containsAnyKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func hasEnoughKeywordMatches(text string, keywords []string) bool {
	keywords = uniqueStrings(keywords)
	if len(keywords) == 0 {
		return false
	}
	matched := 0
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			matched++
		}
	}
	if len(keywords) <= 2 {
		return matched >= 1
	}
	return matched >= 2 || float64(matched)/float64(len(keywords)) >= 0.5
}

var investmentKeywordDict = []string{
	"安全边际", "创业板", "买入机会", "科创", "沪深", "宽基", "周期", "极限", "差值", "估值", "仓位", "账户", "风险", "利率", "通胀", "消费", "收入", "利润", "内卷", "配置", "风控", "踏空",
}

var genericTitleWords = map[string]bool{
	"参考": true, "理解": true, "逻辑": true, "关系": true, "影响": true, "观察": true, "判断": true, "问题": true, "分析": true, "思考": true,
	"如何": true, "为什么": true, "怎么": true, "以及": true, "之间": true, "平衡": true,
}

func coreKeywords(text string) []string {
	text = stripDocumentPrefix(text)
	parts := splitTitleParts(text)
	seen := make(map[string]bool)
	var keywords []string

	addKeyword := func(keyword string) {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" || genericTitleWords[keyword] || !isChinesePhrase(keyword) || seen[keyword] {
			return
		}
		seen[keyword] = true
		keywords = append(keywords, keyword)
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || genericTitleWords[part] || !isChinesePhrase(part) {
			continue
		}
		matchedDict := false
		for _, word := range investmentKeywordDict {
			if strings.Contains(part, word) {
				addKeyword(word)
				matchedDict = true
			}
		}
		if matchedDict {
			continue
		}
		runes := []rune(part)
		if len(runes) >= 2 && len(runes) <= 6 {
			addKeyword(part)
		}
	}

	return keywords
}

func stripDocumentPrefix(text string) string {
	text = strings.TrimSpace(text)
	parts := strings.Split(text, "｜")
	if len(parts) > 1 && looksLikeDocumentID(parts[0]) {
		return strings.Join(parts[1:], "｜")
	}
	return text
}

func looksLikeDocumentID(text string) bool {
	upper := strings.ToUpper(strings.TrimSpace(text))
	prefixes := []string{"RAW-", "QA-", "CR-", "KNOW-", "CASE-", "OBS-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func splitTitleParts(text string) []string {
	separators := []string{"｜", "|", "/", "-", "_", "：", ":", "，", "、", "与", "和", "及", "的", "对", "于", " ", "\t", "\r", "\n"}
	parts := []string{text}
	for _, sep := range separators {
		var next []string
		for _, part := range parts {
			for _, p := range strings.Split(part, sep) {
				p = strings.TrimSpace(p)
				if p != "" {
					next = append(next, p)
				}
			}
		}
		parts = next
	}
	return parts
}

func isChinesePhrase(s string) bool {
	runes := []rune(s)
	if len(runes) < 2 {
		return false
	}
	for _, r := range runes {
		if r < 0x4e00 || r > 0x9fff {
			return false
		}
	}
	return true
}

func macroTopicKeywords(result *model.ExtractionResult) []string {
	keywordsByTopic := map[string][]string{
		"RATE":   []string{"利率", "降息", "加息", "货币政策"},
		"POLICY": []string{"政策", "调控", "财政", "货币政策"},
		"ECON":   []string{"经济", "周期", "复苏", "收入", "利润"},
		"GROW":   []string{"增长", "复苏", "需求", "收入"},
		"DEBT":   []string{"债务", "信用", "杠杆"},
		"CREDIT": []string{"社融", "信用", "流动性"},
	}
	var keywords []string
	keywords = append(keywords, keywordsByTopic[result.TopicCode]...)
	keywords = append(keywords, coreKeywords(result.CoreConclusion)...)
	for _, item := range result.ReusableUnderstanding {
		keywords = append(keywords, coreKeywords(item)...)
	}
	return uniqueStrings(keywords)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
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

// RenderRawMaterial 生成原始材料 RAW Markdown
func RenderRawMaterial(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, rawText string, now time.Time) string {
	var sb strings.Builder

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.RawID, result.Title))

	// 元数据
	sb.WriteString(fmt.Sprintf("来源：%s  \n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s  \n", formatTags(result.Tags)))
	sb.WriteString("整理状态：已整理  \n")
	sb.WriteString(fmt.Sprintf("生成时间：%s  \n", now.Format("2006-01-02")))
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
			sb.WriteString(fmt.Sprintf("对应问答知识卡片：%s\n\n", QaLink(cfg, ids.QAID, result.Title, ids.QAID)))
		}
		sb.WriteString("对应候选规则：\n\n")
		for i, crID := range ids.CandidateIDs {
			if i < len(result.CandidateRules) {
				rule := result.CandidateRules[i]
				heading := JoinCandidateRuleHeading(crID, rule.DomainCode, rule.TopicCode, rule.RuleName)
				sb.WriteString(fmt.Sprintf("- %s\n", CandidateRuleLink(cfg, crID, rule.DomainCode, rule.TopicCode, rule.RuleName, heading)))
			}
		}
		sb.WriteString("\n")
	case "macro_knowledge":
		// 宏观理解型材料：链接到 KNOW 卡（单文件模式，直接 WikiLink）
		if ids.KNOWID != "" {
			sb.WriteString(fmt.Sprintf("对应宏观理解卡：%s\n\n", KnowLink(cfg, ids.KNOWID, result.Title)))
		}
		sb.WriteString("对应问答知识卡片：不生成（macro_knowledge 不生成 QA）\n")
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
		sb.WriteString("对应问答知识卡片：不生成\n")
		sb.WriteString("对应候选规则：不生成\n")
		sb.WriteString("对应规则验证卡：不生成\n\n")
	case "archive_only":
		// 仅存档：标注全部不生成
		sb.WriteString("对应问答知识卡片：不生成\n")
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
	sb.WriteString("---\n\n")
	sb.WriteString("## 原文\n\n")
	sb.WriteString(rawText)
	sb.WriteString("\n\n")
	sb.WriteString(RenderSourceMetaComment(result.SourceMeta))

	return sb.String()
}

// formatTags 格式化标签列表
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, " / ")
}
