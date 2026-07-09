package accountgate

type ActionType string

const (
	ActionBuy       ActionType = "buy"
	ActionAdd       ActionType = "add"
	ActionHold      ActionType = "hold"
	ActionReduce    ActionType = "reduce"
	ActionRebalance ActionType = "rebalance"
	ActionFollow    ActionType = "follow"
)

type PositionState string

const (
	PositionEmpty  PositionState = "empty_position"
	PositionLight  PositionState = "light_position"
	PositionNormal PositionState = "normal_position"
	PositionHigh   PositionState = "high_position"
	PositionFull   PositionState = "full_position"
)

type AccountFlag string

const (
	FlagProfitCushionThick   AccountFlag = "profit_cushion_thick"
	FlagFloatingLossPressure AccountFlag = "floating_loss_pressure"
	FlagCashInsufficient     AccountFlag = "cash_insufficient"
	FlagReserveUnconfirmed   AccountFlag = "reserve_unconfirmed"
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
	SuggestAllowSmallBuy      SuggestedAction = "allow_small_buy"
	SuggestAllowBatchAdd      SuggestedAction = "allow_batch_add"
	SuggestHoldNoAdd          SuggestedAction = "hold_no_add"
	SuggestBlockBuy           SuggestedAction = "block_buy"
	SuggestBlockAdd           SuggestedAction = "block_add"
	SuggestBlockFollow        SuggestedAction = "block_follow"
	SuggestReviewReduce       SuggestedAction = "review_reduce"
	SuggestReduceToLimit      SuggestedAction = "reduce_to_limit"
	SuggestReviewPlan         SuggestedAction = "review_plan"
	SuggestReviewCash         SuggestedAction = "review_cash"
	SuggestReviewRiskExposure SuggestedAction = "review_risk_exposure"
)

type ReasonCode string

const (
	ReasonReserveCashUnconfirmed        ReasonCode = "reserve_cash_unconfirmed"
	ReasonTotalPositionExceedsMax       ReasonCode = "total_position_exceeds_max"
	ReasonSingleIndexExceedsMax         ReasonCode = "single_index_exceeds_max"
	ReasonCashAfterBuyBelowMin          ReasonCode = "cash_after_buy_below_min"
	ReasonPlanMissing                   ReasonCode = "plan_missing"
	ReasonStressLossExceedsTolerance    ReasonCode = "stress_loss_exceeds_tolerance"
	ReasonInfluencedByOthersPosition    ReasonCode = "influenced_by_others_position"
	ReasonFOMOWithHighPosition          ReasonCode = "fomo_with_high_position"
	ReasonPanicReduceWithoutExitSignal  ReasonCode = "panic_reduce_without_exit_signal"
	ReasonProfitCushionSubjectiveReduce ReasonCode = "profit_cushion_subjective_reduce"
	ReasonAccountStateAllowsSmallBuy    ReasonCode = "account_state_allows_small_buy"
	ReasonAccountStateAllowsBatchAdd    ReasonCode = "account_state_allows_batch_add"
)

type AccountGateInput struct {
	ActionType           ActionType    `yaml:"action_type"`
	PrimaryPositionState PositionState `yaml:"primary_position_state"`
	AccountFlags         []AccountFlag `yaml:"account_flags"`

	InvestmentAssetsCNY  float64 `yaml:"investment_assets_cny"`
	InvestableCashCNY    float64 `yaml:"investable_cash_cny"`
	ReserveCashConfirmed bool    `yaml:"reserve_cash_confirmed"`

	TotalPositionPct float64 `yaml:"total_position_pct"`
	IndexPositionPct float64 `yaml:"index_position_pct"`
	PlannedBuyPct    float64 `yaml:"planned_buy_pct"`
	CashAfterBuyPct  float64 `yaml:"cash_after_buy_pct"`

	CostBasis        float64 `yaml:"cost_basis"`
	FloatingPNLPct   float64 `yaml:"floating_pnl_pct"`
	ProfitCushionPct float64 `yaml:"profit_cushion_pct"`

	MaxDrawdownTolerancePct float64 `yaml:"max_drawdown_tolerance_pct"`
	MaxLossToleranceCNY     float64 `yaml:"max_loss_tolerance_cny"`
	StressDrawdownPct       float64 `yaml:"stress_drawdown_pct"`
	LossIfStressCNY         float64 `yaml:"loss_if_stress_cny"`

	UpsidePlanReady          bool `yaml:"upside_plan_ready"`
	DownsidePlanReady        bool `yaml:"downside_plan_ready"`
	ExitPlanReady            bool `yaml:"exit_plan_ready"`
	ExtremeDrawdownPlanReady bool `yaml:"extreme_drawdown_plan_ready"`

	InfluencedByOthersPosition bool `yaml:"influenced_by_others_position"`
	FOMOTriggered              bool `yaml:"fomo_triggered"`
	PanicTriggered             bool `yaml:"panic_triggered"`
}

type AccountGateOutput struct {
	Decision            Decision        `yaml:"decision"`
	SuggestedAction     SuggestedAction `yaml:"suggested_action"`
	ReasonCodes         []ReasonCode    `yaml:"reason_codes"`
	HumanReviewRequired bool            `yaml:"human_review_required"`
}

type Scenario struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name"`
	Input    AccountGateInput  `yaml:"input"`
	Expected AccountGateOutput `yaml:"expected"`
	Note     string            `yaml:"note"`
}
