package plangate

import (
	"path/filepath"
	"testing"
)

func TestPlanGateScenarios(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "plan_probe", "scenarios.yaml")
	scenarios, err := LoadScenarios(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarios) != 6 {
		t.Fatalf("expected 6 scenarios, got %d", len(scenarios))
	}

	for _, scenario := range scenarios {
		t.Run(scenario.ID, func(t *testing.T) {
			if scenario.Expected.Decision == "" {
				t.Fatalf("scenario must have unique expected decision")
			}
			if scenario.Expected.SuggestedAction == "" {
				t.Fatalf("scenario must have unique expected suggested_action")
			}
			actual := Evaluate(scenario.Input, DefaultParams())
			if actual.Decision != scenario.Expected.Decision {
				t.Fatalf("decision = %s, want %s", actual.Decision, scenario.Expected.Decision)
			}
			if actual.SuggestedAction != scenario.Expected.SuggestedAction {
				t.Fatalf("suggested_action = %s, want %s", actual.SuggestedAction, scenario.Expected.SuggestedAction)
			}
			if !sameReasons(actual.ReasonCodes, scenario.Expected.ReasonCodes) {
				t.Fatalf("reason_codes = %#v, want %#v", actual.ReasonCodes, scenario.Expected.ReasonCodes)
			}
		})
	}
}

func sameReasons(a, b []ReasonCode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
