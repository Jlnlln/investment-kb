package screening

import (
	"fmt"
	"strings"
)

func RenderConclusion(d Decision, date string) string {
	var sb strings.Builder
	sb.WriteString("---\n\n")
	sb.WriteString("## 第一轮筛选结论\n\n")
	sb.WriteString(fmt.Sprintf("- 筛选日期：%s\n", date))
	sb.WriteString(fmt.Sprintf("- 筛选分类：%s\n", ClassLabel(d.Class)))
	sb.WriteString(fmt.Sprintf("- 处理动作：%s\n", valueOrDefault(d.Action, "暂未填写")))
	sb.WriteString(fmt.Sprintf("- 合并去向：%s\n", mergeTargetText(d)))
	if text := LinkWatchText(d); text != "" {
		sb.WriteString(fmt.Sprintf("- 联动观察：%s\n", text))
	}
	sb.WriteString(fmt.Sprintf("- 规则定位：%s\n\n", valueOrDefault(d.Position, "暂未填写")))

	sb.WriteString("### 判断理由\n\n")
	writeBullets(&sb, d.Reasons, "暂未填写")
	sb.WriteString("\n")

	sb.WriteString("### 需要优化的地方\n\n")
	writeBullets(&sb, d.Improvements, "暂无")
	sb.WriteString("\n")

	sb.WriteString("### 后续处理\n\n")
	writeBullets(&sb, d.NextSteps, "暂无")
	sb.WriteString("\n")

	sb.WriteString("### 正式规则观察\n\n")
	if d.FormalCandidate {
		sb.WriteString("- 是否具备正式规则潜力：是\n")
	} else {
		sb.WriteString("- 是否具备正式规则潜力：否\n")
	}
	sb.WriteString(fmt.Sprintf("- 建议正式规则方向：%s\n", valueOrDefault(d.FormalRuleSuggestion, "暂不建议")))
	sb.WriteString("- 当前转正式阻碍：\n")
	blockers := nonEmptyList(d.PromoteBlockers)
	if len(blockers) == 0 {
		sb.WriteString("  - 暂无\n")
	} else {
		for _, item := range blockers {
			sb.WriteString("  - " + sentence(item) + "\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n") + "\n"
}

func writeBullets(sb *strings.Builder, items []string, empty string) {
	items = nonEmptyList(items)
	if len(items) == 0 {
		sb.WriteString("- " + empty + "\n")
		return
	}
	for _, item := range items {
		sb.WriteString("- " + sentence(item) + "\n")
	}
}

func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	last := []rune(s)[len([]rune(s))-1]
	switch last {
	case '。', '！', '？', '.', '!', '?':
		return s
	default:
		return s + "。"
	}
}

func valueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func mergeTargetText(d Decision) string {
	if strings.TrimSpace(d.MergeTarget) != "" {
		return strings.TrimSpace(d.MergeTarget)
	}
	return "暂不合并"
}
