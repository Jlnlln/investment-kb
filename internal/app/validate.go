package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/markdown"
	"investment-kb/internal/model"
)

type ValidateReport struct {
	RawCount            int
	QACount             int
	KnowCount           int
	CRCount             int
	ValidationCardCount int
	Issues              []string
}

type docRef struct {
	ID    string
	Title string
	Meta  map[string]string
	Body  string
}

func Validate(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	report := runValidation(cfg)
	printValidateReport(report)
	if len(report.Issues) > 0 {
		return fmt.Errorf("validate failed: %d issue(s)", len(report.Issues))
	}
	return nil
}

func runValidation(cfg *config.Config) ValidateReport {
	var report ValidateReport

	rawDocs := parseAggregateDocs(readVaultFile(cfg, cfg.Files.RawMaterial), "RAW-")
	qaDocs := parseAggregateDocs(readVaultFile(cfg, cfg.Files.QA), "QA-")
	crDocs := parseAggregateDocs(readVaultFile(cfg, cfg.Files.CandidateRule), "CR-")
	knowDocs := scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, cfg.Files.MacroKnowledgeDir), "KNOW-")
	vcDocs := scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, cfg.Files.ValidationCardDir), "CR-")

	report.RawCount = len(rawDocs)
	report.QACount = len(qaDocs)
	report.KnowCount = len(knowDocs)
	report.CRCount = len(crDocs)
	report.ValidationCardCount = len(vcDocs)

	if report.CRCount != report.ValidationCardCount {
		report.Issues = append(report.Issues, fmt.Sprintf("CR 数量(%d) != 验证卡数量(%d)", report.CRCount, report.ValidationCardCount))
	}

	crIDs := make(map[string]bool)
	for _, doc := range crDocs {
		crIDs[doc.ID] = true
	}
	for _, doc := range vcDocs {
		if !crIDs[doc.ID] {
			report.Issues = append(report.Issues, fmt.Sprintf("孤立验证卡：%s", doc.ID))
		}
	}

	rawByID := make(map[string]docRef)
	hashSeen := make(map[string]string)
	for _, raw := range rawDocs {
		rawByID[raw.ID] = raw
		hash := firstNonEmpty(raw.Meta["cleaned_hash"], raw.Meta["raw_hash"], raw.Meta["原文哈希"])
		if hash != "" {
			if existing, ok := hashSeen[hash]; ok {
				report.Issues = append(report.Issues, fmt.Sprintf("重复 raw_hash/cleaned_hash：%s 同时出现在 %s 和 %s", hash, existing, raw.ID))
			} else {
				hashSeen[hash] = raw.ID
			}
		}

		materialType := model.MaterialType(raw.Meta["material_type"])
		if materialType == "" {
			materialType = model.MaterialTypeRuleCandidate
		}
		body := extractRawBody(raw.Body)
		warnings := markdown.ValidateRawConsistency(&model.ExtractionResult{Title: raw.Title, MaterialType: materialType}, body)
		for _, warning := range warnings {
			report.Issues = append(report.Issues, fmt.Sprintf("RAW 标题/正文疑似错配：%s：%s", raw.ID, warning))
		}
	}

	checkSourceGroup := func(kind string, docs []docRef) {
		for _, doc := range docs {
			rawID := doc.Meta["raw_id"]
			if rawID == "" {
				report.Issues = append(report.Issues, fmt.Sprintf("%s 缺少 raw_id：%s", kind, doc.ID))
				continue
			}
			raw, ok := rawByID[rawID]
			if !ok {
				report.Issues = append(report.Issues, fmt.Sprintf("%s 引用不存在的 RAW：%s -> %s", kind, doc.ID, rawID))
				continue
			}
			compareMeta := []string{"source_file", "raw_hash", "cleaned_hash", "material_type"}
			for _, key := range compareMeta {
				if doc.Meta[key] != "" && raw.Meta[key] != "" && doc.Meta[key] != raw.Meta[key] {
					report.Issues = append(report.Issues, fmt.Sprintf("source mismatch：%s %s 字段 %s=%s，RAW %s=%s", kind, doc.ID, key, doc.Meta[key], rawID, raw.Meta[key]))
				}
			}
		}
	}

	checkSourceGroup("QA", qaDocs)
	checkSourceGroup("CR", crDocs)
	checkSourceGroup("验证卡", vcDocs)
	checkSourceGroup("KNOW", knowDocs)

	oldKnowPath := filepath.Join(cfg.ObsidianVaultPath, filepath.Dir(cfg.Files.MacroKnowledgeDir), "宏观理解卡库.md")
	if _, err := os.Stat(oldKnowPath); err == nil {
		report.Issues = append(report.Issues, "存在旧文件："+oldKnowPath)
	}

	return report
}

func printValidateReport(report ValidateReport) {
	fmt.Println("=== validate report ===")
	fmt.Printf("RAW count: %d\n", report.RawCount)
	fmt.Printf("QA count: %d\n", report.QACount)
	fmt.Printf("KNOW count: %d\n", report.KnowCount)
	fmt.Printf("CR count: %d\n", report.CRCount)
	fmt.Printf("validation card count: %d\n", report.ValidationCardCount)
	if len(report.Issues) == 0 {
		fmt.Println("source mismatch: none")
		fmt.Println("status: PASS")
		return
	}
	fmt.Println("issues:")
	for _, issue := range report.Issues {
		fmt.Println("- " + issue)
	}
	fmt.Println("status: FAIL")
}

func readVaultFile(cfg *config.Config, relativePath string) string {
	data, err := os.ReadFile(filepath.Join(cfg.ObsidianVaultPath, relativePath))
	if err != nil {
		return ""
	}
	return string(data)
}

func parseAggregateDocs(content, prefix string) []docRef {
	var docs []docRef
	lines := strings.Split(content, "\n")
	var current *docRef
	headingPrefix := "# " + prefix
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, headingPrefix) {
			if current != nil {
				docs = append(docs, *current)
			}
			id, title := parseHeading(trimmed)
			current = &docRef{ID: id, Title: title, Meta: make(map[string]string)}
			continue
		}
		if current == nil {
			continue
		}
		current.Body += line + "\n"
		if key, value, ok := parseMetaLine(line); ok {
			current.Meta[key] = value
		}
	}
	if current != nil {
		docs = append(docs, *current)
	}
	return docs
}

func scanStandaloneDocs(dir, prefix string) []docRef {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var docs []docRef
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		parsed := parseAggregateDocs(content, prefix)
		frontmatterMeta := parseFrontmatterMeta(content)
		if len(parsed) > 0 {
			for key, value := range frontmatterMeta {
				if parsed[0].Meta[key] == "" {
					parsed[0].Meta[key] = value
				}
			}
			docs = append(docs, parsed[0])
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		if idx := strings.Index(id, "｜"); idx > 0 {
			id = id[:idx]
		}
		doc := docRef{ID: id, Meta: frontmatterMeta, Body: content}
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	return docs
}

func parseHeading(heading string) (string, string) {
	heading = strings.TrimPrefix(strings.TrimSpace(heading), "# ")
	parts := strings.SplitN(heading, "｜", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func parseMetaLine(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimSuffix(trimmed, "  ")
	if strings.Contains(trimmed, ":") {
		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if isKnownMetaKey(key) {
			return key, value, true
		}
	}
	if strings.Contains(trimmed, "：") {
		parts := strings.SplitN(trimmed, "：", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "原文哈希" {
			return key, value, true
		}
	}
	return "", "", false
}

func parseFrontmatterMeta(content string) map[string]string {
	meta := make(map[string]string)
	if !strings.HasPrefix(content, "---") {
		return meta
	}
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return meta
	}
	for _, line := range strings.Split(parts[1], "\n") {
		if key, value, ok := parseMetaLine(line); ok {
			meta[key] = value
		}
	}
	return meta
}

func isKnownMetaKey(key string) bool {
	switch key {
	case "source_file", "raw_hash", "cleaned_hash", "raw_id", "material_type", "uid":
		return true
	default:
		return false
	}
}

func extractRawBody(body string) string {
	idx := strings.Index(body, "## 原文")
	if idx < 0 {
		return body
	}
	return body[idx:]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
