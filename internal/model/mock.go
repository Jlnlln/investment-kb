package model

// MockExtractionResult 返回一个模拟的 ExtractionResult，用于测试和开发
// 基于技术文档第 16 节的 JSON 示例
func MockExtractionResult() *ExtractionResult {
	return &ExtractionResult{
		Title:      "安全边际与错失买入机会如何平衡",
		Source:     "陈老师问答",
		DomainCode: "POS",
		TopicCode:  "SAFETY",
		Tags:       []string{"仓位管理", "安全边际", "踏空风险", "账户状态", "容错计划"},
		Summary:    "这段问答讨论的是投资者如何在追求安全边际和避免错失买入机会之间取得平衡。",
		CoreConclusion: "安全边际不是越低越好，而是在高概率区域内结合自己的账户状态分批参与。",
		CoreLogic: []LogicBlock{
			{
				Title:   "买点不能脱离账户状态",
				Content: "同样是 3800 点，对低成本持仓者和空仓者意义不同。低成本持仓者有浮盈安全垫，空仓者没有，因此不能照搬满仓计划。",
			},
			{
				Title:   "安全边际过高会转化为踏空风险",
				Content: "如果安全边际设置得过于极致，市场没有跌到理想点位就反弹，投资者会完全踏空。",
			},
		},
		ApplicableScenarios: []string{
			"宽基指数出现明显回撤",
			"估值进入合理区或低估区",
			"投资者担心买早又害怕踏空",
			"空仓者迟迟不敢建仓",
		},
		RiskBoundaries: []string{
			"指数估值仍处于高估区",
			"账户已经满仓",
			"没有后续加仓资金",
			"指数长期逻辑发生变化",
		},
		ExtractableRules: []RuleSummary{
			{
				RuleType: "买入规则",
				RuleName: "高概率区间先建底仓",
				Summary:  "当宽基指数进入历史高概率买入区间后，应先建立第一笔仓位。",
			},
		},
		ShouldGenerateCase:     false,
		Case:                   nil,
		CaseInsufficientReason: "原文提到标普回撤，但缺少具体时间、点位、估值状态和后续走势。",
		CandidateRules: []CandidateRule{
			{
				RuleType:              "买入规则",
				RuleName:              "高概率区间先建底仓",
				DomainCode:            "BUY",
				TopicCode:             "SAFETY",
				SuggestedFormalRuleID: "BUY-001",
				RuleContent:           "当宽基指数进入历史高概率买入区间后，应先建立第一笔仓位，不应因为等待极端低点而完全空仓。",
				TriggerConditions: []string{
					"宽基指数出现明显回撤",
					"估值进入合理区或低估区",
					"市场情绪偏悲观",
					"指数长期配置逻辑未被破坏",
					"当前账户不是满仓",
					"仍有后续加仓资金",
				},
				Actions: []string{
					"建立第一笔仓位",
					"第一笔仓位建议为目标仓位的 20%-30%",
					"如果后续继续下跌，再按照预设区间分批加仓",
					"如果后续直接反弹，至少已有底仓参与",
					"买入前必须写出上涨预案和下跌预案",
				},
				NotApplicable: []string{
					"指数估值仍处于高估区",
					"当前账户已经满仓",
					"没有后续加仓资金",
					"指数长期逻辑发生变化",
					"只是因为看到上涨而产生踏空焦虑",
				},
				RiskBoundary: "这条规则最大的风险是误把普通下跌当成高概率买点。必须结合估值、回撤、情绪、账户状态共同判断。",
				QuestionsToConfirm: []string{
					"明显回撤的标准是多少？",
					"估值合理区或低估区使用什么数据源判断？",
					"第一笔仓位是否固定为目标仓位的 20%-30%？",
					"是否适用于所有宽基指数？",
				},
				Recommendation: "建议修改后采纳。正式采用前需要补充高概率买入区间定义。",
			},
			{
				RuleType:              "仓位规则",
				RuleName:              "账户状态决定仓位力度",
				DomainCode:            "POS",
				TopicCode:             "ACCOUNT",
				SuggestedFormalRuleID: "POS-002",
				RuleContent:           "买入决策必须结合账户当前状态。低成本持仓者和空仓者的安全边际策略完全不同。",
				TriggerConditions: []string{
					"评估买入时机",
					"判断账户仓位状态",
				},
				Actions: []string{
					"识别当前是满仓、低持仓、还是空仓",
					"根据账户状态调整买入节奏和力度",
					"空仓者可以使用更激进的安全边际策略",
					"满仓者应更多关注风险控制而非补仓",
				},
				NotApplicable: []string{
					"尚未建立账户状态认知",
					"无法准确判断持仓成本",
				},
				RiskBoundary: "低估自己持仓成本或现金储备会导致错误的账户状态判断。",
				QuestionsToConfirm: []string{
					"如何定义低成本持仓的标准？",
					"账户状态分类是否需要更细致？",
				},
				Recommendation: "建议采纳。这是投资决策的基础前提。",
			},
			{
				RuleType:              "风控规则",
				RuleName:              "买入前必须有上涨和下跌预案",
				DomainCode:            "RISK",
				TopicCode:             "PLAN",
				SuggestedFormalRuleID: "RISK-003",
				RuleContent:           "任何买入操作前，必须写出上涨预案和下跌预案。没有预案的买入是赌博。",
				TriggerConditions: []string{
					"计划执行买入操作",
				},
				Actions: []string{
					"明确买入后的上涨应对方案",
					"明确买入后的下跌应对方案",
					"将预案写入交易笔记",
				},
				NotApplicable: []string{
					"没有交易笔记习惯",
				},
				RiskBoundary: "预案过于笼统无法执行，等同于没有预案。",
				QuestionsToConfirm: []string{
					"预案的具体程度需要多详细？",
					"预案是否需要包含具体的点位？",
				},
				Recommendation: "建议采纳。这是纪律性投资的核心。",
			},
		},
		MyUnderstanding: "这段问答最重要的启发是，投资决策不能只看市场点位，还要看账户状态。完整问题不是当前点位能不能买，而是在我的账户状态下当前点位能买多少。",
	}
}