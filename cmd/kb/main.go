package main

import (
	"flag"
	"fmt"
	"os"

	"investment-kb/internal/app"
)

var (
	version = "0.1.0"
)

func main() {
	versionFlag := flag.Bool("v", false, "显示版本号")
	inputPath := flag.String("input", "", "输入文件路径")
	source := flag.String("source", "", "来源（如：陈老师问答）")
	dryRun := flag.Bool("dry-run", false, "只打印 Markdown，不写入 Obsidian")
	mock := flag.Bool("mock", false, "使用 Mock 数据，不调用 AI")
	allowDuplicate := flag.Bool("allow-duplicate", false, "允许重复导入（默认同一 hash 禁止重复写入）")
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("investment-kb v%s\n", version)
		os.Exit(0)
	}

	// 验证必填参数
	if *inputPath == "" {
		fmt.Fprintf(os.Stderr, "❌ 缺少必填参数：--input\n\n")
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  kb extract --input examples/raw_qa.txt --source 陈老师问答\n")
		fmt.Fprintf(os.Stderr, "  kb extract --input examples/raw_qa.txt --source 陈老师问答 --mock --dry-run\n")
		os.Exit(1)
	}

	if *source == "" {
		fmt.Fprintf(os.Stderr, "❌ 缺少必填参数：--source\n\n")
		os.Exit(1)
	}

	// 执行提取
	opts := &app.ExtractOptions{
		InputPath:      *inputPath,
		Source:         *source,
		DryRun:         *dryRun,
		Mock:           *mock,
		AllowDuplicate: *allowDuplicate,
		ConfigPath:     *configPath,
	}

	if err := app.Extract(opts); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		os.Exit(1)
	}
}