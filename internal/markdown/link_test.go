package markdown

import (
	"strings"
	"testing"

	"investment-kb/internal/config"
)

func TestObsidianHeadingLink(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		heading  string
		alias    string
		want     string
	}{
		{
			name:     "with .md suffix",
			filePath: "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.md",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA-POS-SAFETY-20260617-001",
			want:     "[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]",
		},
		{
			name:     "with .MD suffix",
			filePath: "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.MD",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA-POS-SAFETY-20260617-001",
			want:     "[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]",
		},
		{
			name:     "no suffix",
			filePath: "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA-POS-SAFETY-20260617-001",
			want:     "[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]",
		},
		{
			name:     "with alias",
			filePath: "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.md",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA标题",
			want:     "[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA标题]]",
		},
		{
			name:     "Windows path with backslashes",
			filePath: "G:\\Obsidian\\我的知识库\\03-知识与案例\\问答知识库.md",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA-POS-SAFETY-20260617-001",
			want:     "[[G:/Obsidian/我的知识库/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]",
		},
		{
			name:     "with backslashes and .md",
			filePath: "G:\\Obsidian\\我的知识库\\03-知识与案例\\问答知识库.md",
			heading:  "QA-POS-SAFETY-20260617-001",
			alias:    "QA-POS-SAFETY-20260617-001",
			want:     "[[G:/Obsidian/我的知识库/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ObsidianHeadingLink(tt.filePath, tt.heading, tt.alias)
			if got != tt.want {
				t.Errorf("ObsidianHeadingLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		suffix   string
		expected string
	}{
		{
			name:     "trim .md",
			input:    "问答知识库.md",
			suffix:   ".md",
			expected: "问答知识库",
		},
		{
			name:     "trim .MD",
			input:    "问答知识库.MD",
			suffix:   ".MD",
			expected: "问答知识库",
		},
		{
			name:     "no suffix to trim",
			input:    "问答知识库",
			suffix:   ".md",
			expected: "问答知识库",
		},
		{
			name:     "empty input",
			input:    "",
			suffix:   ".md",
			expected: "",
		},
		{
			name:     "suffix after slash",
			input:    "path/to/file.md",
			suffix:   ".md",
			expected: "path/to/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.TrimSuffix(tt.input, tt.suffix)
			if result != tt.expected {
				t.Errorf("TrimSuffix(%q, %q) = %q, want %q", tt.input, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestReplaceAllBackslashes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows backslashes to forward slashes",
			input:    "G:\\Obsidian\\我的知识库\\03-知识与案例\\问答知识库.md",
			expected: "G:/Obsidian/我的知识库/03-知识与案例/问答知识库.md",
		},
		{
			name:     "already forward slashes",
			input:    "G:/Obsidian/我的知识库/03-知识与案例/问答知识库.md",
			expected: "G:/Obsidian/我的知识库/03-知识与案例/问答知识库.md",
		},
		{
			name:     "no backslashes",
			input:    "问答知识库.md",
			expected: "问答知识库.md",
		},
		{
			name:     "mixed separators",
			input:    "G:\\Obsidian/我的知识库/问答知识库.md",
			expected: "G:/Obsidian/我的知识库/问答知识库.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.ReplaceAll(tt.input, "\\", "/")
			if result != tt.expected {
				t.Errorf("ReplaceAllBackslashes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCandidateRuleStandaloneLink(t *testing.T) {
	cfg := &config.Config{}
	cfg.Files.CandidateRuleDir = "rules/candidate_rules"
	cfg.Files.CandidateRuleIndex = "rules/candidate_rules/index.md"

	got := CandidateRuleLink(cfg, "CR-VALUATION-20260702-001", "VALUATION", "SAFETY", "高概率区间先建底仓", "")
	want := "[[rules/candidate_rules/CR-VALUATION-20260702-001｜VALUATION-SAFETY｜高概率区间先建底仓|CR-VALUATION-20260702-001]]"
	if got != want {
		t.Fatalf("CandidateRuleLink() = %s, want %s", got, want)
	}
	if strings.Contains(got, "#") {
		t.Fatalf("standalone candidate rule link should not use heading anchor: %s", got)
	}
}

func TestKnowLinkUsesPathAndIDAlias(t *testing.T) {
	cfg := &config.Config{}
	cfg.Files.MacroKnowledgeDir = "knowledge/macro_cards"

	got := KnowLink(cfg, "KNOW-L3-RATE-20260702-001", "消费与输入性通胀对利率走向的制约")
	want := "[[knowledge/macro_cards/KNOW-L3-RATE-20260702-001｜消费与输入性通胀对利率走向的制约|KNOW-L3-RATE-20260702-001]]"
	if got != want {
		t.Fatalf("KnowLink() = %s, want %s", got, want)
	}
}

func TestCandidateRuleFileNameSanitizesWindowsInvalidChars(t *testing.T) {
	got := CandidateRuleFileName("CR-RISK-20260702-001", "RISK", "PLAN", `买入/卖出:预案? "检查"`)
	if strings.ContainsAny(got, `\/:*?"<>|`) {
		t.Fatalf("CandidateRuleFileName contains Windows invalid chars: %s", got)
	}
	if !strings.Contains(got, "｜") {
		t.Fatalf("CandidateRuleFileName should preserve fullwidth separators: %s", got)
	}
}
