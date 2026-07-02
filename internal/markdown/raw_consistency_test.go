package markdown

import (
	"testing"

	"investment-kb/internal/model"
)

func TestValidateRawConsistencyChineseSemanticKeywords(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		rawText string
		wantOK  bool
	}{
		{
			name:    "growth board cycle limit spread reference passes",
			title:   "创业板周期极限与差值参考",
			rawText: "这段材料讨论创业板在周期位置上的极限状态，并用差值作为辅助参考。",
			wantOK:  true,
		},
		{
			name:    "safety margin missed opportunity passes",
			title:   "安全边际与错失买入机会如何平衡",
			rawText: "原文讨论安全边际过高时，可能错失买入机会，需要结合账户状态处理。",
			wantOK:  true,
		},
		{
			name:    "unrelated title and body warns",
			title:   "创业板周期极限与差值参考",
			rawText: "这段材料只讨论利率、通胀和消费变化，没有涉及相关指数主题。",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateRawConsistency(&model.ExtractionResult{Title: tt.title}, tt.rawText)
			gotOK := len(warnings) == 0
			if gotOK != tt.wantOK {
				t.Fatalf("ValidateRawConsistency ok = %v, want %v; warnings=%v", gotOK, tt.wantOK, warnings)
			}
		})
	}
}

func TestCoreKeywordsDoesNotExtractPseudoChineseFragments(t *testing.T) {
	keywords := coreKeywords("创业板周期极限与差值参考")
	for _, bad := range []string{"创业板周", "板周期极"} {
		for _, keyword := range keywords {
			if keyword == bad {
				t.Fatalf("coreKeywords extracted pseudo keyword %q from %v", bad, keywords)
			}
		}
	}

	want := map[string]bool{
		"创业板": true,
		"周期":  true,
		"极限":  true,
		"差值":  true,
	}
	for keyword := range want {
		if !containsString(keywords, keyword) {
			t.Fatalf("coreKeywords missing %q from %v", keyword, keywords)
		}
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
