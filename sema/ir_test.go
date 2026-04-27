package sema

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestStmtPosImplemented(t *testing.T) {
	pos := entity.SourceRange{SourceStart: entity.SourcePos{Line: 1}}
	b := &Block{Range: pos}
	if b.Pos().SourceStart.Line != 1 {
		t.Fatalf("Block.Pos lost source range")
	}
}

func TestExprNodeBasics(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	lit := &IntLit{Value: 42, T: intT}
	if lit.GetType() != intT {
		t.Fatalf("IntLit.GetType wrong")
	}
	if lit.GetCategory() != RValue {
		t.Fatalf("IntLit must be RValue")
	}
}

func TestProgramAndDecls(t *testing.T) {
	tt := NewTypeTable()
	st := NewSymbolTable()
	prog := &Program{
		Types:  tt,
		SymTab: st,
		Funcs:  []*FuncDef{{Sym: &Symbol{Name: "main"}}},
	}
	if len(prog.Funcs) != 1 || prog.Funcs[0].Sym.Name != "main" {
		t.Fatalf("Program.Funcs not stored")
	}
}
