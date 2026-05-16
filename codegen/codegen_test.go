package codegen

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"shinya.click/cvm/sema"
)

var updateGolden = flag.Bool("update", false, "regenerate codegen golden files")

func TestGoldenPassBytecode(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "sema", "testdata", "pass", "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("no sema pass cases")
	}
	for _, src := range matches {
		t.Run(filepath.Base(src), func(t *testing.T) {
			source, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			mod := compileModule(t, string(source))
			got := bytecode.PrintModule(mod)
			goldenPath := strings.TrimSuffix(src, ".c") + ".bytecode.golden"
			if *updateGolden {
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v (run with -update to create)", err)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
			}
		})
	}
}

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

func TestGenerateUnaryMinus(t *testing.T) {
	mod := compileModule(t, `int main(void) { return -1; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"I32Const 1", "I32Neg", "I32Return"} {
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

func TestGenerateVolatileAutomaticUsesAddressPath(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	volatile int x = 1;
	x = 2;
	return x;
}`)
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "LoadLocal") || strings.Contains(out, "StoreLocal") {
		t.Fatalf("volatile automatic used local slot ops:\n%s", out)
	}
	for _, want := range []string{"Object #", "I32Store align=4 volatile=true", "I32Load align=4 volatile=true"} {
		if !strings.Contains(out, want) {
			t.Fatalf("volatile automatic bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateImplicitFallthroughTerminals(t *testing.T) {
	mod := compileModule(t, `void f(void) {} int main(void) {}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{`Func #0`, "ReturnVoid", `Func #1`, "I32Const 0", "I32Return"} {
		if !strings.Contains(out, want) {
			t.Fatalf("fallthrough bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGeneratePointerDifference(t *testing.T) {
	mod := compileModule(t, `long f(int *a, int *b) { return b - a; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"PtrDiff elem_size=4", "I64Return"} {
		if !strings.Contains(out, want) {
			t.Fatalf("pointer diff bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateExternFunctionPointerInitializer(t *testing.T) {
	mod := compileModule(t, `int ext(int); int (*fp)(int) = ext; int main(void) { return 0; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{`Global #0 extern name="ext"`, `var name="fp"`, "reloc offset=0 kind=func target=extern#0(\"ext\")"} {
		if !strings.Contains(out, want) {
			t.Fatalf("extern function pointer initializer missing %q:\n%s", want, out)
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

func TestGenerateStructuredControlFlow(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	int x = 0;
	for (int i = 0; i < 3; i = i + 1) {
		if (i == 2) break;
		x = x + i;
	}
	return x;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"L0:",
		"JumpIfZero",
		"Jump L",
		"I32LtS",
		"I32Eq",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateSwitchGotoAndLabels(t *testing.T) {
	mod := compileModule(t, `
int route(int x) {
	int y = 0;
	switch (x) {
	case 1: y = 10; break;
	default: y = 20;
	}
	goto done;
done:
	return y;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{
		"Switch",
		"Jump L",
		"L",
		"I32Return",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateNamedBreakTargetsOuterLoop(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	int x = 0;
outer:
	for (int i = 0; i < 3; i = i + 1) {
		for (int j = 0; j < 3; j = j + 1) {
			x = 1;
			break outer;
		}
		x = 2;
	}
	return x;
}`)
	fn := mod.Functions[0]
	outerEnd := labelBeforeReturnLoad(t, fn, 0)
	namedBreak := jumpAfterStoreConst(t, fn, 1)
	if namedBreak != outerEnd {
		t.Fatalf("break outer jumps to L%d, want outer end L%d:\n%s", namedBreak, outerEnd, bytecode.PrintModule(mod))
	}
}

func TestGenerateNamedContinueTargetsOuterLoop(t *testing.T) {
	mod := compileModule(t, `
int main(void) {
	int x = 0;
outer:
	for (int i = 0; i < 3; i = i + 1) {
		for (int j = 0; j < 3; j = j + 1) {
			x = 1;
			continue outer;
		}
		x = 2;
	}
	return x;
}`)
	fn := mod.Functions[0]
	outerPost := labelBeforePostStore(t, fn, 1)
	namedContinue := jumpAfterStoreConst(t, fn, 1)
	if namedContinue != outerPost {
		t.Fatalf("continue outer jumps to L%d, want outer continue L%d:\n%s", namedContinue, outerPost, bytecode.PrintModule(mod))
	}
}

func TestGenerateNamedBreakTargetsSwitch(t *testing.T) {
	mod := compileModule(t, `
int route(int x) {
	int y = 0;
done:
	switch (x) {
	case 1:
		while (x) {
			y = 10;
			break done;
		}
	default:
		y = 20;
	}
	y = 30;
	return y;
}`)
	fn := mod.Functions[0]
	switchEnd := labelBeforeStoreConst(t, fn, 30)
	namedBreak := jumpAfterStoreConst(t, fn, 10)
	if namedBreak != switchEnd {
		t.Fatalf("break done jumps to L%d, want switch end L%d:\n%s", namedBreak, switchEnd, bytecode.PrintModule(mod))
	}
}

func TestGenerateAssignmentExpressionStatementResult(t *testing.T) {
	compileModule(t, `int main(void) { int x; x = 1; return x; }`)
}

func TestGenerateFileScopeExternVariableMetadata(t *testing.T) {
	mod := compileModule(t, `extern int g; int main(void) { return g; }`)
	if got := mod.Globals[0]; got.Kind != bytecode.GlobalExtern || got.Size != 4 || got.Align <= 0 || got.Init.ZeroFill != 0 {
		t.Fatalf("global g metadata = %#v, want extern size 4 positive align and no zero-fill", got)
	}
	out := bytecode.PrintModule(mod)
	if !strings.Contains(out, `Global #0 extern name="g"`) {
		t.Fatalf("bytecode missing extern global:\n%s", out)
	}
	if strings.Contains(out, `Global #0 var name="g"`) || strings.Contains(out, "init_zero=") {
		t.Fatalf("extern global printed as storage-owning var:\n%s", out)
	}
}

func TestGenerateModuleCompletenessMetadata(t *testing.T) {
	mod := compileModule(t, `
int puts(char *);
int main(void) {
	return puts("ok");
}`)
	if mod.Version != bytecode.CurrentModuleVersion {
		t.Fatalf("module version = %q, want %q", mod.Version, bytecode.CurrentModuleVersion)
	}
	if mod.Entry == nil || mod.Entry.Global < 0 || mod.Globals[mod.Entry.Global].Name != "main" {
		t.Fatalf("entry metadata = %#v, globals = %#v", mod.Entry, mod.Globals)
	}
	var puts *bytecode.Global
	for i := range mod.Globals {
		if mod.Globals[i].Name == "puts" {
			puts = &mod.Globals[i]
			break
		}
	}
	if puts == nil {
		t.Fatalf("puts global not found: %#v", mod.Globals)
	}
	if puts.Kind != bytecode.GlobalExtern || puts.Extern.Name != "puts" || puts.Extern.ABI != bytecode.DefaultExternABI {
		t.Fatalf("puts extern metadata = %#v, want import name puts and ABI %q", *puts, bytecode.DefaultExternABI)
	}
	if puts.Sig < 0 || puts.Sig >= len(mod.Sigs) {
		t.Fatalf("puts signature id = %d, signatures = %#v", puts.Sig, mod.Sigs)
	}
	if sig := mod.Sigs[puts.Sig]; sig.Ret != bytecode.TypeI32 || len(sig.Params) != 1 || sig.Params[0] != bytecode.TypePtr {
		t.Fatalf("puts signature = %#v, want int(ptr)", sig)
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

func TestGenerateStaticInitializersStringsAndRelocations(t *testing.T) {
	mod := compileModule(t, `
int g = 3;
int *gp = &g;
char *s = "hi";
const int cg = 4;
int main(void) { return g; }`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"Global #0 var name=\"g\"", "kind=global target=global#0(\"g\")", "String #0 value=\"hi", "var name=\"cg\"", "readonly=true"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateAggregateAndBitFieldLowering(t *testing.T) {
	mod := compileModule(t, `
struct B { unsigned x:3; unsigned y:5; };
struct P { int a[2]; struct B b; };
int main(void) {
	struct P p = {{1, 2}, {3, 4}};
	struct P q = {{0, 0}, {0, 0}};
	q = p;
	p.b.x = 5;
	return q.a[1] + p.b.x;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"BitFieldStore", "BitFieldLoad", "MemCopy", "Layout #"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "$init") {
		t.Fatalf("initializer list unexpectedly lowered through temporary object:\n%s", out)
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

func TestGenerateUnbracedNestedAggregateStaticInitializer(t *testing.T) {
	mod := compileModule(t, `
struct S { int a[2]; int b; };
struct S s = {1, 2, 3};
int main(void) { return s.a[1] + s.b; }`)
	if got, want := len(mod.Globals[0].Init.Bytes), 12; got != want {
		t.Fatalf("initializer bytes length = %d, want %d: %#v", got, want, mod.Globals[0])
	}
	for i, want := range []uint32{1, 2, 3} {
		if got := binary.LittleEndian.Uint32(mod.Globals[0].Init.Bytes[i*4 : i*4+4]); got != want {
			t.Fatalf("word %d = %d, want %d; bytes=%x", i, got, want, mod.Globals[0].Init.Bytes)
		}
	}
}

func TestGenerateDesignatorContinuationStaticInitializer(t *testing.T) {
	mod := compileModule(t, `
struct T { int x; int y; };
struct T a[2] = { [0].x = 1, 2 };
struct U { int a[3]; int b; };
struct U u = { .a[1] = 5, 6 };
int main(void) { return a[0].y + u.a[2]; }`)
	arrayBytes := mod.Globals[0].Init.Bytes
	for i, want := range []uint32{1, 2, 0, 0} {
		if got := binary.LittleEndian.Uint32(arrayBytes[i*4 : i*4+4]); got != want {
			t.Fatalf("array word %d = %d, want %d; bytes=%x", i, got, want, arrayBytes)
		}
	}
	structBytes := mod.Globals[1].Init.Bytes
	for i, want := range []uint32{0, 5, 6, 0} {
		if got := binary.LittleEndian.Uint32(structBytes[i*4 : i*4+4]); got != want {
			t.Fatalf("struct word %d = %d, want %d; bytes=%x", i, got, want, structBytes)
		}
	}
}

func TestGenerateMixedTypeAggregateInitializerConversions(t *testing.T) {
	mod := compileModule(t, `
struct S { int a[2]; double b; };
struct S s = {1.5, 2.5, 3.5};
struct U { int a[3]; double b; };
struct U u = { .a[1] = 5, 6.5 };
int main(void) {
	struct S ls = {1.5, 2.5, 3.5};
	struct U lu = { .a[1] = 5, 6.5 };
	return s.a[0] + u.a[1] + ls.a[0] + lu.a[1];
}`)
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	staticS := mod.Globals[0].Init.Bytes
	for i, want := range []uint32{1, 2} {
		if got := binary.LittleEndian.Uint32(staticS[i*4 : i*4+4]); got != want {
			t.Fatalf("struct S int word %d = %d, want %d; bytes=%x", i, got, want, staticS)
		}
	}
	if got := binary.LittleEndian.Uint64(staticS[8:16]); got != 0x400c000000000000 {
		t.Fatalf("struct S double bits = %#x, want 3.5; bytes=%x", got, staticS)
	}
	staticU := mod.Globals[1].Init.Bytes
	for i, want := range []uint32{0, 5, 6} {
		if got := binary.LittleEndian.Uint32(staticU[i*4 : i*4+4]); got != want {
			t.Fatalf("struct U int word %d = %d, want %d; bytes=%x", i, got, want, staticU)
		}
	}
}

func TestGenerateUnionDesignatorInitializer(t *testing.T) {
	mod := compileModule(t, `
union U { int i; double d; };
union U u = { .d = 1.5 };
struct S { union U u; int x; };
struct S s = { .u.d = 1.5, 2 };
int main(void) {
	union U lu = { .d = 1.5 };
	struct S ls = { .u.d = 1.5, 2 };
	return s.x + ls.x + (lu.d > 0.0);
}`)
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	unionBytes := mod.Globals[0].Init.Bytes
	if got := binary.LittleEndian.Uint64(unionBytes[:8]); got != 0x3ff8000000000000 {
		t.Fatalf("union double bits = %#x, want 1.5; bytes=%x", got, unionBytes)
	}
	structBytes := mod.Globals[1].Init.Bytes
	if got := binary.LittleEndian.Uint64(structBytes[:8]); got != 0x3ff8000000000000 {
		t.Fatalf("struct union double bits = %#x, want 1.5; bytes=%x", got, structBytes)
	}
	if got := binary.LittleEndian.Uint32(structBytes[8:12]); got != 2 {
		t.Fatalf("struct continuation word = %d, want 2; bytes=%x", got, structBytes)
	}
}

func TestGenerateUnionAggregateDesignatorFirstLeaf(t *testing.T) {
	mod := compileModule(t, `
union U { double d; struct { int a; int b; } s; };
union U u = { .s = 1 };
int main(void) {
	union U lu = { .s = 1 };
	return u.s.a + lu.s.a;
}`)
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	bytes := mod.Globals[0].Init.Bytes
	if got := binary.LittleEndian.Uint32(bytes[:4]); got != 1 {
		t.Fatalf("union struct first field = %d, want 1; bytes=%x", got, bytes)
	}
	if got := binary.LittleEndian.Uint32(bytes[4:8]); got != 0 {
		t.Fatalf("union struct second field = %d, want 0; bytes=%x", got, bytes)
	}
}

func TestGenerateUnbracedNestedAggregateLocalInitializer(t *testing.T) {
	mod := compileModule(t, `
struct S { int a[2]; int b; };
int main(void) {
	struct S s = {1, 2, 3};
	return s.a[1] + s.b;
}`)
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
	}
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "$init") {
		t.Fatalf("local unbraced aggregate initializer unexpectedly lowered through temporary object:\n%s", out)
	}
}

func TestGenerateDesignatorContinuationLocalInitializer(t *testing.T) {
	mod := compileModule(t, `
struct T { int x; int y; };
struct U { int a[3]; int b; };
int main(void) {
	struct T ts[2] = { [0].x = 1, 2 };
	struct U u = { .a[1] = 5, 6 };
	return ts[0].y + u.a[2];
}`)
	if err := bytecode.ValidateModule(mod); err != nil {
		t.Fatalf("validate bytecode: %v\n%s", err, bytecode.PrintModule(mod))
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
		mod:       bytecode.NewModule(),
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

func TestGenerateCallsCastsAndFunctionPointers(t *testing.T) {
	mod := compileModule(t, `
int inc(int x) { return x + 1; }
int main(void) {
	int (*fp)(int) = inc;
	long y = (long)fp(2);
	return (int)y;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"CallIndirect", "SExt", "Trunc", "AddrFunc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateSizeofAndVarargsMetadata(t *testing.T) {
	mod := compileModule(t, `
int f(int n, ...) {
	int a[n];
	return sizeof(a);
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"variadic", "AllocDynamicObject", "DynamicObjectAddr"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateFunctionPointerDerefCall(t *testing.T) {
	mod := compileModule(t, `
int inc(int x) { return x + 1; }
int main(void) {
	int (*fp)(int) = inc;
	return (*fp)(2);
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{"AddrFunc 0", "PtrLoadLocal 0", "CallIndirect"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateExternFunctionAddressUsesAddrFuncGlobal(t *testing.T) {
	mod := compileModule(t, `
int ext(int);
int main(void) {
	int (*fp)(int) = ext;
	return 0;
}`)
	out := bytecode.PrintModule(mod)
	for _, want := range []string{`Global #0 extern name="ext"`, "AddrFunc 0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bytecode missing %q:\n%s", want, out)
		}
	}
}

func TestGenerateAddressOfFunctionUsesAddrFuncGlobal(t *testing.T) {
	mod := compileModule(t, `
int inc(int);
int main(void) {
	int (*fp)(int) = &inc;
	return 0;
}`)
	out := bytecode.PrintModule(mod)
	if !strings.Contains(out, "AddrFunc 0") {
		t.Fatalf("bytecode missing AddrFunc for &function:\n%s", out)
	}
}

func TestGenerateExplicitFloatCastToBoolUsesBoolCast(t *testing.T) {
	mod := compileModule(t, `_Bool main(void) { return (_Bool)1.5; }`)
	out := bytecode.PrintModule(mod)
	if !strings.Contains(out, "Cast f64->bool Bool") {
		t.Fatalf("bytecode missing explicit bool cast:\n%s", out)
	}
}

func TestGenerateGotoOutOfVLAScopeFreesDynamicObject(t *testing.T) {
	mod := compileModule(t, `
int f(int n) {
	{
		int a[n];
		goto done;
	}
done:
	return 0;
}`)
	fn := mod.Functions[0]
	allocPC := instrPC(t, fn, func(i bytecode.Instr) bool { return i.Op == bytecode.OpAllocDynamicObject })
	freePC := instrPC(t, fn, func(i bytecode.Instr) bool { return i.Op == bytecode.OpFreeDynamicObject })
	jumpPC := instrPCAfter(t, fn, allocPC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpJump })
	if !(allocPC < freePC && freePC < jumpPC) {
		t.Fatalf("goto out of VLA scope did not free before jump: alloc=%d free=%d jump=%d\n%s", allocPC, freePC, jumpPC, bytecode.PrintModule(mod))
	}
}

func TestGenerateForInitVLAFreesAtLoopExit(t *testing.T) {
	mod := compileModule(t, `
int f(int n) {
	for (int a[n]; n; n = n - 1) {
		break;
	}
	return 0;
}`)
	fn := mod.Functions[0]
	allocPC := instrPC(t, fn, func(i bytecode.Instr) bool { return i.Op == bytecode.OpAllocDynamicObject })
	breakFreePC := instrPCAfter(t, fn, allocPC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpFreeDynamicObject })
	breakJumpPC := instrPCAfter(t, fn, breakFreePC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpJump })
	endFreePC := instrPCAfter(t, fn, breakJumpPC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpFreeDynamicObject })
	returnPC := instrPCAfter(t, fn, endFreePC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpReturn })
	breakTargetPC := labelPC(t, fn, fn.Instrs[breakJumpPC].Label)
	if !(allocPC < breakFreePC && breakFreePC < breakJumpPC && breakJumpPC < endFreePC && endFreePC < breakTargetPC && breakTargetPC < returnPC) {
		t.Fatalf("for-init VLA cleanup is not emitted on break and normal loop exit: alloc=%d breakFree=%d breakJump=%d endFree=%d return=%d\n%s",
			allocPC, breakFreePC, breakJumpPC, endFreePC, returnPC, bytecode.PrintModule(mod))
	}
}

func TestGenerateForInitVLAContinueDoesNotFreeBeforePost(t *testing.T) {
	mod := compileModule(t, `
int f(int n) {
	int x = 0;
	for (int a[n]; x < 1; x = x + 1) {
		continue;
	}
	return 0;
}`)
	fn := mod.Functions[0]
	allocPC := instrPC(t, fn, func(i bytecode.Instr) bool { return i.Op == bytecode.OpAllocDynamicObject })
	continueJumpPC := instrPCAfter(t, fn, allocPC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpJump })
	storeAfterContinuePC := instrPCAfter(t, fn, continueJumpPC, func(i bytecode.Instr) bool {
		return i.Op == bytecode.OpStoreLocal && i.Slot == 1
	})
	freePC := instrPCAfter(t, fn, storeAfterContinuePC, func(i bytecode.Instr) bool { return i.Op == bytecode.OpFreeDynamicObject })
	if !(allocPC < continueJumpPC && continueJumpPC < storeAfterContinuePC && storeAfterContinuePC < freePC) {
		t.Fatalf("for-init VLA was freed before continue reached post expression: alloc=%d continueJump=%d postStore=%d free=%d\n%s",
			allocPC, continueJumpPC, storeAfterContinuePC, freePC, bytecode.PrintModule(mod))
	}
}

func TestGenerateSizeofSubVLADoesNotReevaluateBound(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m) {
	int a[n][m];
	return sizeof(a[0]);
}`)
	fn := mod.Functions[0]
	loadsOfM := 0
	for _, ins := range fn.Instrs {
		if ins.Op == bytecode.OpLoadLocal && ins.Slot == 1 {
			loadsOfM++
		}
	}
	if loadsOfM != 1 {
		t.Fatalf("sizeof sub-VLA reevaluated m; saw %d loads of slot 1\n%s", loadsOfM, bytecode.PrintModule(mod))
	}
	out := bytecode.PrintModule(mod)
	if !strings.Contains(out, `Local #3 name="a$size$elem" type=i64`) {
		t.Fatalf("bytecode missing saved sub-VLA size slot:\n%s", out)
	}
}

func TestGenerateMultidimensionalVLAIndexUsesDynamicStride(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m) {
	int a[n][m];
	a[1][2] = 7;
	return a[1][2];
}`)
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "PtrAdd elem_size=0") {
		t.Fatalf("bytecode used zero stride for multidimensional VLA index:\n%s", out)
	}
	if !strings.Contains(out, "PtrAddDynamic") {
		t.Fatalf("bytecode missing dynamic stride for multidimensional VLA index:\n%s", out)
	}
}

func TestGenerateVLAParameterIndexUsesDynamicStride(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m, int a[n][m]) {
	return a[1][2] + sizeof(a[0]);
}`)
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "PtrAdd elem_size=0") {
		t.Fatalf("bytecode used zero stride for VLA parameter index:\n%s", out)
	}
	if !strings.Contains(out, "PtrAddDynamic") {
		t.Fatalf("bytecode missing dynamic stride for VLA parameter index:\n%s", out)
	}
	boundLoads := strings.Count(out, "I32LoadLocal 0") + strings.Count(out, "I32LoadLocal 1")
	if boundLoads != 1 {
		t.Fatalf("VLA parameter size should evaluate its sema-provided bound once, saw %d loads:\n%s", boundLoads, out)
	}
}

func TestGeneratePointerToVLALocalIndexUsesDynamicStride(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m) {
	int a[n][m];
	int (*p)[m] = a;
	return (p + 1)[0][0];
}`)
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "PtrAdd elem_size=0") {
		t.Fatalf("bytecode used zero stride for pointer-to-VLA local:\n%s", out)
	}
	if strings.Count(out, "PtrAddDynamic") < 2 {
		t.Fatalf("bytecode missing dynamic strides for pointer-to-VLA local:\n%s", out)
	}
}

func TestGenerateDistinctVLASizeSlotsDoNotAlias(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m, int k) {
	int a[n][m];
	int b[n][k];
	return sizeof(a[0]) + sizeof(b[0]) + a[1][0] + b[1][0];
}`)
	out := bytecode.PrintModule(mod)
	if strings.Contains(out, "PtrAdd elem_size=0") {
		t.Fatalf("bytecode used zero stride for distinct VLA sizes:\n%s", out)
	}
	if strings.Count(out, "I32LoadLocal 1") != 1 || strings.Count(out, "I32LoadLocal 2") != 1 {
		t.Fatalf("distinct VLA bounds should each be evaluated once:\n%s", out)
	}
	if !strings.Contains(out, `name="a$size$elem" type=i64`) || !strings.Contains(out, `name="b$size$elem" type=i64`) {
		t.Fatalf("bytecode missing separate VLA size slots:\n%s", out)
	}
}

func TestGenerateSameBoundVLAInstancesDoNotAlias(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int m) {
	int a[n][m];
	m = m + 1;
	int b[n][m];
	return sizeof(a[0]) + sizeof(b[0]);
}`)
	out := bytecode.PrintModule(mod)
	aSlot := localSlotByName(t, mod.Functions[0], "a$size$elem")
	bSlot := localSlotByName(t, mod.Functions[0], "b$size$elem")
	if aSlot == bSlot {
		t.Fatalf("same-bound VLA instances share a size slot:\n%s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("I64LoadLocal %d", aSlot)) || !strings.Contains(out, fmt.Sprintf("I64LoadLocal %d", bSlot)) {
		t.Fatalf("sizeof should load both VLA instance size slots:\n%s", out)
	}
}

func TestGenerateVLAMetadataAfterAddressTakenParamCopy(t *testing.T) {
	mod := compileModule(t, `
int f(int n, int a[n]) {
	int *p = &n;
	return a[0];
}`)
	fn := mod.Functions[0]
	copyStore := instrPC(t, fn, func(i bytecode.Instr) bool { return i.Op == bytecode.OpStore && i.Type == bytecode.TypeI32 })
	boundLoad := instrPCAfter(t, fn, copyStore, func(i bytecode.Instr) bool { return i.Op == bytecode.OpLoad && i.Type == bytecode.TypeI32 })
	if !(copyStore < boundLoad) {
		t.Fatalf("VLA parameter metadata was prepared before address-taken param copy: copyStore=%d boundLoad=%d\n%s",
			copyStore, boundLoad, bytecode.PrintModule(mod))
	}
}

func TestGenerateSizeofDerefPointerToVLAUsesSavedMetadata(t *testing.T) {
	mod := compileModule(t, `
int f(int m) {
	int (*p)[m];
	return sizeof *p;
}`)
	out := bytecode.PrintModule(mod)
	if strings.Count(out, "I32LoadLocal 0") != 1 {
		t.Fatalf("sizeof *p should not reevaluate pointer-to-VLA bound:\n%s", out)
	}
	slot := localSlotByName(t, mod.Functions[0], "p$size$pointee")
	if !strings.Contains(out, fmt.Sprintf("I64LoadLocal %d", slot)) {
		t.Fatalf("sizeof *p should load saved pointer-to-VLA metadata slot %d:\n%s", slot, out)
	}
}

func instrPC(t *testing.T, fn bytecode.Function, pred func(bytecode.Instr) bool) int {
	t.Helper()
	return instrPCAfter(t, fn, -1, pred)
}

func instrPCAfter(t *testing.T, fn bytecode.Function, after int, pred func(bytecode.Instr) bool) int {
	t.Helper()
	for pc := after + 1; pc < len(fn.Instrs); pc++ {
		if pred(fn.Instrs[pc]) {
			return pc
		}
	}
	t.Fatalf("instruction not found after pc %d", after)
	return -1
}

func localSlotByName(t *testing.T, fn bytecode.Function, name string) int {
	t.Helper()
	for _, l := range fn.Locals {
		if l.Name == name {
			return l.ID
		}
	}
	t.Fatalf("local slot %q not found", name)
	return -1
}

func labelPC(t *testing.T, fn bytecode.Function, label int) int {
	t.Helper()
	for pc, ins := range fn.Instrs {
		if ins.Op == bytecode.OpLabel && ins.Label == label {
			return pc
		}
	}
	t.Fatalf("label L%d not found", label)
	return -1
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

func jumpAfterStoreConst(t *testing.T, fn bytecode.Function, value int64) int {
	t.Helper()

	for i := 0; i+3 < len(fn.Instrs); i++ {
		if fn.Instrs[i].Op == bytecode.OpConst && fn.Instrs[i].Int == value &&
			fn.Instrs[i+1].Op == bytecode.OpDup &&
			fn.Instrs[i+2].Op == bytecode.OpStoreLocal &&
			fn.Instrs[i+3].Op == bytecode.OpPop {
			for j := i + 4; j < len(fn.Instrs); j++ {
				if fn.Instrs[j].Op == bytecode.OpJump {
					return fn.Instrs[j].Label
				}
				if fn.Instrs[j].Op == bytecode.OpLabel {
					break
				}
			}
		}
	}
	t.Fatalf("missing jump after assignment to %d in %#v", value, fn.Instrs)
	return -1
}

func labelBeforeStoreConst(t *testing.T, fn bytecode.Function, value int64) int {
	t.Helper()

	for i := 1; i+3 < len(fn.Instrs); i++ {
		if fn.Instrs[i-1].Op == bytecode.OpLabel &&
			fn.Instrs[i].Op == bytecode.OpConst && fn.Instrs[i].Int == value &&
			fn.Instrs[i+1].Op == bytecode.OpDup &&
			fn.Instrs[i+2].Op == bytecode.OpStoreLocal &&
			fn.Instrs[i+3].Op == bytecode.OpPop {
			return fn.Instrs[i-1].Label
		}
	}
	t.Fatalf("missing label before assignment to %d in %#v", value, fn.Instrs)
	return -1
}

func labelBeforeReturnLoad(t *testing.T, fn bytecode.Function, slot int) int {
	t.Helper()

	for i := 1; i+1 < len(fn.Instrs); i++ {
		if fn.Instrs[i-1].Op == bytecode.OpLabel &&
			fn.Instrs[i].Op == bytecode.OpLoadLocal && fn.Instrs[i].Slot == slot &&
			fn.Instrs[i+1].Op == bytecode.OpReturn {
			return fn.Instrs[i-1].Label
		}
	}
	t.Fatalf("missing label before return load of local slot %d in %#v", slot, fn.Instrs)
	return -1
}

func labelBeforePostStore(t *testing.T, fn bytecode.Function, slot int) int {
	t.Helper()

	for i := 0; i+4 < len(fn.Instrs); i++ {
		if fn.Instrs[i].Op == bytecode.OpLabel &&
			fn.Instrs[i+1].Op == bytecode.OpLoadLocal && fn.Instrs[i+1].Slot == slot &&
			fn.Instrs[i+2].Op == bytecode.OpConst && fn.Instrs[i+2].Int == 1 &&
			fn.Instrs[i+3].Op == bytecode.OpBinary &&
			fn.Instrs[i+4].Op == bytecode.OpDup {
			return fn.Instrs[i].Label
		}
	}
	t.Fatalf("missing post label before increment of local slot %d in %#v", slot, fn.Instrs)
	return -1
}
