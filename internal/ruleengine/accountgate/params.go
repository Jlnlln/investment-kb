package accountgate

type Params struct {
	TotalPositionMaxPct     float64
	SingleIndexMaxPct       float64
	BasePositionMinPct      float64
	BasePositionMaxPct      float64
	MinCashAfterBuyPct      float64
	StressDrawdownPct       float64
	MaxLossTolerancePct     float64
	ProfitCushionThickPct   float64
	FloatingLossPressurePct float64
}

func DefaultParams() Params {
	return Params{
		TotalPositionMaxPct:     80,
		SingleIndexMaxPct:       25,
		BasePositionMinPct:      5,
		BasePositionMaxPct:      15,
		MinCashAfterBuyPct:      20,
		StressDrawdownPct:       30,
		MaxLossTolerancePct:     15,
		ProfitCushionThickPct:   30,
		FloatingLossPressurePct: -10,
	}
}
