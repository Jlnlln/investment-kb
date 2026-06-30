package markdown

import (
	"fmt"
	"strings"

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

func JoinHeading(id, title string) string {
	return fmt.Sprintf("%s｜%s", id, title)
}

func JoinCandidateRuleHeading(crID string, domainCode, topicCode, ruleName string) string {
	shortCode := domainCode + "-" + topicCode
	return fmt.Sprintf("%s｜%s｜%s", crID, shortCode, ruleName)
}
