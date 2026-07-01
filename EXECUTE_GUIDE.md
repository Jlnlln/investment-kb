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
$env:ANTHROPIC_AUTH_TOKEN = "your_api_key_here"
kb.exe -input examples/raw_qa.txt -source "陈老师问答"

# Linux/Mac
export ANTHROPIC_AUTH_TOKEN="your_api_key_here"
./kb -input examples/raw_qa.txt -source "陈老师问答"
```

### 方法 2：通过配置文件设置 API Key

修改 `config.yaml` 文件：

```yaml
obsidian_vault_path: 'G:\Obsidian\我的知识库'

files:
  raw_material: '日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答/原始材料库.md'
  qa: '日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片库.md'
  market_case: '日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/历史案例库/AI提取案例素材库.md'
  candidate_rule: '日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则/候选规则库.md'
  validation_card_template: '日常随笔/股市学习/宽基指数仓位管理系统/99-模板/规则验证卡模板.md'
  validation_card_dir: '日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/规则验证卡'

ai:
  provider: 'custom'
  model: 'glm-5.1'
  base_url: 'https://api.z.ai/api/anthropic'
  api_key_env: 'ANTHROPIC_AUTH_TOKEN'
  timeout_seconds: 300
  temperature: 0

timezone: 'Asia/Beijing'
```

然后设置环境变量（API Key 只能通过环境变量设置，不支持命令行参数）：

```bash
# Windows PowerShell
$env:ANTHROPIC_AUTH_TOKEN = "your_api_key_here"
kb.exe -input examples/raw_qa.txt -source "陈老师问答"

# Linux/Mac
export ANTHROPIC_AUTH_TOKEN="your_api_key_here"
./kb -input examples/raw_qa.txt -source "陈老师问答"
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
| `-allow-duplicate` | 允许重复导入（跳过 hash 检查） | false |
| `-v` | 显示版本号 | false |

---

## 常见问题

### 1. API Key 未设置错误

**错误：**
```
❌ AI 调用失败: 未设置 API Key（环境变量：ANTHROPIC_AUTH_TOKEN）
```

**解决：**
```bash
# Windows PowerShell
$env:ANTHROPIC_AUTH_TOKEN = "your_actual_api_key_here"

# Linux/Mac
export ANTHROPIC_AUTH_TOKEN="your_actual_api_key_here"
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

# RAW-ACCOUNT-20260701-001｜安全边际与错失买入机会如何平衡

来源：陈老师问答
...
对应知识卡片：[[日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片库#QA-ACCOUNT-20260701-001|QA-ACCOUNT-20260701-001]]
...

=== QA ===

---

# QA-ACCOUNT-20260701-001｜安全边际与错失买入机会如何平衡

原始材料：[[日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答/原始材料库#RAW-ACCOUNT-20260701-001|RAW-ACCOUNT-20260701-001]]
...

=== CR ===

---

# CR-VALUATION-20260701-001｜VALUATION-SAFETY｜高概率区间先建底仓

来源知识卡片：[[日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片库#QA-...
来源原文：[[日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答/原始材料库#RAW-ACCOUNT-20260701-001|RAW-ACCOUNT-20260701-001]]
...
```

### 真实 AI 模式输出

```
🤖 正在调用 AI...

✅ RAW-ACCOUNT-20260701-001 生成完成
✅ QA-ACCOUNT-20260701-001 生成完成
✅ CR-VALUATION-20260701-001 生成完成
✅ CR-ACCOUNT-20260701-002 生成完成
✅ CR-RISK-20260701-003 生成完成

📊 生成统计：
  - RAW：1 条
  - QA：1 条
  - CR：3 条
  - 验证卡：3 张
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
   $env:ANTHROPIC_AUTH_TOKEN = "your_api_key"
   
   # 运行
   kb.exe -input examples/raw_qa.txt -source "陈老师问答"
   ```

4. **验证生成的 Markdown 文件**
   - 检查 WikiLink 是否可点击
   - 确认 Obsidian 可以正确跳转
