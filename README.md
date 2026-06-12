# investment-kb

> 投资知识库自动整理工具 V1

---

## 项目简介

把原始投资问答/投资观点/市场材料，自动整理成 Obsidian Markdown 内容。

**核心流程：**

```
原文 → AI 结构化 JSON → RAW / QA / CR Markdown → Obsidian 追加写入
```

---

## V1 功能

- CLI 命令：`kb extract --input raw.txt --source 来源`
- 读取原始投资问答文本
- 调用 AI，要求 AI 返回结构化 JSON
- 根据 JSON 生成语义化编号
- 生成原始材料 RAW Markdown
- 生成知识卡片 QA Markdown
- 生成候选规则 CR Markdown
- 如果案例信息充足，生成 CASE Markdown
- 追加写入 Obsidian 指定文件
- 支持 `--dry-run`，只打印不写入
- 支持 `--mock`，不调用 AI，用内置 mock 数据

---

## V1 不做

- ❌ 数据库
- ❌ Web 页面
- ❌ 账户状态判断器
- ❌ 规则执行器
- ❌ 正式规则确认
- ❌ 自动交易
- ❌ 行情数据抓取

---

## 快速开始

```bash
# Mock 模式 + Dry-run（最快验证流程）
kb extract --input examples/raw_qa.txt --source 陈老师问答 --mock --dry-run

# Mock 模式 + 写入 Obsidian
kb extract --input examples/raw_qa.txt --source 陈老师问答 --mock

# 真实 AI 调用
kb extract --input examples/raw_qa.txt --source 陈老师问答
```

---

## 项目结构

```
investment-kb/
├── cmd/
│   └── kb/
│       └── main.go           # CLI 入口
├── internal/
│   ├── ai/                   # AI 调用
│   ├── app/                  # 业务编排
│   ├── config/               # 配置
│   ├── idgen/                # 编号生成
│   ├── markdown/             # Markdown 渲染
│   ├── model/                # 数据结构
│   ├── obsidian/             # Obsidian 写入
│   └── prompt/               # Prompt 加载
├── prompts/                  # AI Prompt 模板
├── data/                     # 数据文件（编号状态、错误输出）
├── examples/                 # 示例输入
├── docs/                     # 文档
├── config.yaml               # 配置文件
├── go.mod
├── CLAUDE.md                 # 项目规则
└── README.md
```

---

## 配置文件

`config.yaml` 示例：

```yaml
obsidian_vault_path: "G:\\Obsidian\\个人投资训练系统"

files:
  raw_material: "03-知识与案例/原始材料库.md"
  qa: "03-知识与案例/问答知识库.md"
  market_case: "03-知识与案例/市场案例库.md"
  candidate_rule: "04-投资规则/候选规则.md"

ai:
  provider: "custom"
  model: "gml4.7"
  base_url: "你的模型接口地址"
  api_key_env: "AI_API_KEY"
  timeout_seconds: 120

timezone: "Asia/Shanghai"
```

---

## 技术栈

- **语言**：Go 1.23
- **LLM 模型**：gml4.7
- **存储**：Obsidian Markdown
- **开发工具**：VS Code + Claude Code

---

## 参考项目

本项目参考了 `article-pipeline` 中的以下模块（参考后重写，非直接复制）：

- `llm.go` → `internal/ai/`
- `export.go` → `internal/markdown/`
- `main.go` → `cmd/kb/main.go`

---

## 许可证

MIT