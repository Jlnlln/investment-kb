package model

// ExtractionResult 是 AI 返回的结构化提取结果
type ExtractionResult struct {
	Title                  string          `json:"title"`
	Source                 string          `json:"source"`
	DomainCode             string          `json:"domain_code"`
	TopicCode              string          `json:"topic_code"`
	Tags                   []string        `json:"tags"`
	Summary                string          `json:"summary"`
	CoreConclusion         string          `json:"core_conclusion"`
	CoreLogic              []LogicBlock    `json:"core_logic"`
	ApplicableScenarios    []string        `json:"applicable_scenarios"`
	RiskBoundaries         []string        `json:"risk_boundaries"`
	ExtractableRules       []RuleSummary   `json:"extractable_rules"`
	ShouldGenerateCase     bool            `json:"should_generate_case"`
	Case                   *MarketCase     `json:"case"`
	CaseInsufficientReason string          `json:"case_insufficient_reason"`
	CandidateRules         []CandidateRule `json:"candidate_rules"`
	MyUnderstanding        string          `json:"my_understanding"`
	RawHash                string          `json:"-"` // 原文 sha256，程序计算，不由 AI 生成
}

// LogicBlock 是核心逻辑的一个逻辑块
type LogicBlock struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// RuleSummary 是可提炼规则的摘要信息
type RuleSummary struct {
	RuleType string `json:"rule_type"`
	RuleName string `json:"rule_name"`
	Summary  string `json:"summary"`
}