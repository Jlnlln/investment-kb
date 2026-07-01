package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	ObsidianVaultPath string   `yaml:"obsidian_vault_path"`
	Files             Files    `yaml:"files"`
	AI                AI       `yaml:"ai"`
	Timezone          string   `yaml:"timezone"`
}

// Files 文件路径配置
type Files struct {
	RawMaterial             string `yaml:"raw_material"`
	QA                      string `yaml:"qa"`
	MarketCase              string `yaml:"market_case"`
	CandidateRule           string `yaml:"candidate_rule"`
	ValidationCardTemplate  string `yaml:"validation_card_template"`
	ValidationCardDir       string `yaml:"validation_card_dir"`
	MacroKnowledge          string `yaml:"macro_knowledge"`     // 宏观理解卡库
	MarketObservation       string `yaml:"market_observation"` // 市场观察卡库
}

// AI 配置
type AI struct {
	Provider    string `yaml:"provider"`
	Model       string `yaml:"model"`
	BaseURL     string `yaml:"base_url"`
	APIKeyEnv   string `yaml:"api_key_env"`
	TimeoutSec  int    `yaml:"timeout_seconds"`
	Temperature float64 `yaml:"temperature"` // 默认 0，确保输出稳定性
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// GetAPIKey 获取 API Key
func (c *Config) GetAPIKey() string {
	if c.AI.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(c.AI.APIKeyEnv)
}