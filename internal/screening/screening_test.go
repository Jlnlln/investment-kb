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
