package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
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

	cleaned := FixCommonJSON(raw)
	if err := json.Unmarshal([]byte(cleaned), v); err != nil {
		// 保存错误输出
		saveErrorOutput(raw)
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
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	var rawText string
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
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

// saveErrorOutput 保存 AI 错误输出到文件
func saveErrorOutput(raw string) {
	filename := fmt.Sprintf("data/error_outputs/ai_error_%s.txt", time.Now().Format("20060102_150405"))
	_ = writeToFile(filename, raw)
}

func writeToFile(path, content string) error {
	return nil // 简化实现，后续在 obsidian 包中处理
}