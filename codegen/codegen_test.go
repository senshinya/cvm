package codegen

import (
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/sema"
)

func TestGenerateMinimalReturn(t *testing.T) {
	mod := compileModule(t, `int main(void) { return 0; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		`Global #0 func name="main"`,
		"I32Const 0",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func compileModule(t *testing.T, source string) *bytecode.Module {
	t.Helper()

	tokens, err := lexer.NewLexer(source).ScanTokens()
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	prog, err := sema.Analyze(candidates)
	if err != nil {
		t.Fatalf("sema: %v", err)
	}
	mod, err := Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	return mod
}
