package screening

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Decision struct {
	Class                string   `yaml:"class"`
	Batch                string   `yaml:"batch"`
	Queue                string   `yaml:"queue"`
	IntegrationStatus    string   `yaml:"integration_status"`
	FormalizationStatus  string   `yaml:"formalization_status"`
	Position             string   `yaml:"position"`
	Action               string   `yaml:"action"`
	MergeTarget          string   `yaml:"merge_target"`
	MergeWatch           []string `yaml:"merge_watch"`
	FormalCandidate      bool     `yaml:"formal_candidate"`
	FormalRuleSuggestion string   `yaml:"formal_rule_suggestion"`
	PromoteBlockers      []string `yaml:"promote_blockers"`
	Reasons              []string `yaml:"reasons"`
	Improvements         []string `yaml:"improvements"`
	NextSteps            []string `yaml:"next_steps"`
	DiscardReason        string   `yaml:"discard_reason"`
}

func LoadDecisions(path string) (map[string]Decision, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 decisions 失败: %s: %w", path, err)
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("decisions 文件为空: %s", path)
	}
	var decisions map[string]Decision
	if err := yaml.Unmarshal(data, &decisions); err != nil {
		return nil, fmt.Errorf("解析 decisions 失败: %s: %w", path, err)
	}
	return decisions, nil
}

func SelectDecisions(all map[string]Decision, id string) (map[string]Decision, error) {
	if strings.TrimSpace(id) == "" {
		selected := make(map[string]Decision)
		for crID, decision := range all {
			if strings.TrimSpace(decision.Class) == "" {
				continue
			}
			selected[crID] = decision
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("decisions 中没有已填写 class 的筛选条目")
		}
		return selected, nil
	}
	decision, ok := all[id]
	if !ok {
		return nil, fmt.Errorf("decisions 中找不到指定 CR: %s", id)
	}
	return map[string]Decision{id: decision}, nil
}

func ValidateDecision(id string, d Decision) error {
	switch d.Class {
	case "A":
		if strings.TrimSpace(d.MergeTarget) != "" {
			return fmt.Errorf("%s: A 类不能填写 merge_target", id)
		}
	case "B":
	case "C":
		if strings.TrimSpace(d.MergeTarget) == "" {
			return fmt.Errorf("%s: C 类必须填写 merge_target", id)
		}
	case "D":
		if strings.TrimSpace(d.DiscardReason) == "" && len(nonEmptyList(d.Reasons)) == 0 {
			return fmt.Errorf("%s: D 类必须有 discard_reason 或 reasons", id)
		}
	default:
		return fmt.Errorf("%s: class 只能是 A/B/C/D", id)
	}
	if len(nonEmptyList(d.Reasons)) == 0 {
		return fmt.Errorf("%s: reasons 不能为空", id)
	}
	return nil
}

func ValidateDecisions(decisions map[string]Decision) error {
	ids := SortedIDs(decisions)
	for _, id := range ids {
		if err := ValidateDecision(id, decisions[id]); err != nil {
			return err
		}
	}
	return nil
}

func SortedIDs(decisions map[string]Decision) []string {
	ids := make([]string, 0, len(decisions))
	for id := range decisions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func ClassLabel(class string) string {
	switch class {
	case "A":
		return "A｜重点验证"
	case "B":
		return "B｜暂存观察"
	case "C":
		return "C｜合并到其他规则"
	case "D":
		return "D｜废弃"
	default:
		return class
	}
}

func QueueLabel(d Decision) string {
	if strings.TrimSpace(d.Queue) != "" {
		return strings.TrimSpace(d.Queue)
	}
	switch strings.TrimSpace(d.Class) {
	case "A":
		return "A｜待验证"
	case "B":
		return "B｜观察中"
	case "C":
		return "C｜待吸收"
	case "D":
		return "D｜已废弃"
	default:
		return "新增待筛选"
	}
}

func TopAction(d Decision) string {
	action := strings.TrimSpace(d.Action)
	position := strings.TrimSpace(d.Position)
	if d.Class == "C" || d.Class == "D" {
		return valueOrDefault(action, position)
	}
	if idx := strings.Index(action, "，"); idx > 0 {
		action = strings.TrimSpace(action[:idx])
	}
	if action == "" {
		return position
	}
	if position == "" {
		return action
	}
	return action + "，" + position
}

type FrontField struct {
	Key   string
	Value string
}

func ScreeningFrontFields(d Decision) []FrontField {
	fields := []FrontField{
		{Key: "第一轮筛选", Value: ClassLabel(d.Class)},
	}
	if strings.TrimSpace(d.Batch) != "" {
		fields = append(fields, FrontField{Key: "筛选批次", Value: strings.TrimSpace(d.Batch)})
	}
	fields = append(fields, FrontField{Key: "当前处理队列", Value: QueueLabel(d)})
	if strings.TrimSpace(TopAction(d)) != "" {
		fields = append(fields, FrontField{Key: "处理建议", Value: TopAction(d)})
	}
	if d.Class == "C" && strings.TrimSpace(d.MergeTarget) != "" {
		fields = append(fields, FrontField{Key: "合并去向", Value: strings.TrimSpace(d.MergeTarget)})
	}
	if text := LinkWatchText(d); text != "" {
		fields = append(fields, FrontField{Key: "联动观察", Value: text})
	}
	if strings.TrimSpace(d.IntegrationStatus) != "" {
		fields = append(fields, FrontField{Key: "整合状态", Value: strings.TrimSpace(d.IntegrationStatus)})
	}
	if strings.TrimSpace(d.FormalizationStatus) != "" {
		fields = append(fields, FrontField{Key: "正式化状态", Value: strings.TrimSpace(d.FormalizationStatus)})
	}
	return fields
}

func LinkWatchText(d Decision) string {
	watches := make([]string, 0, len(d.MergeWatch))
	for _, item := range d.MergeWatch {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		watches = append(watches, item)
	}
	if len(watches) == 0 {
		return ""
	}
	return "后续与 " + strings.Join(watches, "、") + " 联动验证"
}

func nonEmptyList(items []string) []string {
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
