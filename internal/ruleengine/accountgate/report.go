package accountgate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ScenarioResult struct {
	Scenario Scenario
	Actual   AccountGateOutput
	Passed   bool
}

func RunScenarios(scenarios []Scenario, params Params) []ScenarioResult {
	results := make([]ScenarioResult, 0, len(scenarios))
	for _, scenario := range scenarios {
		actual := Evaluate(scenario.Input, params)
		results = append(results, ScenarioResult{
			Scenario: scenario,
			Actual:   actual,
			Passed:   outputEqual(actual, scenario.Expected),
		})
	}
	return results
}

func WriteMarkdownReport(path string, results []ScenarioResult, runAt time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(RenderMarkdownReport(results, runAt)), 0644)
}

func RenderMarkdownReport(results []ScenarioResult, runAt time.Time) string {
	passCount := 0
	for _, result := range results {
		if result.Passed {
			passCount++
		}
	}
	failCount := len(results) - passCount

	var sb strings.Builder
	sb.WriteString("# P0-Account-Probe｜账户状态规则运行报告\n\n")
	sb.WriteString("## 1. 运行信息\n\n")
	sb.WriteString(fmt.Sprintf("- 运行时间：%s\n", runAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("- 规则：FORMAL-ACCOUNT-001\n")
	sb.WriteString("- 模式：Mock\n")
	sb.WriteString(fmt.Sprintf("- 场景数量：%d\n\n", len(results)))

	sb.WriteString("## 2. 场景结果汇总\n\n")
	for _, result := range results {
		scenario := result.Scenario
		sb.WriteString(fmt.Sprintf("### %s｜%s\n\n", scenario.ID, scenario.Name))
		sb.WriteString("- 输入摘要：")
		sb.WriteString(fmt.Sprintf("action_type=%s, primary_position_state=%s, total_position_pct=%.2f, index_position_pct=%.2f\n",
			scenario.Input.ActionType,
			scenario.Input.PrimaryPositionState,
			scenario.Input.TotalPositionPct,
			scenario.Input.IndexPositionPct,
		))
		sb.WriteString(fmt.Sprintf("- decision：%s\n", result.Actual.Decision))
		sb.WriteString(fmt.Sprintf("- suggested_action：%s\n", result.Actual.SuggestedAction))
		sb.WriteString(fmt.Sprintf("- reason_codes：%s\n", joinReasons(result.Actual.ReasonCodes)))
		sb.WriteString(fmt.Sprintf("- 是否符合 expected：%s\n", passText(result.Passed)))
		if scenario.Note != "" {
			sb.WriteString(fmt.Sprintf("- 解释说明：%s\n", scenario.Note))
		} else {
			sb.WriteString("- 解释说明：按 P0 account_gate 优先级命中对应账户状态闸门。\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 3. 总结\n\n")
	sb.WriteString(fmt.Sprintf("- 通过数量：%d\n", passCount))
	sb.WriteString(fmt.Sprintf("- 失败数量：%d\n", failCount))
	if failCount == 0 {
		sb.WriteString("- 是否证明 account_gate 最小闭环可运行：是\n")
	} else {
		sb.WriteString("- 是否证明 account_gate 最小闭环可运行：否\n")
	}
	sb.WriteString("- 后续限制：本报告仅验证账户状态闸门 account_gate 的 P0 Mock 场景。不接真实行情。不接真实账户。不构成投资建议。不生成买卖信号。\n")
	return sb.String()
}

func outputEqual(a, b AccountGateOutput) bool {
	if a.Decision != b.Decision || a.SuggestedAction != b.SuggestedAction || a.HumanReviewRequired != b.HumanReviewRequired {
		return false
	}
	if len(a.ReasonCodes) != len(b.ReasonCodes) {
		return false
	}
	for i := range a.ReasonCodes {
		if a.ReasonCodes[i] != b.ReasonCodes[i] {
			return false
		}
	}
	return true
}

func joinReasons(reasons []ReasonCode) string {
	if len(reasons) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		parts = append(parts, string(reason))
	}
	return strings.Join(parts, ", ")
}

func passText(pass bool) string {
	if pass {
		return "是"
	}
	return "否"
}
