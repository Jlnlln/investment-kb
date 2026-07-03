package ai

import (
	"context"
	"fmt"
)

// Client 是 AI 调用的通用接口
type Client interface {
	// Complete 发送请求并返回原始文本
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)

	// CompleteJSON 发送请求并将结果解析为结构体
	CompleteJSON(ctx context.Context, systemPrompt string, userPrompt string, v any) error
}

// Config 是 AI 客户端配置
type Config struct {
	Provider    string
	Model       string
	BaseURL     string
	APIKey      string
	TimeoutSec  int
	MaxRetries  int     // 仅用于 Complete 原始文本调用；CompleteJSON 固定最多 3 次尝试。
	Temperature float64 // 默认 0，确保输出稳定性
}

// NewClient 根据配置创建对应的 AI 客户端
func NewClient(cfg *Config) (Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API Key 不能为空")
	}

	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 300
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.Temperature <= 0 {
		cfg.Temperature = 0 // 默认 0，确保输出稳定性
	}

	switch cfg.Provider {
	case "custom":
		return newCustomClient(cfg), nil
	default:
		return nil, fmt.Errorf("不支持的 AI provider: %s", cfg.Provider)
	}
}
