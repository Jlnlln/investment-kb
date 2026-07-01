package dedup

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// 测试用例设计矩阵
// ============================================================
// T1: 完全重复 — 相同 rule_name + 相同 triggers/actions → Level="完全重复"
// T2: 高度相似 — 同 domain-topic，名称重叠≥0.7 且触发条件重叠≥0.4 → Level="高度相似"
// T3: 疑似相似 — 同 domain-topic，名称重叠≥0.5 但触发条件重叠<0.4 → Level="疑似相似"
// T4: 可能相似 — 跨领域，触发条件重叠≥0.5 → Level="可能相似"
// T5: 无匹配 — 同领域但重叠不够 → 空列表
// T6: 空已有列表 → 空列表
// T7: 边界 — 空 triggers/actions/ruleName
// ============================================================

func TestCheckSimilarRules_ExactDuplicate(t *testing.T) {
	// T1: 完全相同的规则
	existing := []RuleFingerprint{{
		CRID:       "CR-VALUATION-20260701-001",
		ShortCode:  "VALUATION-SAFETY",
		DomainCode: "VALUATION",
		TopicCode:  "SAFETY",
		RuleName:   "高概率区间先建底仓",
		Triggers:   []string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		Actions:    []string{"在预设的高概率区间建立第一笔底仓", "保留后续加仓资金"},
		ExactHash:  ComputeExactHash("VALUATION", "高概率区间先建底仓",
			[]string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
			[]string{"在预设的高概率区间建立第一笔底仓", "保留后续加仓资金"}),
	}}

	result := CheckSimilarRules(
		"VALUATION", "SAFETY", "高概率区间先建底仓",
		[]string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		[]string{"在预设的高概率区间建立第一笔底仓", "保留后续加仓资金"},
		existing,
	)

	if len(result) != 1 {
		t.Fatalf("T1 完全重复: 期望 1 条结果，实际 %d 条", len(result))
	}
	if result[0].Level != "完全重复" {
		t.Errorf("T1 完全重复: 期望 Level='完全重复'，实际='%s'", result[0].Level)
	}
	if result[0].CRID != "CR-VALUATION-20260701-001" {
		t.Errorf("T1 完全重复: 期望 CRID='CR-VALUATION-20260701-001'，实际='%s'", result[0].CRID)
	}
}

func TestCheckSimilarRules_HighSimilarity(t *testing.T) {
	// T2: 同 domain-topic，名称和触发条件高度重叠
	existing := []RuleFingerprint{{
		CRID:       "CR-VALUATION-20260701-001",
		ShortCode:  "VALUATION-SAFETY",
		DomainCode: "VALUATION",
		TopicCode:  "SAFETY",
		RuleName:   "高概率区间先建底仓",
		Triggers:   []string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		Actions:    []string{"在预设的高概率区间建立第一笔底仓"},
		ExactHash:  "different-hash-001",
	}}

	// 新规则：名称中"仓位"换"底仓"、"低估值"换"高概率"等，保留大量共同片段
	result := CheckSimilarRules(
		"VALUATION", "SAFETY", "高概率区间先建仓位",
		[]string{"宽基指数出现回调", "估值处于低估区间", "市场情绪悲观"},
		[]string{"在高概率区间建仓", "保留后续加仓资金"},
		existing,
	)

	if len(result) != 1 {
		t.Fatalf("T2 高度相似: 期望 1 条结果，实际 %d 条", len(result))
	}
	if result[0].Level != "高度相似" {
		t.Errorf("T2 高度相似: 期望 Level='高度相似'，实际='%s'", result[0].Level)
	}
	if result[0].CRID != "CR-VALUATION-20260701-001" {
		t.Errorf("T2 高度相似: 期望匹配已有 CR-001")
	}
}

func TestCheckSimilarRules_PossibleSimilar_NameOnly(t *testing.T) {
	// T3: 同 domain-topic，名称重叠≥0.5 但触发条件重叠<0.4
	existing := []RuleFingerprint{{
		CRID:       "CR-STATE-20260701-001",
		ShortCode:  "STATE-REVIEW",
		DomainCode: "STATE",
		TopicCode:  "REVIEW",
		RuleName:   "建立历史极端案例作为风险边界锚点",
		Triggers:   []string{"市场出现重大外部事件", "进入历史罕见状态"},
		Actions:    []string{"记录历史极端数据"},
		ExactHash:  "different-hash-002",
	}}

	// 相同 domain-topic，名称有"历史极端案例"重叠，但触发条件完全不同
	result := CheckSimilarRules(
		"STATE", "REVIEW", "历史极端案例参考与风险边界",
		[]string{"出现新低估值信号", "交易量突然放大"},
		[]string{"参考历史案例调整仓位"},
		existing,
	)

	if len(result) != 1 {
		t.Fatalf("T3 疑似相似: 期望 1 条结果，实际 %d 条", len(result))
	}
	// 名称相似但触发条件重叠低 → 疑似相似（不是高度相似）
	if result[0].Level != "疑似相似" {
		t.Errorf("T3 疑似相似: 期望 Level='疑似相似'，实际='%s'", result[0].Level)
	}
}

func TestCheckSimilarRules_CrossDomain(t *testing.T) {
	// T4: 跨领域，触发条件高度重叠
	existing := []RuleFingerprint{{
		CRID:       "CR-VALUATION-20260701-001",
		ShortCode:  "VALUATION-SAFETY",
		DomainCode: "VALUATION",
		TopicCode:  "SAFETY",
		RuleName:   "高概率区间先建底仓",
		Triggers:   []string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		Actions:    []string{"在预设的高概率区间建立底仓"},
		ExactHash:  "different-hash-003",
	}}

	// 跨到 ACCOUNT 领域，但触发条件与已有 VALUATION 规则高度重叠
	result := CheckSimilarRules(
		"ACCOUNT", "POS", "账户状态决定仓位力度",
		[]string{"宽基指数出现明显回撤", "估值进入合理区", "市场情绪偏悲观"},
		[]string{"根据账户状态调整仓位"},
		existing,
	)

	if len(result) != 1 {
		t.Fatalf("T4 跨领域: 期望 1 条结果，实际 %d 条", len(result))
	}
	if result[0].Level != "可能相似" {
		t.Errorf("T4 跨领域: 期望 Level='可能相似'，实际='%s'", result[0].Level)
	}
	if result[0].CRID != "CR-VALUATION-20260701-001" {
		t.Errorf("T4 跨领域: 期望匹配已有 VALUATION 规则")
	}
}

func TestCheckSimilarRules_NoMatch(t *testing.T) {
	// T5: 完全不同的规则，不应匹配
	existing := []RuleFingerprint{
		{
			CRID:       "CR-VALUATION-20260701-001",
			ShortCode:  "VALUATION-SAFETY",
			DomainCode: "VALUATION",
			TopicCode:  "SAFETY",
			RuleName:   "高概率区间先建底仓",
			Triggers:   []string{"宽基指数出现明显回撤", "估值进入合理区或低估区"},
			Actions:    []string{"建立底仓"},
			ExactHash:  "hash-a",
		},
		{
			CRID:       "CR-RISK-20260701-001",
			ShortCode:  "RISK-DISC",
			DomainCode: "RISK",
			TopicCode:  "DISC",
			RuleName:   "低确定性判断必须分仓应对",
			Triggers:   []string{"判断确定性低于60%", "市场处于关键突破位"},
			Actions:    []string{"分仓操作"},
			ExactHash:  "hash-b",
		},
	}

	result := CheckSimilarRules(
		"ACCOUNT", "POS", "账户状态决定仓位力度",
		[]string{"账户处于空仓状态", "有新增资金到账"},
		[]string{"根据账户状态制定仓位计划"},
		existing,
	)

	if len(result) != 0 {
		t.Errorf("T5 无匹配: 期望 0 条结果，实际 %d 条: %+v", len(result), result)
	}
}

func TestCheckSimilarRules_EmptyExisting(t *testing.T) {
	// T6: 没有已有规则
	result := CheckSimilarRules(
		"VALUATION", "SAFETY", "高概率区间先建底仓",
		[]string{"宽基指数出现明显回撤"},
		[]string{"建立底仓"},
		nil,
	)

	if len(result) != 0 {
		t.Errorf("T6 空列表: 期望 0 条结果，实际 %d 条", len(result))
	}
}

func TestCheckSimilarRules_EdgeCases(t *testing.T) {
	// T7: 边界 — 空 triggers/actions/name
	existing := []RuleFingerprint{{
		CRID:       "CR-RISK-20260701-001",
		ShortCode:  "RISK-DISC",
		DomainCode: "RISK",
		TopicCode:  "DISC",
		RuleName:   "低确定性判断必须分仓应对",
		Triggers:   []string{"判断确定性低于60%"},
		Actions:    []string{"分仓操作"},
		ExactHash:  "hash-c",
	}}

	// 空 name：同一 domain-topic 且触发条件完全相同，算法应检测到相似
	result := CheckSimilarRules(
		"RISK", "DISC", "",
		[]string{"判断确定性低于60%"},
		[]string{"分仓操作"},
		existing,
	)
	// 空 name + 同 domain-topic + trigger 重叠 1.0 → 应检测到相似
	if len(result) != 1 {
		t.Fatalf("T7 空 name+同触发: 期望 1 条结果（触发条件相同），实际 %d 条", len(result))
	}
	if result[0].Level != "疑似相似" {
		t.Errorf("T7 空 name+同触发: 期望 Level='疑似相似'，实际='%s'", result[0].Level)
	}

	// 空 triggers：名称完全相同且在同一个 domain-topic 下 → nameOverlap=1.0
	result = CheckSimilarRules(
		"RISK", "DISC", "低确定性判断必须分仓应对",
		[]string{},
		[]string{"分仓操作"},
		existing,
	)
	// 同一个 domain-topic + 名称完全相同 nameOverlap=1.0 ≥ 0.5
	// triggers 空 → triggerOverlap=0
	// name≥0.7(false, we need 0.7 for 高度相似) → "疑似相似" (fallback)
	if len(result) != 1 {
		t.Fatalf("T7 空 triggers+同名: 期望 1 条结果，实际 %d 条", len(result))
	}
	if result[0].Level != "疑似相似" {
		t.Errorf("T7 空 triggers+同名: 期望 Level='疑似相似'，实际='%s'", result[0].Level)
	}

	// 完全不同：空 name + 不同触发条件 → 应无匹配
	result = CheckSimilarRules(
		"RISK", "DISC", "",
		[]string{"完全不同"},
		[]string{"完全不同的动作"},
		existing,
	)
	if len(result) != 0 {
		t.Errorf("T7 空 name+不同触发: 期望 0 条结果，实际 %d 条", len(result))
	}
}

// ============================================================
// ParseExistingCRs 测试
// ============================================================

func TestParseExistingCRs_ValidFile(t *testing.T) {
	// 构造一个迷你候选规则库文件
	content := `# 候选规则库

---

# CR-VALUATION-20260701-001｜VALUATION-SAFETY｜高概率区间先建底仓

状态：候选

## 1. 规则内容

测试内容

---

## 2. 触发条件

- 宽基指数出现明显回撤
- 估值进入合理区或低估区
- 市场情绪偏悲观

---

## 3. 执行动作

- 在预设的高概率区间建立第一笔底仓
- 保留后续加仓资金

---

## 4. 不适用场景

- 个股不适用

---

# CR-RISK-20260701-001｜RISK-DISC｜低确定性判断必须分仓应对

状态：候选

## 2. 触发条件

- 判断确定性低于60%
- 市场处于关键突破位

---

## 3. 执行动作

- 将仓位分为多份
- 针对不同路径预设退出策略

---

`

	tmpFile := filepath.Join(t.TempDir(), "test_cr_library.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	fps, err := ParseExistingCRs(tmpFile)
	if err != nil {
		t.Fatalf("ParseExistingCRs 失败: %v", err)
	}

	if len(fps) != 2 {
		t.Fatalf("期望解析出 2 条 CR，实际 %d 条", len(fps))
	}

	// 验证第一条
	if fps[0].CRID != "CR-VALUATION-20260701-001" {
		t.Errorf("第一条 CRID 错误: %s", fps[0].CRID)
	}
	if fps[0].DomainCode != "VALUATION" {
		t.Errorf("第一条 DomainCode 错误: %s", fps[0].DomainCode)
	}
	if fps[0].TopicCode != "SAFETY" {
		t.Errorf("第一条 TopicCode 错误: %s", fps[0].TopicCode)
	}
	if fps[0].RuleName != "高概率区间先建底仓" {
		t.Errorf("第一条 RuleName 错误: %s", fps[0].RuleName)
	}
	if len(fps[0].Triggers) != 3 {
		t.Errorf("第一条 Triggers 数量错误: %d (期望 3)", len(fps[0].Triggers))
	}
	if len(fps[0].Actions) != 2 {
		t.Errorf("第一条 Actions 数量错误: %d (期望 2)", len(fps[0].Actions))
	}
	if fps[0].ExactHash == "" {
		t.Error("第一条 ExactHash 不应为空")
	}
	if fps[0].SemanticKey == "" {
		t.Error("第一条 SemanticKey 不应为空")
	}

	// 验证第二条
	if fps[1].CRID != "CR-RISK-20260701-001" {
		t.Errorf("第二条 CRID 错误: %s", fps[1].CRID)
	}
	if fps[1].DomainCode != "RISK" {
		t.Errorf("第二条 DomainCode 错误: %s", fps[1].DomainCode)
	}
	if fps[1].RuleName != "低确定性判断必须分仓应对" {
		t.Errorf("第二条 RuleName 错误: %s", fps[1].RuleName)
	}
	if len(fps[1].Triggers) != 2 {
		t.Errorf("第二条 Triggers 数量错误: %d (期望 2)", len(fps[1].Triggers))
	}
	if len(fps[1].Actions) != 2 {
		t.Errorf("第二条 Actions 数量错误: %d (期望 2)", len(fps[1].Actions))
	}
}

func TestParseExistingCRs_FileNotExist(t *testing.T) {
	fps, err := ParseExistingCRs("/nonexistent/path/cr_library.md")
	if err != nil {
		t.Fatalf("文件不存在应返回 nil,nil，实际: %v", err)
	}
	if fps != nil {
		t.Errorf("文件不存在应返回 nil，实际 %d 条", len(fps))
	}
}

func TestParseExistingCRs_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.md")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("写入空文件失败: %v", err)
	}

	fps, err := ParseExistingCRs(tmpFile)
	if err != nil {
		t.Fatalf("空文件不应报错: %v", err)
	}
	if fps != nil {
		t.Errorf("空文件应返回 nil，实际 %d 条", len(fps))
	}
}

func TestParseExistingCRs_OnlyWhitespace(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "whitespace.md")
	if err := os.WriteFile(tmpFile, []byte("\n  \n\t\n"), 0644); err != nil {
		t.Fatalf("写入空白文件失败: %v", err)
	}

	fps, err := ParseExistingCRs(tmpFile)
	if err != nil {
		t.Fatalf("纯空白文件不应报错: %v", err)
	}
	if fps != nil {
		t.Errorf("纯空白文件应返回 nil，实际 %d 条", len(fps))
	}
}

func TestParseExistingCRs_NoValidEntries(t *testing.T) {
	// 文件存在但没有合法的 CR 标题行
	content := `# 候选规则库

这只是说明文字，没有合法的 CR。

---

结尾。
`

	tmpFile := filepath.Join(t.TempDir(), "no_cr.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	fps, err := ParseExistingCRs(tmpFile)
	if err != nil {
		t.Fatalf("无合法 CR 不应报错: %v", err)
	}
	if len(fps) != 0 {
		t.Errorf("无合法 CR 应返回空列表，实际 %d 条", len(fps))
	}
}

// ============================================================
// 辅助函数测试
// ============================================================

func TestComputeKeywordOverlap(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
		reason   string
	}{
		{"完全一致", "高概率区间先建底仓", "高概率区间先建底仓", 1.0, "相同文本应 100% 重叠"},
		{"部分相似", "高概率区间先建底仓", "高概率区间先建仓位", 0.733, "仅末字不同（底仓→仓位），重叠度应 > 0.7"},
		{"完全不相关", "高概率区间先建底仓", "创业板涨幅极限参考", 0.0, "完全不同领域，重叠度应为 0"},
		{"有交集但低", "高概率区间先建底仓", "低估值区间不要等极限低点", 0.067, "仅\"区间\" 2 字重叠，浮点取近似"},
		{"空文本", "", "任何内容", 0.0, "空文本重叠为 0"},
		{"双方空", "", "", 0.0, "双方空"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeKeywordOverlap(tt.text1, tt.text2)
			// 浮点比较：允许 0.005 误差
			if result < tt.expected-0.005 || result > tt.expected+0.005 {
				t.Errorf("%s: 期望 %.3f，实际 %.3f (%s)", tt.name, tt.expected, result, tt.reason)
			}
		})
	}
}

func TestComputeListOverlap(t *testing.T) {
	// 高度重叠
	result := computeListOverlap(
		[]string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		[]string{"宽基指数出现回调", "估值处于低估区间", "市场情绪悲观"},
	)
	if result < 0.4 {
		t.Errorf("高度重叠列表: 期望 ≥ 0.4，实际 %.3f", result)
	}

	// 完全不重叠
	result = computeListOverlap(
		[]string{"宽基指数出现明显回撤", "估值进入合理区"},
		[]string{"判断确定性低于60%", "市场处于关键突破位"},
	)
	if result >= 0.1 {
		t.Errorf("不重叠列表: 期望 < 0.1，实际 %.3f", result)
	}

	// 空列表
	result = computeListOverlap([]string{}, []string{"任何内容"})
	if result != 0 {
		t.Errorf("空列表: 期望 0，实际 %.3f", result)
	}
}

func TestComputeExactHash_Deterministic(t *testing.T) {
	// 相同输入应产生相同哈希
	h1 := ComputeExactHash("VALUATION", "高概率区间先建底仓",
		[]string{"宽基指数回撤", "估值合理"},
		[]string{"建立底仓", "保留资金"})
	h2 := ComputeExactHash("VALUATION", "高概率区间先建底仓",
		[]string{"宽基指数回撤", "估值合理"},
		[]string{"建立底仓", "保留资金"})
	if h1 != h2 {
		t.Errorf("相同输入应产生相同哈希: %s vs %s", h1, h2)
	}

	// 不同 domain 应产生不同哈希
	h3 := ComputeExactHash("RISK", "高概率区间先建底仓",
		[]string{"宽基指数回撤", "估值合理"},
		[]string{"建立底仓", "保留资金"})
	if h1 == h3 {
		t.Errorf("不同 domain 应产生不同哈希")
	}

	// 不同 actions 应产生不同哈希
	h4 := ComputeExactHash("VALUATION", "高概率区间先建底仓",
		[]string{"宽基指数回撤", "估值合理"},
		[]string{"不同的动作"})
	if h1 == h4 {
		t.Errorf("不同 actions 应产生不同哈希")
	}
}

// ============================================================
// 综合场景：用户提到的真实案例
// ============================================================

func TestCheckSimilarRules_UserScenario(t *testing.T) {
	// 模拟用户描述的场景：
	// 材料 A：高概率区间先建底仓 → CR-VALUATION-20260701-004
	// 材料 B：低估合理区间不要空等极限低点 → 应被相似检查命中

	existing := []RuleFingerprint{{
		CRID:       "CR-VALUATION-20260701-004",
		ShortCode:  "VALUATION-SAFETY",
		DomainCode: "VALUATION",
		TopicCode:  "SAFETY",
		RuleName:   "高概率区间先建底仓",
		Triggers:   []string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
		Actions:    []string{"在预设的高概率区间建立第一笔底仓", "保留后续加仓资金"},
		ExactHash:  ComputeExactHash("VALUATION", "高概率区间先建底仓",
			[]string{"宽基指数出现明显回撤", "估值进入合理区或低估区", "市场情绪偏悲观"},
			[]string{"在预设的高概率区间建立第一笔底仓", "保留后续加仓资金"}),
	}}

	// 第二篇材料：低估合理区间，不要等极限低点 → AI 提取
	result := CheckSimilarRules(
		"VALUATION", "SAFETY", "低估值区间不要等极限低点",
		[]string{"宽基指数处于低估区间", "估值合理偏低", "市场情绪偏悲观"},
		[]string{"在低估区间建仓", "不要等极限低点"},
		existing,
	)

	if len(result) == 0 {
		t.Error("用户场景: '低估值区间不要等极限低点' 应能检测到与 '高概率区间先建底仓' 的相似性，但返回了空列表")
		t.Error("→ 这说明当前算法在规则名差异较大时可能漏检")
	} else {
		t.Logf("用户场景: 检测到 %d 条相似规则:", len(result))
		for _, r := range result {
			t.Logf("  - CRID=%s, Level=%s, Reason=%s", r.CRID, r.Level, r.Reason)
		}
	}
}
