# investment-kb

投资知识库自动整理 CLI。当前阶段：**模块一 V1.3**。

核心流程：

    原文 -> AI/Mock 结构化 JSON -> material_type 分流 -> Obsidian Markdown -> validate 验收

## 当前支持

### rule_candidate

输出：RAW + QA + CR + 规则验证卡。

用于包含明确触发条件、执行动作、禁止条件、仓位约束的规则型材料。

### macro_knowledge

输出：RAW + KNOW 单文件。

用于解释宏观、利率、政策、通胀、经济周期等运行逻辑的材料。macro_knowledge 不生成 QA / CR / 验证卡。

### market_observation

暂缓。

### archive_only

输出：RAW。

## 当前不做

- 不做模块二
- 不做 P0-Lite
- 不做正式规则 R
- 不做自动转正式
- 不做自动交易
- 不做行情数据抓取
- 不做 Web 页面
- 不做数据库
- 不把 macro_knowledge 强行生成 QA / CR
- 不改 CR 编号规则

## 常用命令

    # 查看版本
    .\kb.exe -v

    # rule_candidate mock 导入预览
    .\kb.exe -input testdata/inputs/rule_safety_margin.md -source 陈老师问答 -mock -dry-run

    # macro_knowledge mock 导入预览
    .\kb.exe -input testdata/inputs/know_rate.md -source 陈老师问答 -mock -force-type macro_knowledge -dry-run

    # 第二个 macro_knowledge mock
    .\kb.exe -input testdata/inputs/know_revenue_income.md -source 陈老师问答 -mock -force-type macro_knowledge -mock-index 2 -dry-run

    # 验收检查
    .\kb.exe validate -config config.yaml

    # 兼容形式
    .\kb.exe -validate -config config.yaml

    # 回归脚本
    .\scripts\run_regression.ps1

## 工程验收能力

V1.3 新增重点：

- mock-index 与 testdata 输入绑定，防止 mock result 与输入正文错配。
- RAW 标题/正文一致性校验，非 dry-run 写入前失败即停止。
- 所有主要输出写入 source metadata：source_file、raw_hash、cleaned_hash、raw_id、material_type。
- validate 命令检查 RAW / QA / KNOW / CR / 验证卡数量、一致性、孤立验证卡、重复 hash、旧宏观理解卡库文件和明显标题错配。
- scripts/run_regression.ps1 支持清洁回归。

## 配置

默认读取 config.yaml。回归脚本会生成测试专用配置，将输出写入 testdata/output/vault，不污染真实 Obsidian 库。

## 技术栈

- Go 1.23
- glm-5.1
- Anthropic Messages API 兼容格式
- Obsidian Markdown
