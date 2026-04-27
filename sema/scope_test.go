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

func TestScopeInsertRedeclarationError(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	pos1 := entity.SourcePos{Line: 1}
	pos2 := entity.SourcePos{Line: 2}
	if err := s.InsertChecked("x", &Symbol{Name: "x", Kind: SymVar, T: intT, Pos: pos1}); err != nil {
		t.Fatalf("first insert errored: %v", err)
	}
	if err := s.InsertChecked("x", &Symbol{Name: "x", Kind: SymTypedef, T: intT, Pos: pos2}); err == nil {
		t.Fatalf("second insert with different kind should error")
	}
}

func TestSymbolTableFileScope(t *testing.T) {
	st := NewSymbolTable()
	if st.File == nil {
		t.Fatalf("SymbolTable.File not initialized")
	}
	if st.File.Kind != ScopeFile {
		t.Fatalf("file scope wrong kind: %v", st.File.Kind)
	}
}

func TestScopeInsertTagChecked(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	tag1 := &TagID{}
	tag2 := &TagID{}
	pos1 := entity.SourcePos{Line: 1}
	pos2 := entity.SourcePos{Line: 2}
	st1 := &StructType{Tag: tag1}
	st2 := &StructType{Tag: tag2}
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag1, T: st1}, pos1); err != nil {
		t.Fatalf("first tag insert errored: %v", err)
	}
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag1, T: st1}, pos2); err != nil {
		t.Fatalf("same-tag re-insert errored: %v", err)
	}
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag2, T: st2}, pos2); err == nil {
		t.Fatalf("conflicting tag should error")
	}
}
