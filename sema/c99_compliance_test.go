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
	mustReject(t, `int n; enum { A = (1 ? 1 : n) };`)
	mustReject(t, `int n; enum { A = (1 || n) };`)
	mustReject(t, `int n; enum { A = (0 && n) };`)
	mustReject(t, `int n; int a[(1 || n)];`)
	mustReject(t, `int n; void f(void) { static int a[(1 || n)]; }`)
	mustReject(t, `int n; static int (*p)[n];`)
	mustReject(t, `int n; typedef int (*P)[n];`)
	mustReject(t, `int n; typedef int A[n];`)
	mustReject(t, `int i = -1 << 0;`)
	mustReject(t, `void f(void) { static int i = -1 << 0; }`)
	mustReject(t, `static int i = { -1 << 0 };`)
	mustReject(t, `static int a[1] = { -1 << 0 };`)
	mustReject(t, `void f(void) { static int i = { -1 << 0 }; }`)
	mustReject(t, `void f(void) { static int b[1] = { -1 << 0 }; }`)
	mustReject(t, `void f(void) { int a[1]; static int (*p)[1] = (int (*)[1])a; }`)
	mustReject(t, `static int sa[100]; int f(int n) { static int (*a)[n] = (int (*)[n])sa + 1; return n; }`)
	mustReject(t, `static int sa[100]; int f(int n, int m) { typedef int T; static int (*p)[n] = (T (*)[m++])sa; return n + m; }`)
	mustReject(t, `void f(void) { static int *p = &(int){1}; }`)
	mustReject(t, `int g; static int *p = &(int){g};`)
	mustReject(t, `int f(int n) { switch (n) { case n: return 1; } return 0; }`)
}

func TestC99AcceptsShortCircuitIntegerConstantExpressions(t *testing.T) {
	mustAnalyze(t, `enum { A = 1 || (1/0) };`)
	mustAnalyze(t, `enum { A = 1 || (-1 << 0) };`)
	mustAnalyze(t, `enum { A = 1 ? 1 : (1/0) };`)
	mustAnalyze(t, `enum { A = 1 ? 1 : (-1 << 0) };`)
	mustAnalyze(t, `int f(int n) { switch (n) { case 1 || (1/0): return 1; } return 0; }`)
}

func TestC99RejectsNonConstantOperandsInUnevaluatedIntegerConstantExpression(t *testing.T) {
	mustReject(t, `int n; int f(int x) { switch (x) { case 1 || n: return 1; } return 0; }`)
	mustReject(t, `int n; int f(int x) { switch (x) { case 0 && n: return 1; } return 0; }`)
	mustReject(t, `int n; int f(int x) { switch (x) { case 1 ? 1 : n: return 1; } return 0; }`)
}

func TestC99ShortCircuitArrayBoundCanRemainVLA(t *testing.T) {
	mustAnalyze(t, `int n; void f(void) { int a[(1 || n)]; (void)a; }`)
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
	mustAnalyze(t, `void f(int n) { static int (*p)[n]; (void)p; }`)
	mustAnalyze(t, `static int sa[100]; int f(int n) { static int (*a1)[n] = &sa; return n; }`)
	mustAnalyze(t, `static int sa[100]; int f(int n) { static int (*a2)[n] = (int (*)[n])sa; return n; }`)
	mustAnalyze(t, `static int sa[100]; int f(int n) { static int (*a3)[n] = (int (*)[(int){n}])sa; return n; }`)
	mustAnalyze(t, `static int sa[100]; int f(int n) { typedef int (*vmt)[n]; static vmt a = (vmt)sa; return n; }`)
	mustAnalyze(t, `static int sa[100]; int f(int n, int m) { typedef int (*vmt)[m++]; static int (*a18)[n] = (vmt)sa; return n + m; }`)
	mustAnalyze(t, `static int sa[100]; int f(int m) { typedef int A[m++]; static A *p = (A *)sa; return m; }`)
}

func TestC99RejectsGotoIntoVLAScope(t *testing.T) {
	mustReject(t, `
		void f(int n) {
			goto inside;
			{
				int a[n];
			inside:
				(void)a;
			}
		}
	`)
}

func TestC99AcceptsGotoOutOfVLAScope(t *testing.T) {
	mustAnalyze(t, `
		void f(int n) {
			{
				int a[n];
				(void)a;
				goto out;
			}
		out:
			;
		}
	`)
}

func TestC99RejectsSwitchJumpIntoVLAScope(t *testing.T) {
	mustReject(t, `
		void f(int n, int x) {
			switch (x) {
				int a[n];
			case 1:
				(void)a;
			default:
				;
			}
		}
	`)
}

func TestC99RejectsSwitchJumpIntoVLAScopeThroughLoopCase(t *testing.T) {
	mustReject(t, `
		void f(int n, int x) {
			switch (x) {
				while (1) {
					int a[n];
				case 1:
					(void)a;
				}
			}
		}
	`)
}

func TestC99AcceptsNestedSwitchCaseAfterOuterVLA(t *testing.T) {
	mustAnalyze(t, `
		void f(int n, int x, int y) {
			switch (x) {
				int a[n];
				(void)a;
				switch (y) {
				case 1:
					;
				}
			}
		}
	`)
}

func TestC99RejectsGotoIntoTypedefVMScope(t *testing.T) {
	mustReject(t, `
		void f(int n) {
			goto inside;
			{
				typedef int (*A)[n];
			inside:
				(void)sizeof(A);
			}
		}
	`)
}

func TestC99AcceptsFileScopeCompoundLiteralAddressInitializer(t *testing.T) {
	mustAnalyze(t, `static int *p = &(int){1};`)
}

func TestC99AcceptsArithmeticConstantStaticInitializer(t *testing.T) {
	mustAnalyze(t, `static double x = 1.0 + 2.0;`)
	mustAnalyze(t, `static double x = { 1.0 + 2.0 };`)
	mustAnalyze(t, `static int x = (int)(1.0 + 2.0);`)
	mustAnalyze(t, `static int x = { (int)(1.0 + 2.0) };`)
	mustAnalyze(t, `static int y = { (int)+1.0 };`)
	mustAnalyze(t, `int n; static int x = 1 || n;`)
	mustAnalyze(t, `int n; static int x = 0 && n;`)
	mustAnalyze(t, `int n; static int x = 1 ? 2 : n;`)
}

func TestC99AcceptsUnsignedLeftShiftIntegerConstantExpression(t *testing.T) {
	mustAnalyze(t, `static unsigned x = (unsigned)-1 << 0;`)
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

func TestC99RejectsInvalidRestrictTargets(t *testing.T) {
	mustAnalyze(t, `
		typedef int *ipa[2];
		int *restrict x[2];
		restrict ipa y;
		void f(int *restrict a[2], restrict ipa b, int *restrict c[restrict]);
	`)
	mustAnalyze(t, `void f(int a[restrict]);`)
	mustAnalyze(t, `void f(int a[restrict]) { (void)a; }`)
	mustReject(t, `int restrict x;`)
	mustReject(t, `typedef void (*fp)(void); fp restrict f;`)
	mustReject(t, `void f(int restrict a[3]);`)
	mustReject(t, `restrict struct s;`)
}

func TestC99RejectsBoolBitFieldTooWide(t *testing.T) {
	mustAnalyze(t, `struct s { _Bool : 0; _Bool b : 1; int i : 3; };`)
	mustReject(t, `struct s { _Bool b : 2; };`)
}

func TestC99RejectsComplexIncrementAndDecrement(t *testing.T) {
	mustAnalyze(t, `_Complex double f(_Complex double z) { return z + 1.0; }`)
	mustReject(t, `_Complex double f(_Complex double z) { z++; return z; }`)
	mustReject(t, `_Complex double f(_Complex double z) { ++z; return z; }`)
	mustReject(t, `_Complex double f(_Complex double z) { z--; return z; }`)
	mustReject(t, `_Complex double f(_Complex double z) { --z; return z; }`)
}

func TestC99RejectsInvalidForInitDeclarations(t *testing.T) {
	mustAnalyze(t, `
		void f(void) {
			for (int i = 0; i < 1; i++) {}
			for (auto int j = 0; j < 1; j++) {}
			for (register int k = 0; k < 1; k++) {}
		}
	`)
	mustReject(t, `void f(void) { for (int i = 0, bar(void); ; ) {} }`)
	mustReject(t, `void f(void) { for (static int i = 0; ; ) {} }`)
	mustReject(t, `void f(void) { for (extern int i; ; ) {} }`)
	mustReject(t, `void f(void) { for (enum { FOO } i = FOO; ; ) {} }`)
	mustReject(t, `void f(void) { for (enum BAR { FOO } i = FOO; ; ) {} }`)
	mustReject(t, `void f(void) { for (typedef int T; ; ) {} }`)
	mustReject(t, `void f(void) { for (struct s { int x; } *p = 0; ; ) {} }`)
	mustReject(t, `void f(void) { for (union u { int x; } *p = 0; ; ) {} }`)
}

func TestC99RejectsOldStyleImplicitIntParameter(t *testing.T) {
	mustAnalyze(t, `void f(int a) { (void)a; }`)
	mustReject(t, `void f(a) { }`)
	mustReject(t, `void f(a) int (*p)(int a); { }`)
}

func TestC99RejectsUsedStaticFunctionWithoutDefinition(t *testing.T) {
	mustAnalyze(t, `
		static int f0(void);
		static int f1(void) { return 1; }
		int (*p1)(void) = f1;
		static int f3(void);
		void g3(void) { sizeof(f3()); }
		static int f4(void);
		void g4(void) { sizeof(int (*)[f4()]); }
		static int f6(void);
		void g6(void) { sizeof(sizeof(int [f6()])); }
	`)
	mustReject(t, `static void f0(void); void g0(void) { f0(); }`)
	mustReject(t, `static void f1(void); void g1(void) { if (0) { f1(); } }`)
	mustReject(t, `static int f2(void); void g2(void) { 0 ? f2() : 0; }`)
	mustReject(t, `static int f5(void); void g5(void) { sizeof(int [0 ? f5() : 1]); }`)
	mustReject(t, `static int f(void); int (*p)(void) = f;`)
	mustReject(t, `static int f(void); static int (*p)(void) = f;`)
}

func TestC99RejectsQualifiedEmptyTagRedeclarations(t *testing.T) {
	mustAnalyze(t, `
		struct s0;
		struct s0 { int a; };
		struct s0;
		void f(void) { struct s0; }
		const union u0;
		union u0 { long b; };
		extern struct s1;
		struct s3;
		const struct s3 { int a; };
		union u4;
		extern union u4 { int z; };
		enum e0 { E0 };
		void i(void) { const enum e0 { E1 }; }
		union u5 { int p; };
		void j(void) { extern struct u5 { int q; }; }
	`)
	mustReject(t, `struct s2 { char x; }; const struct s2;`)
	mustReject(t, `union u1; extern union u1;`)
	mustReject(t, `union u2 { long b; }; void g(void) { const union u2; }`)
	mustReject(t, `union u3 { float v; }; void h(void) { const struct u3; }`)
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
