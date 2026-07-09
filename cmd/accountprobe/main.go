package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"investment-kb/internal/ruleengine/accountgate"
)

func main() {
	scenarioPath := flag.String("scenario", "testdata/account_probe/scenarios.yaml", "scenario yaml path")
	reportPath := flag.String("report", "05-运行报告/P0-Account-Probe/账户状态规则运行报告-20260707.md", "markdown report path")
	flag.Parse()

	scenarios, err := accountgate.LoadScenarios(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenarios failed: %v\n", err)
		os.Exit(1)
	}
	results := accountgate.RunScenarios(scenarios, accountgate.DefaultParams())
	if err := accountgate.WriteMarkdownReport(*reportPath, results, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "write report failed: %v\n", err)
		os.Exit(1)
	}

	failCount := 0
	for _, result := range results {
		if !result.Passed {
			failCount++
		}
	}
	fmt.Printf("accountprobe completed: scenarios=%d pass=%d fail=%d report=%s\n", len(results), len(results)-failCount, failCount, *reportPath)
	if failCount > 0 {
		os.Exit(1)
	}
}
