package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"investment-kb/internal/ruleengine/plangate"
)

func main() {
	scenarioPath := flag.String("scenario", "testdata/plan_probe/scenarios.yaml", "scenario yaml path")
	reportPath := flag.String("report", "05-运行报告/P0-Plan-Probe/预案检查规则运行报告-20260709.md", "markdown report path")
	flag.Parse()

	scenarios, err := plangate.LoadScenarios(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenarios failed: %v\n", err)
		os.Exit(1)
	}
	results := plangate.RunScenarios(scenarios, plangate.DefaultParams())
	if err := plangate.WriteMarkdownReport(*reportPath, results, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "write report failed: %v\n", err)
		os.Exit(1)
	}

	failCount := 0
	for _, result := range results {
		if !result.Passed {
			failCount++
		}
	}
	fmt.Printf("planprobe completed: scenarios=%d pass=%d fail=%d report=%s\n", len(results), len(results)-failCount, failCount, *reportPath)
	if failCount > 0 {
		os.Exit(1)
	}
}
