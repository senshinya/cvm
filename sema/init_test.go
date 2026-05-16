package sema

import "testing"

func TestInitializerListAndStaticFold(t *testing.T) {
	r := analyzeSource(t, "int a[3] = {1, 2, 3}; int x = 3 + 4 * 2;")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	if il, ok := r.Program.Globals[0].(*VarDecl).Init.(*InitList); !ok || len(il.Elems) != 3 {
		t.Fatalf("array init wrong: %T %+v", r.Program.Globals[0].(*VarDecl).Init, r.Program.Globals[0].(*VarDecl).Init)
	}
	if lit, ok := r.Program.Globals[1].(*VarDecl).Init.(*IntLit); !ok || lit.Value != 11 {
		t.Fatalf("static init not folded: %T %+v", r.Program.Globals[1].(*VarDecl).Init, r.Program.Globals[1].(*VarDecl).Init)
	}
}

func TestInitializerCompletesUnsizedStringLiteralArrays(t *testing.T) {
	r := analyzeSource(t, `char a[] = "foo"; char b[] = ("bar"); char c[] = { ("baz") };`)
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	for i, want := range []int64{4, 4, 4} {
		vd := r.Program.Globals[i].(*VarDecl)
		at, ok := Unqual(vd.T).(*ArrayType)
		if !ok || at.SizeKind != ArrayConstantSize || at.Size != want {
			t.Fatalf("global %d type = %s, want char[%d]", i, vd.T, want)
		}
	}
}

func TestInitializerDesignatedStruct(t *testing.T) {
	r := analyzeSource(t, "struct S { int x; int y; } s = { .y = 5 };")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	il, ok := r.Program.Globals[0].(*VarDecl).Init.(*InitList)
	if !ok || len(il.Elems) != 1 || len(il.Elems[0].Designators) != 1 || il.Elems[0].Designators[0].Field == nil {
		t.Fatalf("designated init wrong: %T %+v", r.Program.Globals[0].(*VarDecl).Init, r.Program.Globals[0].(*VarDecl).Init)
	}
}

func TestInitializerDesignatedUnionNonFirstMemberType(t *testing.T) {
	r := analyzeSource(t, "union U { int i; double d; } u = { .d = 1.5 };")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	il, ok := r.Program.Globals[0].(*VarDecl).Init.(*InitList)
	if !ok || len(il.Elems) != 1 {
		t.Fatalf("union init wrong: %T %+v", r.Program.Globals[0].(*VarDecl).Init, r.Program.Globals[0].(*VarDecl).Init)
	}
	if got := il.Elems[0].Designators[0].Field.Name; got != "d" {
		t.Fatalf("union designator field = %q, want d", got)
	}
	if got := il.Elems[0].Value.GetType().String(); got != "double" {
		t.Fatalf("union designator value type = %s, want double: elem=%#v", got, il.Elems[0])
	}
}

func TestInitializerUnionAggregateDesignatorUsesSelectedMemberLeaf(t *testing.T) {
	r := analyzeSource(t, "union U { double d; struct { int a; int b; } s; } u = { .s = 1 };")
	if len(r.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", r.Errors)
	}
	il, ok := r.Program.Globals[0].(*VarDecl).Init.(*InitList)
	if !ok || len(il.Elems) != 1 {
		t.Fatalf("union init wrong: %T %+v", r.Program.Globals[0].(*VarDecl).Init, r.Program.Globals[0].(*VarDecl).Init)
	}
	if got := il.Elems[0].Designators[len(il.Elems[0].Designators)-1].Field.Name; got != "a" {
		t.Fatalf("leaf designator field = %q, want a: %#v", got, il.Elems[0])
	}
	if got := il.Elems[0].Value.GetType().String(); got != "int" {
		t.Fatalf("leaf value type = %s, want int: %#v", got, il.Elems[0])
	}
}
