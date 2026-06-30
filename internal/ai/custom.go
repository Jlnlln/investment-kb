package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"investment-kb/internal/model"
)

// customClient 是自定义 AI 客户端（兼容 Anthropic Messages API 格式）
type customClient struct {
	cfg *Config
}

func newCustomClient(cfg *Config) *customClient {
	return &customClient{cfg: cfg}
}

// anthropic Messages API 请求/响应类型
type messagesRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Messages  []messageItem `json:"messages"`
	Temperature float64     `json:"temperature"` // 固定为 0，确保输出稳定性
}

type messageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []contentBlock `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var lastCallTime time.Time

// Complete 发送请求并返回原始文本
func (c *customClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.doCall(ctx, systemPrompt, userPrompt)
}

// CompleteJSON 发送请求并将结果解析为结构体
func (c *customClient) CompleteJSON(ctx context.Context, systemPrompt, userPrompt string, v any) error {
	raw, err := c.doCall(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	// 使用 ExtractJSONFromAIOutput 清洗输出
	cleaned, err := ExtractJSONFromAIOutput(raw)
	if err != nil {
		// 清洗失败，保存原始输出
		saveErrorOutput(raw, "ExtractJSONFromAIOutput", err.Error(), "", "")
		return fmt.Errorf("JSON 提取失败: %w", err)
	}

	// 尝试解析 JSON
	if err := json.Unmarshal([]byte(cleaned), v); err != nil {
		// JSON 解析失败，保存原始输出和清洗后的输出
		saveErrorOutput(raw, "json.Unmarshal", err.Error(), "", "")
		return fmt.Errorf("JSON 解析失败: %w", err)
	}

	return nil
}

func (c *customClient) doCall(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// 速率限制
	if elapsed := time.Since(lastCallTime); elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}

	reqBody := messagesRequest{
		Model:     c.cfg.Model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []messageItem{
			{Role: "user", Content: userPrompt},
		},
		Temperature: c.cfg.Temperature,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	var rawText string
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// 529 过载错误使用更长的退避时间
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			if attempt > 2 {
				backoff *= 2 // 第3次开始，退避时间加倍
			}
			fmt.Printf("   ⏳ 重试 %d/%d (等待 %v)...\n", attempt, c.cfg.MaxRetries, backoff)
			time.Sleep(backoff)
		}

		rawText, lastErr = c.doHTTPRequest(ctx, bodyBytes)
		if lastErr != nil {
			continue
		}

		lastCallTime = time.Now()
		return rawText, nil
	}

	return "", fmt.Errorf("重试 %d 次后仍失败: %w", c.cfg.MaxRetries, lastErr)
}

func (c *customClient) doHTTPRequest(ctx context.Context, body []byte) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(c.cfg.TimeoutSec) * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("触发速率限制 (429)")
	}
	if resp.StatusCode >= 500 {
		// 529 是服务端过载错误，应该重试
		if resp.StatusCode == 529 {
			return "", fmt.Errorf("服务端过载错误 (529): %s", truncate(string(respBody), 200))
		}
		return "", fmt.Errorf("服务端错误 (%d): %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	if resp.StatusCode >= 400 {
		var apiErr apiError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return "", fmt.Errorf("API 错误 (%d): %s", resp.StatusCode, apiErr.Message)
		}
		return "", fmt.Errorf("客户端错误 (%d): %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(msgResp.Content) == 0 {
		return "", fmt.Errorf("响应内容为空")
	}

	return msgResp.Content[0].Text, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// ExtractJSONFromAIOutput 从 AI 原始输出中提取并清洗 JSON
func ExtractJSONFromAIOutput(raw string) (string, error) {
	s := strings.TrimSpace(raw)

	// remove BOM using byte slice to avoid character encoding issues
	bom := []byte{0xEF, 0xBB, 0xBF}
	s = strings.TrimPrefix(s, string(bom))
	s = strings.TrimSpace(s)

	// remove markdown json fence
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}

	// remove BOM again after fence stripping
	s = strings.TrimPrefix(s, string(bom))
	s = strings.TrimSpace(s)

	// extract first JSON object if extra text exists
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}

	s = strings.TrimSpace(s)

	if !strings.HasPrefix(s, "{") {
		preview := s
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return "", fmt.Errorf("AI 输出清洗后仍不是 JSON 对象，开头内容：%s", preview)
	}

	return s, nil
}

// saveErrorOutput 保存 AI 错误输出到文件
func saveErrorOutput(raw, step, errorDetail string, inputPath, source string) {
	filename := fmt.Sprintf("data/error_outputs/ai_error_%s.txt", time.Now().Format("20060102_150405"))
	content := fmt.Sprintf(`执行时间: %s
输入文件: %s
Source: %s
错误步骤: %s
错误原因: %s
---
AI 原始输出开头 (前 500 字符):
%s
...
AI 原始输出结尾 (后 500 字符):
%s
---
完整原始输出:
%s
`,
		time.Now().Format("2006-01-02 15:04:05"),
		inputPath,
		source,
		step,
		errorDetail,
		truncate(raw, 500),
		truncate(raw, 500),
		raw)
	_ = writeToFile(filename, content)
}

func writeToFile(path, content string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	return os.WriteFile(path, []byte(content), 0644)
}

// forbiddenPhrases 禁止表达列表（硬性校验，失败时终止）
var forbiddenPhrases = []string{
	"保证盈利",
	"没有亏损风险",
	"必然上涨",
	"一定赚钱",
	"判断错了也不会亏",
	"只赚不亏",
	"无风险",
	"稳赚",
	"绝对安全",
	"必胜",
}

// 满仓/全仓相关关键词
var overloadingPhrasesKeywords = []string{"满仓", "全仓"}

// 需要警告的表达（非硬性失败）
var overloadingPhrasesWarning = []string{
	"应该全仓押注",
	"可以一次性全仓",
}

// 否定词（允许满仓表达）
var negativeWords = []string{
	"不能满仓",
	"不得满仓",
	"不得直接满仓",
	"不要满仓",
	"不要满仓买入",
	"不可直接满仓",
	"避免满仓",
	"不能一次性全仓押注",
	"不能照搬满仓计划",
	"不能照搬低成本持仓者的满仓计划",
}

// 肯定执行词（失败表达）
var positiveWords = []string{
	"应该满仓",
	"必须满仓",
	"可以满仓",
	"应直接满仓",
	"可直接满仓",
	"满仓买入",
	"直接满仓",
	"高确定性时可直接满仓",
	// "可以一次性全仓",  // 已移到 warningPhrases
	// "应该全仓押注",   // 已移到 warningPhrases
}

// positivePhrasePatterns 预编译正面短语的正则表达式
var positivePhrasePatterns = compilePhrasePatterns(positiveWords)

// negativePhrasePatterns 预编译负面短语的正则表达式
var negativePhrasePatterns = compilePhrasePatterns(negativeWords)

// warningPhrasePatterns 预编译警告短语的正则表达式
var warningPhrasePatterns = compilePhrasePatterns(overloadingPhrasesWarning)

// CheckOverloadingPhrases 智能检查满仓误导表达
// 返回：error - 硬性校验失败；nil - 正常
func ContainsForbiddenPhrases(text string) error {
	// 1. 先检查硬性失败的表达（需要检查上下文中的否定词）
	if hasPositivePhrase(text) {
		phrase := findMatchedPositivePhrase(text)
		if phrase != "" {
			return fmt.Errorf("AI 输出包含满仓误导表达：%s。请调整 Prompt 或人工检查后重试。", phrase)
		}
		return fmt.Errorf("AI 输出包含满仓误导表达。请调整 Prompt 或人工检查后重试。")
	}

	// 2. 检查是否包含满仓/全仓关键词
	if !hasOverloadingKeyword(text) {
		return nil
	}

	// 3. 检查是否包含允许的表达（否定词）
	fmt.Printf("  Checking negative phrases...\n")
	if hasNegativePhrase(text) {
		fmt.Printf("    ✓ Negative phrase found\n")
		return nil
	}

	// 4. 检查是否属于警告表达（非硬性失败）
	if hasWarningPhrase(text) {
		phrase := findMatchedPhrase(text, overloadingPhrasesWarning)
		if phrase != "" {
			fmt.Printf("⚠️  包含疑似满仓误导表达但不属于硬性失败：%s。建议检查上下文。\n", phrase)
		} else {
			fmt.Printf("⚠️  包含疑似满仓误导表达但不属于硬性失败。建议检查上下文。\n")
		}
	}

	// 5. 其他情况：包含满仓关键词但没有明确的否定或肯定表达
	return nil
}

// hasPositivePhrase 检查是否包含正面短语（需要检查上下文中的否定词）
func hasPositivePhrase(text string) bool {
	// 首先检查完整单词
	words := strings.Fields(text)
	for _, word := range words {
		if isPositivePhrase(word) && !hasNegativeContext(text, word) {
			return true
		}
	}

	// 然后检查子串匹配（例如 "可以满仓买入" 应该匹配 "可以满仓"）
	// 但需要检查上下文中的否定词
	for _, phrase := range positiveWords {
		if strings.Contains(text, phrase) && !hasNegativeContext(text, phrase) {
			return true
		}
	}
	return false
}

// debugHasPositivePhrase 调试版本
// hasNegativeContext 检查文本中是否有否定词或警告短语
func hasNegativeContext(text, phrase string) bool {
	// 将文本分割成单词
	words := strings.Fields(text)

	// 检查文本中的所有单词是否包含否定词或警告短语
	for _, word := range words {
		if isNegativePhrase(word) {
			// 找到明确的否定词，阻止正面短语
			return true
		}
		if isWarningPhrase(word) {
			// 警告短语匹配，不阻止流程
			return false
		}
	}

	// 检查整个文本中是否包含否定关键词（没有、不、避免等）
	// 这用于识别上下文中的否定表达（如"没有绝对安全"、"不保证盈利"等）
	negativeContextKeywords := []string{
		"没有", "不", "避免", "杜绝", "禁止", "勿", "别",
	}
	for _, keyword := range negativeContextKeywords {
		if strings.Contains(text, keyword) {
			// 找到否定关键词，阻止正面短语
			return true
		}
	}

	// 没有找到否定上下文，允许正面短语
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// findMatchedPositivePhrase 找出文本中匹配的正面短语（考虑上下文）
func findMatchedPositivePhrase(text string) string {
	var longestMatch string

	// 首先检查完整单词
	words := strings.Fields(text)
	for _, word := range words {
		if isPositivePhrase(word) && !hasNegativeContext(text, word) {
			if len(word) > len(longestMatch) {
				longestMatch = word
			}
		}
	}

	// 然后检查子串匹配，优先匹配更长的短语
	for _, phrase := range positiveWords {
		if strings.Contains(text, phrase) && !hasNegativeContext(text, phrase) {
			if len(phrase) > len(longestMatch) {
				longestMatch = phrase
			}
		}
	}

	return longestMatch
}

// findMatchedPhrase 找出文本中匹配的短语（优先匹配最长的短语）
func findMatchedPhrase(text string, phrases []string) string {
	var longestMatch string

	// 首先检查完整单词
	words := strings.Fields(text)
	for _, word := range words {
		for _, phrase := range phrases {
			if word == phrase && len(phrase) > len(longestMatch) {
				longestMatch = phrase
			}
		}
	}

	// 然后检查子串匹配，优先匹配更长的短语
	for _, phrase := range phrases {
		if strings.Contains(text, phrase) && len(phrase) > len(longestMatch) {
			longestMatch = phrase
		}
	}

	return longestMatch
}

// hasNegativePhrase 检查是否包含任何负面短语
func hasNegativePhrase(text string) bool {
	// 首先按空格分割，检查完整单词
	words := strings.Fields(text)
	fmt.Printf("    hasNegativePhrase: text='%s', words=%v\n", text, words)

	// 检查完整单词匹配
	for _, word := range words {
		if isNegativePhrase(word) {
			fmt.Printf("      Found complete word match: '%s'\n", word)
			return true
		}
	}

	// 检查子串匹配（作为独立单词出现）
	for _, phrase := range negativeWords {
		// 检查短语是否出现在文本中作为独立单词（前后有边界）
		if strings.Contains(text, phrase) {
			// 将短语分割，检查每个词是否都是独立的完整单词
			phraseWords := strings.Fields(phrase)
			if len(phraseWords) == 0 {
				continue
			}

			// 检查所有短语单词是否都在文本的单词列表中
			allWordsPresent := true
			for _, pw := range phraseWords {
				found := false
				for _, w := range words {
					if w == pw {
						found = true
						break
					}
				}
				if !found {
					allWordsPresent = false
					break
				}
			}

			if allWordsPresent {
				fmt.Printf("      Found phrase: '%s' (words: %v), allWordsPresent=%v\n", phrase, phraseWords, allWordsPresent)
				return true
			}
		}
	}

		// 检查整个句子是否完全匹配负面短语（处理"不得直接满仓"这类句子）
		for _, phrase := range negativeWords {
			if text == phrase {
				fmt.Printf("      Found exact phrase match: '%s'\n", phrase)
				return true
			}
		}

		fmt.Printf("      No negative phrase found\n")
		return false
	}



// debugHasNegativePhrase 调试版本（临时）
func debugHasNegativePhrase(text string) bool {
	words := strings.Fields(text)
	fmt.Printf("[DEBUG] hasNegativePhrase: %s\n", text)
	fmt.Printf("  Words: %v\n", words)

	// 首先检查完整单词
	for _, word := range words {
		if isNegativePhrase(word) {
			fmt.Printf("    Found complete word match: %s\n", word)
			return true
		}
	}

	// 检查子串匹配
	for _, phrase := range negativeWords {
		if strings.Contains(text, phrase) {
			phraseWords := strings.Fields(phrase)
			fmt.Printf("    Contains phrase: %s (words: %v)\n", phrase, phraseWords)

			allWordsPresent := true
			for _, pw := range phraseWords {
				found := false
				for _, w := range words {
					if w == pw {
						found = true
						break
					}
				}
				if !found {
					allWordsPresent = false
					break
				}
			}

			if allWordsPresent {
				fmt.Printf("    ALL WORDS PRESENT, returning true\n")
				return true
			}
		}
	}

	fmt.Printf("    No match found\n")
	return false
}

// hasPhraseBoundary 检查短语是否出现在单词边界
func hasPhraseBoundary(text, phrase string) bool {
	// 将文本分割成单词
	words := strings.Fields(text)
	phraseWords := strings.Fields(phrase)

	// 如果短语包含多个单词，检查所有单词是否都独立存在
	if len(phraseWords) > 1 {
		for _, pw := range phraseWords {
			found := false
			for _, w := range words {
				if w == pw {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// 单词短语，检查前后是否有其他单词（避免部分匹配）
	return true
}

// hasWarningPhrase 检查是否包含任何警告短语
func hasWarningPhrase(text string) bool {
	words := strings.Fields(text)
	for _, word := range words {
		if isWarningPhrase(word) {
			return true
		}
	}

	// 检查子串匹配
	for _, phrase := range overloadingPhrasesWarning {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

// isPositivePhrase 检查单词是否在正面短语列表中
func isPositivePhrase(word string) bool {
	return checkPhrase(word, positiveWords)
}

// isNegativePhrase 检查单词是否在负面短语列表中
func isNegativePhrase(word string) bool {
	result := checkPhrase(word, negativeWords)
	if result {
		fmt.Printf("      isNegativePhrase('%s') = true\n", word)
	} else {
		fmt.Printf("      isNegativePhrase('%s') = false\n", word)
	}
	return result
}

// isWarningPhrase 检查单词是否在警告短语列表中
func isWarningPhrase(word string) bool {
	return checkPhrase(word, overloadingPhrasesWarning)
}

// checkPhrase 检查单词是否匹配任意短语
func checkPhrase(word string, phrases []string) bool {
	// 将短语按空格分割成单词
	phraseWords := make(map[string]struct{})
	for _, phrase := range phrases {
		phraseWords[phrase] = struct{}{}
		fmt.Printf("      Added phrase: '%s'\n", phrase)
	}
	_, exists := phraseWords[word]
	fmt.Printf("      checkPhrase('%s', %d phrases) = %v\n", word, len(phrases), exists)
	return exists
}

// compilePhrasePatterns 编译所有短语的分割匹配函数
func compilePhrasePatterns(phrases []string) func(string) bool {
	// 将所有短语按空格/标点分割成单词列表
	wordSet := make(map[string]struct{})
	for _, phrase := range phrases {
		words := strings.Fields(phrase)
		for _, word := range words {
			wordSet[word] = struct{}{}
		}
	}

	// 返回检查函数
	return func(text string) bool {
		// 首先按空格/标点分割文本
		words := strings.Fields(text)
		// 如果文本中某个单词在列表中，直接返回 true
		for _, word := range words {
			if _, ok := wordSet[word]; ok {
				return true
			}
		}

		// 检查是否包含子串（避免误匹配，例如 "可以满仓买入" 应该匹配 "可以满仓"）
		for phrase := range wordSet {
			if strings.Contains(text, phrase) {
				return true
			}
		}

		return false
	}
}

// hasPhrase 检查文本是否匹配任意一个短语模式
func hasPhrase(text string, checkFunc func(string) bool) bool {
	return checkFunc(text)
}

// hasOverloadingKeyword 检查是否包含满仓/全仓关键词（作为独立单词）
func hasOverloadingKeyword(text string) bool {
	// 满仓
	if hasPhraseWithBoundary(text, "满仓") {
		return true
	}
	// 全仓
	if hasPhraseWithBoundary(text, "全仓") {
		return true
	}
	return false
}

// hasPhraseWithBoundary 检查短语是否匹配（单词边界）
func hasPhraseWithBoundary(text, phrase string) bool {
	escaped := regexp.QuoteMeta(phrase)
	pattern := `\b` + escaped + `\b`
	re := regexp.MustCompile(pattern)
	return re.MatchString(text)
}

// ContainsForbiddenPhrases 检查输出中是否包含禁止表达

// ContainsForbiddenPhrasesInResult 检查 ExtractionResult 中所有字段是否包含禁止表达
func ContainsForbiddenPhrasesInResult(result *model.ExtractionResult) error {
	// 检查的字段列表（共18个字段）
	fieldsToCheck := []struct {
		name string
		text string
	}{
		{"title", result.Title},
		{"summary", result.Summary},
		{"core_conclusion", result.CoreConclusion},
	}

	for _, logic := range result.CoreLogic {
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"core_logic.title", logic.Title})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"core_logic.content", logic.Content})
	}

	for _, scenario := range result.ApplicableScenarios {
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"applicable_scenarios", scenario})
	}

	for _, boundary := range result.RiskBoundaries {
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"risk_boundaries", boundary})
	}

	for _, rule := range result.ExtractableRules {
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"extractable_rules.rule_name", rule.RuleName})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"extractable_rules.summary", rule.Summary})
	}

	for _, rule := range result.CandidateRules {
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.rule_name", rule.RuleName})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.rule_content", rule.RuleContent})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.trigger_conditions", strings.Join(rule.TriggerConditions, ";")})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.actions", strings.Join(rule.Actions, ";")})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.not_applicable", strings.Join(rule.NotApplicable, ";")})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.risk_boundary", rule.RiskBoundary})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.questions_to_confirm", strings.Join(rule.QuestionsToConfirm, ";")})
		fieldsToCheck = append(fieldsToCheck, struct {
			name string
			text string
		}{"candidate_rules.recommendation", rule.Recommendation})
	}

	fieldsToCheck = append(fieldsToCheck, struct {
		name string
		text string
	}{"my_understanding", result.MyUnderstanding})

	// 逐个字段检查
	for _, field := range fieldsToCheck {
		if field.text != "" {
			if err := ContainsForbiddenPhrases(field.text); err != nil {
				return fmt.Errorf("在字段「%s」中： %w", field.name, err)
			}
		}
	}

	return nil
}
