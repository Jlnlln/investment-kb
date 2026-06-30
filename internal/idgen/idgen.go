package idgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"investment-kb/internal/model"
)

// IDState 存储编号状态
type IDState struct {
	sync.RWMutex
	Date  map[string]map[string]int // date -> prefix -> sequence
}

var (
	stateFile = "data/id_state.json"
	state     = &IDState{
		Date: make(map[string]map[string]int),
	}
	loadedOnce sync.Once
)

// LoadState 加载编号状态文件
func LoadState() error {
	var err error
	loadedOnce.Do(func() {
		err = loadStateFile()
	})
	return err
}

func loadStateFile() error {
	state.Lock()
	defer state.Unlock()

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，使用空状态
			return nil
		}
		return fmt.Errorf("读取状态文件失败: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &state.Date); err != nil {
		return fmt.Errorf("解析状态文件失败: %w", err)
	}

	return nil
}

// SaveState 保存编号状态
func SaveState() error {
	state.Lock()
	defer state.Unlock()

	data, err := json.MarshalIndent(state.Date, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化状态失败: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("写入状态文件失败: %w", err)
	}

	return nil
}

// GenerateIDs 生成文档编号
func GenerateIDs(result *model.ExtractionResult, now time.Time) (*model.DocumentIDs, error) {
	if err := LoadState(); err != nil {
		return nil, err
	}

	dateStr := now.Format("20060102")

	// RAW 和 QA 共用同一组 domain/topic
	rawPrefix := fmt.Sprintf("RAW-%s-%s", result.DomainCode, result.TopicCode)
	qaPrefix := fmt.Sprintf("QA-%s-%s", result.DomainCode, result.TopicCode)

	rawSeq := nextSequence(dateStr, rawPrefix)
	qaSeq := nextSequence(dateStr, qaPrefix)

	ids := &model.DocumentIDs{
		RawID:        fmt.Sprintf("%s-%s-%03d", rawPrefix, dateStr, rawSeq),
		QAID:         fmt.Sprintf("%s-%s-%03d", qaPrefix, dateStr, qaSeq),
		CandidateIDs: make([]string, 0, len(result.CandidateRules)),
	}

	// CASE ID（如果需要）
	if result.ShouldGenerateCase && result.Case != nil {
		casePrefix := fmt.Sprintf("CASE-%s-%s", result.Case.DomainCode, result.Case.TopicCode)
		caseSeq := nextSequence(dateStr, casePrefix)
		ids.CaseID = fmt.Sprintf("%s-%s-%03d", casePrefix, dateStr, caseSeq)
	}

	// CR IDs（全局递增，不按 domain/topic 分组）
	for range result.CandidateRules {
		crPrefix := "CR"
		crSeq := nextSequence(dateStr, crPrefix)
		crID := fmt.Sprintf("%s-%s-%03d", crPrefix, dateStr, crSeq)
		ids.CandidateIDs = append(ids.CandidateIDs, crID)
	}

	return ids, nil
}

// nextSequence 获取下一个序号
func nextSequence(dateStr, prefix string) int {
	state.Lock()
	defer state.Unlock()

	if state.Date[dateStr] == nil {
		state.Date[dateStr] = make(map[string]int)
	}

	seq := state.Date[dateStr][prefix] + 1
	state.Date[dateStr][prefix] = seq

	return seq
}