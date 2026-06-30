package markdown

import (
	"testing"
	"time"

	"investment-kb/internal/model"
)

func TestRenderRawMaterial(t *testing.T) {
	result := model.MockExtractionResult()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.Local)
	rawText := "这是原始文本"

	ids, err := mockGenerateIDs(result, now)
	if err != nil {
		t.Fatalf("生成 ID 失败: %v", err)
	}

	md := RenderRawMaterial(nil, ids, result, rawText, now)

	// 验证包含必要元素
	if len(md) == 0 {
		t.Error("Markdown 输出为空")
	}
	// 验证包含标题
	if !contains(md, "# RAW-") {
		t.Error("未找到标题")
	}
	// 验证包含原始文本
	if !contains(md, rawText) {
		t.Error("未找到原始文本")
	}
}

func TestRenderKnowledgeCard(t *testing.T) {
	result := model.MockExtractionResult()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.Local)

	ids, err := mockGenerateIDs(result, now)
	if err != nil {
		t.Fatalf("生成 ID 失败: %v", err)
	}

	md := RenderKnowledgeCard(nil, ids, result, now)

	// 验证包含必要元素
	if len(md) == 0 {
		t.Error("Markdown 输出为空")
	}
	// 验证包含 QA- 标题
	if !contains(md, "# QA-") {
		t.Error("未找到标题")
	}
	// 验证包含核心结论
	if !contains(md, "## 2. 核心结论") {
		t.Error("未找到核心结论章节")
	}
}

func TestRenderCandidateRules(t *testing.T) {
	result := model.MockExtractionResult()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.Local)

	ids, err := mockGenerateIDs(result, now)
	if err != nil {
		t.Fatalf("生成 ID 失败: %v", err)
	}

	md := RenderCandidateRules(nil, ids, result, result.CandidateRules)

	// 验证包含必要元素
	if len(md) == 0 {
		t.Error("Markdown 输出为空")
	}
	// 验证包含 CR- 标题
	if !contains(md, "# CR-") {
		t.Error("未找到标题")
	}
	// 验证规则数量
	expectedCount := len(result.CandidateRules)
	actualCount := countOccurrences(md, "# CR-")
	if actualCount != expectedCount {
		t.Errorf("候选规则数量错误: got %d, want %d", actualCount, expectedCount)
	}
}

// 辅助函数

func mockGenerateIDs(result *model.ExtractionResult, now time.Time) (*model.DocumentIDs, error) {
	// 简化的 ID 生成，用于测试
	return &model.DocumentIDs{
		RawID:        "RAW-POS-SAFETY-20260609-001",
		QAID:         "QA-POS-SAFETY-20260609-001",
		CandidateIDs: []string{
			"CR-BUY-SAFETY-20260609-001",
			"CR-POS-ACCOUNT-20260609-002",
			"CR-RISK-PLAN-20260609-003",
		},
	}, nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func countOccurrences(s, substr string) int {
	count := 0
	idx := indexOf(s, substr)
	for idx >= 0 {
		count++
		s = s[idx+len(substr):]
		idx = indexOf(s, substr)
	}
	return count
}