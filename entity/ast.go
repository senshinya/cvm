package entity

import (
	"fmt"
	"github.com/hyphennn/glambda/gslice"
	"strings"
)

type AstNode struct {
	Parent           *AstNode
	Typ              TokenType
	Terminal         *Token
	PossibleBranches []*Branch
	GLRLabel
	SourceRange
}
type GLRLabel struct {
	// Declaration 使用，规约出 Declaration 后消除
	TypeDef      bool     // 是否是 TypeDef
	DeclaratorID []string // 包含的 Identifier
}

type Branch struct {
	Production Production
	Children   []*AstNode
}

func (n *AstNode) GetProductionBranch(left TokenType, index int64) *Branch {
	for _, branch := range n.PossibleBranches {
		if branch.Production.Left == left && branch.Production.Index == index {
			return branch
		}
	}
	return nil
}

func (n *AstNode) SetBranch(production Production, children []*AstNode) {
	branch := &Branch{Production: production, Children: children}
	// set parent later in case of deepcopy error
	n.PossibleBranches = append(n.PossibleBranches, branch)
	// set glr label
	gslice.ForEach(children, func(child *AstNode) {
		if child.TypeDef {
			n.TypeDef = true
		}
		n.DeclaratorID = append(n.DeclaratorID, child.DeclaratorID...)
	})
	// set source range
	n.SourceRange = SourceRange{
		SourceStart: children[0].SourceRange.SourceStart,
		SourceEnd:   children[len(children)-1].SourceRange.SourceEnd,
	}
}

func (n *AstNode) Merge(ast *AstNode) {
	if n.Typ != ast.Typ {
		return
	}
	branches := map[int64]*Branch{}
	for _, branch := range n.PossibleBranches {
		prod := branch.Production
		branches[prod.Index] = branch
	}
	for _, branch := range ast.PossibleBranches {
		prod := branch.Production
		originBranch, ok := branches[prod.Index]
		if ok {
			// same branch, merge recursively
			for i, originNode := range originBranch.Children {
				node := branch.Children[i]
				originNode.Merge(node)
			}
		}
		if !ok {
			// different branch, add
			n.PossibleBranches = append(n.PossibleBranches, branch)
		}
	}
}

type Production struct {
	Left  TokenType
	Index int64
	Right []TokenType
}

func (p Production) String() string {
	var rights []string
	for _, t := range p.Right {
		rights = append(rights, string(t))
	}
	return fmt.Sprintf("%s -> %s", p.Left, strings.Join(rights, " "))
}
