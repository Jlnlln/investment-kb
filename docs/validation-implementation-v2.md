# AI 输出校验和错误保存机制（v2 更新）

## 概述

为 investment-kb 增加了完善的 AI 输出校验机制，确保生成的数据符合业务规则和质量要求。

## 更新内容

### 新增检查功能

#### 1. 满仓误导表达检查（硬性校验）

**函数**：`ContainsForbiddenPhrases()` 和 `ContainsForbiddenPhrasesInResult()` ([internal/ai/custom.go](G:/GoCode/investment-kb/internal/ai/custom.go))

**禁止表达列表（15 个关键词）**：

**绝对化收益表达（10 个）**：
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

**满仓误导表达（6 个）**：
- 可直接满仓
- 应直接满仓
- 可以满仓
- 满仓买入
- 直接满仓
- 高确定性时可直接满仓

**检查字段**：
- summary
- core_conclusion
- core_logic.title
- core_logic.content
- applicable_scenarios
- risk_boundaries
- extractable_rules.summary
- candidate_rules.rule_content
- candidate_rules.trigger_conditions
- candidate_rules.actions
- candidate_rules.not_applicable
- candidate_rules.risk_boundary
- candidate_rules.questions_to_confirm
- candidate_rules.recommendation
- my_understanding

**校验失败行为**：
- ❌ 终止流程
- ❌ 不渲染 Markdown
- ❌ 不写入 Obsidian
- ❌ 不更新 id_state.json
- ❌ 不更新 import_hashes.json
- ❌ Dry-run 模式下也必须失败

**错误示例**：
```
❌ AI 输出包含禁止表达：保证盈利。请调整 Prompt 或人工检查后重试。

❌ 在字段「core_conclusion」中：AI 输出包含禁止表达：可直接满仓。请调整 Prompt 或人工检查后重试。
```

**测试**：`TestContainsForbiddenPhrases` ([internal/app/validation_test.go](G:/GoCode/investment-kb/internal/app/validation_test.go#L127-L282))

#### 2. 买入规则 domain_code 不匹配检查（Warning，不失败）

**函数**：`WarnRuleTypeDomainMismatch()` ([internal/app/extract.go:378-394](G:/GoCode/investment-kb/internal/app/extract.go#L378-L394))

**检查逻辑**：
- 只检查 `rule_type == "买入规则"` 的规则
- 如果 `domain_code != "BUY"`，则打印 warning
- 只打印 warning，不终止流程

**Warning 文案**：
```
⚠️  候选规则分类可能不一致：买入规则「高概率区间先建底仓」的 domain_code=SAFETY，建议检查是否应为 BUY-SAFETY。
```

**示例**：
- ❌ 错误示例：
  - rule_type = "买入规则"
  - domain_code = "SAFETY"
  - topic_code = "SAFETY"
  - rule_name = "高概率区间先建底仓"

- ✅ 正确示例：
  - rule_type = "买入规则"
  - domain_code = "BUY"
  - topic_code = "SAFETY"
  - rule_name = "高概率区间先建底仓"

**注意**：
- 这只是 warning，不会中断程序
- Dry-run 模式下也要显示 warning
- 正式写入模式下也显示 warning，但仍允许继续写入
- 这类 warning 用于提醒人工检查，不作为硬性失败条件

**测试**：`TestWarnRuleTypeDomainMismatch` ([internal/app/validation_test.go](G:/GoCode/investment-kb/internal/app/validation_test.go#L286-L363))

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
4. 禁止表达检查（硬性校验）
   └─ 包含 15 个禁止表达（绝对化收益 + 满仓误导）
   ↓
5. 禁止绝对化收益表达检查
   └─ 检查 10 个绝对化关键词
   ↓
6. candidate_rules 类型集中 warning
   └─ 打印 warning（不终止）
   ↓
7. 买入规则 domain_code 检查（warning，不终止）
   └─ 打印 warning（不终止）
   ↓
8. 渲染 Markdown (只在校验通过后)
   ↓
9. 写入 Obsidian (只在 DryRun=false 时)
```

## 硬性校验 vs Warning

| 类型 | 示例 | 处理方式 | 终止流程 |
|------|------|----------|----------|
| 硬性校验 | 包含"保证盈利"、"可直接满仓" | 返回 error，终止流程 | ✅ 是 |
| Warning | domain_code 不匹配 | 打印 warning，继续执行 | ❌ 否 |

## 错误处理

### 硬性校验失败时

1. **打印错误信息**：
   ```
   ❌ 在字段「core_conclusion」中：AI 输出包含禁止表达：保证盈利。请调整 Prompt 或人工检查后重试。
   ```

2. **保存错误输出**：
   - 文件路径：`data/error_outputs/ai_error_YYYYMMDD_HHMMSS.txt`
   - 包含原始 AI 输出、清洗后输出、错误信息

3. **终止流程**：
   - 不渲染 Markdown
   - 不写入 Obsidian
   - 不更新状态文件

### Warning 时

1. **打印 warning 信息**：
   ```
   ⚠️  候选规则分类可能不一致：买入规则「高概率区间先建底仓」的 domain_code=SAFETY，建议检查是否应为 BUY-SAFETY。
   ```

2. **继续执行**：
   - 可以继续渲染 Markdown
   - 可以写入 Obsidian（正式模式）
   - 可以更新状态文件

## 测试

### 单元测试

```bash
# 运行所有校验测试
go test ./internal/app -v -run "TestValidate"

# 运行禁止表达测试
go test ./internal/app -v -run "TestContainsForbiddenPhrases"

# 运行 domain_code 检查测试
go test ./internal/app -v -run "TestWarnRuleTypeDomainMismatch"
```

### 测试覆盖

- ✅ **硬性校验测试**：17 个测试用例
  - 10 个绝对化表达测试
  - 6 个满仓误导表达测试
  - 不包含表达测试
  - 多个表达测试

- ✅ **Warning 测试**：5 个测试用例
  - 所有 domain_code 都是 BUY
  - 一条规则 domain_code 不匹配
  - 多条规则 domain_code 不匹配
  - 没有买入规则
  - 只有非买入规则

## 文件修改清单

### 新增/修改文件

1. **internal/ai/custom.go**
   - 新增 `forbiddenPhrases` 列表（15 个禁止关键词）
   - 新增 `ContainsForbiddenPhrases()` 函数
   - 新增 `ContainsForbiddenPhrasesInResult()` 函数
   - 导入 `sort` 包（按长度排序）

2. **internal/app/extract.go**
   - 新增 `WarnRuleTypeDomainMismatch()` 函数
   - 修改 `validateExtractionResult()` 函数
     - 增加禁止表达检查
     - 增加买入规则 domain_code 检查

3. **internal/app/validation_test.go**
   - 新增 `TestContainsForbiddenPhrases` 测试
   - 新增 `TestWarnRuleTypeDomainMismatch` 测试
   - 新增 `contains` 和 `findSubstring` 辅助函数

4. **docs/validation-implementation-v2.md**（本文档）
   - v2 更新说明
   - 新增功能详解
   - 测试覆盖说明

## 使用示例

### Mock 模式测试

```bash
# Mock 模式，测试正常流程
./kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"

# Mock 模式，测试禁止表达
# 修改 mock 数据添加禁止表达，应该会失败
```

### 真实 AI 模式测试

```bash
# 真实 AI，正常流程
$env:AI_API_KEY = "your_key"
./kb.exe -input examples/raw_qa.txt -source "陈老师问答" -dry-run

# 如果 AI 输出包含禁止表达，会失败并保存错误
```

## 兼容性说明

- ✅ RAW / QA / CR Markdown 模板未修改
- ✅ Obsidian WikiLink 未修改
- ✅ 编号规则未修改
- ✅ Mock 数据保持不变
- ✅ 现有功能不受影响
- ✅ 向后兼容

## 禁止表达列表更新

### v1 已有（10 个）
1. 保证盈利
2. 没有亏损风险
3. 必然上涨
4. 一定赚钱
5. 判断错了也不会亏
6. 只赚不亏
7. 无风险
8. 稳赚
9. 绝对安全
10. 必胜

### v2 新增（6 个）
11. 可直接满仓
12. 应直接满仓
13. 可以满仓
14. 满仓买入
15. 直接满仓
16. 高确定性时可直接满仓

**总计**：15 个禁止表达

## 注意事项

1. **排序规则**：禁止表达按长度从长到短排序，避免较短的短语先匹配
2. **Warning 不中断**：domain_code 检查的 warning 只是提醒，不会影响执行
3. **硬性失败**：任何禁止表达都会导致流程终止，确保数据质量
4. **字段检查**：覆盖所有相关字段，不留遗漏
