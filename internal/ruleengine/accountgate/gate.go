package accountgate

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Evaluate(input AccountGateInput, params Params) AccountGateOutput {
	if !input.ReserveCashConfirmed && isOneOf(input.ActionType, ActionBuy, ActionAdd, ActionFollow) {
		return output(DecisionBlock, SuggestReviewCash, true, ReasonReserveCashUnconfirmed)
	}
	if input.TotalPositionPct >= params.TotalPositionMaxPct && isOneOf(input.ActionType, ActionBuy, ActionAdd, ActionFollow) {
		return output(DecisionBlock, SuggestReviewReduce, true, ReasonTotalPositionExceedsMax)
	}
	if input.IndexPositionPct >= params.SingleIndexMaxPct && isOneOf(input.ActionType, ActionBuy, ActionAdd, ActionFollow) {
		return output(DecisionBlock, SuggestReviewReduce, true, ReasonSingleIndexExceedsMax)
	}
	if (!input.UpsidePlanReady || !input.DownsidePlanReady) && isOneOf(input.ActionType, ActionBuy, ActionAdd) {
		return output(DecisionBlock, SuggestReviewPlan, true, ReasonPlanMissing)
	}
	if input.FOMOTriggered && isHighOrFull(input.PrimaryPositionState) && isOneOf(input.ActionType, ActionAdd, ActionBuy) {
		return output(DecisionBlock, SuggestBlockAdd, false, ReasonFOMOWithHighPosition)
	}
	if input.InfluencedByOthersPosition && input.ActionType == ActionFollow {
		return output(DecisionBlock, SuggestBlockFollow, false, ReasonInfluencedByOthersPosition)
	}
	if input.ActionType == ActionReduce && hasFlag(input.AccountFlags, FlagProfitCushionThick) {
		return output(DecisionReview, SuggestHoldNoAdd, true, ReasonProfitCushionSubjectiveReduce)
	}
	if input.PrimaryPositionState == PositionEmpty &&
		input.ActionType == ActionBuy &&
		input.ReserveCashConfirmed &&
		input.UpsidePlanReady &&
		input.DownsidePlanReady {
		return output(DecisionLimit, SuggestAllowSmallBuy, false, ReasonAccountStateAllowsSmallBuy)
	}
	if input.PrimaryPositionState == PositionLight &&
		input.ActionType == ActionAdd &&
		input.ReserveCashConfirmed &&
		input.UpsidePlanReady &&
		input.DownsidePlanReady {
		return output(DecisionLimit, SuggestAllowBatchAdd, false, ReasonAccountStateAllowsBatchAdd)
	}
	return output(DecisionReview, SuggestReviewPlan, true)
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

func output(decision Decision, suggested SuggestedAction, review bool, reasons ...ReasonCode) AccountGateOutput {
	return AccountGateOutput{
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

func isHighOrFull(state PositionState) bool {
	return state == PositionHigh || state == PositionFull
}

func hasFlag(flags []AccountFlag, target AccountFlag) bool {
	for _, flag := range flags {
		if flag == target {
			return true
		}
	}
	return false
}
