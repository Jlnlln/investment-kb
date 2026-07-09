package stressgate

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
	DecisionLimit  Decision = "limit"
	DecisionBlock  Decision = "block"
	DecisionReview Decision = "review"
)

type SuggestedAction string

const (
	SuggestAllowWithStressPassed SuggestedAction = "allow_with_stress_passed"
	SuggestReducePlannedBuyPct   SuggestedAction = "reduce_planned_buy_pct"
	SuggestLowerTargetPosition   SuggestedAction = "lower_target_position"
	SuggestReviewStressPlan      SuggestedAction = "review_stress_plan"
	SuggestReviewCashSafety      SuggestedAction = "review_cash_safety"
	SuggestBlockHeavyPosition    SuggestedAction = "block_heavy_position"
	SuggestBlockOperation        SuggestedAction = "block_operation"
)

type ReasonCode string

const (
	ReasonStressTestPassed                 ReasonCode = "stress_test_passed"
	ReasonStressLossExceedsTolerance       ReasonCode = "stress_loss_exceeds_tolerance"
	ReasonStressLossNearTolerance          ReasonCode = "stress_loss_near_tolerance"
	ReasonCashAfterStressBelowMin          ReasonCode = "cash_after_stress_below_min"
	ReasonReserveCashUnconfirmed           ReasonCode = "reserve_cash_unconfirmed"
	ReasonHeavyPositionRequiresStressTest  ReasonCode = "heavy_position_requires_stress_test"
	ReasonPlannedBuyTooLargeUnderStress    ReasonCode = "planned_buy_too_large_under_stress"
	ReasonIndexPositionTooLargeUnderStress ReasonCode = "index_position_too_large_under_stress"
	ReasonFOMOWithStressRisk               ReasonCode = "fomo_with_stress_risk"
	ReasonStressPlanMissing                ReasonCode = "stress_plan_missing"
)

type StressGateInput struct {
	ActionType ActionType `yaml:"action_type"`

	IsHeavyPositionIntent *bool `yaml:"is_heavy_position_intent"`
	IsChasingHigh         *bool `yaml:"is_chasing_high"`
	FOMOTriggered         *bool `yaml:"fomo_triggered"`
	PanicTriggered        *bool `yaml:"panic_triggered"`

	TotalPositionPct    *float64 `yaml:"total_position_pct"`
	IndexPositionPct    *float64 `yaml:"index_position_pct"`
	PlannedBuyPct       *float64 `yaml:"planned_buy_pct"`
	PositionAfterBuyPct *float64 `yaml:"position_after_buy_pct"`

	InvestmentAssetsCNY  *float64 `yaml:"investment_assets_cny"`
	InvestableCashCNY    *float64 `yaml:"investable_cash_cny"`
	ReserveCashCNY       *float64 `yaml:"reserve_cash_cny"`
	ReserveCashConfirmed *bool    `yaml:"reserve_cash_confirmed"`
	CashAfterStressCNY   *float64 `yaml:"cash_after_stress_cny"`
	MinCashRequiredCNY   *float64 `yaml:"min_cash_required_cny"`

	StressDrawdownPct        *float64 `yaml:"stress_drawdown_pct"`
	StressLossCNY            *float64 `yaml:"stress_loss_cny"`
	StressLossPct            *float64 `yaml:"stress_loss_pct"`
	MaxLossToleranceCNY      *float64 `yaml:"max_loss_tolerance_cny"`
	MaxLossTolerancePct      *float64 `yaml:"max_loss_tolerance_pct"`
	ExtremeDrawdownPlanReady *bool    `yaml:"extreme_drawdown_plan_ready"`
}

type StressGateOutput struct {
	Decision            Decision        `yaml:"decision"`
	SuggestedAction     SuggestedAction `yaml:"suggested_action"`
	ReasonCodes         []ReasonCode    `yaml:"reason_codes"`
	HumanReviewRequired bool            `yaml:"human_review_required"`
}

type Scenario struct {
	ID       string           `yaml:"id"`
	Name     string           `yaml:"name"`
	Input    StressGateInput  `yaml:"input"`
	Expected StressGateOutput `yaml:"expected"`
	Note     string           `yaml:"note"`
}
