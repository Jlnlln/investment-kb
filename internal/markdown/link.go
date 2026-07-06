package markdown

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"investment-kb/internal/config"
)

// ObsidianHeadingLink 生成 Obsidian WikiLink 格式的引用
func ObsidianHeadingLink(filePath string, heading string, alias string) string {
	// 1. 替换 Windows 反斜杠为正斜杠
	linkPath := strings.ReplaceAll(filePath, "\\", "/")

	// 2. 去掉末尾的 .md 或 .MD
	linkPath = strings.TrimSuffix(linkPath, ".md")
	linkPath = strings.TrimSuffix(linkPath, ".MD")

	if alias == "" {
		alias = heading
	}

	return fmt.Sprintf("[[%s#%s|%s]]", linkPath, heading, alias)
}

// ObsidianFileLink 生成指向独立 Markdown 文件的 Obsidian WikiLink。
func ObsidianFileLink(filePath string, alias string) string {
	linkPath := strings.ReplaceAll(filePath, "\\", "/")
	linkPath = strings.TrimSuffix(linkPath, ".md")
	linkPath = strings.TrimSuffix(linkPath, ".MD")
	if alias == "" {
		parts := strings.Split(linkPath, "/")
		alias = parts[len(parts)-1]
	}
	return fmt.Sprintf("[[%s|%s]]", linkPath, alias)
}

func GetRawMaterialPath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/个人投资训练系统/03-知识与案例/原始材料库.md"
	}
	return cfg.Files.RawMaterial
}

func GetRawMaterialDir(cfg *config.Config) string {
	if cfg != nil && cfg.Files.RawMaterialDir != "" {
		return cfg.Files.RawMaterialDir
	}
	if cfg != nil && cfg.Files.RawMaterial != "" {
		return strings.TrimSuffix(cfg.Files.RawMaterial, filepath.Ext(cfg.Files.RawMaterial))
	}
	return "日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答"
}

func GetRawMaterialIndexPath(cfg *config.Config) string {
	if cfg != nil && cfg.Files.RawMaterialIndex != "" {
		return cfg.Files.RawMaterialIndex
	}
	return filepath.Join(GetRawMaterialDir(cfg), "原始材料索引.md")
}

func UseStandaloneRawMaterials(cfg *config.Config) bool {
	return cfg != nil && cfg.Files.RawMaterialDir != "" && cfg.Files.RawMaterialIndex != ""
}

func GetQaPath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.md"
	}
	return cfg.Files.QA
}

func GetQaDir(cfg *config.Config) string {
	if cfg != nil && cfg.Files.QADir != "" {
		return cfg.Files.QADir
	}
	if cfg != nil && cfg.Files.QA != "" {
		return strings.TrimSuffix(cfg.Files.QA, filepath.Ext(cfg.Files.QA))
	}
	return "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片"
}

func GetQaIndexPath(cfg *config.Config) string {
	if cfg != nil && cfg.Files.QAIndex != "" {
		return cfg.Files.QAIndex
	}
	return filepath.Join(GetQaDir(cfg), "问答知识卡片索引.md")
}

func UseStandaloneQA(cfg *config.Config) bool {
	return cfg != nil && cfg.Files.QADir != "" && cfg.Files.QAIndex != ""
}

func GetMarketCasePath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/个人投资训练系统/03-知识与案例/市场案例库.md"
	}
	return cfg.Files.MarketCase
}

func GetCandidateRulePath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/个人投资训练系统/04-投资规则/候选规则.md"
	}
	return cfg.Files.CandidateRule
}

func GetCandidateRuleDir(cfg *config.Config) string {
	if cfg != nil && cfg.Files.CandidateRuleDir != "" {
		return cfg.Files.CandidateRuleDir
	}
	if cfg != nil && cfg.Files.CandidateRule != "" {
		return strings.TrimSuffix(cfg.Files.CandidateRule, filepath.Ext(cfg.Files.CandidateRule))
	}
	return "日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则"
}

func GetCandidateRuleIndexPath(cfg *config.Config) string {
	if cfg != nil && cfg.Files.CandidateRuleIndex != "" {
		return cfg.Files.CandidateRuleIndex
	}
	return filepath.Join(GetCandidateRuleDir(cfg), "候选规则索引.md")
}

func UseStandaloneCandidateRules(cfg *config.Config) bool {
	return cfg != nil && cfg.Files.CandidateRuleDir != "" && cfg.Files.CandidateRuleIndex != ""
}

func GetMacroKnowledgeDir(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/宏观理解卡"
	}
	return cfg.Files.MacroKnowledgeDir
}

func GetMacroKnowledgeIndexPath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/宏观理解卡/宏观理解卡索引.md"
	}
	return cfg.Files.MacroKnowledgeIndex
}

func GetMarketObservationDir(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/市场观察卡"
	}
	return cfg.Files.MarketObservationDir
}

func GetMarketObservationIndexPath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/宽基指数仓位管理系统/02-观点/市场观察卡/市场观察卡索引.md"
	}
	return cfg.Files.MarketObservationIndex
}

// GetKnowRelativePath 返回单个 KNOW 文件的相对路径（不含 vault 前缀）
// V1.5.1 起独立文件名只使用稳定 ID，正文标题保留完整标题。
func GetKnowRelativePath(cfg *config.Config, knowID, title string) string {
	dir := GetMacroKnowledgeDir(cfg)
	return filepath.Join(dir, knowID+".md")
}

func KnowLink(cfg *config.Config, knowID, title string) string {
	return ObsidianFileLink(GetKnowRelativePath(cfg, knowID, title), linkAlias(knowID, title))
}

// GetObsRelativePath 返回单个 OBS 文件的相对路径（不含 vault 前缀）
func GetObsRelativePath(cfg *config.Config, obsID, title string) string {
	dir := GetMarketObservationDir(cfg)
	return fmt.Sprintf("%s/%s｜%s.md", dir, obsID, title)
}

func JoinHeading(id, title string) string {
	return fmt.Sprintf("%s｜%s", id, title)
}

func RawMaterialFileName(rawID, title string) string {
	return rawID + ".md"
}

func RawMaterialRelativePath(cfg *config.Config, rawID, title string) string {
	return filepath.Join(GetRawMaterialDir(cfg), RawMaterialFileName(rawID, title))
}

func RawMaterialLink(cfg *config.Config, rawID, title string, alias string) string {
	if UseStandaloneRawMaterials(cfg) {
		if alias == "" || alias == rawID {
			alias = linkAlias(rawID, title)
		}
		return ObsidianFileLink(RawMaterialRelativePath(cfg, rawID, title), alias)
	}
	if alias == "" {
		alias = rawID
	}
	return ObsidianHeadingLink(GetRawMaterialPath(cfg), JoinHeading(rawID, title), alias)
}

func QaFileName(qaID, title string) string {
	return qaID + ".md"
}

func QaRelativePath(cfg *config.Config, qaID, title string) string {
	return filepath.Join(GetQaDir(cfg), QaFileName(qaID, title))
}

func QaLink(cfg *config.Config, qaID, title string, alias string) string {
	if UseStandaloneQA(cfg) {
		if alias == "" || alias == qaID {
			alias = linkAlias(qaID, title)
		}
		return ObsidianFileLink(QaRelativePath(cfg, qaID, title), alias)
	}
	if alias == "" {
		alias = qaID
	}
	return ObsidianHeadingLink(GetQaPath(cfg), JoinHeading(qaID, title), alias)
}

func JoinCandidateRuleHeading(crID string, domainCode, topicCode, ruleName string) string {
	shortCode := domainCode + "-" + topicCode
	return fmt.Sprintf("%s｜%s｜%s", crID, shortCode, ruleName)
}

func CandidateRuleFileName(crID string, domainCode, topicCode, ruleName string) string {
	return crID + ".md"
}

func CandidateRuleRelativePath(cfg *config.Config, crID string, domainCode, topicCode, ruleName string) string {
	return filepath.Join(GetCandidateRuleDir(cfg), CandidateRuleFileName(crID, domainCode, topicCode, ruleName))
}

func CandidateRuleLink(cfg *config.Config, crID string, domainCode, topicCode, ruleName string, alias string) string {
	heading := JoinCandidateRuleHeading(crID, domainCode, topicCode, ruleName)
	if UseStandaloneCandidateRules(cfg) {
		if alias == "" || alias == crID || alias == heading {
			alias = linkAlias(crID, ruleName)
		}
		return ObsidianFileLink(CandidateRuleRelativePath(cfg, crID, domainCode, topicCode, ruleName), alias)
	}
	if alias == "" {
		alias = crID
	}
	return ObsidianHeadingLink(GetCandidateRulePath(cfg), heading, alias)
}

func SimilarRuleLink(cfg *config.Config, crID, shortCode, ruleName string) string {
	alias := linkAlias(crID, ruleName)
	if UseStandaloneCandidateRules(cfg) {
		parts := strings.SplitN(shortCode, "-", 2)
		domainCode := shortCode
		topicCode := ""
		if len(parts) == 2 {
			domainCode = parts[0]
			topicCode = parts[1]
		}
		return ObsidianFileLink(CandidateRuleRelativePath(cfg, crID, domainCode, topicCode, ruleName), alias)
	}
	return ObsidianHeadingLink(GetCandidateRulePath(cfg), crID+"｜"+shortCode+"｜"+ruleName, alias)
}

func ValidationCardLink(cfg *config.Config, crID string, alias string) string {
	if alias == "" {
		alias = crID
	}
	return ObsidianFileLink(GetValidationCardPath(cfg, crID), alias)
}

func linkAlias(id, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return id
	}
	return JoinHeading(id, shortLinkTitle(title))
}

func shortLinkTitle(title string) string {
	const maxRunes = 18
	runes := []rune(strings.TrimSpace(title))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes])
}

func sanitizeMarkdownFileName(name string) string {
	replacer := strings.NewReplacer(
		"\\", "／",
		"/", "／",
		":", "：",
		"*", "＊",
		"?", "？",
		"\"", "'",
		"<", "＜",
		">", "＞",
		"|", "｜",
	)
	name = strings.TrimSpace(replacer.Replace(name))
	const maxRunes = 110
	runes := []rune(name)
	if len(runes) > maxRunes {
		name = string(runes[:maxRunes])
		for !utf8.ValidString(name) {
			name = name[:len(name)-1]
		}
	}
	return name
}
