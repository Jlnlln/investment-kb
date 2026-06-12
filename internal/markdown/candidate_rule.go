package markdown

import (
	"fmt"
	"strings"

	"investment-kb/internal/model"
)

// RenderCandidateRules 生成候选规则 CR Markdown（每条规则一个段落）
func RenderCandidateRules(ids *model.DocumentIDs, rules []model.CandidateRule) string {
	var sb strings.Builder

	for i, rule := range rules {
		crID := ""
		if i < len(ids.CandidateIDs) {
			crID = ids.CandidateIDs[i]
		}
		sb.WriteString(renderSingleCandidateRule(crID, ids.QAID, rule))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderSingleCandidateRule 生成单条候选规则 Markdown
func renderSingleCandidateRule(crID, qaID string, rule model.CandidateRule) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", crID, rule.RuleName))

	// 元数据
	sb.WriteString("状态：待确认\n")
	sb.WriteString(fmt.Sprintf("规则类型：%s\n", rule.RuleType))
	sb.WriteString(fmt.Sprintf("建议正式编号：%s\n", rule.SuggestedFormalRuleID))
	sb.WriteString(fmt.Sprintf("来源知识卡片：%s\n\n", qaID))

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