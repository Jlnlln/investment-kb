# 执行指南

## 使用 Mock 模式（推荐用于测试 WikiLink 格式）

```bash
# Mock 模式：使用 Mock 数据，不调用 AI，只打印 Markdown
kb.exe -input examples/raw_qa.txt -mock -dry-run -source "陈老师问答"
```

**优点：**
- ✅ 快速生成，无需 API Key
- ✅ 适合测试 WikiLink 格式
- ✅ 可以重复运行

**适用场景：**
- 测试 WikiLink 格式是否正确
- 验证 Markdown 生成逻辑
- 快速查看输出效果

---

## 使用真实 AI 模式（需要配置 API Key）

### 方法 1：通过环境变量设置 API Key

```bash
# Windows PowerShell
$env:AI_API_KEY = "your_api_key_here"
kb.exe -input examples/raw_qa.txt -source "陈老师问答"

# Windows CMD
set AI_API_KEY=your_api_key_here
kb.exe -input examples/raw_qa.txt -source "陈老师问答"

# Linux/Mac
export AI_API_KEY="your_api_key_here"
./kb -input examples/raw_qa.txt -source "陈老师问答"
```

### 方法 2：通过配置文件设置 API Key

修改 `config.yaml` 文件：

```yaml
obsidian_vault_path: 'G:\Obsidian\我的知识库'

files:
  raw_material: '日常随笔/股市学习/个人投资训练系统/03-知识与案例/原始材料库.md'
  qa: '日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.md'
  market_case: '日常随笔/股市学习/个人投资训练系统/03-知识与案例/市场案例库.md'
  candidate_rule: '日常随笔/股市学习/个人投资训练系统/04-投资规则/候选规则.md'

ai:
  provider: 'custom'
  model: 'glm-4.7-flash'
  base_url: 'https://api.z.ai/api/anthropic'
  api_key_env: 'AI_API_KEY'  # 设置为环境变量名
  timeout_seconds: 120

timezone: 'Asia/Beijing'
```

然后设置环境变量：

```bash
# Windows PowerShell
$env:AI_API_KEY = "your_api_key_here"
kb.exe -input examples/raw_qa.txt -source "陈老师问答"

# Linux/Mac
export AI_API_KEY="your_api_key_here"
./kb -input examples/raw_qa.txt -source "陈老师问答"
```

### 方法 3：通过命令行参数直接设置（推荐）

```bash
# Windows PowerShell
kb.exe -input examples/raw_qa.txt -source "陈老师问答" -ai-api-key "your_api_key_here"

# Linux/Mac
./kb -input examples/raw_qa.txt -source "陈老师问答" -ai-api-key "your_api_key_here"
```

---

## 其他可用参数

### Dry-run 模式（推荐）

```bash
# Mock 模式 + Dry-run（只打印，不写入 Obsidian）
kb.exe -input examples/raw_qa.txt -mock -dry-run -source "陈老师问答"

# 真实 AI + Dry-run（只打印，不写入 Obsidian）
kb.exe -input examples/raw_qa.txt -source "陈老师问答" -dry-run
```

### 完整参数示例

```bash
# Mock 模式完整示例
kb.exe -input examples/raw_qa.txt -mock -dry-run -source "陈老师问答" -config config.yaml

# 真实 AI 完整示例
kb.exe -input examples/raw_qa.txt -source "陈老师问答" -dry-run -config config.yaml -v
```

---

## 常用参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-input` | 输入文件路径 | 必需 |
| `-mock` | 使用 Mock 数据 | false |
| `-dry-run` | 只打印 Markdown，不写入 Obsidian | false |
| `-source` | 来源（如：陈老师问答） | 必需 |
| `-config` | 配置文件路径 | config.yaml |
| `-v` | 显示版本号 | false |
| `-ai-api-key` | 直接设置 API Key | 从环境变量读取 |

---

## 常见问题

### 1. API Key 未设置错误

**错误：**
```
❌ AI 调用失败: 未设置 API Key（环境变量：AI_API_KEY）
```

**解决：**
```bash
# Windows PowerShell
$env:AI_API_KEY = "your_actual_api_key_here"

# Linux/Mac
export AI_API_KEY="your_actual_api_key_here"
```

### 2. Token 过期错误

**错误：**
```
❌ AI 调用失败: 客户端错误 (401): {"error":{"message":"token expired or incorrect","type":"401"}}
```

**解决：**
- 重新生成 API Key
- 检查 API Key 是否正确
- 确认 API Key 有足够额度

### 3. 文件路径错误

**错误：**
```
❌ 读取配置文件失败: open config.yaml: no such file or directory
```

**解决：**
```bash
# 使用相对路径
kb.exe -input examples/raw_qa.txt -mock -dry-run

# 或指定绝对路径
kb.exe -input examples/raw_qa.txt -mock -dry-run -config G:/GoCode/investment-kb/config.yaml
```

---

## 示例输出

### Mock 模式输出

```
🧪 使用 Mock 数据

=== RAW ===

---

# RAW-POS-SAFETY-20260617-001｜安全边际与错失买入机会如何平衡

来源：陈老师问答
...
对应知识卡片：[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]
...

=== QA ===

---

# QA-POS-SAFETY-20260617-001｜安全边际与错失买入机会如何平衡

原始材料：[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/原始材料库#RAW-POS-SAFETY-20260617-001|RAW-POS-SAFETY-20260617-001]]
...

=== CR ===

---

# CR-20260617-001｜BUY-SAFETY｜高概率区间先建底仓

来源知识卡片：[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库#QA-POS-SAFETY-20260617-001|QA-POS-SAFETY-20260617-001]]
来源原文：[[日常随笔/股市学习/个人投资训练系统/03-知识与案例/原始材料库#RAW-POS-SAFETY-20260617-001|RAW-POS-SAFETY-20260617-001]]
...
```

### 真实 AI 模式输出

```
🤖 正在调用 AI...

✅ QA-POS-SAFETY-20260617-001 生成完成
✅ CR-20260617-001 生成完成
✅ CR-20260617-002 生成完成
✅ CR-20260617-003 生成完成

📊 生成统计：
  - RAW 文件：1 个
  - QA 文件：1 个
  - CR 文件：3 个
  - 总字数：约 5000 字
```

---

## 推荐工作流程

1. **先用 Mock 模式测试**
   ```bash
   kb.exe -input examples/raw_qa.txt -mock -dry-run -source "测试"
   ```

2. **检查 WikiLink 格式是否正确**
   - 确认 .md 后缀已去除
   - 确认路径格式正确

3. **使用真实 AI 模式生成**
   ```bash
   # 设置 API Key
   $env:AI_API_KEY = "your_api_key"
   
   # 运行
   kb.exe -input examples/raw_qa.txt -source "陈老师问答"
   ```

4. **验证生成的 Markdown 文件**
   - 检查 WikiLink 是否可点击
   - 确认 Obsidian 可以正确跳转
