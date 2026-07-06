package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/markdown"
	"investment-kb/internal/model"
)

type ValidateReport struct {
	RawCount               int
	QACount                int
	KnowCount              int
	CRCount                int
	ValidationCardCount    int
	RawMaterialIndex       bool
	QAIndex                bool
	CandidateRuleIndex     bool
	BrokenLinks            []brokenLink
	FrontmatterIssues      []string
	SourceMetaMissing      []string
	OrphanValidationCards  []string
	MissingValidationCards []string
	Warnings               []string
	Issues                 []string
}

type docRef struct {
	ID    string
	Title string
	Path  string
	Meta  map[string]string
	Body  string
}

type brokenLink struct {
	Source string
	Target string
	Reason string
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

	rawDocs := loadRawDocs(cfg)
	qaDocs := loadQADocs(cfg)
	crDocs := loadCandidateRuleDocs(cfg)
	knowDocs := scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, cfg.Files.MacroKnowledgeDir), "KNOW-")
	vcDocs := scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, cfg.Files.ValidationCardDir), "CR-")

	report.RawCount = len(rawDocs)
	report.QACount = len(qaDocs)
	report.KnowCount = len(knowDocs)
	report.CRCount = len(crDocs)
	report.ValidationCardCount = len(vcDocs)

	if markdown.UseStandaloneRawMaterials(cfg) {
		if _, err := os.Stat(filepath.Join(cfg.ObsidianVaultPath, markdown.GetRawMaterialIndexPath(cfg))); err == nil {
			report.RawMaterialIndex = true
		} else {
			report.Issues = append(report.Issues, "原始材料索引不存在："+markdown.GetRawMaterialIndexPath(cfg))
		}
	}
	if markdown.UseStandaloneQA(cfg) {
		if _, err := os.Stat(filepath.Join(cfg.ObsidianVaultPath, markdown.GetQaIndexPath(cfg))); err == nil {
			report.QAIndex = true
		} else {
			report.Issues = append(report.Issues, "问答知识卡片索引不存在："+markdown.GetQaIndexPath(cfg))
		}
	}

	if markdown.UseStandaloneCandidateRules(cfg) {
		if _, err := os.Stat(filepath.Join(cfg.ObsidianVaultPath, markdown.GetCandidateRuleIndexPath(cfg))); err == nil {
			report.CandidateRuleIndex = true
		} else {
			report.Issues = append(report.Issues, "候选规则索引不存在："+markdown.GetCandidateRuleIndexPath(cfg))
		}
	} else {
		report.CandidateRuleIndex = false
	}

	if report.CRCount != report.ValidationCardCount {
		report.Issues = append(report.Issues, fmt.Sprintf("CR 数量(%d) != 验证卡数量(%d)", report.CRCount, report.ValidationCardCount))
	}

	crIDs := make(map[string]bool)
	crByID := make(map[string]docRef)
	for _, doc := range crDocs {
		crIDs[doc.ID] = true
		crByID[doc.ID] = doc
	}
	vcIDs := make(map[string]bool)
	for _, doc := range vcDocs {
		vcIDs[doc.ID] = true
		if !crIDs[doc.ID] {
			report.OrphanValidationCards = append(report.OrphanValidationCards, doc.ID)
			report.Issues = append(report.Issues, fmt.Sprintf("孤立验证卡：%s", doc.ID))
		}
	}
	for _, doc := range crDocs {
		if !vcIDs[doc.ID] {
			report.MissingValidationCards = append(report.MissingValidationCards, doc.ID)
			report.Issues = append(report.Issues, fmt.Sprintf("缺失验证卡：%s", doc.ID))
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
			report.Warnings = append(report.Warnings, fmt.Sprintf("RAW 标题/正文疑似错配：%s：%s", raw.ID, warning))
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
	checkSourceMetaComplete(&report, "RAW", rawDocs)
	checkSourceMetaComplete(&report, "QA", qaDocs)
	checkSourceMetaComplete(&report, "CR", crDocs)
	checkSourceMetaComplete(&report, "验证卡", vcDocs)
	checkSourceMetaComplete(&report, "KNOW", knowDocs)

	qaByRawID := make(map[string][]docRef)
	for _, qa := range qaDocs {
		if qa.Meta["raw_id"] != "" {
			qaByRawID[qa.Meta["raw_id"]] = append(qaByRawID[qa.Meta["raw_id"]], qa)
		}
	}
	for _, cr := range crDocs {
		if model.MaterialType(cr.Meta["material_type"]) == model.MaterialTypeRuleCandidate && len(qaByRawID[cr.Meta["raw_id"]]) == 0 {
			report.Issues = append(report.Issues, fmt.Sprintf("CR 找不到同 raw_id 的 QA：%s -> %s", cr.ID, cr.Meta["raw_id"]))
		}
	}
	for _, vc := range vcDocs {
		cr, ok := crByID[vc.ID]
		if !ok {
			continue
		}
		for _, key := range []string{"source_file", "raw_hash", "cleaned_hash", "raw_id", "material_type"} {
			if vc.Meta[key] != "" && cr.Meta[key] != "" && vc.Meta[key] != cr.Meta[key] {
				report.Issues = append(report.Issues, fmt.Sprintf("source mismatch：验证卡 %s 字段 %s=%s，CR %s=%s", vc.ID, key, vc.Meta[key], vc.ID, cr.Meta[key]))
			}
		}
	}

	oldKnowPath := filepath.Join(cfg.ObsidianVaultPath, filepath.Dir(cfg.Files.MacroKnowledgeDir), "宏观理解卡库.md")
	if _, err := os.Stat(oldKnowPath); err == nil {
		report.Issues = append(report.Issues, "存在旧文件："+oldKnowPath)
	}
	if cfg.Files.RawMaterial != "" {
		legacyRawPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.RawMaterial)
		if _, err := os.Stat(legacyRawPath); err == nil {
			report.Warnings = append(report.Warnings, "legacy raw material library exists: "+legacyRawPath)
		}
	}
	if cfg.Files.QA != "" {
		legacyQAPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.QA)
		if _, err := os.Stat(legacyQAPath); err == nil {
			report.Warnings = append(report.Warnings, "legacy qa library exists: "+legacyQAPath)
		}
	}
	if cfg.Files.CandidateRule != "" {
		legacyCRPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.CandidateRule)
		if _, err := os.Stat(legacyCRPath); err == nil {
			report.Warnings = append(report.Warnings, "legacy candidate rule library exists: "+legacyCRPath)
		}
	}

	checkLinkHygiene(cfg, &report, rawDocs, qaDocs, crDocs, knowDocs, vcDocs)

	return report
}

func checkSourceMetaComplete(report *ValidateReport, kind string, docs []docRef) {
	required := []string{"source_file", "raw_hash", "cleaned_hash", "raw_id", "material_type"}
	for _, doc := range docs {
		for _, key := range required {
			if strings.TrimSpace(doc.Meta[key]) == "" {
				msg := fmt.Sprintf("%s %s 缺少 source_meta.%s", kind, doc.ID, key)
				report.SourceMetaMissing = append(report.SourceMetaMissing, msg)
				report.Issues = append(report.Issues, msg)
			}
		}
	}
}

func checkLinkHygiene(cfg *config.Config, report *ValidateReport, rawDocs, qaDocs, crDocs, knowDocs, vcDocs []docRef) {
	if markdown.UseStandaloneQA(cfg) {
		content := readVaultFile(cfg, markdown.GetQaIndexPath(cfg))
		if hasBareWikiLink(content, "RAW-") {
			report.Issues = append(report.Issues, "QA 索引存在 bare RAW 链接，应指向 RAW 独立文件")
		}
	}
	if markdown.UseStandaloneCandidateRules(cfg) {
		content := readVaultFile(cfg, markdown.GetCandidateRuleIndexPath(cfg))
		if hasBareWikiLink(content, "CR-") {
			report.Issues = append(report.Issues, "候选规则索引存在 bare 验证卡链接，应指向规则验证卡独立文件路径")
		}
	}

	contents := []string{
		readVaultFile(cfg, markdown.GetRawMaterialIndexPath(cfg)),
		readVaultFile(cfg, markdown.GetQaIndexPath(cfg)),
		readVaultFile(cfg, markdown.GetCandidateRuleIndexPath(cfg)),
		readVaultFile(cfg, markdown.GetMacroKnowledgeIndexPath(cfg)),
	}
	for _, docs := range [][]docRef{rawDocs, qaDocs, crDocs, knowDocs, vcDocs} {
		for _, doc := range docs {
			contents = append(contents, doc.Body)
		}
	}
	for _, content := range contents {
		if strings.Contains(content, "原始材料库#") || strings.Contains(content, "问答知识卡片库#") || strings.Contains(content, "候选规则库#") {
			report.Issues = append(report.Issues, "存在旧聚合库 anchor 链接")
			return
		}
	}
	report.BrokenLinks = findBrokenLinks(cfg, rawDocs, qaDocs, crDocs, knowDocs, vcDocs)
	for _, link := range report.BrokenLinks {
		report.Issues = append(report.Issues, fmt.Sprintf("broken link: source=%s target=%s reason=%s", link.Source, link.Target, link.Reason))
	}
	report.FrontmatterIssues = findFrontmatterDelimiterIssues(cfg, rawDocs, qaDocs, crDocs, knowDocs, vcDocs)
	for _, source := range report.FrontmatterIssues {
		report.Issues = append(report.Issues, "frontmatter delimiter issue: "+source+" starts with ---")
	}
}

func hasBareWikiLink(content, prefix string) bool {
	pattern := regexp.MustCompile(`\[\[` + regexp.QuoteMeta(prefix) + `[A-Z0-9-]+\]\]`)
	return pattern.MatchString(content)
}

func findBrokenLinks(cfg *config.Config, groups ...[]docRef) []brokenLink {
	var sources []linkScanSource
	indexPaths := []string{
		markdown.GetRawMaterialIndexPath(cfg),
		markdown.GetQaIndexPath(cfg),
		markdown.GetMacroKnowledgeIndexPath(cfg),
		markdown.GetCandidateRuleIndexPath(cfg),
	}
	for _, relPath := range indexPaths {
		if strings.TrimSpace(relPath) == "" {
			continue
		}
		fullPath := filepath.Join(cfg.ObsidianVaultPath, relPath)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			sources = append(sources, linkScanSource{Path: relPath, Content: string(data)})
		}
	}
	for _, docs := range groups {
		for _, doc := range docs {
			source := doc.Path
			if source == "" {
				source = doc.ID
			}
			sources = append(sources, linkScanSource{Path: source, Content: doc.Body})
		}
	}

	seen := make(map[string]bool)
	var broken []brokenLink
	for _, source := range sources {
		for _, target := range extractWikiLinkTargets(source.Content) {
			if shouldSkipLinkTarget(target) {
				continue
			}
			fileTarget := target
			if idx := strings.Index(fileTarget, "#"); idx >= 0 {
				fileTarget = fileTarget[:idx]
			}
			fileTarget = strings.TrimSpace(fileTarget)
			if fileTarget == "" {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(fileTarget), ".md") {
				fileTarget += ".md"
			}
			fullPath := filepath.Join(cfg.ObsidianVaultPath, filepath.FromSlash(fileTarget))
			if _, err := os.Stat(fullPath); err == nil {
				continue
			}
			key := source.Path + "\x00" + target
			if seen[key] {
				continue
			}
			seen[key] = true
			broken = append(broken, brokenLink{
				Source: source.Path,
				Target: target,
				Reason: "target file not found",
			})
		}
	}
	return broken
}

type linkScanSource struct {
	Path    string
	Content string
}

func extractWikiLinkTargets(content string) []string {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	var targets []string
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		target := strings.TrimSpace(match[1])
		if idx := strings.Index(target, "|"); idx >= 0 {
			target = strings.TrimSpace(target[:idx])
		}
		targets = append(targets, target)
	}
	return targets
}

func shouldSkipLinkTarget(target string) bool {
	target = strings.TrimSpace(strings.ToLower(target))
	if target == "" {
		return true
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(strings.Split(target, "#")[0]))
	if ext != "" && ext != ".md" {
		return true
	}
	return false
}

func findFrontmatterDelimiterIssues(cfg *config.Config, groups ...[]docRef) []string {
	var sources []linkScanSource
	indexPaths := []string{
		markdown.GetRawMaterialIndexPath(cfg),
		markdown.GetQaIndexPath(cfg),
		markdown.GetMacroKnowledgeIndexPath(cfg),
		markdown.GetCandidateRuleIndexPath(cfg),
	}
	for _, relPath := range indexPaths {
		if strings.TrimSpace(relPath) == "" {
			continue
		}
		fullPath := filepath.Join(cfg.ObsidianVaultPath, relPath)
		data, err := os.ReadFile(fullPath)
		if err == nil {
			sources = append(sources, linkScanSource{Path: relPath, Content: string(data)})
		}
	}
	for _, docs := range groups {
		for _, doc := range docs {
			source := doc.Path
			if source == "" {
				source = doc.ID
			}
			sources = append(sources, linkScanSource{Path: source, Content: doc.Body})
		}
	}

	var issues []string
	for _, source := range sources {
		if startsWithFrontmatterDelimiter(source.Content) {
			issues = append(issues, source.Path)
		}
	}
	return issues
}

func startsWithFrontmatterDelimiter(content string) bool {
	content = strings.TrimPrefix(content, "\ufeff")
	if idx := strings.IndexAny(content, "\r\n"); idx >= 0 {
		return strings.TrimSpace(content[:idx]) == "---"
	}
	return strings.TrimSpace(content) == "---"
}

func printValidateReport(report ValidateReport) {
	fmt.Println("=== validate report ===")
	fmt.Printf("RAW count: %d\n", report.RawCount)
	fmt.Printf("QA count: %d\n", report.QACount)
	fmt.Printf("KNOW count: %d\n", report.KnowCount)
	fmt.Printf("CR count: %d\n", report.CRCount)
	fmt.Printf("validation card count: %d\n", report.ValidationCardCount)
	if report.RawMaterialIndex {
		fmt.Println("raw material index: exists")
	} else {
		fmt.Println("raw material index: missing")
	}
	if report.QAIndex {
		fmt.Println("qa index: exists")
	} else {
		fmt.Println("qa index: missing")
	}
	if report.CandidateRuleIndex {
		fmt.Println("candidate rule index: exists")
	} else {
		fmt.Println("candidate rule index: missing")
	}
	if len(report.OrphanValidationCards) == 0 {
		fmt.Println("orphan validation cards: none")
	}
	if len(report.MissingValidationCards) == 0 {
		fmt.Println("missing validation cards: none")
	}
	if len(report.BrokenLinks) == 0 {
		fmt.Println("broken links: none")
	}
	if len(report.FrontmatterIssues) == 0 {
		fmt.Println("frontmatter delimiter issue: none")
	}
	if len(report.SourceMetaMissing) == 0 {
		fmt.Println("source_meta missing: none")
	}
	for _, warning := range report.Warnings {
		fmt.Println("warning: " + warning)
	}
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

func loadCandidateRuleDocs(cfg *config.Config) []docRef {
	if markdown.UseStandaloneCandidateRules(cfg) {
		return scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, markdown.GetCandidateRuleDir(cfg)), "CR-")
	}
	return parseAggregateDocs(readVaultFile(cfg, cfg.Files.CandidateRule), "CR-")
}

func loadRawDocs(cfg *config.Config) []docRef {
	if markdown.UseStandaloneRawMaterials(cfg) {
		return scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, markdown.GetRawMaterialDir(cfg)), "RAW-")
	}
	return parseAggregateDocs(readVaultFile(cfg, cfg.Files.RawMaterial), "RAW-")
}

func loadQADocs(cfg *config.Config) []docRef {
	if markdown.UseStandaloneQA(cfg) {
		return scanStandaloneDocs(filepath.Join(cfg.ObsidianVaultPath, markdown.GetQaDir(cfg)), "QA-")
	}
	return parseAggregateDocs(readVaultFile(cfg, cfg.Files.QA), "QA-")
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
			parsed[0].Path = filepath.Join(dir, entry.Name())
			docs = append(docs, parsed[0])
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		if idx := strings.Index(id, "｜"); idx > 0 {
			id = id[:idx]
		}
		doc := docRef{ID: id, Path: filepath.Join(dir, entry.Name()), Meta: frontmatterMeta, Body: content}
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
