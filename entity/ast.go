package entity

import (
	"fmt"
	"strings"
)

type AstNode struct {
	Parent           *AstNode
	Typ              TokenType
	Terminal         *Token
	PossibleBranches []*Branch
	SourceRange
}

type Branch struct {
	Production Production
	Children   []*AstNode
}

func (n *AstNode) SetBranch(production Production, children []*AstNode) {
	branch := &Branch{Production: production, Children: children}
	// set parent later in case of deepcopy error
	n.PossibleBranches = append(n.PossibleBranches, branch)
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
