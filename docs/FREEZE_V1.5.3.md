# V1.5.3 — 模块一冻结文档

> 模块一：Obsidian 知识沉淀与规则系统
> 状态：**稳定版**
> 冻结日期：2026-07-06

---

## 📌 版本摘要

| 字段 | 值 |
|------|-----|
| 版本号 | **V1.5.3** |
| 模块 | 模块一（Obsidian 知识沉淀与规则系统） |
| 状态 | 通过 |
| 正式 config 全量 validate | **PASS** |
| 备注 | 存在 5 条 RAW 标题/正文一致性 warning，人工确认不阻断 |

### 当前统计

| 指标 | 数量 |
|------|------|
| RAW count | 12 |
| QA count | 5 |
| KNOW count | 7 |
| CR count | 12 |
| validation card count | 12 |
| broken links | none |
| frontmatter delimiter issue | none |
| source_meta missing | none |
| source mismatch | none |

---

## 🔒 冻结内容（自 V1.5.3 起不再变更）

以下规则在 V1.5.3 状态稳定，自此版本起冻结。后续修改需要走「版本变更流程」。

### 1. 目录结构冻结

```
G:\Obsidian\我的知识库\
└── 日常随笔/
    └── 股市学习/
        └── 宽基指数仓位管理系统/
            ├── 00-待处理材料/                    # 输入收件箱
            │   └── 问答/
            ├── 01-源文档/                        # RAW 落地区
            │   └── 问答/
            │       ├── RAW-{ID}.md               # 单文件
            │       └── 原始材料索引.md
            ├── 02-观点/                          # QA / KNOW 落地区
            │   ├── 问答知识卡片/
            │   │   ├── QA-{ID}.md
            │   │   └── 问答知识卡片索引.md
            │   ├── 宏观理解卡/
            │   │   ├── KNOW-{ID}.md
            │   │   └── 宏观理解卡索引.md
            │   └── 市场观察卡/                    # MO 预留
            │       ├── MO-{ID}.md
            │       └── 市场观察卡索引.md
            └── 03-规则/
                ├── 候选规则/
                │   ├── CR-{ID}.md
                │   └── 候选规则索引.md
                └── 规则回溯验证/
                    └── 规则验证卡/
                        └── CR-{ID}.md
```

### 2. 文件命名规则冻结

| 类型 | 格式 | 示例 |
|------|------|------|
| RAW | `RAW-{DOMAIN}-{SUBTYPE}-{YYYYMMDD}-{NNN}.md` | `RAW-ACCOUNT-SAFETY-20260703-001.md` |
| QA | `QA-{DOMAIN}-{SUBTYPE}-{YYYYMMDD}-{NNN}.md` | `QA-ACCOUNT-SAFETY-20260703-001.md` |
| KNOW | `KNOW-{DOMAIN}-{SUBTYPE}-{YYYYMMDD}-{NNN}.md` | `KNOW-L2-ECON-20260703-001.md` |
| MO | `MO-{DOMAIN}-{SUBTYPE}-{YYYYMMDD}-{NNN}.md` | （预留） |
| CR | `CR-{DOMAIN}-{YYYYMMDD}-{NNN}.md` | `CR-ACCOUNT-20260703-001.md` |
| 验证卡 | `CR-{DOMAIN}-{YYYYMMDD}-{NNN}.md` | `CR-ACCOUNT-20260703-001.md` |

- **ID 段顺序固定**：`类型-领域-子类型-日期-序号`
- **序号 NNN**：3 位数字，从 001 开始，按日自增
- **日期段**：YYYYMMDD 格式
- **不允许**：在文件名中加入标题、subtype 之外的自定义字段

### 3. 链接规则冻结

| 链接类型 | 格式 | 说明 |
|----------|------|------|
| 内部 WikiLink | `[[相对路径\|ID]]` | 仅用 ID 作为别名 |
| 内部裸链接 | `[[相对路径]]` | 链接到具体文件 |
| 索引条目 | `- [[相对路径\|ID]]` | 列表项只引用 ID |
| 外部链接 | 标准 Markdown 链接 | 不变 |

**禁止**：
- 在链接别名中使用标题、subtype 等额外信息
- 使用绝对路径作为 WikiLink 目标
- 使用 `[text](path)` 形式链接内部文件

### 4. source_meta 保存方式冻结

每个生成文件必须在 frontmatter 中保存以下字段：

```yaml
---
source_file: "G:\Obsidian\...\input.md"   # 原始输入文件绝对路径
raw_hash: "f87771320d..."                  # 原始内容 SHA-256
cleaned_hash: "189bf519b..."               # 清洗后内容 SHA-256
raw_id: "RAW-XXX-20260703-001"            # 关联的 RAW ID
material_type: "rule_candidate"            # macro_knowledge / rule_candidate / market_observation / archive_only
---
```

**保存要求**：
- 所有 4 类文件（RAW / QA / KNOW / CR / 验证卡）都必须保存完整 source_meta
- hash 必须使用 SHA-256
- 不允许遗漏任何字段
- validate 会检查 `source_meta missing` 和 `source mismatch`

### 5. 文件模板冻结

#### RAW 模板
```markdown
---
[source_meta]
---

# RAW-{ID}

## 元信息
- 标题：{原文标题}
- 来源：{source 名称}
- 日期：{YYYY-MM-DD}
- 主题：{主题}

## 原文内容
{清洗后的原文}
```

#### QA 模板
```markdown
---
[source_meta]
---

# QA-{ID}

## 问题
{AI 提取的问题}

## 答案
{AI 提取的答案}

## 知识定位
- 领域：{domain}
- 子类型：{subtype}
```

#### KNOW 模板
```markdown
---
[source_meta]
---

# KNOW-{ID}

## 主题
{主题}

## 核心观点
{3-5 条核心结论}

## 适用范围
{适用场景说明}

## 边界条件
{限制与不适用的场景}
```

#### CR 模板
```markdown
---
[source_meta]
---

# CR-{ID}

## 规则陈述
{可执行规则}

## 触发条件
- ...

## 执行动作
- ...

## 适用场景
{...}

## 不适用场景
{...}
```

#### 验证卡模板
```markdown
---
[source_meta]
cr_id: "CR-XXX-20260703-001"   # 关联的 CR ID
---

# 验证卡 — CR-{ID}

## 规则回溯
{对原始材料中支持该规则的引用}

## 验证要点
- ...

## 边界提醒
- ...
```

### 6. validate 规则冻结

`./kb.exe validate` 必须检查以下项：

| 检查项 | 失败级别 | 说明 |
|--------|----------|------|
| RAW count vs 文件系统 | info | 统计一致性 |
| QA count vs 文件系统 | info | 统计一致性 |
| KNOW count vs 文件系统 | info | 统计一致性 |
| CR count vs 文件系统 | info | 统计一致性 |
| 验证卡 count vs 文件系统 | info | 统计一致性 |
| 索引存在性 | **fail** | 原始/QA/CR 索引必须存在 |
| orphan validation cards | **fail** | 验证卡必须有对应 CR |
| missing validation cards | **fail** | CR 必须有验证卡 |
| broken links | **fail** | 所有链接目标必须存在 |
| frontmatter delimiter issue | **fail** | YAML 分隔符必须正确 |
| source_meta missing | **fail** | 所有生成文件必须有完整 source_meta |
| source mismatch | **fail** | 关联文件间 source_meta 必须一致 |
| 标题/正文一致性 | **warning** | RAW 标题与原文关键词匹配检查（不阻断） |
| 重复 hash | **fail** | raw_hash / cleaned_hash 不能重复 |

**PASS 条件**：所有 fail 项为 0，warning 项不阻断。

---

## 🔄 版本变更流程

任何对上述冻结内容的修改，需要：

1. **新建 V1.5.x 分支**
2. **记录变更原因 + 影响范围**
3. **重跑全量 validate**
4. **在 AGENTS.md / README.md / docs/ 同步更新**
5. **更新本冻结文档（V1.5.4、V1.5.5…）**

不允许在原 V1.5.3 上热修补。

---

## 📋 历史版本

| 版本 | 状态 | 关键变更 |
|------|------|----------|
| V1.1 | 封板 | KNOW 单文件模式、索引、一致性校验、source 追溯 |
| V1.3 | 封板 | CLI 子命令化、validate 命令、编号去重、material_type 分类优化 |
| V1.4 | 封板 | CR 单文件模式、聚合库去重、链接格式修复、validate 增强 |
| V1.4.2 | 封板 | 真实材料批量处理、AI 调用重试机制 |
| V1.4.3 | 封板 | 修复 AI 输出 JSON 格式问题 |
| V1.5.x | 迭代中 | RAW/QA 单文件模式、清理脚本、单文件去重、source_meta 完善 |
| **V1.5.3** | **稳定版** | **当前冻结版本** |
