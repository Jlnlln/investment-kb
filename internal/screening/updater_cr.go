package screening

import (
	"fmt"
	"regexp"
	"strings"
)

type CRUpdateResult struct {
	FrontFieldsUpdated bool
	ConclusionUpdated  bool
	BeforeSourceMeta   string
	AfterSourceMeta    string
}

func UpdateCRContent(content string, d Decision, date string) (string, CRUpdateResult, error) {
	result := CRUpdateResult{BeforeSourceMeta: ExtractSourceMeta(content)}
	updated, changed := updateFrontFields(content, d)
	result.FrontFieldsUpdated = changed

	updated, changed = upsertConclusion(updated, RenderConclusion(d, date))
	result.ConclusionUpdated = changed
	result.AfterSourceMeta = ExtractSourceMeta(updated)
	if result.BeforeSourceMeta != result.AfterSourceMeta {
		return "", result, fmt.Errorf("source_meta 内容发生变化")
	}
	return updated, result, nil
}

func updateFrontFields(content string, d Decision) (string, bool) {
	lines := strings.Split(content, "\n")
	end := len(lines)
	for i, line := range lines {
		if i > 0 && strings.TrimSpace(line) == "---" {
			end = i
			break
		}
		if i > 0 && strings.HasPrefix(strings.TrimSpace(line), "## ") {
			end = i
			break
		}
	}

	values := make(map[string]string)
	var wantOrder []string
	for _, field := range ScreeningFrontFields(d) {
		values[field.Key] = field.Key + "：" + field.Value + "  "
		wantOrder = append(wantOrder, field.Key)
	}
	legacyKeys := screeningFrontFieldKeys()

	var block []string
	insertAt := end
	for i := 0; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])
		key := metaKey(trimmed)
		if legacyKeys[key] {
			continue
		}
		block = append(block, lines[i])
		if strings.HasPrefix(trimmed, "是否可转正式：") || strings.HasPrefix(trimmed, "是否可转正式:") {
			insertAt = len(block)
		}
	}
	for i := len(wantOrder) - 1; i >= 0; i-- {
		key := wantOrder[i]
		block = append(block[:insertAt], append([]string{values[key]}, block[insertAt:]...)...)
	}

	newLines := append([]string{}, block...)
	newLines = append(newLines, lines[end:]...)
	newContent := strings.Join(newLines, "\n")
	return newContent, newContent != content
}

func upsertConclusion(content, section string) (string, bool) {
	if heading := strings.Index(content, "## 第一轮筛选结论"); heading >= 0 {
		start := heading
		if prefixStart := strings.LastIndex(content[:heading], "\n---\n\n"); prefixStart >= 0 {
			between := content[prefixStart+len("\n---\n\n") : heading]
			if strings.TrimSpace(between) == "" {
				start = prefixStart + 1
			}
		}
		end := len(content)
		for _, marker := range []string{"\n---\n\n## ", "\n<!--\nsource_meta:"} {
			if idx := strings.Index(content[heading+len("## 第一轮筛选结论"):], marker); idx >= 0 {
				candidate := heading + len("## 第一轮筛选结论") + idx
				if candidate < end {
					end = candidate
				}
			}
		}
		replaced := content[:start]
		if !strings.HasSuffix(replaced, "\n\n") {
			replaced = strings.TrimRight(replaced, "\n") + "\n\n"
		}
		replaced += section
		replaced += content[end:]
		return replaced, replaced != content
	}

	appendAt := len(content)
	if idx := strings.Index(content, "\n<!--\nsource_meta:"); idx >= 0 {
		appendAt = idx
	}
	prefix := strings.TrimRight(content[:appendAt], "\n")
	suffix := content[appendAt:]
	updated := prefix + "\n\n" + section
	if suffix != "" {
		updated += "\n" + strings.TrimLeft(suffix, "\n")
	}
	return updated, true
}

func metaKey(line string) string {
	line = strings.TrimSpace(strings.TrimSuffix(line, "  "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
	for _, sep := range []string{"：", ":"} {
		if idx := strings.Index(line, sep); idx > 0 {
			return strings.TrimSpace(line[:idx])
		}
	}
	return ""
}

func screeningFrontFieldKeys() map[string]bool {
	return map[string]bool{
		"第一轮筛选":  true,
		"筛选批次":   true,
		"当前处理队列": true,
		"处理建议":   true,
		"合并观察":   true,
		"合并去向":   true,
		"联动观察":   true,
		"整合状态":   true,
		"正式化状态":  true,
	}
}

func ExtractSourceMeta(content string) string {
	re := regexp.MustCompile(`(?ms)<!--\s*\nsource_meta:\n.*?\n-->`)
	return re.FindString(content)
}
