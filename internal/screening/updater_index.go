package screening

import (
	"fmt"
	"strings"
)

func UpdateIndexContent(content, id string, d Decision) (string, bool, error) {
	lines := strings.Split(content, "\n")
	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [[") && strings.Contains(trimmed, id) {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false, fmt.Errorf("候选规则索引中找不到条目: %s", id)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- [[") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			end = i
			break
		}
	}

	block := []string{lines[start]}
	insertAt := 1
	for _, line := range lines[start+1 : end] {
		key := metaKey(line)
		switch key {
		case "第一轮筛选", "处理建议", "合并观察":
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		block = append(block, normalizeIndexLine(line))
		if strings.HasPrefix(strings.TrimSpace(line), "- 状态：") || strings.HasPrefix(strings.TrimSpace(line), "- 状态:") {
			insertAt = len(block)
		}
	}

	inserts := []string{
		"  - 第一轮筛选：" + ClassLabel(d.Class),
		"  - 处理建议：" + TopAction(d),
		"  - 合并观察：" + MergeObservation(d),
	}
	block = append(block[:insertAt], append(inserts, block[insertAt:]...)...)

	newLines := append([]string{}, lines[:start]...)
	newLines = append(newLines, block...)
	newLines = append(newLines, lines[end:]...)
	updated := strings.Join(newLines, "\n")
	return updated, updated != content, nil
}

func normalizeIndexLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "- ") {
		return "  " + trimmed
	}
	return line
}
