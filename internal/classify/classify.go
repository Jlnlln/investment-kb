package classify

import (
	"fmt"
	"strings"

	"investment-kb/internal/model"
)

// ClassifyDomain 基于 trigger_conditions + rule_content + actions 的关键词二次分类
// 优先级：ACCOUNT > CASH > VALUATION > STATE > RISK > REBALANCE > EXPOSURE
// 原则：如果一条规则同时涉及多个领域，优先选择"真正决定是否允许执行"的领域
//
// 返回值：
//   - 如果关键词证据足够强，返回映射后的领域
//   - 如果关键词证据不足，返回空字符串（表示保留 AI 原始分类）

// DomainKeywords 领域关键词表，按优先级排列
// 重要：VALUATION 的触发条件（估值/安全边际）是"主判断"，ACCOUNT 的约束（账户状态）是"执行约束"
// 原则：主触发条件决定领域，执行约束放在 not_applicable 里
var DomainKeywords = []struct {
	Domain   string
	Keywords []string
	Threshold int // 最少命中关键词数量（提高阈值，避免误映射）
}{
	{
		Domain: "ACCOUNT",
		Keywords: []string{
			"账户硬约束", "现金不足", "大额操作前", "账户适配",
			"满仓", "空仓", "仓位上限", "是否能买",
			"容错计划", "账户状态检查",
		},
		Threshold: 2, // 提高阈值，避免把"分批建仓"误判为 ACCOUNT
	},
	{
		Domain: "CASH",
		Keywords: []string{
			"现金管理", "流动性管理", "可用资金比例", "赎回纪律", "现金储备",
		},
		Threshold: 1,
	},
	{
		Domain: "VALUATION",
		Keywords: []string{
			"估值分位", "安全边际", "高概率区间", "低估区间", "高估区间", "PE分位",
			"PB分位", "估值赔率", "概率区间", "低估高估",
			"极限低点", "极端低估", "合理估值", "估值锚",
			"底仓", "建仓", // 这些词出现在 VALUATION 规则里，不应该映射到 ACCOUNT
		},
		Threshold: 1, // 降低阈值，VALUATION 关键词一旦出现就应该识别
	},
	{
		Domain: "STATE",
		Keywords: []string{
			"市场状态", "历史阶段", "大周期位置", "周期位置", "市场阶段划分",
			"S0", "S1", "S2", "S3", "S4", "涨幅极限参考", "历史锚点",
			"参考锚点", "市场位置判断", "周期判断", "牛市阶段", "熊市阶段",
		},
		Threshold: 2,
	},
	{
		Domain: "RISK",
		Keywords: []string{
			"下跌预案", "上涨预案", "踏空风险", "情绪控制", "误用风险",
			"不可预测", "纪律约束", "黑天鹅", "恐惧下跌", "追涨",
			"负反馈", "情绪控制规则", "风险预案", "风险边界",
			"多指标综合判断", "不能单独决策", "机械规则",
		},
		Threshold: 2,
	},
	{
		Domain: "REBALANCE",
		Keywords: []string{
			"再平衡", "超配低配", "偏离目标权重", "目标权重", "权重偏离",
			"资产配置调整", "组合配置", "指数配置比例",
		},
		Threshold: 1,
	},
	{
		Domain: "EXPOSURE",
		Keywords: []string{
			"成长类敞口", "成长敞口", "A股港股集中度", "集中度过高", "风格比例",
		},
		Threshold: 1,
	},
	// SCORE, TARGET, ETF 不设置关键词 — 这些领域由 AI 直接判断即可
}

// RuleNameDomainOverrides rule_name → domain_code 硬覆盖表
// 当规则名称完全匹配时，跳过关键词分类，直接返回指定领域
var RuleNameDomainOverrides = map[string]string{
	"高概率区间先建底仓":    "VALUATION",
	"结合多指标综合判断风险":  "RISK",
	"无论市场是否突破，都需执行减仓动作": "RISK",
}

// ClassifyDomain 对一条候选规则进行领域二次分类
// 注意：不包含 rule_name 硬覆盖逻辑，硬覆盖由 ClassifyDomainWithLog 处理
func ClassifyDomain(rule model.CandidateRule) string {
	// 将所有相关文本合并成一个搜索空间
	// 注意：只搜索 trigger_conditions + rule_content + actions
	// 不搜索 not_applicable（禁止条件不应该影响主领域判断）
	text := strings.Join(rule.TriggerConditions, " ") + " " +
		rule.RuleContent + " " +
		rule.RuleName + " " +
		strings.Join(rule.Actions, " ")

	// 按优先级逐个领域检查
	for _, domain := range DomainKeywords {
		hitCount := 0
		for _, kw := range domain.Keywords {
			if strings.Contains(text, kw) {
				hitCount++
			}
		}
		if hitCount >= domain.Threshold {
			return domain.Domain
		}
	}

	// 关键词证据不足，返回空字符串表示保留 AI 分类
	return ""
}

// ClassifyDomainWithLog 对一条候选规则进行领域二次分类，带日志输出
func ClassifyDomainWithLog(rule model.CandidateRule, aiDomain string) string {
	// 0. rule_name 硬覆盖
	if overrideDomain, ok := RuleNameDomainOverrides[rule.RuleName]; ok {
		if overrideDomain != aiDomain {
			fmt.Printf("   🎯 规则名覆盖：AI 建议 %s → 程序覆盖 %s（规则名精确匹配）\n", aiDomain, overrideDomain)
		} else {
			fmt.Printf("   ✅ 规则名确认：%s（与覆盖表一致）\n", aiDomain)
		}
		return overrideDomain
	}

	finalDomain := ClassifyDomain(rule)

	if finalDomain == "" {
		// 关键词证据不足，保留 AI 分类
		fmt.Printf("   📎 领域保留：%s（关键词证据不足，保留 AI 分类）\n", aiDomain)
		return aiDomain
	}

	if finalDomain != aiDomain {
		fmt.Printf("   🔄 领域映射：AI 建议 %s → 程序映射 %s（关键词证据充足）\n", aiDomain, finalDomain)
	} else {
		fmt.Printf("   ✅ 领域确认：%s（关键词与 AI 分类一致）\n", aiDomain)
	}

	return finalDomain
}
