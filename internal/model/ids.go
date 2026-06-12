package model

// DocumentIDs 是生成的文档编号
type DocumentIDs struct {
	RawID        string   // 原始材料编号，如 RAW-POS-SAFETY-20260609-001
	QAID         string   // 知识卡片编号，如 QA-POS-SAFETY-20260609-001
	CaseID       string   // 市场案例编号，如 CASE-INDEX-DD-20260609-001（可能为空）
	CandidateIDs []string // 候选规则编号列表
}