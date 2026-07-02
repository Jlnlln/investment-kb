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

func GetQaPath(cfg *config.Config) string {
	if cfg == nil {
		return "日常随笔/股市学习/个人投资训练系统/03-知识与案例/问答知识库.md"
	}
	return cfg.Files.QA
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
// 格式：宏观理解卡目录/KNOW-ID｜title.md
func GetKnowRelativePath(cfg *config.Config, knowID, title string) string {
	dir := GetMacroKnowledgeDir(cfg)
	return fmt.Sprintf("%s/%s｜%s.md", dir, knowID, title)
}

func KnowLink(cfg *config.Config, knowID, title string) string {
	return ObsidianFileLink(GetKnowRelativePath(cfg, knowID, title), knowID)
}

// GetObsRelativePath 返回单个 OBS 文件的相对路径（不含 vault 前缀）
func GetObsRelativePath(cfg *config.Config, obsID, title string) string {
	dir := GetMarketObservationDir(cfg)
	return fmt.Sprintf("%s/%s｜%s.md", dir, obsID, title)
}

func JoinHeading(id, title string) string {
	return fmt.Sprintf("%s｜%s", id, title)
}

func JoinCandidateRuleHeading(crID string, domainCode, topicCode, ruleName string) string {
	shortCode := domainCode + "-" + topicCode
	return fmt.Sprintf("%s｜%s｜%s", crID, shortCode, ruleName)
}

func CandidateRuleFileName(crID string, domainCode, topicCode, ruleName string) string {
	return sanitizeMarkdownFileName(JoinCandidateRuleHeading(crID, domainCode, topicCode, ruleName)) + ".md"
}

func CandidateRuleRelativePath(cfg *config.Config, crID string, domainCode, topicCode, ruleName string) string {
	return filepath.Join(GetCandidateRuleDir(cfg), CandidateRuleFileName(crID, domainCode, topicCode, ruleName))
}

func CandidateRuleLink(cfg *config.Config, crID string, domainCode, topicCode, ruleName string, alias string) string {
	heading := JoinCandidateRuleHeading(crID, domainCode, topicCode, ruleName)
	if alias == "" {
		alias = crID // 使用 CR-ID 作为别名（更简洁）
	}
	if UseStandaloneCandidateRules(cfg) {
		return ObsidianFileLink(CandidateRuleRelativePath(cfg, crID, domainCode, topicCode, ruleName), alias)
	}
	return ObsidianHeadingLink(GetCandidateRulePath(cfg), heading, alias)
}

func SimilarRuleLink(cfg *config.Config, crID, shortCode, ruleName string) string {
	alias := crID // 使用 CR-ID 作为别名
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
