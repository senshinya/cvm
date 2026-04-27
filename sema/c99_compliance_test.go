package sema

import "testing"

func TestC99FuncIdentifierIsImplicitInFunctionScope(t *testing.T) {
	prog := mustAnalyze(t, `
		extern int strcmp(const char *, const char *);
		int main(void) {
			return strcmp(__func__, "main");
		}
	`)
	if len(prog.Funcs) != 1 {
		t.Fatalf("expected one function, got %d", len(prog.Funcs))
	}
}

func TestC99StringLiteralInitializesCharacterArray(t *testing.T) {
	mustAnalyze(t, `
		char global[] = { "foo" };
		int main(void) {
			char local[] = { "bar" };
			return global[0] + local[0];
		}
	`)
}

func TestC99UnbracedStringLiteralInitializesCharacterArrays(t *testing.T) {
	mustAnalyze(t, `
		char a[] = "foo";
		signed char b[] = "foo";
		unsigned char c[] = "foo";
		int main(void) {
			return a[0] + b[0] + c[0];
		}
	`)
}

func TestC99ForwardEnumWarningOnlyIsAccepted(t *testing.T) {
	mustAnalyze(t, `
		enum e1;
		enum e2 { E };
		enum e2;
	`)
}

func TestC99ForwardEnumPedanticErrorsRejects(t *testing.T) {
	mustRejectWithOptions(t, `enum e1;`, SemaOptions{PedanticErrors: true})
}

func TestC99StaticAssertWarningOnlySyntaxIsAccepted(t *testing.T) {
	mustAnalyze(t, `_Static_assert(1);`)
}

func TestC99StaticAssertPedanticErrorsRejects(t *testing.T) {
	mustRejectWithOptions(t, `_Static_assert(1);`, SemaOptions{PedanticErrors: true})
}

func TestC99StaticAssertZeroRejects(t *testing.T) {
	mustReject(t, `_Static_assert(0);`)
}

func TestC99FuncIdentifierPreservesConstForPointerAssignment(t *testing.T) {
	mustReject(t, `
		void foo(void) {
			char *p = __func__;
		}
	`)
}

func TestC99StringLiteralDoesNotInitializeUnsizedArrayMember(t *testing.T) {
	mustReject(t, `
		typedef char T[];
		struct foo { int x; T y; };
		struct foo b[] = { { 1, "abc" } };
	`)
}

func mustAnalyze(t *testing.T, src string) *Program {
	t.Helper()
	candidates := parseCandidates(t, src)
	prog, err := Analyze(candidates)
	if err != nil {
		t.Fatalf("Analyze rejected source: %v", err)
	}
	return prog
}

func mustReject(t *testing.T, src string) {
	t.Helper()
	mustRejectWithOptions(t, src, SemaOptions{})
}

func mustRejectWithOptions(t *testing.T, src string, opts SemaOptions) {
	t.Helper()
	candidates := parseCandidates(t, src)
	if _, err := AnalyzeWithOptions(candidates, opts); err == nil {
		t.Fatalf("Analyze accepted source, want rejection")
	}
}
