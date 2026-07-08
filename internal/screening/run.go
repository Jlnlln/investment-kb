package screening

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Options struct {
	KBRoot       string
	DecisionsRel string
	ID           string
	DryRun       bool
	Apply        bool
	Init         bool
	Date         string
}

func Run(opts Options) error {
	paths, err := NewPaths(opts.KBRoot, opts.DecisionsRel)
	if err != nil {
		return err
	}
	if opts.Init {
		return InitDecisions(paths)
	}
	decisionsPath, err := paths.DecisionsPath()
	if err != nil {
		return err
	}
	all, err := LoadDecisions(decisionsPath)
	if err != nil {
		return err
	}
	filledDecisions, err := SelectDecisions(all, "")
	if err != nil {
		return err
	}
	decisions, err := SelectDecisions(all, opts.ID)
	if err != nil {
		return err
	}
	if err := ValidateInputs(paths, decisions); err != nil {
		return fail(err)
	}
	date := opts.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if !opts.Apply || opts.DryRun {
		PrintDryRun(paths, opts.ID, decisions)
		return nil
	}
	return apply(paths, decisions, filledDecisions, date)
}

func InitDecisions(paths Paths) error {
	indexPath, err := paths.IndexPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("读取候选规则索引失败: %s: %w", indexPath, err)
	}
	target, err := paths.DecisionsPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(target); err == nil {
		generated, err := paths.GeneratedDecisionsPath()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(generated), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(generated, []byte(renderEmptyDecisions(string(data))), 0644); err != nil {
			return err
		}
		fmt.Printf("[INIT] decisions already exists: %s\n", target)
		fmt.Printf("[INIT] generated template: %s\n", generated)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(target, []byte(renderEmptyDecisions(string(data))), 0644); err != nil {
		return err
	}
	fmt.Printf("[INIT] generated: %s\n", target)
	return nil
}

func PrintDryRun(paths Paths, id string, decisions map[string]Decision) {
	fmt.Printf("[DRY-RUN] kb-root: %s\n", paths.KBRoot)
	if strings.TrimSpace(id) != "" {
		fmt.Printf("[DRY-RUN] target id: %s\n", id)
	}
	for _, crID := range SortedIDs(decisions) {
		d := decisions[crID]
		fmt.Printf("\n[PLAN] %s\n", crID)
		fmt.Println("  - update CR front fields")
		fmt.Println("  - append/update 第一轮筛选结论")
		fmt.Println("  - update 候选规则索引 entry")
		for _, field := range ScreeningFrontFields(d) {
			fmt.Printf("[FIELD] %s：%s\n", field.Key, field.Value)
		}
		fmt.Printf("\n[CHECK] class = %s\n", ClassLabel(d.Class))
		if strings.TrimSpace(d.MergeTarget) != "" {
			fmt.Printf("[CHECK] merge_target = %s\n", strings.TrimSpace(d.MergeTarget))
		}
		fmt.Println("[CHECK] validation_status remains 待验证")
		fmt.Println("[CHECK] can_promote_formal remains 否")
		fmt.Println("[CHECK] formal rule generation disabled")
	}
}

func apply(paths Paths, decisions map[string]Decision, allDecisions map[string]Decision, date string) error {
	fmt.Printf("[APPLY] kb-root: %s\n", paths.KBRoot)
	indexPath, err := paths.IndexPath()
	if err != nil {
		return err
	}
	files := []string{indexPath}
	beforeMeta := make(map[string]string)
	for _, id := range SortedIDs(decisions) {
		crPath, err := paths.CRPath(id)
		if err != nil {
			return err
		}
		files = append(files, crPath)
		data, err := os.ReadFile(crPath)
		if err != nil {
			return err
		}
		beforeMeta[id] = ExtractSourceMeta(string(data))
	}
	backupRoot, err := BackupFiles(paths, uniqueStrings(files), Timestamp())
	if err != nil {
		return fail(err)
	}
	fmt.Printf("[BACKUP] %s\n", backupRoot)

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	indexContent := removeQueueSection(string(indexData))
	indexContent = EnsureIndexEntries(paths, indexContent)
	for _, id := range SortedIDs(decisions) {
		crPath, _ := paths.CRPath(id)
		data, err := os.ReadFile(crPath)
		if err != nil {
			return err
		}
		updated, result, err := UpdateCRContent(string(data), decisions[id], date)
		if err != nil {
			return fmt.Errorf("[FAIL] %w; backup: %s", err, backupRoot)
		}
		if updated != string(data) {
			if err := os.WriteFile(crPath, []byte(updated), 0644); err != nil {
				return fmt.Errorf("[FAIL] 写入 CR 失败: %w; backup: %s", err, backupRoot)
			}
		}
		fmt.Printf("\n[UPDATE] %s\n", filepath.Base(crPath))
		if result.FrontFieldsUpdated {
			fmt.Println("  - front fields updated")
		}
		if result.ConclusionUpdated {
			fmt.Println("  - 第一轮筛选结论 updated")
		}
		var changed bool
		indexContent, changed, err = UpdateIndexContent(indexContent, id, decisions[id])
		if err != nil {
			return fmt.Errorf("[FAIL] %w; backup: %s", err, backupRoot)
		}
		if changed {
			fmt.Println("  - index entry planned")
		}
	}
	var queueChanged bool
	indexContent, queueChanged = UpdateQueueSectionWithItems(indexContent, CandidateRuleItems(paths, indexContent), allDecisions)
	if queueChanged {
		fmt.Println("  - current processing queue planned")
	}
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("[FAIL] 写入索引失败: %w; backup: %s", err, backupRoot)
	}
	fmt.Printf("\n[UPDATE] %s\n", filepath.Base(indexPath))
	fmt.Println("  - index entry updated")

	if err := ValidateApplied(paths, beforeMeta, decisions); err != nil {
		return fmt.Errorf("[FAIL] %w; backup: %s", err, backupRoot)
	}
	fmt.Println("\n[PASS] each CR has exactly one 第一轮筛选结论")
	fmt.Println("[PASS] index entries updated")
	fmt.Println("[PASS] validation_status not promoted unexpectedly")
	fmt.Println("[PASS] can_promote_formal not promoted unexpectedly")
	fmt.Println("[PASS] source_meta unchanged")
	return nil
}

func renderEmptyDecisions(index string) string {
	re := regexp.MustCompile(`CR-[A-Z]+-\d{8}-\d{3}`)
	seen := make(map[string]bool)
	for _, id := range re.FindAllString(index, -1) {
		seen[id] = true
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var sb strings.Builder
	for _, id := range ids {
		sb.WriteString(id + ":\n")
		sb.WriteString("  class: \n")
		sb.WriteString("  batch: \n")
		sb.WriteString("  queue: \n")
		sb.WriteString("  integration_status: \n")
		sb.WriteString("  formalization_status: \n")
		sb.WriteString("  position: \n")
		sb.WriteString("  action: \n")
		sb.WriteString("  merge_target: \n")
		sb.WriteString("  merge_watch: []\n")
		sb.WriteString("  formal_candidate: false\n")
		sb.WriteString("  formal_rule_suggestion: \n")
		sb.WriteString("  promote_blockers: []\n")
		sb.WriteString("  reasons: []\n")
		sb.WriteString("  improvements: []\n")
		sb.WriteString("  next_steps: []\n\n")
	}
	return sb.String()
}

func fail(err error) error {
	return fmt.Errorf("[FAIL] %w", err)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}
