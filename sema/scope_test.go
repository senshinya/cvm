package sema

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestSymbolFields(t *testing.T) {
	pos := entity.SourcePos{Line: 1, Column: 2}
	sym := &Symbol{
		Name:    "x",
		Kind:    SymVar,
		T:       NewTypeTable().Builtin(Int),
		Storage: StorageAuto,
		Linkage: LinkageNone,
		Pos:     pos,
	}
	if sym.Name != "x" || sym.Kind != SymVar {
		t.Fatalf("symbol fields not stored: %+v", sym)
	}
}

func TestScopeLookup(t *testing.T) {
	root := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	root.Insert("x", &Symbol{Name: "x", Kind: SymVar, T: intT})

	inner := NewScope(ScopeBlock, root)
	if got := inner.Lookup("x", NSOrdinary); got == nil {
		t.Fatalf("inner.Lookup did not walk parent chain")
	}
	if got := inner.LookupCurrent("x", NSOrdinary); got != nil {
		t.Fatalf("inner.LookupCurrent returned outer symbol")
	}
}

func TestScopeNamespaceIsolation(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	s.Insert("foo", &Symbol{Name: "foo", Kind: SymVar, T: intT})
	tagId := NewTypeTable().NewTagID()
	s.InsertTag("foo", &TagInfo{Tag: tagId})
	if s.Lookup("foo", NSOrdinary) == nil {
		t.Fatalf("ordinary 'foo' lost")
	}
	if s.LookupTag("foo") == nil {
		t.Fatalf("tag 'foo' lost")
	}
}
