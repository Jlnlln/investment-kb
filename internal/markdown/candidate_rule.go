package markdown

import (
	"fmt"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/dedup"
	"investment-kb/internal/idgen"
	"investment-kb/internal/model"
)

// RenderCandidateRules 生成候选规则 CR Markdown（每条规则一个段落）
func RenderCandidateRules(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, rules []model.CandidateRule, similarData [][]dedup.SimilarRule) string {
	var sb strings.Builder

	for i, rule := range rules {
		crID := ""
		if i < len(ids.CandidateIDs) {
			crID = ids.CandidateIDs[i]
		}
		var similarRules []dedup.SimilarRule
		if i < len(similarData) {
			similarRules = similarData[i]
		}
		sb.WriteString(renderSingleCandidateRule(cfg, crID, ids.QAID, ids.RawID, result, ids, rule, similarRules))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderSingleCandidateRule 生成单条候选规则 Markdown
func renderSingleCandidateRule(cfg *config.Config, crID, qaID, rawID string, result *model.ExtractionResult, ids *model.DocumentIDs, rule model.CandidateRule, similarRules []dedup.SimilarRule) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题：CR-日期-序数｜DOMAIN-TOPIC｜规则名称
	shortCode := rule.DomainCode + "-" + rule.TopicCode
	sb.WriteString(fmt.Sprintf("# %s｜%s｜%s\n\n", crID, shortCode, rule.RuleName))

	// 元数据
	sb.WriteString("状态：候选  \n")
	sb.WriteString("验证状态：待验证  \n")
	sb.WriteString(fmt.Sprintf("规则验证卡：%s  \n", GetValidationCardLink(cfg, crID, crID)))
	sb.WriteString("是否可转正式：否  \n")

	// 领域信息（显示原始和映射）
	mappedDomain := idgen.MapCRDomain(rule.DomainCode)
	sb.WriteString(fmt.Sprintf("建议正式领域：%s  \n", mappedDomain))
	if rule.OriginalDomainCode != "" && rule.OriginalDomainCode != rule.DomainCode {
		sb.WriteString(fmt.Sprintf("原始领域（AI 建议）：%-s → 映射领域：%s  \n", rule.OriginalDomainCode, rule.DomainCode))
	} else {
		sb.WriteString(fmt.Sprintf("领域分类：%s  \n", rule.DomainCode))
	}

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
	sb.WriteString("\n\n")

	// 8. 相似规则检查（始终渲染）
	sb.WriteString("---\n\n")
	sb.WriteString("## 8. 相似规则检查\n\n")
	if len(similarRules) > 0 {
		sb.WriteString("相似候选规则：\n\n")
		for _, sr := range similarRules {
			sb.WriteString(fmt.Sprintf("- [[%s#%s|%s]]\n", GetCandidateRulePath(cfg), sr.CRID+"｜"+sr.ShortCode+"｜"+sr.RuleName, sr.CRID))
			sb.WriteString(fmt.Sprintf("  - 相似原因：%s\n", sr.Reason))
			sb.WriteString(fmt.Sprintf("  - 相似级别：%s\n", sr.Level))
		}
		sb.WriteString("\n处理建议：\n\n")
		sb.WriteString("- [ ] 新建独立规则\n")
		sb.WriteString("- [ ] 合并到已有规则（作为补充来源）\n")
		sb.WriteString("- [ ] 保留但标记可能与已有规则冲突\n")
		sb.WriteString("- [ ] 废弃（与已有规则重复）\n")
		sb.WriteString("\n")
	} else {
		sb.WriteString("相似候选规则：暂无\n\n")
	}

	return sb.String()
}

// getCaseText 获取关联案例文本
func getCaseText(ids *model.DocumentIDs, result *model.ExtractionResult) string {
	if result.ShouldGenerateCase && result.Case != nil {
		return fmt.Sprintf("见：%s｜%s", ids.CaseID, result.Case.CaseName)
	}
	return "暂无"
}
