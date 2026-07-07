package screening

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDecisionsPath = "03-规则/候选规则/cr_screening_decisions.yaml"
	CandidateRuleIndex   = "03-规则/候选规则/候选规则索引.md"
	CandidateRuleDir     = "03-规则/候选规则"
)

type Paths struct {
	KBRoot       string
	DecisionsRel string
}

func NewPaths(kbRoot, decisions string) (Paths, error) {
	kbRoot = strings.TrimSpace(kbRoot)
	if kbRoot == "" {
		return Paths{}, fmt.Errorf("缺少必填参数: --kb-root")
	}
	absRoot, err := filepath.Abs(kbRoot)
	if err != nil {
		return Paths{}, err
	}
	decisions = strings.TrimSpace(decisions)
	if decisions == "" {
		decisions = DefaultDecisionsPath
	}
	return Paths{KBRoot: absRoot, DecisionsRel: filepath.FromSlash(decisions)}, nil
}

func (p Paths) Resolve(rel string) (string, error) {
	rel = filepath.Clean(filepath.FromSlash(rel))
	full := filepath.Join(p.KBRoot, rel)
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	root := filepath.Clean(p.KBRoot)
	if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("路径越界: %s", rel)
	}
	return abs, nil
}

func (p Paths) DecisionsPath() (string, error) {
	return p.Resolve(p.DecisionsRel)
}

func (p Paths) IndexPath() (string, error) {
	return p.Resolve(CandidateRuleIndex)
}

func (p Paths) CRPath(id string) (string, error) {
	return p.Resolve(filepath.ToSlash(filepath.Join(CandidateRuleDir, id+".md")))
}

func (p Paths) BackupRoot(timestamp string) (string, error) {
	return p.Resolve(filepath.ToSlash(filepath.Join(".backup", "cr_screening_"+timestamp)))
}

func (p Paths) GeneratedDecisionsPath() (string, error) {
	return p.Resolve(filepath.ToSlash(filepath.Join(".generated", "cr_screening_decisions.yaml")))
}
