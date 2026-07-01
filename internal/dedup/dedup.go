package dedup

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SimilarRule 表示一条与新规则相似的已有候选规则
type SimilarRule struct {
	CRID      string
	ShortCode string // domain-topic
	RuleName  string
	Reason    string // 相似原因
	Level     string // 完全重复 / 高度相似 / 可能相似
}

// RuleFingerprint 规则指纹，用于去重判断
type RuleFingerprint struct {
	CRID           string
	ShortCode      string
	DomainCode     string
	TopicCode      string
	RuleName       string
	Triggers       []string
	Actions        []string
	ExactHash      string // 精确指纹：sha256(domain + normalized_name + normalized_triggers + normalized_actions)
	SemanticKey    string // 语义指纹：domain-topic + 核心关键词
}

// normalizeText 标准化文本：去空格、去标点、小写
func normalizeText(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	// 去掉常见标点
	for _, ch := range []string{",", ".", "，", "。", "、", "！", "！", "？", "?", "：", ":", "；", ";"} {
		text = strings.ReplaceAll(text, ch, "")
	}
	text = strings.ReplaceAll(text, " ", "")
	return text
}

// ComputeExactHash 计算精确指纹
func ComputeExactHash(domain, ruleName string, triggers, actions []string) string {
	normalized := normalizeText(domain) + normalizeText(ruleName) +
		normalizeText(strings.Join(triggers, "")) +
		normalizeText(strings.Join(actions, ""))
	hash := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", hash)[:16] // 取前 16 位够用
}

// ComputeSemanticKey 计算语义指纹
func ComputeSemanticKey(domain, topicCode string, triggers []string) string {
	// 从 trigger_conditions 中提取核心关键词
	// 简化版：取前 3 个触发条件的前 4 个字
	coreKeywords := make([]string, 0, 3)
	for i, trigger := range triggers {
		if i >= 3 {
			break
		}
		runes := []rune(trigger)
		if len(runes) > 4 {
			coreKeywords = append(coreKeywords, string(runes[:4]))
		} else {
			coreKeywords = append(coreKeywords, trigger)
		}
	}
	return domain + "-" + topicCode + "-" + strings.Join(coreKeywords, "-")
}

// ParseExistingCRs 从候选规则库 markdown 文件中解析已有 CR 的指纹
func ParseExistingCRs(crLibraryPath string) ([]RuleFingerprint, error) {
	data, err := os.ReadFile(crLibraryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在 = 没有已有规则
		}
		return nil, fmt.Errorf("读取候选规则库失败: %w", err)
	}

	content := string(data)
	if content == "" || strings.TrimSpace(content) == "" {
		return nil, nil
	}

	// 按 --- 分割，找到每个 CR 段
	// 每个 CR 段以 "# CR-XXX｜DOMAIN-TOPIC｜rule_name" 开头
	// 候选规则库使用 --- 分隔章节，同一条 CR 的标题、触发条件、动作
	// 可能被拆到不同 section。因此改为逐行解析，跟踪当前 CR 上下文。
	crTitlePattern := regexp.MustCompile(`^#\s+(CR-[A-Z]+-\d{8}-\d{3})｜([A-Z]+-[A-Z]+)｜(.+)$`)

	fingerprints := make([]RuleFingerprint, 0)
	lines := strings.Split(content, "\n")

	var crID, shortCode, ruleName string
	var domainCode, topicCode string
	var triggers, actions []string
	currentSection := ""

	flushCurrentCR := func() {
		if crID == "" {
			return
		}
		fp := RuleFingerprint{
			CRID:        crID,
			ShortCode:   shortCode,
			DomainCode:  domainCode,
			TopicCode:   topicCode,
			RuleName:    ruleName,
			Triggers:    triggers,
			Actions:     actions,
			ExactHash:   ComputeExactHash(domainCode, ruleName, triggers, actions),
			SemanticKey: ComputeSemanticKey(domainCode, topicCode, triggers),
		}
		fingerprints = append(fingerprints, fp)
		// 重置 CR 级变量
		crID = ""
		shortCode = ""
		ruleName = ""
		domainCode = ""
		topicCode = ""
		triggers = nil
		actions = nil
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// --- 是 CR 内部章节分隔符，不表示 CR 结束。跳过即可。
		if line == "---" {
			currentSection = ""
			continue
		}

		// 匹配 CR 标题行 → 新 CR 开始，先 flush 前一个
		matches := crTitlePattern.FindStringSubmatch(line)
		if len(matches) == 4 {
			flushCurrentCR()
			crID = matches[1]
			shortCode = matches[2]
			ruleName = matches[3]
			parts := strings.SplitN(shortCode, "-", 2)
			if len(parts) == 2 {
				domainCode = parts[0]
				topicCode = parts[1]
			}
			currentSection = ""
			continue
		}

		// 跳过无 CR 上下文的行
		if crID == "" {
			continue
		}

		// 识别当前章节
		if strings.HasPrefix(line, "## 2. 触发条件") {
			currentSection = "triggers"
			continue
		}
		if strings.HasPrefix(line, "## 3. 执行动作") {
			currentSection = "actions"
			continue
		}
		if strings.HasPrefix(line, "## ") {
			currentSection = ""
			continue
		}

		// 收集章节内的条目
		if strings.HasPrefix(line, "- ") {
			item := strings.TrimPrefix(line, "- ")
			switch currentSection {
			case "triggers":
				triggers = append(triggers, item)
			case "actions":
				actions = append(actions, item)
			}
		}
	}

	// 最后一条 CR（没有尾部 --- 的情况）
	flushCurrentCR()

	return fingerprints, nil
}

// CheckSimilarRules 检查新规则与已有规则的相似性
// 返回相似规则列表，按相似度排序
func CheckSimilarRules(
	newDomainCode, newTopicCode, newRuleName string,
	newTriggers, newActions []string,
	existingCRs []RuleFingerprint,
) []SimilarRule {
	similarRules := make([]SimilarRule, 0)

	newExactHash := ComputeExactHash(newDomainCode, newRuleName, newTriggers, newActions)
	_ = ComputeSemanticKey(newDomainCode, newTopicCode, newTriggers) // 语义指纹暂未使用，保留计算能力

	for _, existing := range existingCRs {
		// 1. 精确指纹匹配 = 完全重复
		if newExactHash == existing.ExactHash {
			similarRules = append(similarRules, SimilarRule{
				CRID:      existing.CRID,
				ShortCode: existing.ShortCode,
				RuleName:  existing.RuleName,
				Reason:    "精确指纹完全一致，可能是同一规则的不同版本",
				Level:     "完全重复",
			})
			continue
		}

		// 2. 同 domain-topic + rule_name 关键词重叠 = 高度相似
		if existing.DomainCode == newDomainCode && existing.TopicCode == newTopicCode {
			nameOverlap := computeKeywordOverlap(newRuleName, existing.RuleName)
			triggerOverlap := computeListOverlap(newTriggers, existing.Triggers)

			if nameOverlap >= 0.5 || triggerOverlap >= 0.4 {
				reason := fmt.Sprintf("同属 %s-%s，规则名称关键词重叠 %.0f%%，触发条件重叠 %.0f%%",
					newDomainCode, newTopicCode, nameOverlap*100, triggerOverlap*100)
				level := "疑似相似"
				// 高度相似：名称相似 + 触发条件也相似
				if nameOverlap >= 0.7 && triggerOverlap >= 0.4 {
					level = "高度相似"
				} else if nameOverlap >= 0.5 && triggerOverlap < 0.4 {
					// 名称相似但触发条件重叠低，需人工确认
					level = "疑似相似"
					reason += "。名称相似但触发条件重叠低，需人工确认是否真重复"
				}
				similarRules = append(similarRules, SimilarRule{
					CRID:      existing.CRID,
					ShortCode: existing.ShortCode,
					RuleName:  existing.RuleName,
					Reason:    reason,
					Level:     level,
				})
				continue
			}
		}

		// 3. 不同 domain-topic 但触发条件关键词重叠 = 可能相似
		triggerOverlap := computeListOverlap(newTriggers, existing.Triggers)
		if triggerOverlap >= 0.5 && existing.DomainCode != newDomainCode {
			reason := fmt.Sprintf("触发条件关键词重叠 %.0f%%，但领域不同（%s vs %s），可能是跨领域相关规则",
				triggerOverlap*100, newDomainCode+"-"+newTopicCode, existing.ShortCode)
			similarRules = append(similarRules, SimilarRule{
				CRID:      existing.CRID,
				ShortCode: existing.ShortCode,
				RuleName:  existing.RuleName,
				Reason:    reason,
				Level:     "可能相似",
			})
		}
	}

	return similarRules
}

// computeKeywordOverlap 计算两个短文本的关键词重叠度
// 返回 0-1 之间的值，1 表示完全重叠
func computeKeywordOverlap(text1, text2 string) float64 {
	// 简化版：把文本拆成 2-4 字的片段，计算重叠比例
	words1 := extractFragments(text1)
	words2 := extractFragments(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	overlap := 0
	for w := range words1 {
		if words2[w] {
			overlap++
		}
	}

	// 重叠度 = 重叠数 / 较小集合的大小
	smallerSize := len(words1)
	if len(words2) < smallerSize {
		smallerSize = len(words2)
	}
	if smallerSize == 0 {
		return 0
	}

	return float64(overlap) / float64(smallerSize)
}

// extractFragments 从文本中提取 2 字和 3 字片段
func extractFragments(text string) map[string]bool {
	fragments := make(map[string]bool)
	runes := []rune(text)

	// 提取 2 字片段
	for i := 0; i <= len(runes)-2; i++ {
		fragments[string(runes[i:i+2])] = true
	}
	// 提取 3 字片段
	for i := 0; i <= len(runes)-3; i++ {
		fragments[string(runes[i:i+3])] = true
	}

	return fragments
}

// computeListOverlap 计算两个字符串列表的关键词重叠度
func computeListOverlap(list1, list2 []string) float64 {
	if len(list1) == 0 || len(list2) == 0 {
		return 0
	}

	// 合并所有文本，然后计算关键词重叠
	text1 := strings.Join(list1, " ")
	text2 := strings.Join(list2, " ")

	return computeKeywordOverlap(text1, text2)
}
