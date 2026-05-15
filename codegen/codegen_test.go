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

func TestGenerateScalarComparisonResultTypes(t *testing.T) {
	tests := []string{
		`int main(void) { return 1 < 2; }`,
		`int main(void) { int x; x = 1 < 2; return x; }`,
		`_Bool main(void) { return 1 < 2; }`,
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			compileModule(t, source)
		})
	}
}

func TestGenerateScalarShiftNormalizesRHS(t *testing.T) {
	compileModule(t, `long main(void) { long x = 1; return x << 1; }`)
}

func TestGenerateScalarLogicalOps(t *testing.T) {
	tests := []string{
		`int main(void) { return 1 && 0; }`,
		`int main(void) { return 0 || 2; }`,
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			compileModule(t, source)
		})
	}
}

func TestGenerateAssignmentExpressionStatementResult(t *testing.T) {
	compileModule(t, `int main(void) { int x; x = 1; return x; }`)
}

func TestGenerateFileScopeExternVariableMetadata(t *testing.T) {
	mod := compileModule(t, `extern int g; int main(void) { return g; }`)
	if got := mod.Globals[0]; got.Kind != bytecode.GlobalExtern {
		t.Fatalf("global g kind = %v, want GlobalExtern: %#v", got.Kind, got)
	}
	out := bytecode.PrintModule(mod)
	if !strings.Contains(out, `Global #0 extern name="g"`) {
		t.Fatalf("bytecode missing extern global:\n%s", out)
	}
	if strings.Contains(out, `Global #0 var name="g"`) || strings.Contains(out, "init_zero=") {
		t.Fatalf("extern global printed as storage-owning var:\n%s", out)
	}
}

func TestGenerateGlobalsPointersAndArrayIndex(t *testing.T) {
	mod := compileModule(t, `
int g;
int a[3];
int main(void) {
	int *p = &g;
	*p = 4;
	return a[1] + g;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		`Global #0 var name="g"`,
		`Global #1 var name="a"`,
		"AddrGlobal 0",
		"I32Store",
		"I32Load",
		"PtrAdd",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
	if got := mod.Globals[0]; got.Size != 4 || got.Align != 4 {
		t.Fatalf("global g metadata = size %d align %d, want size 4 align 4", got.Size, got.Align)
	}
	if got := mod.Globals[1]; got.Size != 12 || got.Align != 4 {
		t.Fatalf("global a metadata = size %d align %d, want size 12 align 4", got.Size, got.Align)
	}
	var arrayLayout *bytecode.ObjectLayout
	for i := range mod.Layouts {
		if mod.Layouts[i].ElemSize == 4 {
			arrayLayout = &mod.Layouts[i]
			break
		}
	}
	if arrayLayout == nil {
		t.Fatalf("missing array layout with elem size 4: %#v", mod.Layouts)
	}
}

func TestGenerateAddressTakenScalarLocal(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	int x;
	int *p = &x;
	*p = 7;
	return x;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"AddrLocalObject",
		"PtrStoreLocal",
		"I32Store",
		"I32Load",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
	if got := mod.Functions[0].Objects; len(got) == 0 || got[0].Size != 4 || got[0].Align != 4 {
		t.Fatalf("address-taken local object metadata = %#v, want first object size 4 align 4", got)
	}
}

func TestGeneratePointerArithmeticOutsideIndexing(t *testing.T) {
	mod := compileModule(t, `
int a[3];
int main(void) {
	int *p = a;
	return *(p + 1);
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"PtrAdd elem_size=4",
		"I32Load",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGeneratePackedStructArrayLayoutMetadata(t *testing.T) {
	mod := compileModule(t, `
struct S { char c; int i; };
struct S a[2];
int main(void) {
	return a[1].i;
}`)
	var structLayout, arrayLayout *bytecode.ObjectLayout
	for i := range mod.Layouts {
		layout := &mod.Layouts[i]
		if len(layout.Fields) == 2 && layout.Fields[0].Name == "c" && layout.Fields[1].Name == "i" {
			structLayout = layout
		}
		if layout.ElemSize != 0 {
			arrayLayout = layout
		}
	}
	if structLayout == nil {
		t.Fatalf("missing struct layout: %#v", mod.Layouts)
	}
	if got := structLayout.Fields[1].Offset; got != 1 {
		t.Fatalf("field i offset = %d, want sema offset 1: %#v", got, structLayout)
	}
	if structLayout.Size != 5 || structLayout.Align != 1 {
		t.Fatalf("struct layout size/align = %d/%d, want sema-compatible 5/1: %#v", structLayout.Size, structLayout.Align, structLayout)
	}
	if arrayLayout == nil {
		t.Fatalf("missing array layout: %#v", mod.Layouts)
	}
	if arrayLayout.ElemSize != 5 || arrayLayout.Align != 1 {
		t.Fatalf("array layout elem size/align = %d/%d, want sema-compatible 5/1: %#v", arrayLayout.ElemSize, arrayLayout.Align, arrayLayout)
	}
	if got := mod.Globals[0].Size; got != arrayLayout.ElemSize*2 {
		t.Fatalf("global array size = %d, want elem size * 2 = %d", got, arrayLayout.ElemSize*2)
	}
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
}

func TestGenerateBitFieldOnlyStructObjectSize(t *testing.T) {
	mod := compileModule(t, `
struct S { int b:1; };
struct S g;
int main(void) { return 0; }`)
	if got := mod.Globals[0].Size; got < 4 {
		t.Fatalf("bit-field-only global size = %d, want at least sizeof(int)", got)
	}
	var layout *bytecode.ObjectLayout
	for i := range mod.Layouts {
		if len(mod.Layouts[i].Bit) == 1 && mod.Layouts[i].Bit[0].Name == "b" {
			layout = &mod.Layouts[i]
			break
		}
	}
	if layout == nil {
		t.Fatalf("missing bit-field struct layout: %#v", mod.Layouts)
	}
	if layout.Size < 4 {
		t.Fatalf("bit-field-only layout size = %d, want at least sizeof(int): %#v", layout.Size, layout)
	}
}

func TestGenerateTerminalBitFieldStructArrayObjectSize(t *testing.T) {
	mod := compileModule(t, `
struct S { int x; int b:1; };
struct S a[2];
int main(void) { return 0; }`)
	var structLayout, arrayLayout *bytecode.ObjectLayout
	for i := range mod.Layouts {
		layout := &mod.Layouts[i]
		if len(layout.Fields) == 1 && layout.Fields[0].Name == "x" && len(layout.Bit) == 1 && layout.Bit[0].Name == "b" {
			structLayout = layout
		}
		if layout.ElemSize != 0 {
			arrayLayout = layout
		}
	}
	if structLayout == nil {
		t.Fatalf("missing terminal bit-field struct layout: %#v", mod.Layouts)
	}
	if structLayout.Size < 8 {
		t.Fatalf("terminal bit-field struct size = %d, want at least 8: %#v", structLayout.Size, structLayout)
	}
	if arrayLayout == nil {
		t.Fatalf("missing terminal bit-field array layout: %#v", mod.Layouts)
	}
	if arrayLayout.ElemSize != structLayout.Size {
		t.Fatalf("array elem size = %d, want struct size %d", arrayLayout.ElemSize, structLayout.Size)
	}
	if got := mod.Globals[0].Size; got != structLayout.Size*2 {
		t.Fatalf("global array size = %d, want struct size * 2 = %d", got, structLayout.Size*2)
	}
}

func TestGenerateStructMemberAccess(t *testing.T) {
	mod := compileModule(t, `
struct P { int x; int y; };
int main(void) {
	struct P p;
	p.x = 1;
	p.y = 2;
	return p.x + p.y;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"Layout #",
		"FieldAddr",
		"I32Store",
		"I32Load",
		"I32Add",
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
		stringMap: map[string]int{},
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
