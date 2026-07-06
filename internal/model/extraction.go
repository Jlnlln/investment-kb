package model

// MaterialType 材料类型
type MaterialType string

const (
	MaterialTypeRuleCandidate     MaterialType = "rule_candidate"     // 规则型材料
	MaterialTypeMacroKnowledge    MaterialType = "macro_knowledge"    // 宏观理解型
	MaterialTypeMarketObservation MaterialType = "market_observation" // 市场状态观察型
	MaterialTypeArchiveOnly       MaterialType = "archive_only"       // 仅存档
)

// ExtractionResult 是 AI 返回的结构化提取结果
type ExtractionResult struct {
	Title                   string             `json:"title"`
	Source                  string             `json:"source"`
	MaterialType            MaterialType       `json:"material_type"`             // 新增：材料类型
	GenerateQA              bool               `json:"generate_qa"`               // 新增：是否生成 QA
	GenerateCandidateRules  bool               `json:"generate_candidate_rules"`  // 新增：是否生成 CR
	GenerateValidationCards bool               `json:"generate_validation_cards"` // 新增：是否生成验证卡
	GenerateKnowledgeCard   bool               `json:"generate_knowledge_card"`   // 新增：是否生成 KNOW 卡
	GenerateObservationCard bool               `json:"generate_observation_card"` // 新增：是否生成 OBS 卡
	NoRuleReason            string             `json:"no_rule_reason"`            // 新增：不生成规则的原因
	ReusableUnderstanding   []string           `json:"reusable_understanding"`    // 可复用理解（macro_knowledge 用）
	DomainCode              string             `json:"domain_code"`
	TopicCode               string             `json:"topic_code"`
	Tags                    []string           `json:"tags"`
	Summary                 string             `json:"summary"`
	CoreConclusion          string             `json:"core_conclusion"`
	CoreLogic               []LogicBlock       `json:"core_logic"`
	AccountProfiles         AccountProfiles    `json:"account_profiles"`
	BehaviorCorrection      BehaviorCorrection `json:"behavior_correction"`
	ApplicableScenarios     []string           `json:"applicable_scenarios"`
	RiskBoundaries          []string           `json:"risk_boundaries"`
	ExtractableRules        []RuleSummary      `json:"extractable_rules"`
	PotentialRules          []PotentialRule    `json:"potential_rules"`
	ShouldGenerateCase      bool               `json:"should_generate_case"`
	Case                    *MarketCase        `json:"case"`
	CaseInsufficientReason  string             `json:"case_insufficient_reason"`
	CandidateRules          []CandidateRule    `json:"candidate_rules"`
	MyUnderstanding         string             `json:"my_understanding"`
	RawHash                 string             `json:"-"` // 原文 sha256，程序计算，不由 AI 生成
	SourceMeta              SourceMeta         `json:"-"` // 来源追溯信息，程序计算，不由 AI 生成
}

// SourceMeta 记录所有输出对象都需要携带的来源追溯信息。
type SourceMeta struct {
	SourceFile   string
	RawHash      string
	CleanedHash  string
	RawID        string
	MaterialType MaterialType
}

// LogicBlock 是核心逻辑的一个逻辑块
type LogicBlock struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AccountProfiles struct {
	MentionedStates     []AccountStateNote `json:"mentioned_states"`
	RecommendationDiffs []string           `json:"recommendation_diffs"`
	Reason              string             `json:"reason"`
}

type AccountStateNote struct {
	State string `json:"state"`
	Note  string `json:"note"`
}

type BehaviorCorrection struct {
	WrongBehaviors         []string `json:"wrong_behaviors"`
	BehaviorConstraints    []string `json:"behavior_constraints"`
	CounterintuitiveAdvice string   `json:"counterintuitive_advice"`
}

// RuleSummary 是可提炼规则的摘要信息
type RuleSummary struct {
	RuleType string `json:"rule_type"`
	RuleName string `json:"rule_name"`
	Summary  string `json:"summary"`
}
type PotentialRule struct {
	RuleDraft         string   `json:"rule_draft"`
	DomainCode        string   `json:"domain_code"`
	OriginalEvidence  string   `json:"original_evidence"`
	ApplicableObjects []string `json:"applicable_objects"`
	PreventedError    string   `json:"prevented_error"`
	ShouldGenerateCR  string   `json:"should_generate_cr"`
	NoGenerateReason  string   `json:"no_generate_reason"`
}
