package plangate

type ActionType string

const (
	ActionBuy       ActionType = "buy"
	ActionAdd       ActionType = "add"
	ActionHold      ActionType = "hold"
	ActionReduce    ActionType = "reduce"
	ActionRebalance ActionType = "rebalance"
	ActionFollow    ActionType = "follow"
)

type Decision string

const (
	DecisionAllow  Decision = "allow"
	DecisionBlock  Decision = "block"
	DecisionReview Decision = "review"
)

type SuggestedAction string

const (
	SuggestAllowWithPlan              SuggestedAction = "allow_with_plan"
	SuggestReviewPlan                 SuggestedAction = "review_plan"
	SuggestRequireDownsidePlan        SuggestedAction = "require_downside_plan"
	SuggestRequireExitPlan            SuggestedAction = "require_exit_plan"
	SuggestRequireExtremeDrawdownPlan SuggestedAction = "require_extreme_drawdown_plan"
	SuggestBlockOperation             SuggestedAction = "block_operation"
	SuggestRejectEmotionalPlanChange  SuggestedAction = "reject_emotional_plan_change"
	SuggestReviewExitCondition        SuggestedAction = "review_exit_condition"
	SuggestReviewRiskExposure         SuggestedAction = "review_risk_exposure"
)

type ReasonCode string

const (
	ReasonPlanComplete                     ReasonCode = "plan_complete"
	ReasonDownsidePlanMissing              ReasonCode = "downside_plan_missing"
	ReasonUpsidePlanMissing                ReasonCode = "upside_plan_missing"
	ReasonExitPlanMissing                  ReasonCode = "exit_plan_missing"
	ReasonExtremeDrawdownPlanMissing       ReasonCode = "extreme_drawdown_plan_missing"
	ReasonExitConditionMissing             ReasonCode = "exit_condition_missing"
	ReasonMaxLossMissing                   ReasonCode = "max_loss_missing"
	ReasonRiskExposureUnknown              ReasonCode = "risk_exposure_unknown"
	ReasonPlanNotLockedBeforeAction        ReasonCode = "plan_not_locked_before_action"
	ReasonEmotionalPlanChangeDetected      ReasonCode = "emotional_plan_change_detected"
	ReasonHeavyPositionRequiresExtremePlan ReasonCode = "heavy_position_requires_extreme_plan"
	ReasonChasingHighRequiresExitPlan      ReasonCode = "chasing_high_requires_exit_plan"
)

type PlanGateInput struct {
	ActionType ActionType `yaml:"action_type"`

	UpsidePlanReady          *bool `yaml:"upside_plan_ready"`
	DownsidePlanReady        *bool `yaml:"downside_plan_ready"`
	ExitPlanReady            *bool `yaml:"exit_plan_ready"`
	ExtremeDrawdownPlanReady *bool `yaml:"extreme_drawdown_plan_ready"`

	ExitConditionDefined *bool   `yaml:"exit_condition_defined"`
	MaxLossDefined       *bool   `yaml:"max_loss_defined"`
	MaxLossPct           float64 `yaml:"max_loss_pct"`
	RiskExposureKnown    *bool   `yaml:"risk_exposure_known"`

	PlanLockedBeforeAction    *bool `yaml:"plan_locked_before_action"`
	PlanChangedAfterAction    *bool `yaml:"plan_changed_after_action"`
	EmotionalOverrideDetected *bool `yaml:"emotional_override_detected"`

	IsHeavyPositionIntent *bool `yaml:"is_heavy_position_intent"`
	IsChasingHigh         *bool `yaml:"is_chasing_high"`
}

type PlanGateOutput struct {
	Decision            Decision        `yaml:"decision"`
	SuggestedAction     SuggestedAction `yaml:"suggested_action"`
	ReasonCodes         []ReasonCode    `yaml:"reason_codes"`
	HumanReviewRequired bool            `yaml:"human_review_required"`
}

type Scenario struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Input    PlanGateInput  `yaml:"input"`
	Expected PlanGateOutput `yaml:"expected"`
	Note     string         `yaml:"note"`
}
