package main

import (
	"flag"
	"fmt"
	"os"

	"investment-kb/internal/screening"
)

func main() {
	kbRoot := flag.String("kb-root", "", "知识库根目录")
	decisions := flag.String("decisions", screening.DefaultDecisionsPath, "筛选决策 YAML 路径")
	id := flag.String("id", "", "只处理指定 CR")
	dryRun := flag.Bool("dry-run", false, "只打印计划，不写入")
	apply := flag.Bool("apply", false, "写入文件")
	init := flag.Bool("init", false, "根据候选规则索引生成空 decisions 文件")
	flag.Parse()

	opts := screening.Options{
		KBRoot:       *kbRoot,
		DecisionsRel: *decisions,
		ID:           *id,
		DryRun:       *dryRun,
		Apply:        *apply,
		Init:         *init,
	}
	if err := screening.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
