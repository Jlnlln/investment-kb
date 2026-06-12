package model

// MarketCase 是市场案例
type MarketCase struct {
	CaseName             string   `json:"case_name"`
	DomainCode           string   `json:"domain_code"`
	TopicCode            string   `json:"topic_code"`
	TimeBackground       string   `json:"time_background"`
	Assets               []string `json:"assets"`
	MarketStatus         string   `json:"market_status"`
	KeyDecisionQuestion   string   `json:"key_decision_question"`
	AlternativeSolutions []string `json:"alternative_solutions"`
	FinalInsight         string   `json:"final_insight"`
	ExtractedRules       []string `json:"extracted_rules"`
}