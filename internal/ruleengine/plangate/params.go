package plangate

type Params struct {
	MaxLossPctRequired                          bool
	RiskExposureRequiredForChasingHigh          bool
	ExtremeDrawdownPlanRequiredForHeavyPosition bool
	PlanLockRequiredBeforeBuy                   bool
}

func DefaultParams() Params {
	return Params{
		MaxLossPctRequired:                          true,
		RiskExposureRequiredForChasingHigh:          true,
		ExtremeDrawdownPlanRequiredForHeavyPosition: true,
		PlanLockRequiredBeforeBuy:                   true,
	}
}
