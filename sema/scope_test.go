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

func TestScopeOrdinaryIdentifierShadowing(t *testing.T) {
	root := NewScope(ScopeFile, nil)
	inner := NewScope(ScopeBlock, root)
	intT := NewTypeTable().Builtin(Int)
	root.Insert("x", &Symbol{Name: "x", Kind: SymVar, T: intT, Pos: entity.SourcePos{Line: 1}})
	inner.Insert("x", &Symbol{Name: "x", Kind: SymVar, T: intT, Pos: entity.SourcePos{Line: 2}})
	if got := inner.Lookup("x", NSOrdinary); got == nil || got.Pos.Line != 2 {
		t.Fatalf("inner scope did not shadow ordinary identifier: %+v", got)
	}
	if got := root.Lookup("x", NSOrdinary); got == nil || got.Pos.Line != 1 {
		t.Fatalf("outer scope was overwritten by shadow: %+v", got)
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

func TestScopeIgnoresUnsupportedNamespaces(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	s.Insert("label", &Symbol{Name: "label", Kind: SymVar})
	if got := s.Lookup("label", NSLabel); got != nil {
		t.Fatalf("Scope should not manage C labels: %+v", got)
	}
	if got := s.LookupCurrent("label", NSMember); got != nil {
		t.Fatalf("Scope should not manage struct members: %+v", got)
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

func TestScopeInsertSameKindMergesDefs(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	firstDecl := &FuncDecl{}
	secondDecl := &FuncDecl{}
	first := &Symbol{Name: "f", Kind: SymFunc, T: intT, Decl: firstDecl}
	second := &Symbol{Name: "f", Kind: SymFunc, T: intT, Decl: secondDecl}
	if err := s.InsertChecked("f", first); err != nil {
		t.Fatalf("first insert errored: %v", err)
	}
	if err := s.InsertChecked("f", second); err != nil {
		t.Fatalf("same-kind insert errored: %v", err)
	}
	if got := s.LookupCurrent("f", NSOrdinary); got != first {
		t.Fatalf("same-kind merge replaced original symbol")
	}
	if len(first.Defs) != 1 || first.Defs[0] != secondDecl {
		t.Fatalf("same-kind merge did not append second declaration: %+v", first.Defs)
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

func TestScopeInsertTagCheckedRejectsC99TagKindMismatch(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	tag := &TagID{}
	if err := s.InsertTagChecked("T", &TagInfo{Tag: tag, T: &StructType{Tag: tag}}, entity.SourcePos{Line: 1}); err != nil {
		t.Fatalf("struct tag insert errored: %v", err)
	}
	if err := s.InsertTagChecked("T", &TagInfo{Tag: tag, T: &UnionType{Tag: tag}}, entity.SourcePos{Line: 2}); err == nil {
		t.Fatalf("same C tag name changed from struct to union without error")
	}
	if err := s.InsertTagChecked("T", &TagInfo{Tag: tag, T: &EnumType{Tag: tag}}, entity.SourcePos{Line: 3}); err == nil {
		t.Fatalf("same C tag name changed from struct to enum without error")
	}
}

func TestScopeTagShadowingUsesCurrentScopeOnlyForInsert(t *testing.T) {
	root := NewScope(ScopeFile, nil)
	inner := NewScope(ScopeBlock, root)
	outerTag := &TagID{}
	innerTag := &TagID{}
	if err := root.InsertTagChecked("S", &TagInfo{Tag: outerTag, T: &StructType{Tag: outerTag}}, entity.SourcePos{Line: 1}); err != nil {
		t.Fatalf("outer tag insert errored: %v", err)
	}
	if err := inner.InsertTagChecked("S", &TagInfo{Tag: innerTag, T: &StructType{Tag: innerTag}}, entity.SourcePos{Line: 2}); err != nil {
		t.Fatalf("inner tag should shadow outer tag, got: %v", err)
	}
	if got := inner.LookupTag("S"); got == nil || got.Tag != innerTag {
		t.Fatalf("inner lookup did not find shadowing tag: %+v", got)
	}
	if got := root.LookupTag("S"); got == nil || got.Tag != outerTag {
		t.Fatalf("outer tag was overwritten: %+v", got)
	}
}
