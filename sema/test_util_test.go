package sema

import (
	"testing"

	"shinya.click/cvm/entity"
)

func findFuncDef(t *testing.T, prog *Program, name string) *FuncDef {
	t.Helper()
	for _, fn := range prog.Funcs {
		if fn.Sym != nil && fn.Sym.Name == name {
			return fn
		}
	}
	t.Fatalf("function %q not found", name)
	return nil
}

func findFirstNode(node *entity.AstNode, typ entity.TokenType) *entity.AstNode {
	if node == nil {
		return nil
	}
	if node.Typ == typ {
		return node
	}
	for _, c := range node.Children {
		if got := findFirstNode(c, typ); got != nil {
			return got
		}
	}
	return nil
}
