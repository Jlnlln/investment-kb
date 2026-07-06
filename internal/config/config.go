package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	ObsidianVaultPath string `yaml:"obsidian_vault_path"`
	Files             Files  `yaml:"files"`
	AI                AI     `yaml:"ai"`
	Timezone          string `yaml:"timezone"`
}

// Files 文件路径配置
type Files struct {
	RawMaterial            string `yaml:"raw_material"`
	RawMaterialDir         string `yaml:"raw_material_dir"`
	RawMaterialIndex       string `yaml:"raw_material_index"`
	RawInputInboxDir       string `yaml:"raw_input_inbox_dir"`
	QA                     string `yaml:"qa"`
	QADir                  string `yaml:"qa_dir"`
	QAIndex                string `yaml:"qa_index"`
	MarketCase             string `yaml:"market_case"`
	CandidateRule          string `yaml:"candidate_rule"`
	CandidateRuleDir       string `yaml:"candidate_rule_dir"`
	CandidateRuleIndex     string `yaml:"candidate_rule_index"`
	ValidationCardTemplate string `yaml:"validation_card_template"`
	ValidationCardDir      string `yaml:"validation_card_dir"`
	MacroKnowledgeDir      string `yaml:"macro_knowledge_dir"`      // 宏观理解卡目录（单文件模式）
	MacroKnowledgeIndex    string `yaml:"macro_knowledge_index"`    // 宏观理解卡索引文件
	MarketObservationDir   string `yaml:"market_observation_dir"`   // 市场观察卡目录（单文件模式）
	MarketObservationIndex string `yaml:"market_observation_index"` // 市场观察卡索引文件
}

// AI 配置
type AI struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url"`
	APIKeyEnv   string  `yaml:"api_key_env"`
	TimeoutSec  int     `yaml:"timeout_seconds"`
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
