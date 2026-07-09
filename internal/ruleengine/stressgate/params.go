package stressgate

type Params struct {
	DefaultStressDrawdownPct      float64
	MaxLossTolerancePct           float64
	NearToleranceThresholdPct     float64
	HeavyPositionThresholdPct     float64
	LargePlannedBuyThresholdPct   float64
	HighIndexPositionThresholdPct float64
}

func DefaultParams() Params {
	return Params{
		DefaultStressDrawdownPct:      30,
		MaxLossTolerancePct:           15,
		NearToleranceThresholdPct:     80,
		HeavyPositionThresholdPct:     70,
		LargePlannedBuyThresholdPct:   10,
		HighIndexPositionThresholdPct: 25,
	}
}
