package stressgate

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Evaluate(input StressGateInput, params Params) StressGateOutput {
	if isTrue(input.IsHeavyPositionIntent) && isFalse(input.ExtremeDrawdownPlanReady) {
		return output(DecisionBlock, SuggestReviewStressPlan, true, ReasonStressPlanMissing, ReasonHeavyPositionRequiresStressTest)
	}
	if isFalse(input.ReserveCashConfirmed) && isOneOf(input.ActionType, ActionBuy, ActionAdd) {
		return output(DecisionBlock, SuggestReviewCashSafety, true, ReasonReserveCashUnconfirmed)
	}
	if bothProvided(input.StressLossCNY, input.MaxLossToleranceCNY) && *input.StressLossCNY > *input.MaxLossToleranceCNY {
		return output(DecisionBlock, SuggestReducePlannedBuyPct, true, ReasonStressLossExceedsTolerance)
	}
	if bothProvided(input.StressLossPct, input.MaxLossTolerancePct) && *input.StressLossPct > *input.MaxLossTolerancePct {
		return output(DecisionBlock, SuggestReducePlannedBuyPct, true, ReasonStressLossExceedsTolerance)
	}
	if bothProvided(input.CashAfterStressCNY, input.MinCashRequiredCNY) && *input.CashAfterStressCNY < *input.MinCashRequiredCNY {
		return output(DecisionBlock, SuggestReviewCashSafety, true, ReasonCashAfterStressBelowMin)
	}
	if bothProvided(input.StressLossCNY, input.MaxLossToleranceCNY) &&
		*input.MaxLossToleranceCNY > 0 &&
		(*input.StressLossCNY / *input.MaxLossToleranceCNY * 100) >= params.NearToleranceThresholdPct {
		return output(DecisionReview, SuggestReviewStressPlan, true, ReasonStressLossNearTolerance)
	}
	return output(DecisionAllow, SuggestAllowWithStressPassed, false, ReasonStressTestPassed)
}

func LoadScenarios(path string) ([]Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var scenarios []Scenario
	if err := yaml.Unmarshal(data, &scenarios); err != nil {
		return nil, err
	}
	return scenarios, nil
}

func output(decision Decision, suggested SuggestedAction, review bool, reasons ...ReasonCode) StressGateOutput {
	return StressGateOutput{
		Decision:            decision,
		SuggestedAction:     suggested,
		ReasonCodes:         reasons,
		HumanReviewRequired: review,
	}
}

func isOneOf(value ActionType, candidates ...ActionType) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func isTrue(value *bool) bool {
	return value != nil && *value
}

func isFalse(value *bool) bool {
	return value != nil && !*value
}

func bothProvided(a, b *float64) bool {
	return a != nil && b != nil
}
