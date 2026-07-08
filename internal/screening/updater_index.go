package screening

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func UpdateIndexContent(content, id string, d Decision) (string, bool, error) {
	lines := strings.Split(content, "\n")
	start := -1
	searchEnd := len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == "## 当前处理队列" {
			searchEnd = i
			break
		}
	}
	for i, line := range lines[:searchEnd] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [[") && strings.Contains(trimmed, id) {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false, fmt.Errorf("候选规则索引中找不到条目: %s", id)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- [[") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			end = i
			break
		}
	}

	block := []string{lines[start]}
	insertAt := 1
	legacyKeys := screeningFrontFieldKeys()
	for _, line := range lines[start+1 : end] {
		key := metaKey(line)
		if legacyKeys[key] {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		block = append(block, normalizeIndexLine(line))
		if strings.HasPrefix(strings.TrimSpace(line), "- 状态：") || strings.HasPrefix(strings.TrimSpace(line), "- 状态:") {
			insertAt = len(block)
		}
	}

	var inserts []string
	for _, field := range ScreeningFrontFields(d) {
		inserts = append(inserts, "  - "+field.Key+"："+field.Value)
	}
	block = append(block[:insertAt], append(inserts, block[insertAt:]...)...)

	newLines := append([]string{}, lines[:start]...)
	newLines = append(newLines, block...)
	newLines = append(newLines, lines[end:]...)
	updated := strings.Join(newLines, "\n")
	return updated, updated != content, nil
}

func normalizeIndexLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "- ") {
		return "  " + trimmed
	}
	return line
}

type indexRuleItem struct {
	ID    string
	Link  string
	Title string
	Batch string
}

func UpdateQueueSection(content string, decisions map[string]Decision) (string, bool) {
	content = removeManagedIndexSections(content)
	items := parseIndexRuleItems(content)
	return UpdateQueueSectionWithItems(content, items, decisions)
}

func UpdateQueueSectionWithItems(content string, items []indexRuleItem, decisions map[string]Decision) (string, bool) {
	if len(items) == 0 {
		return content, false
	}
	section := renderManagedSections(items, decisions)
	updated := insertQueueSectionAfterDomain(content, section)
	updated = insertLifecycleBeforeAllRules(updated, renderLifecycleSection(len(items)))
	return updated, updated != content
}

func CandidateRuleItems(paths Paths, indexContent string) []indexRuleItem {
	byID := make(map[string]indexRuleItem)
	for _, item := range parseIndexRuleItems(removeManagedIndexSections(indexContent)) {
		byID[item.ID] = item
	}

	dir, err := paths.Resolve(CandidateRuleDir)
	if err != nil {
		return sortedIndexRuleItems(byID)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return sortedIndexRuleItems(byID)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "CR-") || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		title, batch := crFileMeta(filepath.Join(dir, entry.Name()), id)
		rel := filepath.ToSlash(filepath.Join(CandidateRuleDir, entry.Name()))
		rel = strings.TrimSuffix(rel, ".md")
		if existing, ok := byID[id]; ok {
			existing.Title = valueOrDefault(existing.Title, title)
			existing.Batch = batch
			byID[id] = existing
			continue
		}
		byID[id] = indexRuleItem{ID: id, Link: fmt.Sprintf("[[%s|%s]]", rel, title), Title: title, Batch: batch}
	}
	return sortedIndexRuleItems(byID)
}

func EnsureIndexEntries(paths Paths, content string) string {
	content = removeManagedIndexSections(content)
	existing := make(map[string]bool)
	for _, item := range parseIndexRuleItems(content) {
		existing[item.ID] = true
	}
	for _, item := range CandidateRuleItems(paths, content) {
		if existing[item.ID] {
			continue
		}
		content = insertDomainEntry(content, item)
		existing[item.ID] = true
	}
	return content
}

func insertDomainEntry(content string, item indexRuleItem) string {
	domain := domainFromCRID(item.ID)
	lines := strings.Split(content, "\n")
	domainHeading := "### " + domain
	domainStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == domainHeading {
			domainStart = i
			break
		}
	}
	block := []string{
		"- " + item.Link,
		"  - 状态：候选",
		"  - 验证状态：待验证",
	}
	if domainStart < 0 {
		insertAt := findAfterDomainHeading(lines)
		newBlock := []string{domainHeading, ""}
		newBlock = append(newBlock, block...)
		newBlock = append(newBlock, "")
		return insertLines(lines, insertAt, newBlock)
	}
	insertAt := len(lines)
	for i := domainStart + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "### ") || strings.HasPrefix(trimmed, "## ") || trimmed == "---" {
			insertAt = i
			break
		}
	}
	newBlock := append([]string{}, block...)
	newBlock = append(newBlock, "")
	return insertLines(lines, insertAt, newBlock)
}

func findAfterDomainHeading(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == "## 按领域" {
			return i + 1
		}
	}
	return len(lines)
}

func insertLines(lines []string, at int, block []string) string {
	out := append([]string{}, lines[:at]...)
	out = append(out, block...)
	out = append(out, lines[at:]...)
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func domainFromCRID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "UNKNOWN"
}

func renderQueueSection(items []indexRuleItem, decisions map[string]Decision) string {
	queues := []string{"A｜待验证", "B｜观察中", "C｜待吸收", "D｜已废弃", "新增待筛选"}
	grouped := make(map[string][]indexRuleItem)
	for _, item := range items {
		decision, ok := decisions[item.ID]
		queue := "新增待筛选"
		if ok {
			queue = QueueLabel(decision)
		}
		grouped[queue] = append(grouped[queue], item)
	}

	var sb strings.Builder
	sb.WriteString("---\n\n")
	sb.WriteString("## 当前处理队列\n\n")
	for _, queue := range queues {
		sb.WriteString("### " + queue + "\n\n")
		queueItems := grouped[queue]
		sort.Slice(queueItems, func(i, j int) bool { return queueItems[i].ID < queueItems[j].ID })
		if len(queueItems) == 0 {
			sb.WriteString("暂无\n\n")
			continue
		}
		for _, item := range queueItems {
			decision, ok := decisions[item.ID]
			if ok && decision.Class == "C" && strings.TrimSpace(decision.MergeTarget) != "" {
				sb.WriteString(fmt.Sprintf("- %s → %s\n", item.Link, strings.TrimSpace(decision.MergeTarget)))
				continue
			}
			sb.WriteString("- " + item.Link + "\n")
			if ok && (decision.Class == "A" || decision.Class == "B") {
				if watch := LinkWatchText(decision); watch != "" {
					sb.WriteString("  - 联动观察：" + watch + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n") + "\n\n"
}

func renderBatchSection(items []indexRuleItem, decisions map[string]Decision) string {
	groups := make(map[string][]indexRuleItem)
	for _, item := range items {
		batch := batchLabel(item, decisions[item.ID])
		groups[batch] = append(groups[batch], item)
	}
	batches := make([]string, 0, len(groups))
	for batch := range groups {
		batches = append(batches, batch)
	}
	sort.Strings(batches)

	var sb strings.Builder
	sb.WriteString("---\n\n")
	sb.WriteString("## 按批次\n\n")
	for _, batch := range batches {
		batchItems := groups[batch]
		sort.Slice(batchItems, func(i, j int) bool { return batchItems[i].ID < batchItems[j].ID })
		stats := batchStats(batchItems, decisions)
		sb.WriteString("### " + batch + "\n\n")
		sb.WriteString(fmt.Sprintf("- 总数：%d\n", len(batchItems)))
		sb.WriteString(fmt.Sprintf("- A｜重点验证：%d\n", stats["A"]))
		sb.WriteString(fmt.Sprintf("- B｜暂存观察：%d\n", stats["B"]))
		sb.WriteString(fmt.Sprintf("- C｜合并到其他规则：%d\n", stats["C"]))
		sb.WriteString(fmt.Sprintf("- D｜废弃：%d\n", stats["D"]))
		sb.WriteString(fmt.Sprintf("- 新增待筛选：%d\n", stats["NEW"]))
		if stats["NEW"] > 0 {
			sb.WriteString("- 批次状态：存在新增待筛选\n\n")
		} else {
			sb.WriteString("- 批次状态：第一轮筛选已完成\n\n")
		}
		for _, item := range batchItems {
			decision, ok := decisions[item.ID]
			sb.WriteString("- " + item.Link + "\n")
			if ok {
				sb.WriteString("  - 第一轮筛选：" + ClassLabel(decision.Class) + "\n")
				sb.WriteString("  - 当前处理队列：" + QueueLabel(decision) + "\n")
				if decision.Class == "C" && strings.TrimSpace(decision.MergeTarget) != "" {
					sb.WriteString("  - 合并去向：" + strings.TrimSpace(decision.MergeTarget) + "\n")
				}
			} else {
				sb.WriteString("  - 第一轮筛选：未筛选\n")
				sb.WriteString("  - 当前处理队列：新增待筛选\n")
			}
		}
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n") + "\n\n"
}

func renderLifecycleSection(total int) string {
	var sb strings.Builder
	sb.WriteString("---\n\n")
	sb.WriteString("## 生命周期状态\n\n")
	sb.WriteString("### 候选\n\n")
	sb.WriteString(fmt.Sprintf("- 数量：%d\n", total))
	sb.WriteString("- 说明：当前全部 CR 仍处于候选生命周期，具体处理进度以“当前处理队列”为准。\n\n")
	sb.WriteString("### 正式草案\n\n")
	sb.WriteString("暂无\n\n")
	sb.WriteString("### 正式规则\n\n")
	sb.WriteString("暂无\n\n")
	sb.WriteString("### 已废弃\n\n")
	sb.WriteString("暂无\n\n")
	return sb.String()
}

func batchStats(items []indexRuleItem, decisions map[string]Decision) map[string]int {
	stats := map[string]int{"A": 0, "B": 0, "C": 0, "D": 0, "NEW": 0}
	for _, item := range items {
		decision, ok := decisions[item.ID]
		if !ok || strings.TrimSpace(decision.Class) == "" {
			stats["NEW"]++
			continue
		}
		if _, exists := stats[decision.Class]; exists {
			stats[decision.Class]++
		} else {
			stats["NEW"]++
		}
	}
	return stats
}

func batchLabel(item indexRuleItem, decision Decision) string {
	if strings.TrimSpace(decision.Batch) != "" {
		return strings.TrimSpace(decision.Batch)
	}
	if strings.TrimSpace(item.Batch) != "" {
		return strings.TrimSpace(item.Batch)
	}
	return "未分批"
}

func parseIndexRuleItems(content string) []indexRuleItem {
	re := regexp.MustCompile(`CR-[A-Z]+-\d{8}-\d{3}`)
	seen := make(map[string]bool)
	var items []indexRuleItem
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- [[") {
			continue
		}
		id := re.FindString(trimmed)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		link := strings.TrimPrefix(trimmed, "- ")
		items = append(items, indexRuleItem{ID: id, Link: link, Title: titleFromLink(link, id)})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func sortedIndexRuleItems(byID map[string]indexRuleItem) []indexRuleItem {
	items := make([]indexRuleItem, 0, len(byID))
	for _, item := range byID {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func crFileMeta(path, fallbackID string) (string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallbackID, ""
	}
	title := fallbackID
	batch := ""
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# CR-") {
			title = strings.TrimPrefix(trimmed, "# ")
		}
		if strings.HasPrefix(trimmed, "筛选批次：") || strings.HasPrefix(trimmed, "筛选批次:") {
			batch = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "筛选批次："), "筛选批次:"))
		}
	}
	return title, batch
}

func titleFromLink(link, fallbackID string) string {
	if idx := strings.LastIndex(link, "|"); idx >= 0 {
		title := strings.TrimSuffix(strings.TrimSpace(link[idx+1:]), "]]")
		if title != "" {
			return title
		}
	}
	return fallbackID
}

func removeQueueSection(content string) string {
	return removeSection(content, "## 当前处理队列")
}

func removeManagedIndexSections(content string) string {
	for _, heading := range []string{"## 当前处理队列", "## 按批次", "## 生命周期状态", "## 按状态"} {
		content = removeSection(content, heading)
	}
	return content
}

func removeSection(content, heading string) string {
	for {
		lines := strings.Split(content, "\n")
		start := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == heading {
				start = i
				if start >= 2 && strings.TrimSpace(lines[start-1]) == "" && strings.TrimSpace(lines[start-2]) == "---" {
					start -= 2
				}
				break
			}
		}
		if start < 0 {
			return content
		}
		end := len(lines)
		for i := start + 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "## ") && trimmed != heading {
				end = i
				if end >= 2 && strings.TrimSpace(lines[end-1]) == "" && strings.TrimSpace(lines[end-2]) == "---" {
					end -= 2
				}
				break
			}
		}
		newLines := append([]string{}, lines[:start]...)
		newLines = append(newLines, lines[end:]...)
		content = strings.TrimRight(strings.Join(newLines, "\n"), "\n") + "\n"
	}
}

func insertQueueSectionAfterDomain(content, section string) string {
	return insertManagedSections(content, section)
}

func insertManagedSections(content, queueSection string) string {
	lines := strings.Split(content, "\n")
	insertAt := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## 按领域" {
			insertAt = i
			break
		}
	}
	if insertAt < 0 {
		return strings.TrimRight(content, "\n") + "\n\n" + queueSection
	}
	prefix := strings.TrimRight(strings.Join(lines[:insertAt], "\n"), "\n")
	suffix := strings.TrimLeft(strings.Join(lines[insertAt:], "\n"), "\n")
	if suffix == "" {
		return prefix + "\n\n" + queueSection
	}
	return prefix + "\n\n" + queueSection + suffix
}

func renderManagedSections(items []indexRuleItem, decisions map[string]Decision) string {
	var sb strings.Builder
	sb.WriteString(renderQueueSection(items, decisions))
	sb.WriteString(renderBatchSection(items, decisions))
	return sb.String()
}

func insertLifecycleBeforeAllRules(content string, section string) string {
	lines := strings.Split(content, "\n")
	insertAt := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## 全部候选规则" {
			insertAt = i
			break
		}
	}
	if insertAt < 0 {
		return strings.TrimRight(content, "\n") + "\n\n" + section
	}
	if insertAt >= 2 && strings.TrimSpace(lines[insertAt-1]) == "" && strings.TrimSpace(lines[insertAt-2]) == "---" {
		insertAt -= 2
	}
	prefix := strings.TrimRight(strings.Join(lines[:insertAt], "\n"), "\n")
	suffix := strings.TrimLeft(strings.Join(lines[insertAt:], "\n"), "\n")
	return prefix + "\n\n" + section + suffix
}
