# investment-kb 项目规则

> 本文件为 investment-kb 项目的特定规则，与全局规则共同生效。
> 与全局规则冲突时，以本文件为准。

---

## 项目概况

- **项目名称**：investment-kb
- **项目类型**：CLI 工具
- **Go 版本**：1.23
- **当前阶段**：模块一 V1.3
- **核心目标**：原文 -> AI 结构化 JSON -> 按 material_type 分流 -> Obsidian Markdown + 验收检查
- **版本**：0.1.0

---

## 当前支持的 material_type 分流

1. **rule_candidate**
   - 输出：RAW + QA + CR + 规则验证卡
   - 用于可沉淀为候选规则的材料。

2. **macro_knowledge**
   - 输出：RAW + KNOW 单文件
   - 不生成 QA / CR / 验证卡。
   - KNOW 卡采用单文件模式，并维护宏观理解卡索引。

3. **market_observation**
   - 暂缓。
   - 当前不作为继续扩展重点。

4. **archive_only**
   - 输出：RAW。

---

## 当前重点

模块一 V1.3 的重点不是继续扩功能，而是补工程验收能力：

- mock 输入与 mock result 必须绑定，避免 RAW 正文与 QA/CR/KNOW 语义错配。
- RAW 标题与正文必须做一致性校验，非 dry-run 写入前必须通过。
- 所有输出对象必须携带 source metadata：source_file、raw_hash、cleaned_hash、raw_id、material_type。
- 提供 validate 命令检查输出库的一致性。
- 提供回归脚本完成清洁重跑验证。

---

## 明确不做

- 不做模块二。
- 不做 P0-Lite。
- 不做正式规则 R。
- 不做自动转正式规则。
- 不做自动交易。
- 不做行情数据抓取。
- 不做 Web 页面。
- 不做数据库。
- 不把 macro_knowledge 强行生成 QA / CR。
- 不自动合并相似规则。
- 不改 CR 编号规则。

---

## 技术栈与接口

- **LLM 模型**：glm-5.1
- **AI 接口**：Anthropic Messages API 兼容格式（通过 config.yaml 配置）
- **存储**：暂不使用数据库，只写入 Obsidian Markdown 文件
- **输出格式**：Obsidian Markdown
- **Temperature**：固定为 0，确保输出稳定

AI Client 通用接口：

    type Client interface {
        Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
        CompleteJSON(ctx context.Context, systemPrompt string, userPrompt string, v any) error
    }

---

## 编号规则

采用：类型-领域-日期-序号｜标题

类型前缀：

- RAW：原始材料，RAW-{Domain}-{Topic}-{YYYYMMDD}-{NNN}
- QA：知识卡片，QA-{Domain}-{Topic}-{YYYYMMDD}-{NNN}
- KNOW：宏观理解卡，KNOW-{Layer}-{Topic}-{YYYYMMDD}-{NNN}
- CASE：市场案例，CASE-{Domain}-{Topic}-{YYYYMMDD}-{NNN}
- CR：候选规则，CR-{Domain}-{YYYYMMDD}-{NNN}

CR 领域映射：程序会将 BUY -> VALUATION、POS -> ACCOUNT、ALLOC -> REBALANCE，另有 rule_name 硬覆盖表。

---

## 验收命令

- 导入：kb.exe -input raw.txt -source 来源 [-config config.yaml] [-mock] [-dry-run] [-allow-duplicate] [-force-type macro_knowledge]
- 验收：kb.exe validate -config config.yaml
- 兼容验收：kb.exe -validate -config config.yaml
- 回归：scripts/run_regression.ps1
