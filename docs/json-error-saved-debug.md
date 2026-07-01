# JSON 错误输出保存优化

> **当前状态**：`saveErrorOutput` 函数的完整原始输出保存逻辑仍然有效。当前实现在 `internal/ai/custom.go:251-277`。

## 问题

AI 返回 JSON 解析失败时，错误输出文件只保存了清洗后的 JSON 部分，没有完整的原始输出，难以排查问题。

## 修改

### 修改 saveErrorOutput 函数 (internal/ai/custom.go:241-262)

**修改前**：
```go
content := fmt.Sprintf(`执行时间: %s
输入文件: %s
Source: %s
错误步骤: %s
错误原因: %s
---
AI 原始输出 (前 2000 字符):
%s
`,
    time.Now().Format("2006-01-02 15:04:05"),
    inputPath,
    source,
    step,
    errorDetail,
    truncate(raw, 2000))
```

**修改后**：
```go
content := fmt.Sprintf(`执行时间: %s
输入文件: %s
Source: %s
错误步骤: %s
错误原因: %s
---
AI 原始输出开头 (前 500 字符):
%s
...
AI 原始输出结尾 (后 500 字符):
%s
---
完整原始输出:
%s
`,
    time.Now().Format("2006-01-02 15:04:05"),
    inputPath,
    source,
    step,
    errorDetail,
    truncate(raw, 500),
    truncate(raw, 500),
    raw)
```

## 改进

1. **保存完整原始输出**：增加 `完整原始输出` 部分，包含完整的 AI 响应
2. **显示开头和结尾**：增加 AI 原始输出开头和结尾的预览（各 500 字符）
3. **便于调试**：可以通过查看文件开头和结尾快速定位问题

## 错误输出文件格式

```
执行时间: 2026-06-24 09:46:10
输入文件: examples/raw_qa.txt
Source: 陈老师问答
错误步骤: json.Unmarshal
错误原因: JSON 解析失败: invalid character 'â' looking for beginning of value

AI 原始输出开头 (前 500 字符):
{"title": "高概率区间先建底仓 账户状态决定仓位力度",...
...
AI 原始输出结尾 (后 500 字符):
...
"recommendation": "建议修改后采纳。需要明确不同账户状态下的具体仓位策略边界和量化标准。"
---
完整原始输出:
{"title": "高概率区间先建底仓 账户状态决定仓位力度", "source": "陈老师问答", ...完整内容...}
```

## 下一步

当遇到 JSON 解析错误时：
1. 查看 `data/error_outputs/ai_error_YYYYMMDD_HHMMSS.txt`
2. 检查 "完整原始输出" 部分是否有 BOM、markdown 代码块等问题
3. 检查 "AI 原始输出开头" 和 "结尾" 是否一致

## 文件修改

- ✅ `internal/ai/custom.go` - 增强 saveErrorOutput 函数
- ✅ 编译成功
