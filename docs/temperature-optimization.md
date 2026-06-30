# AI 输出稳定性优化

## 概述

通过固定 AI 请求的 temperature 参数为 0，确保同一篇文章多次运行时，domain_code、topic_code、candidate_rules 等关键字段保持稳定。

## 修改内容

### 一、添加 Temperature 参数

#### 1. AI 客户端配置 (internal/ai/client.go)

**修改前**：
```go
type Config struct {
	Provider   string
	Model      string
	BaseURL    string
	APIKey     string
	TimeoutSec int
	MaxRetries int
}
```

**修改后**：
```go
type Config struct {
	Provider    string
	Model       string
	BaseURL     string
	APIKey      string
	TimeoutSec  int
	MaxRetries  int
	Temperature float64 // 默认 0，确保输出稳定性
}
```

#### 2. 默认值设置 (internal/ai/client.go:51-55)

```go
if cfg.Temperature <= 0 {
	cfg.Temperature = 0 // 默认 0，确保输出稳定性
}
```

#### 3. Extract 函数集成 (internal/app/extract.go:224-231)

```go
client, err := ai.NewClient(&ai.Config{
	Provider:    cfg.AI.Provider,
	Model:       cfg.AI.Model,
	BaseURL:     cfg.AI.BaseURL,
	APIKey:      apiKey,
	TimeoutSec:  cfg.AI.TimeoutSec,
	MaxRetries:  3,
	Temperature: cfg.AI.Temperature,
})
```

### 二、配置文件

#### config.yaml

```yaml
ai:
  provider: 'custom'
  model: 'glm-4.7-flash'
  base_url: 'https://api.z.ai/api/anthropic'
  api_key_env: 'ANTHROPIC_AUTH_TOKEN'
  timeout_seconds: 120
  temperature: 0
```

**Temperature: 0 的作用**：
- 消除随机性
- 确保同一输入产生相同的输出
- 提高数据一致性和可复现性

### 三、my_understanding 空值处理

#### 1. 校验逻辑 (internal/app/extract.go:314-317)

```go
// 6. my_understanding 空值检查（warning，不终止）
if result.MyUnderstanding == "" {
	fmt.Printf("⚠️  my_understanding 为空，已在 Markdown 中使用「待补充。」\n")
}
```

**行为**：
- Dry-run 和正式运行时都打印 warning
- 不终止流程（warning，不是 error）

#### 2. Markdown 渲染 (internal/markdown/qa.go:117-121)

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
- 如果 my_understanding 为空，输出"待补充。"
- 如果不为空，输出实际内容
- 保持其他部分结构不变

## 测试结果

### 编译测试
```bash
go build ./...
# ✓ Build successful
```

### 单元测试
```bash
go test ./internal/app -v
# PASS (33 tests passed)
```

### Mock 模式测试
```bash
./kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"
# ✓ 无 warning（my_understanding 不为空）
```

## 使用示例

### Mock 模式（快速测试）
```bash
./kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"
```

### 真实 AI 模式
```bash
$env:ANTHROPIC_AUTH_TOKEN = "your_api_key"
./kb.exe -input examples/raw_qa.txt -source "陈老师问答" -dry-run
```

## 影响

### 正面影响
1. **输出一致性**：同一文章多次运行产生相同的 domain_code、topic_code
2. **数据质量**：减少因随机性导致的数据不稳定
3. **调试便利**：更容易追踪和重现问题

### 负面影响
1. **灵活性降低**：完全消除随机性，无法探索不同可能性
2. **创意受限**：某些场景下可能需要更高温度获得多样性

### 适用场景
- **推荐**：生产环境、数据清洗、重复任务
- **不推荐**：需要创意生成的场景、探索性分析

## 文件修改清单

### 修改的文件
1. `internal/ai/client.go`
   - Config 结构体添加 Temperature 字段
   - NewClient 函数设置默认值

2. `internal/app/extract.go`
   - 传递 Temperature 参数给 AI 客户端

3. `internal/config/config.go`
   - AI 结构体添加 Temperature 字段

4. `config.yaml`
   - 添加 temperature: 0 配置

5. `internal/markdown/qa.go`
   - 处理 my_understanding 空值

6. `internal/app/extract.go`
   - 添加 my_understanding 空值检查

### 未修改的部分
- ✅ Prompt
- ✅ Markdown 模板（RAW / QA / CR）
- ✅ 编号规则
- ✅ Obsidian WikiLink 规则
- ✅ 测试用例（通过率 100%）

## 注意事项

1. **Temperature=0**：完全消除随机性，适合需要一致性的场景
2. **my_understanding 为空**：会打印 warning 但不会终止流程
3. **向后兼容**：现有功能不受影响

## 验证方法

1. **运行两次相同的输入**：
   ```bash
   ./kb.exe -input test.txt -mock -dry-run -source "测试" > out1.txt
   ./kb.exe -input test.txt -mock -dry-run -source "测试" > out2.txt
   diff out1.txt out2.txt
   # 应该无差异
   ```

2. **检查 my_understanding 处理**：
   - 如果 AI 输出的 my_understanding 为空，会输出 warning
   - Markdown 文件中"## 8. 我的理解"部分会显示"待补充。"

## 总结

✅ AI 输出稳定性优化完成
✅ Temperature 固定为 0
✅ my_understanding 空值处理完成
✅ 所有测试通过
✅ 向后兼容，无破坏性变更
