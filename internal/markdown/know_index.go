package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"investment-kb/internal/config"
)

// layerTopicOrder 分层标签排序（L2 在前，L3 在后，同层按字母序）
var layerTopicOrder = map[string]int{
	"L2": 1,
	"L3": 2,
}

// knowIndexEntry 索引条目
type knowIndexEntry struct {
	Layer  string
	Topic  string
	Title  string
	KNOWID string
	Path   string
}

// layerTopicNames 分层-主题中文映射
var layerTopicNames = map[string]map[string]string{
	"L2": {
		"ECON": "经济周期",
		"GROW": "增长/复苏",
		"DEBT": "债务/信用",
	},
	"L3": {
		"RATE":   "利率",
		"POLICY": "政策调控",
	},
}

// getLayerTopicCN 获取分层主题的中文名
func getLayerTopicCN(layer, topic string) string {
	if names, ok := layerTopicNames[layer]; ok {
		if name, ok := names[topic]; ok {
			return name
		}
	}
	return topic
}

// getLayerCN 获取分层的中文名
func getLayerCN(layer string) string {
	switch layer {
	case "L2":
		return "L2 经济周期"
	case "L3":
		return "L3 政策与流动性"
	default:
		return layer
	}
}

// ScanKnowCards 扫描 KNOW 目录中已有的 KNOW 卡文件，提取 ID 和标题
func ScanKnowCards(vaultPath, knowDir string) []knowIndexEntry {
	fullDir := filepath.Join(vaultPath, knowDir)
	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return nil
	}

	var result []knowIndexEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// 跳过索引文件
		if strings.HasPrefix(entry.Name(), "宏观理解卡索引") {
			continue
		}
		// V1.5.1 起文件名只保留 KNOW-ID，标题从正文 H1 读取。
		knowID := strings.TrimSuffix(entry.Name(), ".md")
		if !strings.HasPrefix(knowID, "KNOW-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(fullDir, entry.Name()))
		if err != nil {
			continue
		}
		title := titleFromFirstHeading(string(data), knowID)

		// 从 KNOW-ID 解析 layer 和 topic：KNOW-L3-RATE-...
		idParts := strings.SplitN(knowID, "-", 4)
		var layer, topic string
		if len(idParts) >= 3 {
			layer = idParts[1]
			topic = idParts[2]
		}

		result = append(result, knowIndexEntry{
			Layer:  layer,
			Topic:  topic,
			Title:  title,
			KNOWID: knowID,
			Path:   filepath.Join(knowDir, entry.Name()),
		})
	}

	// 按 layer→topic→KNOWID 排序
	sort.Slice(result, func(i, j int) bool {
		li, lj := layerTopicOrder[result[i].Layer], layerTopicOrder[result[j].Layer]
		if li != lj {
			return li < lj
		}
		if result[i].Topic != result[j].Topic {
			return result[i].Topic < result[j].Topic
		}
		return result[i].KNOWID < result[j].KNOWID
	})

	return result
}

// RenderKnowIndex 渲染宏观理解卡索引文件
func RenderKnowIndex(entries []knowIndexEntry) string {
	var sb strings.Builder

	sb.WriteString("# 宏观理解卡索引\n\n")

	if len(entries) == 0 {
		sb.WriteString("（暂无宏观理解卡）\n")
		return sb.String()
	}

	// 按 layer 分组
	currentLayer := ""
	currentTopic := ""
	for _, entry := range entries {
		// 新 layer → 写分层标题
		if entry.Layer != currentLayer {
			currentLayer = entry.Layer
			currentTopic = ""
			sb.WriteString(fmt.Sprintf("## %s\n\n", getLayerCN(entry.Layer)))
		}
		// 新 topic → 写主题标题
		if entry.Topic != currentTopic {
			currentTopic = entry.Topic
			sb.WriteString(fmt.Sprintf("### %s\n\n", getLayerTopicCN(entry.Layer, entry.Topic)))
		}
		// 写条目
		sb.WriteString(fmt.Sprintf("- %s\n", ObsidianFileLink(entry.Path, linkAlias(entry.KNOWID, entry.Title))))
	}

	sb.WriteString("\n")
	return sb.String()
}

// UpdateKnowIndex 更新宏观理解卡索引文件（扫描目录 + 重新生成）
func UpdateKnowIndex(cfg *config.Config) error {
	vaultPath := cfg.ObsidianVaultPath
	knowDir := cfg.Files.MacroKnowledgeDir

	entries := ScanKnowCards(vaultPath, knowDir)
	indexContent := RenderKnowIndex(entries)

	indexPath := filepath.Join(vaultPath, cfg.Files.MacroKnowledgeIndex)
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return fmt.Errorf("创建索引目录失败: %w", err)
	}
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("写入索引文件失败: %w", err)
	}

	fmt.Printf("   ✅ 宏观理解卡索引已更新（%d 条）\n", len(entries))
	return nil
}

// CheckSimilarKnowCards 检查 KNOW 卡相似去重（轻量级）
// 规则：同 layer + topic 下的 KNOW，标题关键词重叠 > 70% 则提示疑似重复
func CheckSimilarKnowCards(vaultPath, knowDir string, newKnowID, newTitle, newLayer, newTopic string) []string {
	entries := ScanKnowCards(vaultPath, knowDir)
	var warnings []string

	for _, entry := range entries {
		// 只比较同 layer + topic 的 KNOW
		if entry.Layer != newLayer || entry.Topic != newTopic {
			continue
		}
		// 跳过自身
		if entry.KNOWID == newKnowID {
			continue
		}
		// 标题关键词重叠检查
		newWords := extractKeywords(newTitle)
		existWords := extractKeywords(entry.Title)
		overlap := countOverlap(newWords, existWords)
		if overlap > 0 && float64(overlap)/float64(maxLen(len(newWords), len(existWords))) > 0.5 {
			overlapPct := float64(overlap) * 100 / float64(maxLen(len(newWords), len(existWords)))
			warnings = append(warnings, fmt.Sprintf("疑似重复：%s（标题重叠度 %.0f%%）",
				ObsidianFileLink(entry.Path, linkAlias(entry.KNOWID, entry.Title)), overlapPct))
		}
	}

	return warnings
}

// extractKeywords 从标题中提取关键词
func extractKeywords(title string) []string {
	// 简单分词：按空格、逗号、和常见分隔符拆分
	seps := []string{" ", ",", "，", "、", "｜", "|", "：", ":", "的", "与", "对", "和", "在"}
	result := []string{title}
	for _, sep := range seps {
		var newResult []string
		for _, s := range result {
			parts := strings.Split(s, sep)
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" && len(p) >= 2 {
					newResult = append(newResult, p)
				}
			}
		}
		result = newResult
	}
	return result
}

// countOverlap 计算两组关键词的重叠数
func countOverlap(a, b []string) int {
	count := 0
	bSet := make(map[string]bool, len(b))
	for _, w := range b {
		bSet[w] = true
	}
	for _, w := range a {
		if bSet[w] {
			count++
		}
	}
	return count
}

func maxLen(a, b int) int {
	if a > b {
		return a
	}
	return b
}
