package main

import (
	"fmt"
	"investment-kb/internal/dedup"
)

func main() {
	crLibraryPath := "G:/Obsidian/我的知识库/日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则/候选规则库.md"
	existingCRs, err := dedup.ParseExistingCRs(crLibraryPath)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("解析到 %d 条已有 CR:\n", len(existingCRs))
	for i, fp := range existingCRs {
		fmt.Printf("%d. %s | %s | %s\n", i+1, fp.CRID, fp.ShortCode, fp.RuleName)
	}

	// 测试相似检查：用一条新规则检查是否与已有规则相似
	fmt.Println("\n--- 相似规则检查测试 ---")
	testRule := dedup.RuleFingerprint{
		CRID:       "CR-ACCOUNT-20260701-005",
		ShortCode:  "ACCOUNT-SAFETY",
		DomainCode: "ACCOUNT",
		TopicCode:  "SAFETY",
		RuleName:   "高概率区间建仓",
		Triggers:   []string{"安全边际足够", "估值进入合理区"},
		Actions:    []string{"先建底仓", "分批买入"},
	}

	similarRules := dedup.CheckSimilarRules(
		testRule.DomainCode, testRule.TopicCode, testRule.RuleName,
		testRule.Triggers, testRule.Actions,
		existingCRs,
	)
	fmt.Printf("测试规则: %s\n", testRule.RuleName)
	fmt.Printf("找到 %d 条相似规则:\n", len(similarRules))
	for i, sr := range similarRules {
		fmt.Printf("  %d. %s | %s | 原因: %s\n", i+1, sr.CRID, sr.RuleName, sr.Reason)
	}
}
