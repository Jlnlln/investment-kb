package markdown

import (
	"fmt"
	"strings"
	"time"

	"investment-kb/internal/config"
	"investment-kb/internal/model"
)

// RenderKnowledgeCard 生成知识卡片 QA Markdown
func RenderKnowledgeCard(cfg *config.Config, ids *model.DocumentIDs, result *model.ExtractionResult, now time.Time) string {
	var sb strings.Builder

	// 分隔线
	sb.WriteString("---\n\n")

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.QAID, result.Title))

	// 元数据（使用 WikiLink）
	sb.WriteString(fmt.Sprintf("原始材料：%s\n", ObsidianHeadingLink(GetRawMaterialPath(cfg), JoinHeading(ids.RawID, result.Title), ids.RawID)))
	sb.WriteString(fmt.Sprintf("来源：%s\n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s\n", formatTags(result.Tags)))
	sb.WriteString(fmt.Sprintf("整理时间：%s\n", now.Format("2006-01-02")))
	sb.WriteString(RenderSourceMetaLines(result.SourceMeta))
	sb.WriteString("\n")

	// 关联候选规则（使用 WikiLink）
	sb.WriteString("关联候选规则：\n\n")
	for i, crID := range ids.CandidateIDs {
		if i < len(result.CandidateRules) {
			rule := result.CandidateRules[i]
			sb.WriteString(fmt.Sprintf("- %s\n", CandidateRuleLink(cfg, crID, rule.DomainCode, rule.TopicCode, rule.RuleName, crID)))
		}
	}
	sb.WriteString("\n")

	// 关联案例（使用 WikiLink）
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("关联案例：%s\n", ObsidianHeadingLink(GetMarketCasePath(cfg), JoinHeading(ids.CaseID, result.Case.CaseName), ids.CaseID)))
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

	// 按固定规则类型顺序输出，保证编号稳定
	ruleTypeOrder := []string{
		"买入规则", "加仓规则", "减仓规则", "卖出规则",
		"仓位规则", "账户适配规则", "风控规则", "情绪控制规则",
		"复盘规则", "配置规则",
	}
	sectionNum := 1
	for _, ruleType := range ruleTypeOrder {
		indices, ok := rulesByType[ruleType]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("### 6.%d %s\n\n", sectionNum, ruleType))
		sectionNum++
		for _, idx := range indices {
			rule := result.CandidateRules[idx]
			crID := ids.CandidateIDs[idx]
			sb.WriteString(fmt.Sprintf("- %s\n\n", CandidateRuleLink(cfg, crID, rule.DomainCode, rule.TopicCode, rule.RuleName, crID)))
			sb.WriteString(rule.RuleContent)
			sb.WriteString("\n\n")
		}
	}

	// 7. 关联案例
	sb.WriteString("---\n\n")
	sb.WriteString("## 7. 关联案例\n\n")
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("见：%s\n\n", ObsidianHeadingLink(GetMarketCasePath(cfg), JoinHeading(ids.CaseID, result.Case.CaseName), ids.CaseID)))
	} else {
		sb.WriteString("暂不单独生成市场案例。\n\n")
		sb.WriteString("原因：")
		sb.WriteString(result.CaseInsufficientReason)
		sb.WriteString("\n\n")
	}

	// 8. 我的理解
	sb.WriteString("---\n\n")
	sb.WriteString("## 8. 我的理解\n\n")
	if result.MyUnderstanding == "" {
		sb.WriteString("待补充。")
	} else {
		sb.WriteString(result.MyUnderstanding)
	}
	sb.WriteString("\n")

	return sb.String()
}
