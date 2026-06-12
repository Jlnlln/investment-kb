package ai

import "strings"

// FixCommonJSON 修复 LLM 输出中常见的 JSON 格式错误
func FixCommonJSON(raw string) string {
	s := strings.TrimSpace(raw)

	// 去除 ```json ... ``` 代码块包裹
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s[3:], "\n"); idx >= 0 {
			s = s[3+idx+1:]
		}
	}
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	s = strings.TrimSpace(s)

	// 去除 } 或 ] 前的尾随逗号
	s = fixTrailingCommas(s)

	return s
}

// fixTrailingCommas 去除 } 或 ] 前面的逗号
func fixTrailingCommas(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == ',' {
			j := i + 1
			for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n' || runes[j] == '\r') {
				j++
			}
			if j < len(runes) && (runes[j] == '}' || runes[j] == ']') {
				continue
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}