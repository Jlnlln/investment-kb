package plangate

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Evaluate(input PlanGateInput, params Params) PlanGateOutput {
	if isTrue(input.PlanChangedAfterAction) && isTrue(input.EmotionalOverrideDetected) {
		return output(DecisionBlock, SuggestRejectEmotionalPlanChange, false, ReasonEmotionalPlanChangeDetected)
	}
	if params.PlanLockRequiredBeforeBuy && isOneOf(input.ActionType, ActionBuy, ActionAdd) && isFalse(input.PlanLockedBeforeAction) {
		return output(DecisionBlock, SuggestReviewPlan, true, ReasonPlanNotLockedBeforeAction)
	}
	if isOneOf(input.ActionType, ActionBuy, ActionAdd) && isFalse(input.DownsidePlanReady) {
		return output(DecisionBlock, SuggestRequireDownsidePlan, true, ReasonDownsidePlanMissing)
	}
	if isOneOf(input.ActionType, ActionBuy, ActionAdd) && isFalse(input.ExitPlanReady) {
		reasons := []ReasonCode{ReasonExitPlanMissing}
		if isTrue(input.IsChasingHigh) {
			reasons = appendChasingHighMissingReasons(reasons, input, params)
		}
		return output(DecisionBlock, SuggestRequireExitPlan, true, reasons...)
	}
	if isTrue(input.IsChasingHigh) && hasChasingHighMissing(input, params) {
		return output(DecisionBlock, SuggestRequireExitPlan, true, appendChasingHighMissingReasons(nil, input, params)...)
	}
	if params.ExtremeDrawdownPlanRequiredForHeavyPosition && isTrue(input.IsHeavyPositionIntent) && isFalse(input.ExtremeDrawdownPlanReady) {
		return output(DecisionBlock, SuggestRequireExtremeDrawdownPlan, true, ReasonExtremeDrawdownPlanMissing, ReasonHeavyPositionRequiresExtremePlan)
	}
	if params.MaxLossPctRequired && isOneOf(input.ActionType, ActionBuy, ActionAdd) && isTrue(input.ExitPlanReady) && isFalse(input.MaxLossDefined) {
		return output(DecisionReview, SuggestReviewRiskExposure, true, ReasonMaxLossMissing)
	}
	if input.ActionType == ActionReduce && isTrue(input.EmotionalOverrideDetected) {
		return output(DecisionReview, SuggestReviewExitCondition, true, ReasonEmotionalPlanChangeDetected)
	}
	if isSupportedAction(input.ActionType) && planComplete(input) {
		return output(DecisionAllow, SuggestAllowWithPlan, false, ReasonPlanComplete)
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

func output(decision Decision, suggested SuggestedAction, review bool, reasons ...ReasonCode) PlanGateOutput {
	return PlanGateOutput{
		Decision:            decision,
		SuggestedAction:     suggested,
		ReasonCodes:         reasons,
		HumanReviewRequired: review,
	}
}

func appendChasingHighMissingReasons(reasons []ReasonCode, input PlanGateInput, params Params) []ReasonCode {
	if isFalse(input.ExitConditionDefined) {
		reasons = append(reasons, ReasonExitConditionMissing)
	}
	if params.MaxLossPctRequired && isFalse(input.MaxLossDefined) {
		reasons = append(reasons, ReasonMaxLossMissing)
	}
	if params.RiskExposureRequiredForChasingHigh && isFalse(input.RiskExposureKnown) {
		reasons = append(reasons, ReasonRiskExposureUnknown)
	}
	if !containsReason(reasons, ReasonChasingHighRequiresExitPlan) {
		reasons = append(reasons, ReasonChasingHighRequiresExitPlan)
	}
	return reasons
}

func hasChasingHighMissing(input PlanGateInput, params Params) bool {
	return isFalse(input.ExitConditionDefined) ||
		(params.MaxLossPctRequired && isFalse(input.MaxLossDefined)) ||
		(params.RiskExposureRequiredForChasingHigh && isFalse(input.RiskExposureKnown))
}

func planComplete(input PlanGateInput) bool {
	if isOneOf(input.ActionType, ActionBuy, ActionAdd) {
		return isTrue(input.UpsidePlanReady) &&
			isTrue(input.DownsidePlanReady) &&
			isTrue(input.ExitPlanReady) &&
			isTrue(input.ExitConditionDefined) &&
			isTrue(input.MaxLossDefined) &&
			isTrue(input.RiskExposureKnown) &&
			isTrue(input.PlanLockedBeforeAction)
	}
	return !isFalse(input.PlanLockedBeforeAction)
}

func isOneOf(value ActionType, candidates ...ActionType) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func isSupportedAction(value ActionType) bool {
	return isOneOf(value, ActionBuy, ActionAdd, ActionHold, ActionReduce, ActionRebalance, ActionFollow)
}

func isTrue(value *bool) bool {
	return value != nil && *value
}

func isFalse(value *bool) bool {
	return value != nil && !*value
}

func containsReason(reasons []ReasonCode, target ReasonCode) bool {
	for _, reason := range reasons {
		if reason == target {
			return true
		}
	}
	return false
}
