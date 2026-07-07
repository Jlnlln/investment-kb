package screening

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Plan struct {
	Paths     Paths
	Decisions map[string]Decision
	Date      string
}

func ValidateInputs(paths Paths, decisions map[string]Decision) error {
	if err := ValidateDecisions(decisions); err != nil {
		return err
	}
	indexPath, err := paths.IndexPath()
	if err != nil {
		return err
	}
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("候选规则索引不存在或不可读: %s: %w", indexPath, err)
	}
	index := string(indexData)
	for _, id := range SortedIDs(decisions) {
		crPath, err := paths.CRPath(id)
		if err != nil {
			return err
		}
		if _, err := os.Stat(crPath); err != nil {
			return fmt.Errorf("CR 文件不存在: %s: %w", crPath, err)
		}
		if !strings.Contains(index, id) {
			return fmt.Errorf("候选规则索引中找不到 CR: %s", id)
		}
	}
	return nil
}

func ValidateApplied(paths Paths, beforeMeta map[string]string, decisions map[string]Decision) error {
	indexPath, err := paths.IndexPath()
	if err != nil {
		return err
	}
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	index := string(indexData)
	for _, id := range SortedIDs(decisions) {
		crPath, err := paths.CRPath(id)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(crPath)
		if err != nil {
			return err
		}
		content := string(data)
		if countHeading(content, "## 第一轮筛选结论") != 1 {
			return fmt.Errorf("%s: 第一轮筛选结论数量不是 1", id)
		}
		if strings.Contains(content, "验证状态：已验证") || strings.Contains(content, "验证状态: 已验证") {
			return fmt.Errorf("%s: 验证状态被错误提升为已验证", id)
		}
		if strings.Contains(content, "是否可转正式：是") || strings.Contains(content, "是否可转正式: 是") {
			return fmt.Errorf("%s: 是否可转正式被错误改为是", id)
		}
		if beforeMeta[id] != ExtractSourceMeta(content) {
			return fmt.Errorf("%s: source_meta 内容发生变化", id)
		}
		if !strings.Contains(index, "第一轮筛选："+ClassLabel(decisions[id].Class)) {
			return fmt.Errorf("%s: 索引未写入第一轮筛选", id)
		}
		if !strings.Contains(index, id) {
			return fmt.Errorf("%s: 索引缺少 CR 条目", id)
		}
	}
	return nil
}

func countHeading(content, heading string) int {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(heading) + `\s*$`)
	return len(re.FindAllStringIndex(content, -1))
}
