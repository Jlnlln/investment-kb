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

	// 标题
	sb.WriteString(fmt.Sprintf("# %s｜%s\n\n", ids.QAID, result.Title))

	// 元数据（使用 WikiLink）
	sb.WriteString(fmt.Sprintf("原始材料：%s\n", RawMaterialLink(cfg, ids.RawID, result.Title, ids.RawID)))
	sb.WriteString(fmt.Sprintf("来源：%s\n", result.Source))
	sb.WriteString(fmt.Sprintf("主题标签：%s\n", formatTags(result.Tags)))
	sb.WriteString(fmt.Sprintf("整理时间：%s\n", now.Format("2006-01-02")))
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

	sb.WriteString("---\n\n")
	renderAccountProfiles(&sb, result.AccountProfiles)

	sb.WriteString("---\n\n")
	renderBehaviorCorrection(&sb, result.BehaviorCorrection)

	// 6. 适用场景
	sb.WriteString("---\n\n")
	sb.WriteString("## 6. 适用场景\n\n")
	for _, scenario := range result.ApplicableScenarios {
		sb.WriteString(fmt.Sprintf("- %s\n", scenario))
	}
	sb.WriteString("\n")

	// 7. 风险边界
	sb.WriteString("---\n\n")
	sb.WriteString("## 7. 风险边界\n\n")
	for _, boundary := range result.RiskBoundaries {
		sb.WriteString(fmt.Sprintf("- %s\n", boundary))
	}
	sb.WriteString("\n")

	// 8. 可提炼规则
	sb.WriteString("---\n\n")
	sb.WriteString("## 8. 可提炼规则\n\n")
	sb.WriteString("### 8.1 已生成候选规则\n\n")
	if len(result.CandidateRules) == 0 {
		sb.WriteString("本次未直接生成候选规则。\n\n")
	} else {
		for i, rule := range result.CandidateRules {
			if i >= len(ids.CandidateIDs) {
				continue
			}
			crID := ids.CandidateIDs[i]
			sb.WriteString(fmt.Sprintf("- %s\n", CandidateRuleLink(cfg, crID, rule.DomainCode, rule.TopicCode, rule.RuleName, crID)))
			sb.WriteString(fmt.Sprintf("  - 规则类型：%s\n", rule.RuleType))
			sb.WriteString(fmt.Sprintf("  - 规则内容：%s\n", rule.RuleContent))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### 8.2 潜在但未生成规则\n\n")
	renderPotentialRules(&sb, result.PotentialRules)

	// 9. 关联案例
	sb.WriteString("---\n\n")
	sb.WriteString("## 9. 关联案例\n\n")
	if ids.CaseID != "" {
		sb.WriteString(fmt.Sprintf("见：%s\n\n", ObsidianHeadingLink(GetMarketCasePath(cfg), JoinHeading(ids.CaseID, result.Case.CaseName), ids.CaseID)))
	} else {
		sb.WriteString("暂不单独生成市场案例。\n\n")
		sb.WriteString("原因：")
		sb.WriteString(result.CaseInsufficientReason)
		sb.WriteString("\n\n")
	}

	// 10. 我的理解
	sb.WriteString("---\n\n")
	sb.WriteString("## 10. 我的理解\n\n")
	if result.MyUnderstanding == "" {
		sb.WriteString("待补充。")
	} else {
		sb.WriteString(result.MyUnderstanding)
	}
	sb.WriteString("\n\n")
	sb.WriteString(RenderSourceMetaComment(result.SourceMeta))

	return sb.String()
}

func renderAccountProfiles(sb *strings.Builder, profiles model.AccountProfiles) {
	sb.WriteString("## 4. 账户画像与状态差异\n\n")
	sb.WriteString("### 4.1 文中涉及的账户状态\n\n")
	stateNotes := make(map[string]string)
	for _, item := range profiles.MentionedStates {
		stateNotes[strings.TrimSpace(item.State)] = strings.TrimSpace(item.Note)
	}
	states := []string{"空仓者", "亏损者", "盈利者", "低成本持仓者", "高成本持仓者", "小白投资者", "经验投资者"}
	for _, state := range states {
		note := stateNotes[state]
		if note == "" {
			note = "原文未明确涉及"
		}
		sb.WriteString(fmt.Sprintf("- %s：%s\n", state, note))
	}
	sb.WriteString("\n")

	sb.WriteString("### 4.2 不同状态下的建议差异\n\n")
	if len(profiles.RecommendationDiffs) == 0 {
		sb.WriteString("- 原文未明确涉及\n")
	} else {
		for _, diff := range profiles.RecommendationDiffs {
			sb.WriteString(fmt.Sprintf("- %s\n", diff))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("### 4.3 差异化建议的原因\n\n")
	if strings.TrimSpace(profiles.Reason) == "" {
		sb.WriteString("原文未明确涉及。\n\n")
	} else {
		sb.WriteString(profiles.Reason)
		sb.WriteString("\n\n")
	}
}

func renderBehaviorCorrection(sb *strings.Builder, correction model.BehaviorCorrection) {
	sb.WriteString("## 5. 行为纠偏点\n\n")
	sb.WriteString("### 5.1 原文想防止的错误行为\n\n")
	writeStringListOrDefault(sb, correction.WrongBehaviors)

	sb.WriteString("### 5.2 对应的行为约束\n\n")
	writeStringListOrDefault(sb, correction.BehaviorConstraints)

	sb.WriteString("### 5.3 是否存在反直觉建议\n\n")
	if strings.TrimSpace(correction.CounterintuitiveAdvice) == "" {
		sb.WriteString("原文未明确涉及。\n\n")
	} else {
		sb.WriteString(correction.CounterintuitiveAdvice)
		sb.WriteString("\n\n")
	}
}

func renderPotentialRules(sb *strings.Builder, rules []model.PotentialRule) {
	if len(rules) == 0 {
		sb.WriteString("本次未识别出需要保留但暂不生成 CR 的潜在规则。\n\n")
		return
	}
	for _, rule := range rules {
		sb.WriteString("- 规则雏形：")
		sb.WriteString(emptyAsNotMentioned(rule.RuleDraft))
		sb.WriteString("\n")
		sb.WriteString("  - 所属领域：")
		sb.WriteString(emptyAsNotMentioned(rule.DomainCode))
		sb.WriteString("\n")
		sb.WriteString("  - 原文依据：")
		sb.WriteString(emptyAsNotMentioned(rule.OriginalEvidence))
		sb.WriteString("\n")
		sb.WriteString("  - 适用对象：")
		if len(rule.ApplicableObjects) == 0 {
			sb.WriteString("原文未明确涉及")
		} else {
			sb.WriteString(strings.Join(rule.ApplicableObjects, " / "))
		}
		sb.WriteString("\n")
		sb.WriteString("  - 防止的错误：")
		sb.WriteString(emptyAsNotMentioned(rule.PreventedError))
		sb.WriteString("\n")
		sb.WriteString("  - 是否建议生成 CR：")
		sb.WriteString(emptyAsNotMentioned(rule.ShouldGenerateCR))
		sb.WriteString("\n")
		sb.WriteString("  - 不生成原因：")
		sb.WriteString(emptyAsNotMentioned(rule.NoGenerateReason))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func writeStringListOrDefault(sb *strings.Builder, items []string) {
	if len(items) == 0 {
		sb.WriteString("- 原文未明确涉及\n\n")
		return
	}
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	sb.WriteString("\n")
}

func emptyAsNotMentioned(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "原文未明确涉及"
	}
	return value
}
