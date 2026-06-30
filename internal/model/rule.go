package model

// CandidateRule 是候选规则
type CandidateRule struct {
	RuleType              string   `json:"rule_type"`
	RuleName              string   `json:"rule_name"`
	DomainCode            string   `json:"domain_code"`
	TopicCode             string   `json:"topic_code"`
	SuggestedFormalRuleID string   `json:"suggested_formal_rule_id"`
	RuleContent           string   `json:"rule_content"`
	TriggerConditions     []string `json:"trigger_conditions"`
	Actions               []string `json:"actions"`
	NotApplicable         []string `json:"not_applicable"`
	RiskBoundary          string   `json:"risk_boundary"`
	QuestionsToConfirm    []string `json:"questions_to_confirm"`
	Recommendation        string   `json:"recommendation"`
	ApplicableObjects     []string `json:"applicable_objects"` // 适用对象
}