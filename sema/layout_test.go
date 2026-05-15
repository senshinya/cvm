package sema

import (
	"strings"
	"testing"
)

func TestProgramLayoutAssignsStableGlobalAndLocalIDs(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, `
int g;
int f(int a) {
	int x;
	static int y;
	return a + x + y;
}
`))
	if err != nil {
		t.Fatal(err)
	}
	g := prog.Globals[0].(*VarDecl)
	f := prog.Funcs[0]
	if g.Sym.GlobalID != 0 {
		t.Fatalf("global g id = %d, want 0", g.Sym.GlobalID)
	}
	if f.Sym.GlobalID != 1 {
		t.Fatalf("function f id = %d, want 1", f.Sym.GlobalID)
	}
	if f.Params[0].Sym.SlotID != 0 {
		t.Fatalf("parameter a slot = %d, want 0", f.Params[0].Sym.SlotID)
	}
	if f.Locals[0].Sym.SlotID != 1 {
		t.Fatalf("local x slot = %d, want 1", f.Locals[0].Sym.SlotID)
	}
	if f.Locals[1].Sym.GlobalID != 2 {
		t.Fatalf("static local y global id = %d, want 2", f.Locals[1].Sym.GlobalID)
	}
}

func TestPrintProgramShowsCodegenLayoutAndVLABound(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, `
int f(int n) {
	int a[n];
	return sizeof(a);
}
`))
	if err != nil {
		t.Fatal(err)
	}
	out := PrintProgram(prog)
	for _, want := range []string{"global=0", "slot=0", "slot=1", "vla-bound:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("PrintProgram output missing %q:\n%s", want, out)
		}
	}
}

func TestVoidParameterListDoesNotCreateFrameSlot(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, `int main(void) { return 0; }`))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(prog.Funcs[0].Params); got != 0 {
		t.Fatalf("main(void) produced %d IR params, want 0", got)
	}
}

func TestBlockExternDeclarationDoesNotUseFrameSlotOrFakeGlobalID(t *testing.T) {
	prog, err := Analyze(parseCandidates(t, `
int f(void) {
	extern int ext;
	int local;
	return local;
}
`))
	if err != nil {
		t.Fatal(err)
	}
	fn := prog.Funcs[0]
	if fn.Locals[0].Sym.SlotID != -1 {
		t.Fatalf("extern local slot = %d, want -1", fn.Locals[0].Sym.SlotID)
	}
	if fn.Locals[0].Sym.GlobalID != -1 {
		t.Fatalf("extern local global id = %d, want -1", fn.Locals[0].Sym.GlobalID)
	}
	if fn.Locals[1].Sym.SlotID != 0 {
		t.Fatalf("automatic local slot = %d, want 0", fn.Locals[1].Sym.SlotID)
	}
	out := PrintProgram(prog)
	if strings.Contains(out, "VarDecl name=\"ext\" type=int storage=4 slot=") ||
		strings.Contains(out, "VarDecl name=\"ext\" type=int storage=4 global=") {
		t.Fatalf("extern local printed storage layout it does not own:\n%s", out)
	}
}
