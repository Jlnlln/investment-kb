package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"investment-kb/internal/config"
)

type standaloneIndexEntry struct {
	ID           string
	Title        string
	Path         string
	Source       string
	MaterialType string
	CleanedHash  string
	RawID        string
	CRCount      int
}

func UpdateRawMaterialIndex(cfg *config.Config) error {
	entries := scanStandaloneIndexEntries(cfg.ObsidianVaultPath, GetRawMaterialDir(cfg), "RAW-")
	content := renderRawMaterialIndex(entries)
	return writeStandaloneIndex(filepath.Join(cfg.ObsidianVaultPath, GetRawMaterialIndexPath(cfg)), content)
}

func UpdateQaIndex(cfg *config.Config) error {
	entries := scanStandaloneIndexEntries(cfg.ObsidianVaultPath, GetQaDir(cfg), "QA-")
	rawEntries := scanStandaloneIndexEntries(cfg.ObsidianVaultPath, GetRawMaterialDir(cfg), "RAW-")
	rawByID := make(map[string]standaloneIndexEntry, len(rawEntries))
	for _, entry := range rawEntries {
		rawByID[entry.ID] = entry
	}
	content := renderQaIndex(entries, rawByID)
	return writeStandaloneIndex(filepath.Join(cfg.ObsidianVaultPath, GetQaIndexPath(cfg)), content)
}

func scanStandaloneIndexEntries(vaultPath, dir, prefix string) []standaloneIndexEntry {
	fullDir := filepath.Join(vaultPath, dir)
	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return nil
	}
	var result []standaloneIndexEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		content := readIndexEntryContent(filepath.Join(fullDir, entry.Name()))
		name := strings.TrimSuffix(entry.Name(), ".md")
		id := name
		title := titleFromFirstHeading(content, id)
		result = append(result, standaloneIndexEntry{
			ID:           id,
			Title:        title,
			Path:         filepath.Join(dir, entry.Name()),
			Source:       parseIndexMeta(content, "来源"),
			MaterialType: parseIndexMeta(content, "material_type"),
			CleanedHash:  parseIndexMeta(content, "cleaned_hash"),
			RawID:        parseIndexMeta(content, "raw_id"),
			CRCount:      countUniqueIDs(content, "CR-"),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func titleFromFirstHeading(content, id string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		heading := strings.TrimPrefix(line, "# ")
		if !strings.HasPrefix(heading, id) {
			continue
		}
		if parts := strings.SplitN(heading, "｜", 2); len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return ""
	}
	return ""
}

func renderRawMaterialIndex(entries []standaloneIndexEntry) string {
	var sb strings.Builder
	sb.WriteString("# 原始材料索引\n\n")
	sb.WriteString(fmt.Sprintf("更新时间：%s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("原始材料总数：%d\n\n", len(entries)))
	sb.WriteString("## 全部原始材料\n\n")
	for _, entry := range entries {
		sb.WriteString(fmt.Sprintf("- %s\n", ObsidianFileLink(entry.Path, linkAlias(entry.ID, entry.Title))))
		if entry.MaterialType != "" {
			sb.WriteString(fmt.Sprintf("  - 类型：%s\n", entry.MaterialType))
		}
		if entry.Source != "" {
			sb.WriteString(fmt.Sprintf("  - 来源：%s\n", entry.Source))
		}
		if entry.CleanedHash != "" {
			sb.WriteString(fmt.Sprintf("  - cleaned_hash: %s\n", entry.CleanedHash))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func renderQaIndex(entries []standaloneIndexEntry, rawByID map[string]standaloneIndexEntry) string {
	var sb strings.Builder
	sb.WriteString("# 问答知识卡片索引\n\n")
	sb.WriteString(fmt.Sprintf("更新时间：%s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("问答知识卡片总数：%d\n\n", len(entries)))
	sb.WriteString("## 全部问答知识卡片\n\n")
	for _, entry := range entries {
		sb.WriteString(fmt.Sprintf("- %s\n", ObsidianFileLink(entry.Path, linkAlias(entry.ID, entry.Title))))
		if entry.RawID != "" {
			raw, ok := rawByID[entry.RawID]
			if ok {
				sb.WriteString(fmt.Sprintf("  - 原始材料：%s\n", ObsidianFileLink(raw.Path, linkAlias(raw.ID, raw.Title))))
			} else {
				sb.WriteString(fmt.Sprintf("  - 原始材料：[[%s|%s]]\n", entry.RawID, entry.RawID))
			}
		}
		sb.WriteString(fmt.Sprintf("  - 关联候选规则：%d 条\n", entry.CRCount))
	}
	sb.WriteString("\n")
	return sb.String()
}

func readIndexEntryContent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func parseIndexMeta(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "  "))
		if strings.HasPrefix(trimmed, key+":") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
		}
		if strings.HasPrefix(trimmed, key+"：") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, key+"："))
		}
	}
	return ""
}

func countUniqueIDs(content, prefix string) int {
	seen := make(map[string]bool)
	pattern := regexp.MustCompile(regexp.QuoteMeta(prefix) + `[A-Z]+-\d{8}-\d{3}`)
	for _, id := range pattern.FindAllString(content, -1) {
		seen[id] = true
	}
	return len(seen)
}

func writeStandaloneIndex(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建索引目录失败: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入索引文件失败: %w", err)
	}
	return nil
}
