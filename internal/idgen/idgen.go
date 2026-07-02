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

// MapCRDomain 将旧/中间领域代码映射到新系统正式领域代码
// 用于候选规则 CR 编号前缀，使其符合 v1.5 规则体系
func MapCRDomain(oldDomain string) string {
	switch oldDomain {
	case "BUY":
		return "VALUATION"
	case "POS":
		return "ACCOUNT"
	case "ALLOC":
		return "REBALANCE"
	case "CASH", "VALUATION", "REBALANCE", "EXPOSURE", "SCORE", "STATE", "TARGET", "ETF", "ACCOUNT", "RISK", "MACRO":
		return oldDomain
	default:
		return oldDomain
	}
}

// KnowTopicToLayerTopic 将 macro_knowledge 的 topic_code 映射到分层编码
// 用于 KNOW 卡编号前缀，如 KNOW-L3-RATE
func KnowTopicToLayerTopic(domainCode, topicCode string) (layer, topic string) {
	// 宏观理解分层编码表
	layerTopicMap := map[string]map[string][2]string{
		"STATE": {
			"ALLOC": {"L3", "RATE"},   // 利率/货币政策
			"PLAN":  {"L3", "POLICY"},  // 政策调控
			"RATE":  {"L3", "RATE"},    // 利率
			"POLICY": {"L3", "POLICY"}, // 政策
		},
		"MACRO": {
			"RATE":   {"L3", "RATE"},    // 利率
			"POLICY": {"L3", "POLICY"},  // 政策调控
			"ECON":   {"L2", "ECON"},    // 经济周期
			"GROW":   {"L2", "GROW"},    // 增长/复苏
			"DEBT":   {"L2", "DEBT"},    // 债务/信用
			"CREDIT": {"L3", "CREDIT"},  // 社融/信用扩张（L3 政策与流动性）
		},
	}

	if layers, ok := layerTopicMap[domainCode]; ok {
		if pair, ok := layers[topicCode]; ok {
			return pair[0], pair[1]
		}
	}
	// 兜底：如果没有匹配到，使用原始 domain/topic
	return domainCode, topicCode
}

// GenerateIDs 生成文档编号
func GenerateIDs(result *model.ExtractionResult, now time.Time) (*model.DocumentIDs, error) {
	if err := LoadState(); err != nil {
		return nil, err
	}

	dateStr := now.Format("20060102")

	ids := &model.DocumentIDs{
		CandidateIDs: make([]string, 0, len(result.CandidateRules)),
	}

	materialType := string(result.MaterialType)
	if materialType == "" {
		materialType = "rule_candidate"
	}

	switch materialType {
	case "rule_candidate":
		// RAW 和 QA 共用同一组 domain/topic
		rawPrefix := fmt.Sprintf("RAW-%s-%s", result.DomainCode, result.TopicCode)
		qaPrefix := fmt.Sprintf("QA-%s-%s", result.DomainCode, result.TopicCode)

		rawSeq := nextSequence(dateStr, rawPrefix)
		qaSeq := nextSequence(dateStr, qaPrefix)

		ids.RawID = fmt.Sprintf("%s-%s-%03d", rawPrefix, dateStr, rawSeq)
		ids.QAID = fmt.Sprintf("%s-%s-%03d", qaPrefix, dateStr, qaSeq)

	case "macro_knowledge":
		// KNOW 卡使用分层编码：RAW-L3-RATE-YYYYMMDD-001
		layer, topic := KnowTopicToLayerTopic(result.DomainCode, result.TopicCode)
		rawPrefix := fmt.Sprintf("RAW-%s-%s", layer, topic)
		knowPrefix := fmt.Sprintf("KNOW-%s-%s", layer, topic)

		rawSeq := nextSequence(dateStr, rawPrefix)
		knowSeq := nextSequence(dateStr, knowPrefix)

		ids.RawID = fmt.Sprintf("%s-%s-%03d", rawPrefix, dateStr, rawSeq)
		ids.KNOWID = fmt.Sprintf("%s-%s-%03d", knowPrefix, dateStr, knowSeq)

	case "market_observation":
		// OBS 卡使用分层编码：RAW-L2-ECON-YYYYMMDD-001
		layer, topic := KnowTopicToLayerTopic(result.DomainCode, result.TopicCode)
		rawPrefix := fmt.Sprintf("RAW-%s-%s", layer, topic)
		obsPrefix := fmt.Sprintf("OBS-%s-%s", layer, topic)

		rawSeq := nextSequence(dateStr, rawPrefix)
		obsSeq := nextSequence(dateStr, obsPrefix)

		ids.RawID = fmt.Sprintf("%s-%s-%03d", rawPrefix, dateStr, rawSeq)
		ids.OBSID = fmt.Sprintf("%s-%s-%03d", obsPrefix, dateStr, obsSeq)

	case "archive_only":
		// 仅生成 RAW 编号
		rawPrefix := fmt.Sprintf("RAW-%s-%s", result.DomainCode, result.TopicCode)
		rawSeq := nextSequence(dateStr, rawPrefix)
		ids.RawID = fmt.Sprintf("%s-%s-%03d", rawPrefix, dateStr, rawSeq)
	}

	// CASE ID（如果需要）
	if result.ShouldGenerateCase && result.Case != nil {
		casePrefix := fmt.Sprintf("CASE-%s-%s", result.Case.DomainCode, result.Case.TopicCode)
		caseSeq := nextSequence(dateStr, casePrefix)
		ids.CaseID = fmt.Sprintf("%s-%s-%03d", casePrefix, dateStr, caseSeq)
	}

	// CR IDs（按映射后的新系统领域 + 日期单独递增）
	for _, rule := range result.CandidateRules {
		crPrefix := fmt.Sprintf("CR-%s", MapCRDomain(rule.DomainCode))
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