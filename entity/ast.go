package entity

import (
	"fmt"
	"github.com/hyphennn/glambda/gslice"
	"strings"
)

type AstNode struct {
	Parent     *AstNode
	Typ        TokenType
	Terminal   *Token
	Production Production
	Children   []*AstNode
	GLRLabel
	SourceRange
}
type GLRLabel struct {
	// Declaration 使用，规约出 Declaration 后消除
	TypeDef      bool     // 是否是 TypeDef
	DeclaratorID []*Token // 包含的 Identifier
}

func (n *AstNode) ReducedBy(left TokenType, index int64) bool {
	return n.Production.Left == left && n.Production.Index == index
}

func (n *AstNode) SetChildren(production Production, children []*AstNode) {
	// set parent later in case of deepcopy error
	n.Production = production
	n.Children = children
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
