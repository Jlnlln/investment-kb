package screening

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateCRContentAppendsThenReplacesConclusion(t *testing.T) {
	first, result, err := UpdateCRContent(sampleCR("待验证", "否", ""), sampleDecision(), "2026-07-06")
	if err != nil {
		t.Fatal(err)
	}
	if !result.FrontFieldsUpdated || !result.ConclusionUpdated {
		t.Fatalf("expected CR update flags, got %+v", result)
	}
	if countHeading(first, "## 第一轮筛选结论") != 1 {
		t.Fatalf("expected one conclusion, got:\n%s", first)
	}
	if !strings.Contains(first, "第一轮筛选：A｜重点验证") {
		t.Fatalf("missing front screening field:\n%s", first)
	}

	second, _, err := UpdateCRContent(first, sampleDecision(), "2026-07-06")
	if err != nil {
		t.Fatal(err)
	}
	if countHeading(second, "## 第一轮筛选结论") != 1 {
		t.Fatalf("expected idempotent conclusion replacement, got:\n%s", second)
	}
	if strings.Count(second, "第一轮筛选：A｜重点验证") != 1 {
		t.Fatalf("expected one front screening field, got:\n%s", second)
	}
}

func TestRenderConclusionIncludesLinkWatchForAAndC(t *testing.T) {
	aDecision := sampleDecision()
	aRendered := RenderConclusion(aDecision, "2026-07-07")
	if !strings.Contains(aRendered, "- 联动观察：后续与 CR-ACCOUNT-20260706-001｜禁止同步他人仓位操作、CR-ACCOUNT-20260706-003｜风险敞口承受力由账户状态决定 联动验证") {
		t.Fatalf("expected A conclusion to include link watch:\n%s", aRendered)
	}

	cDecision := sampleDecision()
	cDecision.Class = "C"
	cDecision.MergeTarget = "CR-ACCOUNT-20260706-002｜账户状态决定仓位力度"
	cRendered := RenderConclusion(cDecision, "2026-07-07")
	if !strings.Contains(cRendered, "- 联动观察：后续与 CR-ACCOUNT-20260706-001｜禁止同步他人仓位操作、CR-ACCOUNT-20260706-003｜风险敞口承受力由账户状态决定 联动验证") {
		t.Fatalf("expected C conclusion to include link watch:\n%s", cRendered)
	}
}

func TestRenderConclusionOmitsLinkWatchWhenEmpty(t *testing.T) {
	decision := sampleDecision()
	decision.MergeWatch = nil
	rendered := RenderConclusion(decision, "2026-07-07")
	if strings.Contains(rendered, "联动观察") {
		t.Fatalf("expected no link watch when merge_watch is empty:\n%s", rendered)
	}
}

func TestUpdateIndexContentInsertsAndUpdatesFields(t *testing.T) {
	first, _, err := UpdateIndexContent(sampleIndex(), "CR-ACCOUNT-20260706-002", sampleDecision())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(first, "第一轮筛选：A｜重点验证") != 1 {
		t.Fatalf("expected inserted screening field:\n%s", first)
	}
	updatedDecision := sampleDecision()
	updatedDecision.Class = "B"
	updatedDecision.Position = "暂存观察"
	second, _, err := UpdateIndexContent(first, "CR-ACCOUNT-20260706-002", updatedDecision)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(second, "第一轮筛选：") != 1 || !strings.Contains(second, "第一轮筛选：B｜暂存观察") {
		t.Fatalf("expected updated screening field without duplicate:\n%s", second)
	}
}

func TestValidateDecisionRejectsInvalidMergeRules(t *testing.T) {
	cDecision := sampleDecision()
	cDecision.Class = "C"
	cDecision.MergeTarget = ""
	if err := ValidateDecision("CR-ACCOUNT-20260706-002", cDecision); err == nil {
		t.Fatal("expected C without merge_target to fail")
	}
	aDecision := sampleDecision()
	aDecision.MergeTarget = "CR-ACCOUNT-20260706-001"
	if err := ValidateDecision("CR-ACCOUNT-20260706-002", aDecision); err == nil {
		t.Fatal("expected A with merge_target to fail")
	}
}

func TestSelectDecisionsSkipsEmptyTemplatesWithoutID(t *testing.T) {
	selected, err := SelectDecisions(map[string]Decision{
		"CR-EMPTY-20260706-001":   {},
		"CR-ACCOUNT-20260706-002": sampleDecision(),
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 {
		t.Fatalf("expected one selected decision, got %d", len(selected))
	}
	if _, ok := selected["CR-ACCOUNT-20260706-002"]; !ok {
		t.Fatalf("expected filled decision to be selected: %#v", selected)
	}
}

func TestQueueLabelDefaultsByClass(t *testing.T) {
	cases := []struct {
		class string
		want  string
	}{
		{class: "A", want: "A｜待验证"},
		{class: "B", want: "B｜观察中"},
		{class: "C", want: "C｜待吸收"},
		{class: "D", want: "D｜已废弃"},
		{class: "", want: "新增待筛选"},
	}
	for _, tc := range cases {
		if got := QueueLabel(Decision{Class: tc.class}); got != tc.want {
			t.Fatalf("QueueLabel(%q) = %q, want %q", tc.class, got, tc.want)
		}
	}
}

func TestUpdateQueueSectionGroupsRules(t *testing.T) {
	index := `# 候选规则索引

## 按领域

### ACCOUNT

- [[rules/CR-A-20260706-001|CR-A-20260706-001｜A rule]]
- [[rules/CR-B-20260706-001|CR-B-20260706-001｜B rule]]
- [[rules/CR-C-20260706-001|CR-C-20260706-001｜C rule]]
- [[rules/CR-D-20260706-001|CR-D-20260706-001｜D rule]]
- [[rules/CR-NEW-20260706-001|CR-NEW-20260706-001｜new rule]]
`
	updated, changed := UpdateQueueSection(index, map[string]Decision{
		"CR-A-20260706-001": {Class: "A", Reasons: []string{"a"}},
		"CR-B-20260706-001": {Class: "B", Reasons: []string{"b"}},
		"CR-C-20260706-001": {Class: "C", MergeTarget: "CR-A-20260706-001｜A rule", Reasons: []string{"c"}},
		"CR-D-20260706-001": {Class: "D", Reasons: []string{"d"}},
	})
	if !changed {
		t.Fatal("expected queue section to be inserted")
	}
	required := []string{
		"## 当前处理队列",
		"### A｜待验证",
		"### B｜观察中",
		"### C｜待吸收",
		"[[rules/CR-C-20260706-001|CR-C-20260706-001｜C rule]] → CR-A-20260706-001｜A rule",
		"### D｜已废弃",
		"### 新增待筛选",
		"[[rules/CR-NEW-20260706-001|CR-NEW-20260706-001｜new rule]]",
	}
	for _, want := range required {
		if !strings.Contains(updated, want) {
			t.Fatalf("expected queue section to contain %q:\n%s", want, updated)
		}
	}
}

func TestNewCRMovesFromUnscreenedToCQueue(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(CandidateRuleDir))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	index := `# 候选规则索引

## 按领域

### RISK

- [[03-规则/候选规则/CR-RISK-20260706-002|CR-RISK-20260706-002｜低确定性判断必须分仓应对]]
`
	if err := os.WriteFile(filepath.Join(dir, "候选规则索引.md"), []byte(index), 0644); err != nil {
		t.Fatal(err)
	}
	newCR := sampleNewCR()
	newPath := filepath.Join(dir, "CR-RISK-20260710-001.md")
	if err := os.WriteFile(newPath, []byte(newCR), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err := NewPaths(root, DefaultDecisionsPath)
	if err != nil {
		t.Fatal(err)
	}
	items := CandidateRuleItems(paths, index)
	unscreened, _ := UpdateQueueSectionWithItems(index, items, map[string]Decision{})
	if !strings.Contains(unscreened, "### 新增待筛选") || !strings.Contains(unscreened, "CR-RISK-20260710-001｜RISK-DISC｜测试新增规则闭环") {
		t.Fatalf("expected new CR in unscreened queue:\n%s", unscreened)
	}
	if strings.Contains(sectionBetween(unscreened, "### C｜待吸收", "### D｜已废弃"), "CR-RISK-20260710-001") {
		t.Fatalf("new CR should not be in C queue before decision:\n%s", unscreened)
	}

	decision := sampleNewCRDecision()
	decisions := map[string]Decision{"CR-RISK-20260710-001": decision}
	screened, _ := UpdateQueueSectionWithItems(index, items, decisions)
	cSection := sectionBetween(screened, "### C｜待吸收", "### D｜已废弃")
	if !strings.Contains(cSection, "CR-RISK-20260710-001") || !strings.Contains(cSection, "→ CR-RISK-20260706-002｜低确定性判断必须分仓应对") {
		t.Fatalf("expected screened CR in C queue with merge target:\n%s", screened)
	}
	if strings.Contains(sectionBetween(screened, "### 新增待筛选", "## 按批次"), "CR-RISK-20260710-001") {
		t.Fatalf("screened CR should leave unscreened queue:\n%s", screened)
	}

	updated, result, err := UpdateCRContent(newCR, decision, "2026-07-07")
	if err != nil {
		t.Fatal(err)
	}
	if !result.FrontFieldsUpdated || !result.ConclusionUpdated {
		t.Fatalf("expected CR content update, got %+v", result)
	}
	if ExtractSourceMeta(newCR) != ExtractSourceMeta(updated) {
		t.Fatal("source_meta changed")
	}
	if strings.Contains(updated, "验证状态：已验证") || strings.Contains(updated, "是否可转正式：是") {
		t.Fatal("protected fields changed")
	}
	required := []string{
		"第一轮筛选：C｜合并到其他规则",
		"当前处理队列：C｜待吸收",
		"合并去向：CR-RISK-20260706-002｜低确定性判断必须分仓应对",
		"联动观察：后续与 CR-ACCOUNT-20260706-002｜账户状态决定仓位力度 联动验证",
	}
	for _, want := range required {
		if !strings.Contains(updated, want) {
			t.Fatalf("updated CR missing %q:\n%s", want, updated)
		}
	}
	again, _, err := UpdateCRContent(updated, decision, "2026-07-07")
	if err != nil {
		t.Fatal(err)
	}
	if countHeading(again, "## 第一轮筛选结论") != 1 {
		t.Fatalf("expected idempotent conclusion, got:\n%s", again)
	}
	for _, field := range []string{"第一轮筛选：", "当前处理队列：", "合并去向：", "联动观察："} {
		if strings.Count(again, field) != strings.Count(updated, field) {
			t.Fatalf("field count changed after repeated update for %s", field)
		}
	}
}

func TestApplyRejectsPromotedFieldsAndPreservesSourceMeta(t *testing.T) {
	root := t.TempDir()
	paths := writeCase(t, root, sampleCR("待验证", "否", ""))
	decisions := map[string]Decision{"CR-ACCOUNT-20260706-002": sampleDecision()}
	beforePath, _ := paths.CRPath("CR-ACCOUNT-20260706-002")
	beforeData, _ := os.ReadFile(beforePath)
	beforeMeta := map[string]string{"CR-ACCOUNT-20260706-002": ExtractSourceMeta(string(beforeData))}
	if err := Run(Options{KBRoot: root, ID: "CR-ACCOUNT-20260706-002", Apply: true, Date: "2026-07-06"}); err != nil {
		t.Fatal(err)
	}
	afterData, _ := os.ReadFile(beforePath)
	if beforeMeta["CR-ACCOUNT-20260706-002"] != ExtractSourceMeta(string(afterData)) {
		t.Fatal("source_meta changed")
	}

	root = t.TempDir()
	paths = writeCase(t, root, sampleCR("已验证", "否", ""))
	if err := ValidateApplied(paths, beforeMeta, decisions); err == nil {
		t.Fatal("expected promoted validation status to fail")
	}

	root = t.TempDir()
	paths = writeCase(t, root, sampleCR("待验证", "是", ""))
	if err := ValidateApplied(paths, beforeMeta, decisions); err == nil {
		t.Fatal("expected formal promotion to fail")
	}
}

func TestIndexIncludesBatchAndLifecycleSections(t *testing.T) {
	index := "# 候选规则索引\n\n## 按领域\n\n### ACCOUNT\n\n"
	decisions := make(map[string]Decision)
	add := func(id, title, class, batch, queue, mergeTarget string) {
		index += "- [[03-规则/候选规则/" + id + "|" + id + "｜" + title + "]]\n"
		decisions[id] = Decision{
			Class:       class,
			Batch:       batch,
			Queue:       queue,
			MergeTarget: mergeTarget,
			Reasons:     []string{"reason"},
		}
	}
	add("CR-ACCOUNT-20260706-001", "规则1", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-ACCOUNT-20260706-002", "规则2", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-ACCOUNT-20260706-003", "规则3", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-ACCOUNT-20260706-004", "规则4", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-RISK-20260706-001", "规则5", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-RISK-20260706-002", "规则6", "A", "20260706-BATCH-001", "A｜待验证", "")
	add("CR-RISK-20260706-003", "规则7", "B", "20260706-BATCH-001", "B｜观察中", "")
	add("CR-RISK-20260706-004", "规则8", "B", "20260706-BATCH-001", "B｜观察中", "")
	add("CR-RISK-20260706-005", "规则9", "B", "20260706-BATCH-001", "B｜观察中", "")
	add("CR-RISK-20260706-006", "规则10", "C", "20260706-BATCH-001", "C｜待吸收", "CR-RISK-20260706-002｜低确定性判断必须分仓应对")
	add("CR-RISK-20260706-007", "规则11", "C", "20260706-BATCH-001", "C｜待吸收", "CR-RISK-20260706-002｜低确定性判断必须分仓应对")
	add("CR-VALUATION-20260706-001", "规则12", "C", "20260706-BATCH-001", "C｜待吸收", "CR-RISK-20260706-002｜低确定性判断必须分仓应对")
	add("CR-VALUATION-20260706-002", "规则13", "C", "20260706-BATCH-001", "C｜待吸收", "CR-RISK-20260706-002｜低确定性判断必须分仓应对")
	add("CR-ACCOUNT-20260707-001", "持续亏损者必须退回宽基指数策略", "B", "20260707-BATCH-002", "B｜观察中", "")
	add("CR-RISK-20260707-001", "趋势投资买入点必须控制止损敞口", "C", "20260707-BATCH-002", "C｜待吸收", "CR-RISK-20260706-006｜买入前必须预设退出策略")
	index += "\n---\n\n## 按状态\n\n### 候选\n\n- legacy\n\n## 全部候选规则\n\n- legacy\n"

	updated, changed := UpdateQueueSection(index, decisions)
	if !changed {
		t.Fatal("expected index to change")
	}
	required := []string{
		"## 当前处理队列",
		"## 按批次",
		"### 20260706-BATCH-001",
		"- 总数：13",
		"- A｜重点验证：6",
		"- B｜暂存观察：3",
		"- C｜合并到其他规则：4",
		"- D｜废弃：0",
		"### 20260707-BATCH-002",
		"- 总数：2",
		"- A｜重点验证：0",
		"- B｜暂存观察：1",
		"- C｜合并到其他规则：1",
		"## 生命周期状态",
		"### 新增待筛选\n\n暂无",
		"CR-RISK-20260707-001｜趋势投资买入点必须控制止损敞口]] → CR-RISK-20260706-006｜买入前必须预设退出策略",
		"- 合并去向：CR-RISK-20260706-006｜买入前必须预设退出策略",
	}
	for _, want := range required {
		if !strings.Contains(updated, want) {
			t.Fatalf("expected updated index to contain %q:\n%s", want, updated)
		}
	}
	if strings.Contains(updated, "## 按状态") {
		t.Fatalf("old status section should be removed:\n%s", updated)
	}
}

func TestBasicCaseFixture(t *testing.T) {
	root := t.TempDir()
	copyDir(t, filepath.Join("..", "..", "testdata", "crscreen", "basic_case", "input"), root)
	if err := Run(Options{KBRoot: root, ID: "CR-ACCOUNT-20260706-002", Apply: true, Date: "2026-07-06"}); err != nil {
		t.Fatal(err)
	}
	assertFileEquals(t,
		filepath.Join(root, filepath.FromSlash(CandidateRuleIndex)),
		filepath.Join("..", "..", "testdata", "crscreen", "basic_case", "expected", filepath.FromSlash(CandidateRuleIndex)),
	)
	assertFileEquals(t,
		filepath.Join(root, filepath.FromSlash(CandidateRuleDir), "CR-ACCOUNT-20260706-002.md"),
		filepath.Join("..", "..", "testdata", "crscreen", "basic_case", "expected", filepath.FromSlash(CandidateRuleDir), "CR-ACCOUNT-20260706-002.md"),
	)
}

func writeCase(t *testing.T, root, crContent string) Paths {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(CandidateRuleDir))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "候选规则索引.md"), []byte(sampleIndex()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CR-ACCOUNT-20260706-002.md"), []byte(crContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cr_screening_decisions.yaml"), []byte(sampleDecisionYAML()), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err := NewPaths(root, DefaultDecisionsPath)
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFileEquals(t *testing.T, actualPath, expectedPath string) {
	t.Helper()
	actual, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(actual)) != strings.TrimSpace(string(expected)) {
		t.Fatalf("file mismatch\nactual: %s\nexpected: %s\n\nactual content:\n%s\n\nexpected content:\n%s", actualPath, expectedPath, actual, expected)
	}
}

func sampleDecision() Decision {
	return Decision{
		Class:                "A",
		Position:             "账户状态类主规则候选",
		Action:               "保留，进入账户类重点验证池",
		MergeWatch:           []string{"CR-ACCOUNT-20260706-001｜禁止同步他人仓位操作", "CR-ACCOUNT-20260706-003｜风险敞口承受力由账户状态决定"},
		FormalCandidate:      true,
		FormalRuleSuggestion: "ACCOUNT-001｜账户状态决定仓位力度",
		PromoteBlockers:      []string{"账户状态字段未量化"},
		Reasons:              []string{"能直接影响买入、加仓、重仓和仓位力度决策"},
		Improvements:         []string{"后续需补充账户状态量化字段"},
		NextSteps:            []string{"暂不转正式规则"},
	}
}

func sampleDecisionYAML() string {
	return `CR-ACCOUNT-20260706-002:
  class: A
  position: 账户状态类主规则候选
  action: 保留，进入账户类重点验证池
  merge_target:
  merge_watch:
    - CR-ACCOUNT-20260706-001｜禁止同步他人仓位操作
    - CR-ACCOUNT-20260706-003｜风险敞口承受力由账户状态决定
  formal_candidate: true
  formal_rule_suggestion: ACCOUNT-001｜账户状态决定仓位力度
  promote_blockers:
    - 账户状态字段未量化
  reasons:
    - 能直接影响买入、加仓、重仓和仓位力度决策
  improvements:
    - 后续需补充账户状态量化字段
  next_steps:
    - 暂不转正式规则
`
}

func sampleNewCRDecision() Decision {
	return Decision{
		Batch:               "20260710-BATCH-002",
		Class:               "C",
		Queue:               "C｜待吸收",
		Position:            "测试新增规则闭环的合并逻辑",
		Action:              "不独立保留为主规则，后续合并入低确定性分仓应对规则",
		MergeTarget:         "CR-RISK-20260706-002｜低确定性判断必须分仓应对",
		MergeWatch:          []string{"CR-ACCOUNT-20260706-002｜账户状态决定仓位力度"},
		IntegrationStatus:   "待吸收",
		FormalizationStatus: "不独立转正式",
		PromoteBlockers:     []string{"测试规则，不具备独立转正式价值"},
		Reasons:             []string{"用于测试新增 CR 从新增待筛选进入 C｜待吸收队列"},
		Improvements:        []string{"测试完成后应删除 sandbox 测试文件或恢复 sandbox"},
		NextSteps:           []string{"不转正式规则"},
	}
}

func sampleNewCR() string {
	return `# CR-RISK-20260710-001｜RISK-DISC｜测试新增规则闭环

状态：候选  
验证状态：待验证  
规则验证卡：[[03-规则/规则回溯验证/规则验证卡/CR-RISK-20260710-001|CR-RISK-20260710-001｜验证卡]]  
是否可转正式：否  
筛选批次：20260710-BATCH-002  
第一轮筛选：未筛选  
当前处理队列：新增待筛选  
建议正式领域：RISK  
领域分类：RISK  
来源知识卡片：[[02-观点/问答知识卡片/QA-TEST-20260710-001|QA-TEST-20260710-001｜测试新增规则闭环]]  
来源原文：[[01-源文档/问答/RAW-TEST-20260710-001|RAW-TEST-20260710-001｜测试新增规则闭环]]  
关联案例：暂无  
适用对象：宽基指数  
---

## 1. 规则内容

这是用于测试新增 CR 增量筛选闭环的模拟规则。

<!--
source_meta:
source_file: sandbox_test_new_article.md
raw_hash: sandbox_test_raw_hash
cleaned_hash: sandbox_test_cleaned_hash
raw_id: RAW-TEST-20260710-001
material_type: rule_candidate
-->
`
}

func sectionBetween(content, start, end string) string {
	startIdx := strings.Index(content, start)
	if startIdx < 0 {
		return ""
	}
	section := content[startIdx:]
	if end == "" {
		return section
	}
	if endIdx := strings.Index(section, end); endIdx >= 0 {
		return section[:endIdx]
	}
	return section
}

func sampleIndex() string {
	return `# 候选规则索引

## 按领域

### ACCOUNT

- [[03-规则/候选规则/CR-ACCOUNT-20260706-002|CR-ACCOUNT-20260706-002｜账户状态决定仓位力度]]
  - 状态：候选
  - 验证状态：待验证
  - 验证卡：[[03-规则/规则回溯验证/规则验证卡/CR-ACCOUNT-20260706-002|CR-ACCOUNT-20260706-002｜验证卡]]
`
}

func sampleCR(validationStatus, canPromote, conclusion string) string {
	return `# CR-ACCOUNT-20260706-002｜ACCOUNT-POS｜账户状态决定仓位力度

状态：候选  
验证状态：` + validationStatus + `  
规则验证卡：[[03-规则/规则回溯验证/规则验证卡/CR-ACCOUNT-20260706-002|CR-ACCOUNT-20260706-002｜验证卡]]  
是否可转正式：` + canPromote + `  
建议正式领域：ACCOUNT  
来源知识卡片：[[02-观点/问答知识卡片/QA-ACCOUNT-SAFETY-20260706-001|QA-ACCOUNT-SAFETY-20260706-001｜安全边际]]  
来源原文：[[01-源文档/问答/RAW-ACCOUNT-SAFETY-20260706-001|RAW-ACCOUNT-SAFETY-20260706-001｜安全边际]]  
---

## 1. 规则内容

账户状态决定仓位力度。

<!--
source_meta:
source_file: input.md
raw_hash: raw
cleaned_hash: cleaned
raw_id: RAW-ACCOUNT-SAFETY-20260706-001
material_type: rule_candidate
-->
` + conclusion
}
