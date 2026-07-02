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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "validate":
			runValidate(os.Args[2:])
			return
		case "extract":
			runExtract(os.Args[2:])
			return
		}
	}

	runExtract(os.Args[1:])
}

func runValidate(args []string) {
	validateFlags := flag.NewFlagSet("validate", flag.ExitOnError)
	configPath := validateFlags.String("config", "config.yaml", "配置文件路径")
	_ = validateFlags.Parse(args)
	if err := app.Validate(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		os.Exit(1)
	}
}

func runExtract(args []string) {
	extractFlags := flag.NewFlagSet("extract", flag.ExitOnError)
	versionFlag := extractFlags.Bool("v", false, "显示版本号")
	inputPath := extractFlags.String("input", "", "输入文件路径")
	source := extractFlags.String("source", "", "来源（如：陈老师问答）")
	dryRun := extractFlags.Bool("dry-run", false, "只打印 Markdown，不写入 Obsidian")
	mock := extractFlags.Bool("mock", false, "使用 Mock 数据，不调用 AI")
	mockIndex := extractFlags.Int("mock-index", 1, "Mock 数据变体编号（仅 --mock 模式有效，默认 1）")
	forceType := extractFlags.String("force-type", "", "强制指定材料类型（rule_candidate/macro_knowledge/market_observation/archive_only），跳过 AI 判断")
	allowDuplicate := extractFlags.Bool("allow-duplicate", false, "允许重复导入（默认同一 hash 禁止重复写入）")
	validateFlag := extractFlags.Bool("validate", false, "运行输出验收检查")
	configPath := extractFlags.String("config", "config.yaml", "配置文件路径")
	_ = extractFlags.Parse(args)

	if *versionFlag {
		fmt.Printf("investment-kb v%s\n", version)
		os.Exit(0)
	}

	if *validateFlag {
		if err := app.Validate(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
			os.Exit(1)
		}
		return
	}

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

	opts := &app.ExtractOptions{
		InputPath:      *inputPath,
		Source:         *source,
		DryRun:         *dryRun,
		Mock:           *mock,
		MockIndex:      *mockIndex,
		ForceType:      *forceType,
		AllowDuplicate: *allowDuplicate,
		ConfigPath:     *configPath,
	}

	if err := app.Extract(opts); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		os.Exit(1)
	}
}
