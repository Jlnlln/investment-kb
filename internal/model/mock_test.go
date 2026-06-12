package model

import (
	"testing"
)

func TestMockExtractionResult(t *testing.T) {
	result := MockExtractionResult()

	// 验证基本信息
	if result.Title == "" {
		t.Error("MockExtractionResult: Title 为空")
	}
	if result.Source == "" {
		t.Error("MockExtractionResult: Source 为空")
	}

	// 验证核心逻辑
	if len(result.CoreLogic) == 0 {
		t.Error("MockExtractionResult: CoreLogic 为空")
	}

	// 验证候选规则
	if len(result.CandidateRules) == 0 {
		t.Error("MockExtractionResult: CandidateRules 为空")
	}

	// 验证每条规则有必要的字段
	for i, rule := range result.CandidateRules {
		if rule.RuleName == "" {
			t.Errorf("MockExtractionResult: 规则 %d 的 RuleName 为空", i)
		}
		if len(rule.TriggerConditions) == 0 {
			t.Errorf("MockExtractionResult: 规则 %d 的 TriggerConditions 为空", i)
		}
		if len(rule.Actions) == 0 {
			t.Errorf("MockExtractionResult: 规则 %d 的 Actions 为空", i)
		}
	}
}