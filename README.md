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

- CLI 命令：`kb.exe -input raw.txt -source 来源`
- 读取原始投资问答文本
- 调用 AI，要求 AI 返回结构化 JSON
- 根据 JSON 生成语义化编号
- 生成原始材料 RAW Markdown
- 生成知识卡片 QA Markdown
- 生成候选规则 CR Markdown
- 如果案例信息充足，生成 CASE Markdown
- 生成规则验证卡（每个候选规则一个独立验证文件）
- 跨文章相似规则检测
- 原文哈希去重
- 追加写入 Obsidian 指定文件
- 支持 `--dry-run`，只打印不写入
- 支持 `--mock`，不调用 AI，用内置 mock 数据
- 支持 `--allow-duplicate`，跳过哈希检查强制导入

---

## V1 不做

- ❌ 数据库
- ❌ Web 页面
- ❌ 账户状态判断器
- ❌ 规则执行器
- ❌ 正式规则确认
- ❌ 自动交易
- ❌ 行情数据抓取
- ❌ 规则自动合并（相似规则仅标记）

---

## 快速开始

```bash
# 1. 查看版本
kb.exe -v

# 2. Mock 模式测试（不调用 AI，快速验证流程）
kb.exe -input examples/raw_qa.txt -source 陈老师问答 -mock -dry-run

# 3. Mock 模式写入 Obsidian（不调用 AI，实际写入文件）
kb.exe -input examples/raw_qa.txt -source 陈老师问答 -mock

# 4. 真实 AI 调用（需要设置环境变量 ANTHROPIC_AUTH_TOKEN）
kb.exe -input examples/raw_qa.txt -source 陈老师问答

# 5. 强制重新导入（跳过哈希检查）
kb.exe -input examples/raw_qa.txt -source 陈老师问答 -allow-duplicate
```

**参数说明：**
- `-input`：输入文件路径（必需）
- `-source`：来源标识，如"陈老师问答"（必需）
- `-mock`：使用内置 Mock 数据，不调用 AI
- `-dry-run`：只打印 Markdown，不写入 Obsidian
- `-allow-duplicate`：允许重复导入（跳过哈希检查）
- `-config`：配置文件路径（默认：config.yaml）
- `-v`：显示版本号

**输出说明：**
- 程序会在 Obsidian 库中生成/更新以下文件：
  - `01-源文档/问答/原始材料库.md`：原始材料（RAW）
  - `02-观点/问答知识卡片库.md`：知识卡片（QA）
  - `03-规则/候选规则/候选规则库.md`：候选规则（CR）
  - `03-规则/规则回溯验证/规则验证卡/CR-*.md`：每个 CR 的独立验证卡
  - `03-规则/规则回溯验证/历史案例库/AI提取案例素材库.md`：市场案例（CASE，可选）

---

## 项目结构

```
investment-kb/
├── cmd/
│   └── kb/
│       └── main.go           # CLI 入口
├── internal/
│   ├── ai/                   # AI 调用、禁止表达检查
│   ├── app/                  # 业务编排（Extract 主流程 + 校验）
│   ├── classify/             # 程序领域分类映射
│   ├── config/               # 配置
│   ├── dedup/                # 跨文章相似规则检测
│   ├── idgen/                # 编号生成 + 领域映射
│   ├── markdown/             # Markdown 渲染（RAW/QA/CR/CASE/验证卡）
│   ├── model/                # 数据结构
│   ├── obsidian/             # Obsidian 写入
│   └── prompt/               # Prompt 加载
├── prompts/                  # AI Prompt 模板
├── data/                     # 数据文件（编号状态、哈希记录、错误输出）
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
obsidian_vault_path: "G:\\Obsidian\\我的知识库"

files:
  raw_material: "日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答/原始材料库.md"
  qa: "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片库.md"
  market_case: "日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/历史案例库/AI提取案例素材库.md"
  candidate_rule: "日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则/候选规则库.md"
  validation_card_template: "日常随笔/股市学习/宽基指数仓位管理系统/99-模板/规则验证卡模板.md"
  validation_card_dir: "日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/规则验证卡"

ai:
  provider: "custom"
  model: "glm-5.1"
  base_url: "https://api.z.ai/api/anthropic"
  api_key_env: "ANTHROPIC_AUTH_TOKEN"
  timeout_seconds: 300
  temperature: 0

timezone: "Asia/Beijing"
```

---

## 技术栈

- **语言**：Go 1.23
- **LLM 模型**：glm-5.1
- **存储**：Obsidian Markdown
- **开发工具**：VS Code + WorkBuddy

---

## 参考项目

本项目参考了 `article-pipeline` 中的以下模块（参考后重写，非直接复制）：

- `llm.go` → `internal/ai/`
- `export.go` → `internal/markdown/`
- `main.go` → `cmd/kb/main.go`

---

## 许可证

MIT