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
		return all, nil
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

func TopAction(d Decision) string {
	action := strings.TrimSpace(d.Action)
	position := strings.TrimSpace(d.Position)
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

func MergeObservation(d Decision) string {
	watchIDs := make([]string, 0, len(d.MergeWatch))
	for _, item := range d.MergeWatch {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if idx := strings.Index(item, "｜"); idx >= 0 {
			item = item[:idx]
		}
		watchIDs = append(watchIDs, item)
	}
	if len(watchIDs) > 0 {
		return "后续可能吸收 " + strings.Join(watchIDs, "、")
	}
	if strings.TrimSpace(d.MergeTarget) != "" {
		return "合并至 " + strings.TrimSpace(d.MergeTarget)
	}
	return "暂不合并"
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
