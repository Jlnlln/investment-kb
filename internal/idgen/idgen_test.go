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
	expectedRawPrefix := "RAW-POS-SAFETY-20260609-001"
	if ids.RawID != expectedRawPrefix {
		t.Errorf("RawID 格式错误: got %s, want %s", ids.RawID, expectedRawPrefix)
	}

	// 验证 QA ID 格式
	if ids.QAID == "" {
		t.Error("QAID 为空")
	}
	expectedQAPrefix := "QA-POS-SAFETY-20260609-001"
	if ids.QAID != expectedQAPrefix {
		t.Errorf("QAID 格式错误: got %s, want %s", ids.QAID, expectedQAPrefix)
	}

	// 验证 CASE ID 为空（因为 ShouldGenerateCase = false）
	if ids.CaseID != "" {
		t.Errorf("CaseID 应为空，实际为: %s", ids.CaseID)
	}

	// 验证 CR IDs
	if len(ids.CandidateIDs) != len(result.CandidateRules) {
		t.Errorf("CandidateIDs 数量错误: got %d, want %d", len(ids.CandidateIDs), len(result.CandidateRules))
	}

	// 验证第一条 CR ID
	expectedCRPrefix := "CR-BUY-SAFETY-20260609-001"
	if len(ids.CandidateIDs) > 0 && ids.CandidateIDs[0] != expectedCRPrefix {
		t.Errorf("CR ID 格式错误: got %s, want %s", ids.CandidateIDs[0], expectedCRPrefix)
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