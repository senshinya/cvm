package sema

import (
	"testing"

	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
)

func analyzeSource(t *testing.T, src string) *SemaResult {
	t.Helper()
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	survivors, _ := PreFilter(candidates)
	if len(survivors) == 0 {
		t.Fatal("no surviving AST candidate")
	}
	return NewSema().Analyze(survivors[0])
}

func TestAnalyzeTopLevelDecls(t *testing.T) {
	r := analyzeSource(t, "int x; double y;")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if len(r.Program.Globals) != 2 {
		t.Fatalf("expected 2 globals, got %d", len(r.Program.Globals))
	}
	v0 := r.Program.Globals[0].(*VarDecl)
	if v0.Sym.Name != "x" {
		t.Fatalf("first var name wrong: %+v", v0)
	}
}

func TestAnalyzeTypedefAndFunctionDef(t *testing.T) {
	r := analyzeSource(t, "typedef int Int; Int f(Int x) { return x; }")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if len(r.Program.Funcs) != 1 || r.Program.Funcs[0].Sym.Name != "f" {
		t.Fatalf("function not registered: %+v", r.Program.Funcs)
	}
	if len(r.Program.Funcs[0].Params) != 1 || r.Program.Funcs[0].Params[0].Sym.Name != "x" {
		t.Fatalf("params wrong: %+v", r.Program.Funcs[0].Params)
	}
}
