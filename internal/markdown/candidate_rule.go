package markdown

import (
	"fmt"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// RenderCandidateRules 生成候选规则 CR Markdown（每条规则一个段落）
func RenderCandidateRules(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, rules []model.CandidateRule) string {
	var sb strings.Builder

	for i, rule := range rules {
		crID := ""
		if i < len(ids.CandidateIDs) {
			crID = ids.CandidateIDs[i]
		}
		sb.WriteString(renderSingleCandidateRule(cfg, crID, ids.QAID, ids.RawID, result, ids, rule))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderSingleCandidateRule 生成单条候选规则 Markdown
func renderSingleCandidateRule(cfg *config.Config, crID, qaID, rawID string, result *model.ExtractionResult, ids *model.DocumentIDs, rule model.CandidateRule) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题：CR-日期-序数｜DOMAIN-TOPIC｜规则名称
	shortCode := rule.DomainCode + "-" + rule.TopicCode
	sb.WriteString(fmt.Sprintf("# %s｜%s｜%s\n\n", crID, shortCode, rule.RuleName))

	// 元数据
	sb.WriteString("状态：待确认  \n")
	sb.WriteString(fmt.Sprintf("规则类型：%s  \n", rule.RuleType))
	sb.WriteString(fmt.Sprintf("建议正式编号：%s  \n", rule.SuggestedFormalRuleID))
	sb.WriteString(fmt.Sprintf("来源知识卡片：%s  \n", ObsidianHeadingLink(GetQaPath(cfg), JoinHeading(ids.QAID, result.Title), JoinHeading(ids.QAID, result.Title))))
	sb.WriteString(fmt.Sprintf("来源原文：%s  \n", ObsidianHeadingLink(GetRawMaterialPath(cfg), JoinHeading(ids.RawID, result.Title), JoinHeading(ids.RawID, result.Title))))

	caseText := getCaseText(ids, result)
	if caseText == "暂无" {
		sb.WriteString(fmt.Sprintf("关联案例：%s  \n", caseText))
	} else {
		sb.WriteString(fmt.Sprintf("关联案例：%s  \n", ObsidianHeadingLink(GetMarketCasePath(cfg), JoinHeading(ids.CaseID, result.Case.CaseName), JoinHeading(ids.CaseID, result.Case.CaseName))))
	}

	if rule.ApplicableObjects != nil && len(rule.ApplicableObjects) > 0 {
		sb.WriteString(fmt.Sprintf("适用对象：%s  \n", strings.Join(rule.ApplicableObjects, " / ")))
	}

	// 分隔线
	sb.WriteString("---\n\n")

	// 1. 规则内容
	sb.WriteString("## 1. 规则内容\n\n")
	sb.WriteString(rule.RuleContent)
	sb.WriteString("\n\n")

	// 2. 触发条件
	sb.WriteString("---\n\n")
	sb.WriteString("## 2. 触发条件\n\n")
	for _, cond := range rule.TriggerConditions {
		sb.WriteString(fmt.Sprintf("- %s\n", cond))
	}
	sb.WriteString("\n")

	// 3. 执行动作
	sb.WriteString("---\n\n")
	sb.WriteString("## 3. 执行动作\n\n")
	for _, action := range rule.Actions {
		sb.WriteString(fmt.Sprintf("- %s\n", action))
	}
	sb.WriteString("\n")

	// 4. 不适用场景
	sb.WriteString("---\n\n")
	sb.WriteString("## 4. 不适用场景\n\n")
	for _, na := range rule.NotApplicable {
		sb.WriteString(fmt.Sprintf("- %s\n", na))
	}
	sb.WriteString("\n")

	// 5. 风险边界
	sb.WriteString("---\n\n")
	sb.WriteString("## 5. 风险边界\n\n")
	sb.WriteString(rule.RiskBoundary)
	sb.WriteString("\n\n")

	// 6. 待确认问题
	sb.WriteString("---\n\n")
	sb.WriteString("## 6. 待确认问题\n\n")
	for i, q := range rule.QuestionsToConfirm {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q))
	}
	sb.WriteString("\n")

	// 7. 建议处理
	sb.WriteString("---\n\n")
	sb.WriteString("## 7. 建议处理\n\n")
	sb.WriteString(rule.Recommendation)
	sb.WriteString("\n")

	return sb.String()
}

// getCaseText 获取关联案例文本
func getCaseText(ids *model.DocumentIDs, result *model.ExtractionResult) string {
	if result.ShouldGenerateCase && result.Case != nil {
		return fmt.Sprintf("见：%s｜%s", ids.CaseID, result.Case.CaseName)
	}
	return "暂无"
}
