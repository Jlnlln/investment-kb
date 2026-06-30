# My Understanding 空值优化

## 优化内容

### 问题

如果 AI 输出中的 `my_understanding` 为空字符串，QA 文档的"## 8. 我的理解"部分会留空，影响文档完整性。

### 解决方案

#### 1. 空值检查（Warning）

**位置**：`internal/app/extract.go:313-317`

```go
// 6. my_understanding 空值检查（warning，不终止）
if result.MyUnderstanding == "" {
    fmt.Printf("⚠️  my_understanding 为空，已在 Markdown 中使用「待补充。」\n")
}
```

**行为**：
- ✅ 在 dry-run 和正式运行时都打印 warning
- ✅ 不终止流程
- ✅ 只作为提醒

#### 2. Markdown 渲染优化

**位置**：`internal/markdown/qa.go:114-122`

```go
// 8. 我的理解
sb.WriteString("---\n\n")
sb.WriteString("## 8. 我的理解\n\n")
if result.MyUnderstanding == "" {
    sb.WriteString("待补充。")
} else {
    sb.WriteString(result.MyUnderstanding)
}
sb.WriteString("\n")
```

**行为**：
- ✅ 如果 `my_understanding` 为空，输出"待补充。"
- ✅ 如果不为空，输出实际内容
- ✅ 保持文档完整性

## 使用示例

### Mock 模式（my_understanding 为空）

```bash
./kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"
```

**输出**：
```
⚠️  my_understanding 为空，已在 Markdown 中使用「待补充。」

--- (部分省略)

## 8. 我的理解

待补充。

```

### 真实 AI 模式

如果 AI 输出 `my_understanding` 为空：
```
⚠️  my_understanding 为空，已在 Markdown 中使用「待补充。」

--- (部分省略)

## 8. 我的理解

待补充。

```

如果 AI 输出 `my_understanding` 有内容：
```
--- (部分省略)

## 8. 我的理解

这段问答最重要的启发是，投资决策不能只看市场点位，还要看账户状态。完整问题不是当前点位能不能买，而是在我的账户状态下当前点位能买多少。

```

## 限制

- ✅ 不修改 RAW / QA / CR Markdown 模板的其他部分
- ✅ 不修改编号规则
- ✅ 不修改 Obsidian WikiLink 规则
- ✅ 不修改 Prompt
- ✅ 只在"## 8. 我的理解"部分处理空值

## 测试

### 单元测试

```bash
go test ./internal/app -v -run "TestValidate"
```

### 集成测试

```bash
# Mock 模式测试
./kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"
```

## 文件修改

1. **internal/app/extract.go**
   - 新增 `my_understanding` 空值检查

2. **internal/markdown/qa.go**
   - 修改 `RenderKnowledgeCard` 函数，处理空值

## 兼容性

- ✅ 完全向后兼容
- ✅ 不影响现有功能
- ✅ 所有测试通过
