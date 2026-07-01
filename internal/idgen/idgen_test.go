package idgen

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"investment-kb/internal/model"
)

func TestGenerateIDs(t *testing.T) {
	// 清理测试状态文件
	testStateFile := filepath.Join("..", "data", "id_state_test.json")
	os.Remove(testStateFile)

	result := model.MockExtractionResult()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.Local)

	ids, err := GenerateIDs(result, now)
	if err != nil {
		t.Fatalf("GenerateIDs 失败: %v", err)
	}

	// 验证 RAW ID 格式
	if ids.RawID == "" {
		t.Error("RawID 为空")
	}
	expectedRawPrefix := "RAW-ACCOUNT-SAFETY-20260609-001"
	if ids.RawID != expectedRawPrefix {
		t.Errorf("RawID 格式错误: got %s, want %s", ids.RawID, expectedRawPrefix)
	}

	// 验证 QA ID 格式
	if ids.QAID == "" {
		t.Error("QAID 为空")
	}
	expectedQAPrefix := "QA-ACCOUNT-SAFETY-20260609-001"
	if ids.QAID != expectedQAPrefix {
		t.Errorf("QAID 格式错误: got %s, want %s", ids.QAID, expectedQAPrefix)
	}

	// 验证 CASE ID 为空（因为 ShouldGenerateCase = false）
	if ids.CaseID != "" {
		t.Errorf("CaseID 应为空，实际为: %s", ids.CaseID)
	}

	// 验证 CR IDs（按映射后的新系统领域 + 日期单独递增）
	expectedCR1 := "CR-VALUATION-20260609-001"
	expectedCR2 := "CR-ACCOUNT-20260609-001"
	expectedCR3 := "CR-RISK-20260609-001"
	if len(ids.CandidateIDs) < 3 {
		t.Fatalf("CR IDs 数量不足: got %d, want >= 3", len(ids.CandidateIDs))
	}
	if ids.CandidateIDs[0] != expectedCR1 {
		t.Errorf("CR ID 1 错误: got %s, want %s", ids.CandidateIDs[0], expectedCR1)
	}
	if ids.CandidateIDs[1] != expectedCR2 {
		t.Errorf("CR ID 2 错误: got %s, want %s", ids.CandidateIDs[1], expectedCR2)
	}
	if ids.CandidateIDs[2] != expectedCR3 {
		t.Errorf("CR ID 3 错误: got %s, want %s", ids.CandidateIDs[2], expectedCR3)
	}
}

func TestNextSequence(t *testing.T) {
	dateStr := "20260609"
	prefix := "TEST"

	// 第一次调用
	seq1 := nextSequence(dateStr, prefix)
	if seq1 != 1 {
		t.Errorf("第一次调用应该返回 1，实际为: %d", seq1)
	}

	// 第二次调用
	seq2 := nextSequence(dateStr, prefix)
	if seq2 != 2 {
		t.Errorf("第二次调用应该返回 2，实际为: %d", seq2)
	}

	// 不同前缀应该独立计数
	otherSeq := nextSequence(dateStr, "OTHER")
	if otherSeq != 1 {
		t.Errorf("不同前缀应该从 1 开始，实际为: %d", otherSeq)
	}
}
