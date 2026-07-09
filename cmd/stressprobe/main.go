package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"investment-kb/internal/ruleengine/stressgate"
)

func main() {
	scenarioPath := flag.String("scenario", "testdata/stress_probe/scenarios.yaml", "scenario yaml path")
	reportPath := flag.String("report", "05-运行报告/P0-Stress-Probe/极限下跌压力测试运行报告-20260709.md", "markdown report path")
	flag.Parse()

	scenarios, err := stressgate.LoadScenarios(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenarios failed: %v\n", err)
		os.Exit(1)
	}
	results := stressgate.RunScenarios(scenarios, stressgate.DefaultParams())
	if err := stressgate.WriteMarkdownReport(*reportPath, results, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "write report failed: %v\n", err)
		os.Exit(1)
	}

	failCount := 0
	for _, result := range results {
		if !result.Passed {
			failCount++
		}
	}
	fmt.Printf("stressprobe completed: scenarios=%d pass=%d fail=%d report=%s\n", len(results), len(results)-failCount, failCount, *reportPath)
	if failCount > 0 {
		os.Exit(1)
	}
}
