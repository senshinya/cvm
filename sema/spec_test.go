package sema

import (
	"testing"

	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func TestParseSpecC99ArithmeticAndQualifiers(t *testing.T) {
	tokens, _ := lexer.NewLexer("typedef const unsigned long T;").ScanTokens()
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	decl := findFirstNode(candidates[0], parser.Declaration)
	got := NewSema().parseSpec(decl.Children[0])
	if !got.IsTypedef {
		t.Fatal("expected typedef")
	}
	q, ok := got.Type.(*QualType)
	if !ok || !q.Const {
		t.Fatalf("expected const-qualified type, got %T", got.Type)
	}
	bt, ok := q.Base.(*BuiltinType)
	if !ok || bt.Kind != ULong {
		t.Fatalf("expected unsigned long, got %v", q.Base)
	}
}

func TestBuildStructAndEnum(t *testing.T) {
	r := analyzeSource(t, "struct S { int x; int y; } s; enum E { A, B = 5, C } e;")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if len(r.Program.Globals) != 2 {
		t.Fatalf("expected 2 globals, got %d", len(r.Program.Globals))
	}
	st := r.Program.Globals[0].(*VarDecl).T.(*StructType)
	if !st.Complete || len(st.Fields) != 2 || st.Fields[1].Name != "y" {
		t.Fatalf("struct fields wrong: %+v", st.Fields)
	}
	et := r.Program.Globals[1].(*VarDecl).T.(*EnumType)
	if len(et.Enumerators) != 3 || et.Enumerators[1].Value != 5 || et.Enumerators[2].Value != 6 {
		t.Fatalf("enum values wrong: %+v", et.Enumerators)
	}
}
