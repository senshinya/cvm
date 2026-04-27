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

func TestC99ForwardEnumDeclaratorPedanticErrorsRejects(t *testing.T) {
	mustRejectWithOptions(t, `
		void f(void) {
			enum e1 *x;
			(void)x;
		}
	`, SemaOptions{PedanticErrors: true})
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

func TestC99FuncIdentifierPreservesConstForVoidPointerAssignment(t *testing.T) {
	mustReject(t, `
		void foo(void) {
			void *p = __func__;
			(void)p;
		}
	`)
}

func TestC99RejectsDiscardedPointerQualifiers(t *testing.T) {
	mustReject(t, `
		struct s { const int a[1]; };
		void f(struct s s) {
			int *p = s.a;
			(void)p;
		}
	`)
}

func TestC99RejectsNestedPointerQualifierConversion(t *testing.T) {
	mustReject(t, `
		void f(void) {
			int **ppi;
			const int **ppci = ppi;
			(void)ppci;
		}
	`)
}

func TestC99RejectsNestedArrayQualifierConversion(t *testing.T) {
	mustReject(t, `
		void f(void) {
			int (*pa)[1];
			const int (*pca)[1] = pa;
			(void)pca;
		}
	`)
}

func TestC99FunctionPointerAcceptsVoidPointerNullAssignment(t *testing.T) {
	mustAnalyze(t, `
		void (*fp)(void);
		void f(void) {
			fp = (void *)0;
		}
	`)
}

func TestC99RejectsFunctionPointerNonNullVoidPointerAssignment(t *testing.T) {
	mustReject(t, `
		void *p;
		void (*fp)(void);
		void f(void) {
			fp = p;
		}
	`)
}

func TestC99RejectsConditionalVoidPointerQualifierLoss(t *testing.T) {
	mustReject(t, `
		int cond;
		const int *p;
		void *q;
		void f(void) {
			int *dst = cond ? (const int *)p : (void *)q;
			(void)dst;
		}
	`)
}

func TestC99RejectsDifferentStructTagPointerAssignment(t *testing.T) {
	mustReject(t, `
		struct a { int x; };
		struct b { int x; };
		void f(struct a *x, struct b *y) {
			x = y;
		}
	`)
}

func TestC99EnumDefinitionCanShadowOuterTag(t *testing.T) {
	mustAnalyze(t, `
		enum e { A };
		void f(void) {
			enum e { B };
			enum e x = B;
			(void)x;
		}
	`)
}

func TestC99RejectsDifferentEnumTagPointerAssignment(t *testing.T) {
	mustReject(t, `
		enum e { A };
		void f(void) {
			enum e *outerp;
			enum e { B } inner = B;
			outerp = &inner;
		}
	`)
}

func TestC99FunctionPointerCompatibleNoPrototypeAssignment(t *testing.T) {
	mustAnalyze(t, `
		int f();
		void g(void) {
			int (*p)(int);
			p = f;
		}
	`)
}

func TestC99RejectsIncompatibleNoPrototypeFunctionPointerAssignment(t *testing.T) {
	mustReject(t, `
		int f();
		void g(void) {
			int (*p)(float);
			p = f;
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

func TestC99FlexibleArrayAllowsPreviousNamedBitField(t *testing.T) {
	mustAnalyze(t, `struct s { int a:1; int x[]; };`)
}

func TestC99FlexibleArrayAllowsStandaloneUnionContainingFAMAggregate(t *testing.T) {
	mustAnalyze(t, `
		struct flex { int a; int b[]; };
		union rf1 { struct flex a; int b; };
		union rf2 { int a; union rf1 b; };
	`)
}

func TestC99RejectsInvalidFlexibleArrayPlacement(t *testing.T) {
	mustReject(t, `struct s1 { int x[]; };`)
	mustReject(t, `struct s2 { int x[]; int y; };`)
	mustReject(t, `union u { int a; char b[]; };`)
}

func TestC99RejectsInvalidFlexibleArrayTypedefPlacement(t *testing.T) {
	mustReject(t, `typedef int A[]; struct s1 { A x; };`)
	mustReject(t, `typedef int A[]; struct s2 { A x; int y; };`)
	mustReject(t, `typedef char T[]; union u { int a; T b; };`)
}

func TestC99RejectsArrayOfNonObjectTypes(t *testing.T) {
	mustReject(t, `struct incomplete a[1];`)
	mustReject(t, `int f(void)[1];`)
}

func TestC99AcceptsPrototypeStarArrayObjectTypes(t *testing.T) {
	mustAnalyze(t, `void f(int a[*][*]);`)
}

func TestC99RejectsNonPrototypeStarArrayDeclarations(t *testing.T) {
	mustReject(t, `int a[*][*];`)
	mustReject(t, `struct s { int x[*][*]; };`)
}

func TestC99RejectsInvalidIntegerConstantExpressions(t *testing.T) {
	mustReject(t, `int f(int n) { int a[-1]; return 0; }`)
	mustReject(t, `void f(void) { int a[(int)-1.0]; }`)
	mustReject(t, `void f(void) { int a[(int)(double)0.0]; }`)
	mustReject(t, `enum E { A = -1 << 0 };`)
	mustReject(t, `enum { A = (int)(double)1.0 };`)
	mustReject(t, `int i = -1 << 0;`)
	mustReject(t, `void f(void) { static int i = -1 << 0; }`)
	mustReject(t, `static int i = { -1 << 0 };`)
	mustReject(t, `static int a[1] = { -1 << 0 };`)
	mustReject(t, `void f(void) { static int i = { -1 << 0 }; }`)
	mustReject(t, `void f(void) { static int b[1] = { -1 << 0 }; }`)
	mustReject(t, `int f(int n) { switch (n) { case n: return 1; } return 0; }`)
}

func TestC99AcceptsShortCircuitIntegerConstantExpressions(t *testing.T) {
	mustAnalyze(t, `enum { A = 1 || (1/0) };`)
	mustAnalyze(t, `int f(int n) { switch (n) { case 1 || (1/0): return 1; } return 0; }`)
}

func TestC99SizeofPointerToVLAIsIntegerConstantExpression(t *testing.T) {
	mustAnalyze(t, `void f(int n) { enum { A = sizeof(int (*)[n]) }; }`)
}

func TestC99RejectsSizeofVLAAsIntegerConstantExpression(t *testing.T) {
	mustReject(t, `void f(int n) { enum { A = sizeof(int[n]) }; }`)
	mustReject(t, `void f(int n) { enum { A = sizeof(*(int (*)[n])0) }; }`)
}

func TestC99RejectsNonIntegerArrayDesignator(t *testing.T) {
	mustReject(t, `int a[] = { [(void *)0] = 1 };`)
}

func TestC99RejectsNegativeArrayDesignator(t *testing.T) {
	mustReject(t, `int a[] = { [-1] = 1 };`)
}

func TestC99AcceptsVLASizeAndNullPointerInitializers(t *testing.T) {
	mustAnalyze(t, `
		void f(int n) {
			int a[n];
			int *p = (void *)0;
			(void)a;
			(void)p;
		}
	`)
}

func TestC99AcceptsArithmeticConstantStaticInitializer(t *testing.T) {
	mustAnalyze(t, `static int y = { (int)+1.0 };`)
}

func TestC99UnaryFloatCastArrayBoundsRemainVLA(t *testing.T) {
	mustAnalyze(t, `
		void f(int m) {
			int a5[(int)+1.0];
			int a6[(int)+2.0];
			void *p = m ? &a5 : &a6;
			(void)p;
		}
	`)
}

func TestC99NestedFloatCastArrayBoundsRemainVLA(t *testing.T) {
	mustAnalyze(t, `
		void f(int m) {
			int a5[(int)(double)1.0];
			int a6[(int)(double)2.0];
			void *p = m ? &a5 : &a6;
			(void)p;
		}
	`)
}

func TestC99DirectFloatCastArrayBoundsAreFixed(t *testing.T) {
	mustReject(t, `
		void f(int m) {
			int a5[(int)1.0];
			int a6[(int)2.0];
			void *p = m ? &a5 : &a6;
			(void)p;
		}
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
