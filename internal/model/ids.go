package model

// DocumentIDs 是生成的文档编号
type DocumentIDs struct {
	RawID        string   // 原始材料编号，如 RAW-POS-SAFETY-20260609-001
	QAID         string   // 知识卡片编号，如 QA-POS-SAFETY-20260609-001
	KNOWID       string   // 宏观理解卡编号，如 KNOW-L3-RATE-20260701-001
	OBSID        string   // 市场观察卡编号，如 OBS-L2-POLICY-20260701-001
	CaseID       string   // 市场案例编号，如 CASE-INDEX-DD-20260609-001（可能为空）
	CandidateIDs []string // 候选规则编号列表
}