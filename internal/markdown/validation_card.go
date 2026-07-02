package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"investment-kb/internal/config"
	"investment-kb/internal/dedup"
	"investment-kb/internal/idgen"
	"investment-kb/internal/model"
)

// DefaultValidationCardTemplate 默认验证卡模板（当配置文件未指定或模板读取失败时使用）
const DefaultValidationCardTemplate = `# 规则验证卡：{{RULE_NAME}}

> **规则 ID**: {{RULE_ID}}
> **规则短码**: {{SHORT_CODE}}
> **原始领域**: {{ORIGINAL_DOMAIN_CODE}}-{{TOPIC_CODE}}
> **映射领域**: {{DOMAIN_CODE}}-{{TOPIC_CODE}}
> **建议正式领域**: {{SUGGESTED_FORMAL_DOMAIN}}
> **来源知识卡片**: {{SOURCE_QA_LINK}}
> **来源原文**: {{SOURCE_RAW_LINK}}
> **source_file**: {{SOURCE_FILE}}
> **raw_hash**: {{RAW_HASH}}
> **cleaned_hash**: {{CLEANED_HASH}}
> **raw_id**: {{RAW_ID}}
> **material_type**: {{MATERIAL_TYPE}}

source_file: {{SOURCE_FILE}}
raw_hash: {{RAW_HASH}}
cleaned_hash: {{CLEANED_HASH}}
raw_id: {{RAW_ID}}
material_type: {{MATERIAL_TYPE}}

---

## 规则摘要

- **rule_id**: {{RULE_ID}}
- **rule_name**: {{RULE_NAME}}
- **rule_type**: {{RULE_TYPE}}
- **原始领域（AI 建议）**: {{ORIGINAL_DOMAIN_CODE}}-{{TOPIC_CODE}}
- **映射领域（程序确认）**: {{DOMAIN_CODE}}-{{TOPIC_CODE}}
- **建议正式领域**: {{SUGGESTED_FORMAL_DOMAIN}}

### 触发条件

{{TRIGGER_CONDITIONS}}

### 适用范围

{{APPLICABLE_OBJECTS}}

### 禁止条件

{{NOT_APPLICABLE}}

### 风险边界

{{RISK_BOUNDARY}}

---

## 领域复核

| 字段 | 值 |
|---|---|
| original_domain_code | {{ORIGINAL_DOMAIN_CODE}} |
| reviewed_domain_code | （待人工确认，默认与映射领域一致） |
| domain_review_status | 待复核 |
| domain_review_note | （如分类有误，在此注明原因和应归属的领域） |

> 复核原则：如果规则的触发条件核心不是估值，而是账户硬约束或风险预案，应调整领域。

---

## 验证结论

| 字段 | 值 |
|---|---|
| final_decision | 待定 |
| backtest_validation_done | false |
| live_review_done | false |

### 转正式检查清单

| # | 检查项 | 是否满足 |
|---|---|---|
| 1 | 明确触发条件 | 否 |
| 2 | 明确适用范围 | 否 |
| 3 | 明确禁止条件 | 否 |
| 4 | 至少 1 个来源观点链接 | 是 |
| 5 | 不与现有规则冲突 | 否 |
| 6 | 有人为确认记录 | 否 |
| 7 | 至少完成 1 个相似案例 + 1 个反例/边界案例验证 | 否 |

---

## Case 1：相似案例

（待人工补充）

---

## Case 2：反例/边界案例

（待人工补充）

---

## 问题与修订

（待人工补充）

---

## 相似规则检查

{{SIMILAR_RULES_CHECK}}

---

## 最终决定

| 字段 | 值 |
|---|---|
| final_decision | 待定 |
| 决定日期 | |
| 确认人 | |
| 备注 | |
`

// RenderValidationCard 生成规则验证卡 Markdown 内容
func RenderValidationCard(cfg *config.Config, crID, qaID, rawID string, result *model.ExtractionResult, rule model.CandidateRule, similarRules []dedup.SimilarRule) (string, string) {
	content := loadValidationCardTemplate(cfg)
	content = fillValidationCardTemplate(content, cfg, crID, qaID, rawID, result, rule, similarRules)

	// 生成文件名：CR-<领域>-<日期>-<序号>.md
	fileName := crID + ".md"
	relativePath := filepath.Join(GetValidationCardDir(cfg), fileName)

	return content, relativePath
}

// loadValidationCardTemplate 读取验证卡模板，失败时返回默认模板
func loadValidationCardTemplate(cfg *config.Config) string {
	if cfg == nil || cfg.Files.ValidationCardTemplate == "" {
		return DefaultValidationCardTemplate
	}

	fullPath := filepath.Join(cfg.ObsidianVaultPath, cfg.Files.ValidationCardTemplate)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return DefaultValidationCardTemplate
	}

	return string(data)
}

// fillValidationCardTemplate 填充模板中的占位符
func fillValidationCardTemplate(content string, cfg *config.Config, crID, qaID, rawID string, result *model.ExtractionResult, rule model.CandidateRule, similarRules []dedup.SimilarRule) string {
	shortCode := rule.DomainCode + "-" + rule.TopicCode
	mappedDomain := idgen.MapCRDomain(rule.DomainCode)

	replacements := map[string]string{
		"{{RULE_ID}}":              crID,
		"{{RULE_NAME}}":            rule.RuleName,
		"{{RULE_TYPE}}":            rule.RuleType,
		"{{DOMAIN_CODE}}":          rule.DomainCode,
		"{{ORIGINAL_DOMAIN_CODE}}": rule.OriginalDomainCode,
		"{{TOPIC_CODE}}":           rule.TopicCode,
		"{{SHORT_CODE}}":           shortCode,
		"{{SUGGESTED_FORMAL_DOMAIN}}": mappedDomain,
		"{{RULE_CONTENT}}":         rule.RuleContent,
		"{{TRIGGER_CONDITIONS}}":   joinBulletList(rule.TriggerConditions),
		"{{ACTIONS}}":              joinBulletList(rule.Actions),
		"{{NOT_APPLICABLE}}":       joinBulletList(rule.NotApplicable),
		"{{RISK_BOUNDARY}}":        rule.RiskBoundary,
		"{{QUESTIONS_TO_CONFIRM}}": joinNumberedList(rule.QuestionsToConfirm),
		"{{RECOMMENDATION}}":       rule.Recommendation,
		"{{SOURCE_QA_LINK}}":       ObsidianHeadingLink(GetQaPath(cfg), JoinHeading(qaID, result.Title), qaID),
		"{{SOURCE_RAW_LINK}}":      ObsidianHeadingLink(GetRawMaterialPath(cfg), JoinHeading(rawID, result.Title), rawID),
		"{{APPLICABLE_OBJECTS}}":   joinSimpleList(rule.ApplicableObjects),
		"{{SOURCE_FILE}}":          result.SourceMeta.SourceFile,
		"{{RAW_HASH}}":             result.SourceMeta.RawHash,
		"{{CLEANED_HASH}}":         result.SourceMeta.CleanedHash,
		"{{RAW_ID}}":               result.SourceMeta.RawID,
		"{{MATERIAL_TYPE}}":        string(result.SourceMeta.MaterialType),
	}

	// 生成相似规则检查文本
	var similarRulesCheck string
	if len(similarRules) > 0 {
		var sb strings.Builder
		sb.WriteString("相似候选规则：\n\n")
		for _, sr := range similarRules {
			sb.WriteString(fmt.Sprintf("- [[%s#%s|%s]]\n", GetCandidateRulePath(cfg), sr.CRID+"｜"+sr.ShortCode+"｜"+sr.RuleName, sr.CRID))
			sb.WriteString(fmt.Sprintf("  - 相似原因：%s\n", sr.Reason))
			sb.WriteString(fmt.Sprintf("  - 相似级别：%s\n", sr.Level))
		}
		sb.WriteString("\n处理建议：\n\n")
		sb.WriteString("- [ ] 新建独立规则\n")
		sb.WriteString("- [ ] 合并到已有规则（作为补充来源）\n")
		sb.WriteString("- [ ] 保留但标记可能与已有规则冲突\n")
		sb.WriteString("- [ ] 废弃（与已有规则重复）\n")
		similarRulesCheck = sb.String()
	} else {
		similarRulesCheck = "相似候选规则：暂无\n"
	}
	replacements["{{SIMILAR_RULES_CHECK}}"] = similarRulesCheck

	for placeholder, value := range replacements {
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// 强制替换标题行：无论模板是什么，都改成以 rule_id 开头的正确标题
	// 匹配以 # 开头的任意标题行，替换为标准标题
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			lines[i] = "# " + crID + "｜规则验证卡"
			break // 只替换第一个标题行
		}
	}
	content = strings.Join(lines, "\n")
	content = ensureValidationSourceMetadata(content, result.SourceMeta)

	return content
}


func ensureValidationSourceMetadata(content string, meta model.SourceMeta) string {
	if strings.Contains(content, "source_file:") && strings.Contains(content, "cleaned_hash:") && strings.Contains(content, "raw_id:") {
		return content
	}
	block := "\nsource_file: " + meta.SourceFile + "  \n" +
		"raw_hash: " + meta.RawHash + "  \n" +
		"cleaned_hash: " + meta.CleanedHash + "  \n" +
		"raw_id: " + meta.RawID + "  \n" +
		"material_type: " + string(meta.MaterialType) + "  \n"
	if idx := strings.Index(content, "\n"); idx >= 0 {
		return content[:idx+1] + block + content[idx+1:]
	}
	return content + block
}

// joinBulletList 将字符串数组转换为 Markdown 无序列表
func joinBulletList(items []string) string {
	if len(items) == 0 {
		return "（待补充）"
	}
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	return sb.String()
}

// joinNumberedList 将字符串数组转换为 Markdown 有序列表
func joinNumberedList(items []string) string {
	if len(items) == 0 {
		return "（待补充）"
	}
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	return sb.String()
}

// joinSimpleList 将字符串数组用 / 分隔
func joinSimpleList(items []string) string {
	if len(items) == 0 {
		return "（待补充）"
	}
	return strings.Join(items, " / ")
}

// GetValidationCardDir 获取验证卡目录
func GetValidationCardDir(cfg *config.Config) string {
	if cfg == nil || cfg.Files.ValidationCardDir == "" {
		return "日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/规则验证卡"
	}
	return cfg.Files.ValidationCardDir
}

// GetValidationCardPath 获取单个验证卡文件的相对路径
func GetValidationCardPath(cfg *config.Config, crID string) string {
	return filepath.Join(GetValidationCardDir(cfg), crID+".md")
}

// GetValidationCardLink 生成验证卡 WikiLink
func GetValidationCardLink(cfg *config.Config, crID string, alias string) string {
	return ObsidianHeadingLink(GetValidationCardPath(cfg, crID), crID, alias)
}
