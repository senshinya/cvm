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

func TestGenerateReturnCastsIntLiteralToLong(t *testing.T) {
	mod := compileModule(t, `long main(void) { return 0; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"Cast i32->i64 SExt",
		"I64Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateReturnCastsIntLiteralToBool(t *testing.T) {
	mod := compileModule(t, `_Bool main(void) { return 0; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"Cast i32->bool Bool",
		"BoolReturn",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateLocalArithmetic(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	int x = 1;
	int y = 2;
	return x + y * 3;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"I32StoreLocal 0",
		"I32StoreLocal 1",
		"I32LoadLocal 0",
		"I32LoadLocal 1",
		"I32Mul",
		"I32Add",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestCollectGlobalsIncludesStaticLocals(t *testing.T) {
	prog := analyzeProgram(t, `
int g;
int main(void) {
	static int s;
	return 0;
}
`)
	g := &generator{
		prog:      prog,
		mod:       &bytecode.Module{Target: bytecode.DefaultTarget()},
		globalMap: map[*sema.Symbol]int{},
		sigMap:    map[string]int{},
		layoutMap: map[sema.Type]int{},
	}
	if err := g.collectGlobals(); err != nil {
		t.Fatal(err)
	}
	if got, want := len(g.mod.Globals), 3; got != want {
		t.Fatalf("global count = %d, want %d: %#v", got, want, g.mod.Globals)
	}
	for i, global := range g.mod.Globals {
		if global.Name == "" {
			t.Fatalf("global %d is an empty placeholder: %#v", i, g.mod.Globals)
		}
	}
	staticID := prog.Funcs[0].Locals[0].Sym.GlobalID
	if got := g.mod.Globals[staticID]; got.Kind != bytecode.GlobalVar || got.Name != "s" {
		t.Fatalf("static local global = %#v, want var named s", got)
	}
}

func TestGenerateReusesFunctionPrototypeGlobal(t *testing.T) {
	mod := compileModule(t, `
int main(void);
int main(void) { return 0; }
`)
	out := bytecode.PrintModule(mod)
	if got := strings.Count(out, `Global #0 func name="main"`); got != 1 {
		t.Fatalf("main global count = %d, want 1:\n%s", got, out)
	}
	if !strings.Contains(out, "Func #0 global=0") {
		t.Fatalf("definition did not reuse prototype global:\n%s", out)
	}
}

func compileModule(t *testing.T, source string) *bytecode.Module {
	t.Helper()

	prog := analyzeProgram(t, source)
	mod, err := Generate(prog)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	return mod
}

func analyzeProgram(t *testing.T, source string) *sema.Program {
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
	return prog
}
