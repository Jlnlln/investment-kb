package app

import (
	"testing"

	"investment-kb/internal/ai"
	"investment-kb/internal/model"
)

func TestValidateExtractionResult_CaseValidation(t *testing.T) {
	tests := []struct {
		name         string
		shouldGenCase bool
		case_        *model.MarketCase
		caseReason   string
		expectError  bool
	}{
		{
			name:         "ShouldGenerateCase=false 时 Case 为 nil（合法）",
			shouldGenCase: false,
			case_:        nil,
			caseReason:   "原因",
			expectError:  false,
		},
		{
			name:         "ShouldGenerateCase=false 时 CaseInsufficientReason 为空（错误）",
			shouldGenCase: false,
			case_:        nil,
			caseReason:   "",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=false 时 Case 不为 nil（错误）",
			shouldGenCase: false,
			case_:        &model.MarketCase{CaseName: "测试"},
			caseReason:   "原因",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=true 时 Case 为 nil（错误）",
			shouldGenCase: true,
			case_:        nil,
			caseReason:   "原因",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=true 时 CaseName 为空（错误）",
			shouldGenCase: true,
			case_:        &model.MarketCase{},
			caseReason:   "原因",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=true 时 DomainCode 为空（错误）",
			shouldGenCase: true,
			case_: &model.MarketCase{
				CaseName:      "测试案例",
				DomainCode:    "",
				TopicCode:     "交易",
				KeyDecisionQuestion: "如何决策?",
				FinalInsight:  "心得",
			},
			caseReason:   "原因",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=true 时所有字段都为空（错误）",
			shouldGenCase: true,
			case_: &model.MarketCase{
				CaseName:      "",
				DomainCode:    "",
				TopicCode:     "",
				KeyDecisionQuestion: "",
				FinalInsight:  "",
			},
			caseReason:   "原因",
			expectError:  true,
		},
		{
			name:         "ShouldGenerateCase=true 时所有字段都合法",
			shouldGenCase: true,
			case_: &model.MarketCase{
				CaseName:      "测试案例",
				DomainCode:    "买入",
				TopicCode:     "交易",
				KeyDecisionQuestion: "如何决策?",
				FinalInsight:  "心得",
			},
			caseReason:   "原因",
			expectError:  false,
		},
		{
			name:         "ShouldGenerateCase=false 时所有字段都合法",
			shouldGenCase: false,
			case_:        nil,
			caseReason:   "原因",
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := &model.ExtractionResult{
				ShouldGenerateCase: tc.shouldGenCase,
				Case:               tc.case_,
				CaseInsufficientReason: tc.caseReason,
			}

			err := validateExtractionResult(result, "test", "测试")
			if tc.expectError {
				if err == nil {
					t.Errorf("期望错误但未返回错误")
				}
			} else {
				if err != nil {
					t.Errorf("期望成功但返回错误: %v", err)
				}
			}
		})
	}
}

func TestCheckAbsoluteClaims(t *testing.T) {
	positiveResult := &model.ExtractionResult{
		Summary:         "这个方法一定能赚钱",
		RiskBoundaries:  []string{"保证盈利", "无风险"},
	}

	negativeResult := &model.ExtractionResult{
		Summary:         "这个方法可能会赚钱，也有可能亏损",
		RiskBoundaries:  []string{"可能会亏损"},
	}

	t.Run("包含绝对化表达应返回错误", func(t *testing.T) {
		err := checkAbsoluteClaims(positiveResult)
		if err == nil {
			t.Errorf("期望错误但未返回错误")
		}
		if err != nil && err.Error() == "" {
			t.Errorf("错误信息为空")
		}
	})

	t.Run("不包含绝对化表达应通过", func(t *testing.T) {
		err := checkAbsoluteClaims(negativeResult)
		if err != nil {
			t.Errorf("期望成功但返回错误: %v", err)
		}
	})
}

func TestContainsForbiddenPhrases(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		expectError bool
		errContains string
	}{
		{
			name:  "保证盈利",
			text:  "这个方法保证盈利",
			expectError: true,
			errContains: "保证盈利",
		},
		{
			name:  "没有亏损风险",
			text:  "没有亏损风险",
			expectError: true,
			errContains: "没有亏损风险",
		},
		{
			name:  "必然上涨",
			text:  "必然上涨",
			expectError: true,
			errContains: "必然上涨",
		},
		{
			name:  "一定赚钱",
			text:  "一定赚钱",
			expectError: true,
			errContains: "一定赚钱",
		},
		{
			name:  "判断错了也不会亏",
			text:  "判断错了也不会亏",
			expectError: true,
			errContains: "判断错了也不会亏",
		},
		{
			name:  "只赚不亏",
			text:  "只赚不亏",
			expectError: true,
			errContains: "只赚不亏",
		},
		{
			name:  "无风险",
			text:  "无风险",
			expectError: true,
			errContains: "无风险",
		},
		{
			name:  "稳赚",
			text:  "稳赚",
			expectError: true,
			errContains: "稳赚",
		},
		{
			name:  "绝对安全",
			text:  "绝对安全",
			expectError: true,
			errContains: "绝对安全",
		},
		{
			name:  "必胜",
			text:  "必胜",
			expectError: true,
			errContains: "必胜",
		},
		{
			name:  "可直接满仓",
			text:  "可直接满仓",
			expectError: true,
			errContains: "可直接满仓",
		},
		{
			name:  "应直接满仓",
			text:  "应直接满仓",
			expectError: true,
			errContains: "应直接满仓",
		},
		{
			name:  "可以满仓",
			text:  "可以满仓",
			expectError: true,
			errContains: "可以满仓",
		},
		{
			name:  "满仓买入",
			text:  "满仓买入",
			expectError: true,
			errContains: "满仓买入",
		},
		{
			name:  "直接满仓",
			text:  "直接满仓",
			expectError: true,
			errContains: "直接满仓",
		},
		{
			name:  "高确定性时可直接满仓",
			text:  "高确定性时可直接满仓",
			expectError: true,
			errContains: "高确定性时可直接满仓",
		},
		{
			name:  "不包含任何禁止表达",
			text:  "这个方法可能会赚钱，也有可能亏损",
			expectError: false,
		},
		{
			name:  "多个禁止表达",
			text:  "保证盈利且没有亏损风险",
			expectError: true,
			errContains: "没有亏损风险",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ai.ContainsForbiddenPhrases(tt.text)
			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但未返回错误")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("错误信息应包含「%s」，实际：「%s」", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("期望成功但返回错误: %v", err)
				}
			}
		})
	}
}

func TestWarnRuleTypeDomainMismatch(t *testing.T) {
	tests := []struct {
		name         string
		rules        []model.CandidateRule
		expectWarnings int
	}{
		{
			name:         "所有 domain_code 都是 BUY",
			rules: []model.CandidateRule{
				{RuleType: "买入规则", DomainCode: "BUY", TopicCode: "SAFETY", RuleName: "高概率区间先建底仓"},
				{RuleType: "买入规则", DomainCode: "BUY", TopicCode: "POSITION", RuleName: "仓位管理"},
			},
			expectWarnings: 0,
		},
		{
			name: "一条规则 domain_code 不匹配",
			rules: []model.CandidateRule{
				{RuleType: "买入规则", DomainCode: "BUY", TopicCode: "SAFETY", RuleName: "高概率区间先建底仓"},
				{RuleType: "买入规则", DomainCode: "SAFETY", TopicCode: "SAFETY", RuleName: "安全边际"},
			},
			expectWarnings: 1,
		},
		{
			name: "多条规则 domain_code 不匹配",
			rules: []model.CandidateRule{
				{RuleType: "买入规则", DomainCode: "BUY", TopicCode: "SAFETY", RuleName: "高概率区间先建底仓"},
				{RuleType: "买入规则", DomainCode: "RISK", TopicCode: "控制", RuleName: "风险控制"},
				{RuleType: "买入规则", DomainCode: "ACCOUNT", TopicCode: "资金", RuleName: "资金管理"},
			},
			expectWarnings: 2,
		},
		{
			name:         "没有买入规则",
			rules:        []model.CandidateRule{},
			expectWarnings: 0,
		},
		{
			name: "只有非买入规则",
			rules: []model.CandidateRule{
				{RuleType: "仓位规则", DomainCode: "POSITION", TopicCode: "仓位", RuleName: "仓位调整"},
				{RuleType: "风控规则", DomainCode: "RISK", TopicCode: "风控", RuleName: "止损规则"},
			},
			expectWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := WarnRuleTypeDomainMismatch(&model.ExtractionResult{
				CandidateRules: tt.rules,
			})
			if len(warnings) != tt.expectWarnings {
				t.Errorf("期望 %d 条 warning，实际 %d 条", tt.expectWarnings, len(warnings))
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
