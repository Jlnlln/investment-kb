package idgen

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"investment-kb/internal/model"
)

func resetStateForTest(t *testing.T, path string) {
	t.Helper()
	oldStateFile := stateFile
	oldStateDate := state.Date
	oldLoadedOnce := loadedOnce

	stateFile = path
	state.Date = make(map[string]map[string]int)
	loadedOnce = sync.Once{}

	t.Cleanup(func() {
		stateFile = oldStateFile
		state.Date = oldStateDate
		loadedOnce = oldLoadedOnce
	})
}

func TestLoadStateFile(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		create  bool
		want    map[string]map[string]int
	}{
		{
			name:    "normal json",
			content: []byte(`{"20260702":{"RAW-ACCOUNT-SAFETY":3}}`),
			create:  true,
			want: map[string]map[string]int{
				"20260702": map[string]int{"RAW-ACCOUNT-SAFETY": 3},
			},
		},
		{
			name:    "json with utf8 bom",
			content: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"20260702":{"QA-ACCOUNT-SAFETY":2}}`)...),
			create:  true,
			want: map[string]map[string]int{
				"20260702": map[string]int{"QA-ACCOUNT-SAFETY": 2},
			},
		},
		{
			name:    "blank file",
			content: []byte{0x20, 0x0A, 0x09, 0x20},
			create:  true,
			want:    map[string]map[string]int{},
		},
		{
			name:   "missing file",
			create: false,
			want:   map[string]map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "id_state.json")
			resetStateForTest(t, path)
			state.Date = map[string]map[string]int{"stale": map[string]int{"OLD": 9}}

			if tt.create {
				if err := os.WriteFile(path, tt.content, 0644); err != nil {
					t.Fatalf("write state file: %v", err)
				}
			}

			if err := loadStateFile(); err != nil {
				t.Fatalf("loadStateFile() error = %v", err)
			}
			assertStateDate(t, tt.want)
		})
	}
}

func TestLoadStateFileInvalidJSONIncludesPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "id_state.json")
	resetStateForTest(t, path)
	if err := os.WriteFile(path, []byte("not-json"), 0644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	err := loadStateFile()
	if err == nil {
		t.Fatal("loadStateFile() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error %q does not include path %q", err.Error(), path)
	}
}

func TestGenerateIDs(t *testing.T) {
	resetStateForTest(t, filepath.Join(t.TempDir(), "id_state.json"))

	result := model.MockExtractionResult()
	now := time.Date(2026, 6, 9, 0, 0, 0, 0, time.Local)

	ids, err := GenerateIDs(result, now)
	if err != nil {
		t.Fatalf("GenerateIDs 失败: %v", err)
	}

	if ids.RawID == "" {
		t.Error("RawID 为空")
	}
	expectedRawPrefix := "RAW-ACCOUNT-SAFETY-20260609-001"
	if ids.RawID != expectedRawPrefix {
		t.Errorf("RawID 格式错误: got %s, want %s", ids.RawID, expectedRawPrefix)
	}

	if ids.QAID == "" {
		t.Error("QAID 为空")
	}
	expectedQAPrefix := "QA-ACCOUNT-SAFETY-20260609-001"
	if ids.QAID != expectedQAPrefix {
		t.Errorf("QAID 格式错误: got %s, want %s", ids.QAID, expectedQAPrefix)
	}

	if ids.CaseID != "" {
		t.Errorf("CaseID 应为空，实际为: %s", ids.CaseID)
	}

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
	resetStateForTest(t, filepath.Join(t.TempDir(), "id_state.json"))
	dateStr := "20260609"
	prefix := "TEST"

	seq1 := nextSequence(dateStr, prefix)
	if seq1 != 1 {
		t.Errorf("第一次调用应该返回 1，实际为: %d", seq1)
	}

	seq2 := nextSequence(dateStr, prefix)
	if seq2 != 2 {
		t.Errorf("第二次调用应该返回 2，实际为: %d", seq2)
	}

	otherSeq := nextSequence(dateStr, "OTHER")
	if otherSeq != 1 {
		t.Errorf("不同前缀应该从 1 开始，实际为: %d", otherSeq)
	}
}

func assertStateDate(t *testing.T, want map[string]map[string]int) {
	t.Helper()
	if len(state.Date) != len(want) {
		t.Fatalf("len(state.Date) = %d, want %d; state=%v", len(state.Date), len(want), state.Date)
	}
	for date, wantPrefixes := range want {
		gotPrefixes, ok := state.Date[date]
		if !ok {
			t.Fatalf("missing date %q in state %v", date, state.Date)
		}
		if len(gotPrefixes) != len(wantPrefixes) {
			t.Fatalf("len(state.Date[%q]) = %d, want %d", date, len(gotPrefixes), len(wantPrefixes))
		}
		for prefix, wantSeq := range wantPrefixes {
			if got := gotPrefixes[prefix]; got != wantSeq {
				t.Fatalf("state.Date[%q][%q] = %d, want %d", date, prefix, got, wantSeq)
			}
		}
	}
}
