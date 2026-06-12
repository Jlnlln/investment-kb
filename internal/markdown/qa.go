package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/model"
)

// RenderKnowledgeCard 生成知识卡片 QA Markdown
func RenderKnowledgeCard(ids *model.DocumentIDs, result *model.ExtractionResult, now time.Time) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.QAID, result.Title))

	// 元数据
	sb.WriteString(fmt.Sprintf("原始材料：%s｜%s\n", ids.RawID, result.Title))
	sb.WriteString(fmt.Sprintf("来源：%s\n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s\n", formatTags(result.Tags)))
	sb.WriteString(fmt.Sprintf("整理时间：%s\n\n", now.Format("2006-01-02")))

	// 关联候选规则
	sb.WriteString("关联候选规则：\n\n")
	for i, crID := range ids.CandidateIDs {
		if i < len(result.CandidateRules) {
			sb.WriteString(fmt.Sprintf("- %s｜%s\n", crID, result.CandidateRules[i].RuleName))
		}
	}
	sb.WriteString("\n")

	// 关联案例
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("关联案例：%s｜%s\n", ids.CaseID, result.Case.CaseName))
	} else {
		sb.WriteString("关联案例：暂不单独生成市场案例")
	}
	sb.WriteString("\n关联正式规则：待确认\n\n")

	// 分隔线
	sb.WriteString("---\n\n")

	// 1. 问题摘要
	sb.WriteString("## 1. 问题摘要\n\n")
	sb.WriteString(result.Summary)
	sb.WriteString("\n\n")

	// 2. 核心结论
	sb.WriteString("---\n\n")
	sb.WriteString("## 2. 核心结论\n\n")
	sb.WriteString(result.CoreConclusion)
	sb.WriteString("\n\n")

	// 3. 核心逻辑
	sb.WriteString("---\n\n")
	sb.WriteString("## 3. 核心逻辑\n\n")
	for i, logic := range result.CoreLogic {
		sb.WriteString(fmt.Sprintf("### 3.%d %s\n\n", i+1, logic.Title))
		sb.WriteString(logic.Content)
		sb.WriteString("\n\n")
	}

	// 4. 适用场景
	sb.WriteString("---\n\n")
	sb.WriteString("## 4. 适用场景\n\n")
	for _, scenario := range result.ApplicableScenarios {
		sb.WriteString(fmt.Sprintf("- %s\n", scenario))
	}
	sb.WriteString("\n")

	// 5. 风险边界
	sb.WriteString("---\n\n")
	sb.WriteString("## 5. 风险边界\n\n")
	for _, boundary := range result.RiskBoundaries {
		sb.WriteString(fmt.Sprintf("- %s\n", boundary))
	}
	sb.WriteString("\n")

	// 6. 可提炼规则
	sb.WriteString("---\n\n")
	sb.WriteString("## 6. 可提炼规则\n\n")

	// 按规则类型分组
	rulesByType := make(map[string][]int)
	for i, rule := range result.CandidateRules {
		rulesByType[rule.RuleType] = append(rulesByType[rule.RuleType], i)
	}

	for ruleType, indices := range rulesByType {
		sb.WriteString(fmt.Sprintf("### 6.%s %s\n\n", getRuleTypeSuffix(ruleType), ruleType))
		for _, idx := range indices {
			sb.WriteString(fmt.Sprintf("- %s｜%s\n\n", ids.CandidateIDs[idx], result.CandidateRules[idx].RuleName))
			sb.WriteString(result.CandidateRules[idx].RuleContent)
			sb.WriteString("\n\n")
		}
	}

	// 7. 关联案例
	sb.WriteString("---\n\n")
	sb.WriteString("## 7. 关联案例\n\n")
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("见：%s｜%s\n\n", ids.CaseID, result.Case.CaseName))
	} else {
		sb.WriteString("暂不单独生成市场案例。\n\n")
		sb.WriteString("原因：")
		sb.WriteString(result.CaseInsufficientReason)
		sb.WriteString("\n\n")
	}

	// 8. 我的理解
	sb.WriteString("---\n\n")
	sb.WriteString("## 8. 我的理解\n\n")
	sb.WriteString(result.MyUnderstanding)
	sb.WriteString("\n")

	return sb.String()
}

// getRuleTypeSuffix 获取规则类型后缀（用于排序）
func getRuleTypeSuffix(ruleType string) string {
	// 简单映射，实际可以更复杂
	return "1"
}