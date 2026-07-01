# AI 输出校验和错误保存机制

> **当前状态**：本文档描述的校验体系已全部实现并随代码迭代演进。部分函数已从 `internal/app/extract.go` 迁移至 `internal/ai/custom.go`，且新增了否定语境放行（`checkAbsoluteClaimInText` + `negationMarkers`）和智能满仓检查（`ContainsForbiddenPhrases` 三级体系）。行号引用可能因代码演进略有偏移，以当前代码为准。

## 概述

为 investment-kb 增加了完善的 AI 输出校验机制，确保生成的数据符合业务规则和质量要求。

## 主要功能

### 1. JSON 解析错误保存

**问题**：AI 返回内容可能包含 UTF-8 BOM，导致 JSON 解析失败。

**解决方案**：
- 实现 `ExtractJSONFromAIOutput()` 函数 ([internal/ai/custom.go:190-233](G:/GoCode/investment-kb/internal/ai/custom.go#L190-L233))
- 自动去除 UTF-8/UTF-16 BOM
- 自动处理 markdown 代码块标记
- 提取第一个 JSON 对象
- 清洗后验证 JSON 格式

**错误输出文件**：`data/error_outputs/ai_error_YYYYMMDD_HHMMSS.txt`

**文件内容**：
```
执行时间: 2026-06-19 10:30:45
输入文件: examples/raw_qa.txt
Source: 陈老师问答
错误步骤: json.Unmarshal
错误原因: JSON 解析失败: invalid character 'ï' looking for beginning of value
---
AI 原始输出 (前 2000 字符):
﻿{"key": "value"}
```

### 2. CASE 校验

**函数**：`validateExtractionResult()` ([internal/app/extract.go:262-317](G:/GoCode/investment-kb/internal/app/extract.go#L262-L317))

**校验规则**：
- `ShouldGenerateCase=false` 时：
  - `Case` 必须为 `nil`
  - `CaseInsufficientReason` 必须不为空
- `ShouldGenerateCase=true` 时：
  - `Case` 必须不为 `nil`
  - `Case.CaseName` 不为空
  - `Case.DomainCode` 不为空
  - `Case.TopicCode` 不为空
  - `Case.KeyDecisionQuestion` 不为空
  - `Case.FinalInsight` 不为空

**测试**：`TestValidateExtractionResult_CaseValidation` ([internal/app/validation_test.go:13-104](G:/GoCode/investment-kb/internal/app/validation_test.go#L13-L104))

### 3. 绝对化收益表达检查

**函数**：`checkAbsoluteClaims()` ([internal/app/extract.go:319-355](G:/GoCode/investment-kb/internal/app/extract.go#L319-L355))

**检查关键词**：
- 保证盈利
- 没有亏损风险
- 必然上涨
- 一定赚钱
- 判断错了也不会亏
- 只赚不亏
- 无风险
- 稳赚
- 绝对安全
- 必胜

**检查范围**：
- `summary`
- `core_conclusion`
- `core_logic.title`
- `core_logic.content`
- `risk_boundaries`
- `extractable_rules.summary`
- `candidate_rules.rule_content`
- `candidate_rules.risk_boundary`
- `candidate_rules.actions`
- `candidate_rules.trigger_conditions`
- `candidate_rules.not_applicable`
- `my_understanding`

**错误示例**：
```
❌ AI 输出包含绝对化收益表达：没有亏损风险。请调整 Prompt 或人工检查后重试。
```

**测试**：`TestCheckAbsoluteClaims` ([internal/app/validation_test.go:106-128](G:/GoCode/investment-kb/internal/app/validation_test.go#L106-L128))

### 4. candidate_rules 类型集中 Warning

**函数**：`warnOnConsistentRuleTypes()` ([internal/app/extract.go:357-372](G:/GoCode/investment-kb/internal/app/extract.go#L357-L372))

**触发条件**：
- `candidate_rules` 数量 >= 3
- 所有 `rule_type` 相同
- 某种类型占比 >= 50%

**Warning 文案**：
```
⚠️  候选规则全部为同一类型，请检查是否遗漏仓位规则、风控规则或账户适配规则。
```

**注意**：这只是 warning，不会终止流程。

### 5. dry-run 模式行为

**修改**：在渲染 Markdown 之前执行完整校验 ([internal/app/extract.go:162-171](G:/GoCode/investment-kb/internal/app/extract.go#L162-L171))

**行为**：
- ✅ Dry-run 模式下也执行校验
- ✅ 校验失败时**不打印 Markdown**
- ✅ 校验失败时**不写入 Obsidian**
- ✅ 校验失败时**不更新** `id_state.json`
- ✅ 校验失败时**不更新** `import_hashes.json`

## 执行流程

```
1. 调用 AI (callAI)
   ↓
2. JSON 清洗与解析 (ExtractJSONFromAIOutput)
   ├─ 去除 BOM
   ├─ 处理 markdown 代码块
   └─ 提取 JSON 对象
   ↓
3. CASE 校验 (validateExtractionResult)
   ├─ ShouldGenerateCase 检查
   └─ Case 字段完整性检查
   ↓
4. 绝对化表达检查 (checkAbsoluteClaims)
   └─ 检查所有关键词
   ↓
5. candidate_rules 类型检查 (warnOnConsistentRuleTypes)
   └─ 打印 warning（不终止）
   ↓
6. 渲染 Markdown (只在校验通过后)
   ↓
7. 写入 Obsidian (只在 DryRun=false 时)
```

## 错误处理

### 校验失败时

1. **打印错误信息**：
   ```
   ❌ CASE 校验失败: ShouldGenerateCase=false 时，CaseInsufficientReason 不能为空
   ❌ 绝对化表达检查失败: AI 输出包含绝对化收益表达：保证盈利
   ```

2. **保存错误输出**：
   - 文件路径：`data/error_outputs/ai_error_YYYYMMDD_HHMMSS.txt`
   - 包含原始 AI 输出、清洗后输出、错误信息

3. **终止流程**：
   - 不渲染 Markdown
   - 不写入 Obsidian
   - 不更新状态文件

## 测试

### 单元测试

```bash
# 运行所有校验测试
go test ./internal/app -v -run "TestValidate"

# 运行 CASE 校验测试
go test ./internal/app -v -run "TestValidateExtractionResult_CaseValidation"

# 运行绝对化表达测试
go test ./internal/app -v -run "TestCheckAbsoluteClaims"
```

### 集成测试

```bash
# Mock 模式测试
kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"

# 真实 AI 模式测试
$env:AI_API_KEY = "your_key"
kb.exe -input examples/raw_qa.txt -source "陈老师问答" -dry-run
```

## 文件修改清单

### 新增文件

1. `internal/app/validation_test.go` - 校验测试文件
2. `EXECUTE_GUIDE.md` - 执行指南（文档）

### 修改文件

1. `internal/ai/custom.go`
   - 新增 `ExtractJSONFromAIOutput()` 函数
   - 修改 `saveErrorOutput()` 函数签名
   - 修改 `CompleteJSON()` 函数调用

2. `internal/app/extract.go`
   - 新增 `validateExtractionResult()` 函数
   - 新增 `checkAbsoluteClaims()` 函数
   - 新增 `warnOnConsistentRuleTypes()` 函数
   - 修改 `callAI()` 函数，增加校验调用
   - 修改 `Extract()` 函数，dry-run 模式先校验

## 向后兼容性

- ✅ RAW / QA / CR Markdown 模板未修改
- ✅ Obsidian WikiLink 未修改
- ✅ 编号规则未修改
- ✅ Mock 数据保持不变
- ✅ 现有功能不受影响

## 待优化项

1. **可配置性**：当前校验规则硬编码，未来可考虑从配置文件读取
2. **错误分类**：可以增加更细致的错误码，方便后续自动化处理
3. **预警阈值**：candidate_rules 类型集中的百分比可以配置化
