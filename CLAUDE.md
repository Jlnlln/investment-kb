# investment-kb 项目规则

> 本文件为 `investment-kb` 项目的特定规则，与全局 `~/.claude/CLAUDE.md` 共同生效。
> 与全局规则冲突时，以本文件为准。

---

## 项目概况

- **项目名称**：investment-kb
- **项目类型**：CLI 工具
- **Go 版本**：1.23
- **当前阶段**：V1 封板（功能完整，已通过清洁重跑验证）
- **核心目标**：原文 → AI 结构化 JSON → RAW / QA / CR / CASE → Obsidian + 验证卡
- **版本**：0.1.0

---

## 项目特定约定

### 技术栈

- **LLM 模型**：glm-5.1
- **AI 接口**：Anthropic Messages API 兼容格式（通过 config.yaml 配置）
- **存储**：暂不使用数据库，只写入 Obsidian Markdown 文件
- **输出格式**：Obsidian Markdown

### V1 范围

**只做：**
- CLI 命令：`kb.exe -input raw.txt -source 来源 [-config config.yaml] [-mock] [-dry-run] [-allow-duplicate]`
- 读取原始投资问答文本
- 调用 AI，要求 AI 返回结构化 JSON
- 根据 JSON 生成语义化编号
- 生成 RAW / QA / CR / CASE Markdown
- 追加写入 Obsidian 指定文件
- 生成规则验证卡（每个 CR 一个独立 .md 文件）
- AI 输出校验：禁止表达检查、绝对化收益检查、CASE 完整性校验
- 跨文章相似规则检测（ParseExistingCRs + CheckSimilarRules）
- 程序领域分类映射（classify，含 rule_name 硬覆盖表）
- 候选规则同篇去重
- 原文哈希去重（`import_hashes.json`）
- 支持 `--dry-run`，只打印不写入
- 支持 `--mock`，不调用 AI，用内置 mock 数据
- 支持 `--allow-duplicate`，跳过哈希检查强制导入
- Temperature 固定为 0，确保输出稳定

**不做：**
- 数据库
- Web 页面
- 账户状态判断器
- 规则执行器
- 正式规则确认
- 自动交易
- 行情数据抓取
- 规则自动合并（相似规则仅标记，不自动合并）

### 文档分层制度

V1 阶段：最小文档集
- 必须：docs/project-status.md（首次会话创建）
- 按需：docs/templates.md（已复制，用于文档格式参考）
- 技术文档：G:\Obsidian\我的知识库\日常随笔\股市学习\个人投资训练系统\98-想法\程序开发\V1版本\投资知识库自动整理工具 V1 技术文档.md

### 编号规则

采用：`类型-领域-日期-序号｜标题`

类型前缀：
- RAW：原始材料（`RAW-{Domain}-{Topic}-{YYYYMMDD}-{NNN}`）
- QA：知识卡片（`QA-{Domain}-{Topic}-{YYYYMMDD}-{NNN}`）
- CASE：市场案例（`CASE-{Domain}-{Topic}-{YYYYMMDD}-{NNN}`）
- CR：候选规则（`CR-{Domain}-{YYYYMMDD}-{NNN}`）

CR 领域映射：程序会将 `BUY`→`VALUATION`、`POS`→`ACCOUNT`、`ALLOC`→`REBALANCE`，另有 rule_name 硬覆盖表。

### 禁止表达体系

三层防护（代码 + AI Prompt）：
1. **AI Prompt 层**：Prompt 第 12 章明确列出 23 种禁止表达，禁止 AI 输出
2. **绝对化收益硬校验**（`checkAbsoluteClaims`）：含否定语境放行逻辑
3. **满仓/全仓智能检查**（`ContainsForbiddenPhrases`）：区分肯定/否定/警告三级，肯定表达硬失败

### AI Client 设计

定义通用接口（参考 article-pipeline 的 llm.go 思路）：

```go
type Client interface {
    Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
    CompleteJSON(ctx context.Context, systemPrompt string, userPrompt string, v any) error
}
```

V1 实现一个 `custom` client，用于调用 gml4.7 接口。

---

## 旧代码复用策略

### 复用（参考后重写，不直接复制）

| 旧项目模块 | 新项目位置 | 说明 |
|-----------|-----------|------|
| llm.go | internal/ai/ | AI 调用、JSON 解析、重试机制 |
| export.go | internal/markdown/ | Markdown 生成思路 |
| main.go | cmd/kb/main.go | CLI 参数解析思路 |
| dedup.go | internal/dedup/ | 跨文章相似规则检测（V1 中期启用） |
| classify.go | internal/classify/ | 程序领域二次分类（V1 中期启用） |

### 不复用（V1 不需要）

| 旧项目模块 | 不复用原因 |
|-----------|-----------|
| parser.go | V1 不做文章批量解析 |
| cluster.go | V1 不做聚类 |
| compile.go | V1 不做文章编译 |
| db.go | V1 不用数据库 |
| search.go | V1 不做全文搜索 |
| signal.go | V1 不做信号监控 |
| verify.go | V1 不做逻辑链验证 |

---

## 开发顺序

按技术文档中的步骤：

1. 创建项目结构和 go.mod ✅
2. 实现 model 结构体 ✅
3. 实现 mock ExtractionResult ✅
4. 实现 idgen ✅
5. 实现 Markdown renderer ✅
6. 实现 Obsidian writer ✅
7. 实现 CLI 和 app.Extract ✅
8. 跑通 --mock --dry-run ✅
9. 跑通 --mock 写入 Obsidian ✅
10. 接入 glm-5.1 AI Client ✅
11. 实现禁止表达校验体系 ✅
12. 实现领域分类映射 classify ✅
13. 实现跨文章相似规则检测 dedup ✅
14. 实现规则验证卡 ✅
15. 实现原文哈希去重 ✅
16. 清洁重跑验证通过 ✅

**V1 开发已完成，程序封板。**

---

## 小任务豁免

V1 开发属于全新项目创建，适用小任务豁免规则（全局规则 Section 1.3.4），无需 Issue 也可执行。