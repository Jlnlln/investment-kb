package ai

import (
	"testing"
)

func TestCheckOverloadingPhrases_AllowNegative(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  error
	}{
		{
			name:  "不能满仓",
			text:  "判断错了不会亏，不能满仓",
			want:  nil,
		},
		{
			name:  "不得满仓",
			text:  "不得直接满仓",
			want:  nil,
		},
		{
			name:  "不要满仓",
			text:  "不要满仓买入",
			want:  nil,
		},
		{
			name:  "不可直接满仓",
			text:  "不可直接满仓",
			want:  nil,
		},
		{
			name:  "避免满仓",
			text:  "避免满仓",
			want:  nil,
		},
		{
			name:  "不能一次性全仓押注",
			text:  "不能一次性全仓押注",
			want:  nil,
		},
		{
			name:  "不得照搬满仓计划",
			text:  "不得照搬满仓计划",
			want:  nil,
		},
		{
			name:  "不能照搬低成本持仓者的满仓计划",
			text:  "不能照搬低成本持仓者的满仓计划",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if tt.want == nil {
				if got != nil {
					t.Errorf("期望成功但返回错误: %v", got)
				}
			} else {
				if got == nil {
					t.Errorf("期望错误但返回成功")
				}
			}
		})
	}
}

func TestCheckOverloadingPhrases_FailPositive(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		errContains string
	}{
		{
			name:  "可以满仓",
			text:  "可以满仓买入",
			errContains: "可以满仓",
		},
		{
			name:  "应直接满仓",
			text:  "应直接满仓",
			errContains: "应直接满仓",
		},
		{
			name:  "可直接满仓",
			text:  "可直接满仓",
			errContains: "可直接满仓",
		},
		{
			name:  "满仓买入",
			text:  "满仓买入",
			errContains: "满仓买入",
		},
		{
			name:  "直接满仓",
			text:  "直接满仓",
			errContains: "直接满仓",
		},
		{
			name:  "高确定性时可直接满仓",
			text:  "高确定性时可直接满仓",
			errContains: "高确定性时可直接满仓",
		},
		// 以下两个是警告短语，不在 FailPositive 测试中（在 WarningPhrases 测试中）
		// {
		// 	name:  "可以一次性全仓",
		// 	text:  "可以一次性全仓",
		// 	errContains: "可以一次性全仓",
		// },
		// {
		// 	name:  "应该全仓押注",
		// 	text:  "应该全仓押注",
		// 	errContains: "应该全仓押注",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if got == nil {
				t.Errorf("期望错误但返回成功")
			} else if tt.errContains != "" && !contains(got.Error(), tt.errContains) {
				t.Errorf("错误信息应包含「%s」，实际：「%s」", tt.errContains, got.Error())
			}
		})
	}
}

func TestCheckOverloadingPhrases_WarningPhrases(t *testing.T) {
	tests := []struct {
		name  string
		text  string
	}{
		{
			name:  "应该全仓押注（warning）",
			text:  "判断对了就应该全仓押注",
		},
		{
			name:  "可以一次性全仓（warning）",
			text:  "可以一次性全仓押注",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if got != nil {
				t.Errorf("期望成功但返回错误: %v", got)
			}
		})
	}
}

func TestCheckOverloadingPhrases_OnlyKeywordWithoutContext(t *testing.T) {
	tests := []struct {
		name  string
		text  string
	}{
		{
			name:  "只有满仓关键词",
			text:  "现在的仓位是满仓",
		},
		{
			name:  "只有全仓关键词",
			text:  "全仓押注是一个好策略",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if got != nil {
				t.Errorf("期望成功但返回错误: %v", got)
			}
		})
	}
}

func TestCheckOverloadingPhrases_NegativeAfterKeyword(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  error
	}{
		{
			name:  "满仓但不能",
			text:  "满仓但不能满仓",
			want:  nil,
		},
		{
			name:  "全仓不要",
			text:  "全仓不要全仓",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if tt.want == nil {
				if got != nil {
					t.Errorf("期望成功但返回错误: %v", got)
				}
			} else {
				if got == nil {
					t.Errorf("期望错误但返回成功")
				}
			}
		})
	}
}

func TestCheckOverloadingPhrases_PositiveBeforeKeyword(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		errContains string
	}{
		{
			name:  "应该满仓",
			text:  "应该满仓",
			errContains: "应该满仓",
		},
		{
			name:  "必须满仓",
			text:  "必须满仓",
			errContains: "必须满仓",
		},
		{
			name:  "可以满仓（前后都有词）",
			text:  "可以满仓买入",
			errContains: "可以满仓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if got == nil {
				t.Errorf("期望错误但返回成功")
			} else if tt.errContains != "" && !contains(got.Error(), tt.errContains) {
				t.Errorf("错误信息应包含「%s」，实际：「%s」", tt.errContains, got.Error())
			}
		})
	}
}

func TestCheckOverloadingPhrases_WithOtherPhrases(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  error
	}{
		{
			name:  "不能满仓且有保证盈利",
			text:  "不能满仓，保证盈利",
			want:  nil,
		},
		{
			name:  "应该满仓但没有绝对安全",
			text:  "应该满仓，没有绝对安全",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsForbiddenPhrases(tt.text)
			if tt.want == nil {
				if got != nil {
					t.Errorf("期望成功但返回错误: %v", got)
				}
			} else {
				if got == nil {
					t.Errorf("期望错误但返回成功")
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
