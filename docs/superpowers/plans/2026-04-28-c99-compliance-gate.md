# C99 Compliance Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every non-preprocessor-dependent GCC `gcc.dg/c99*.c` case pass through cvm's current `lexer -> parser -> sema.Analyze` pipeline according to GCC's accept/reject expectation.

**Architecture:** Keep the existing sema typed AST as the frontend contract. Tighten the GCC C99 fixture runner so `skipped` is allowed only for cases that truly require preprocessing, system headers, or GCC predefined macro/type expansion. Then close all remaining lexer/parser/sema mismatches in focused semantic slices, committing after each slice when the full relevant test set is green.

**Tech Stack:** Go stdlib, existing cvm `lexer`, `parser`, and `sema` packages, local GCC sparse checkout at `/tmp/cvm-gcc-c99` when available.

**Out of scope:** Intermediate representation generation, backend lowering, runtime execution of GCC `dg-do run` programs, and a full C preprocessor.

---

## Acceptance Criteria

- `GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99' -count=1 -v` passes.
- `GOCACHE=/tmp/cvm-go-cache go test ./... -count=1` passes.
- `sema/testdata/gcc-c99/manifest.tsv` has no `skipped` row whose reason is `current semantic layer accepts this GCC reject case`.
- `sema/testdata/gcc-c99/manifest.tsv` has no `skipped` row whose reason starts with `current lexer+parser+sema rejects this GCC accept case`.
- Every remaining `skipped` row has the reason `requires preprocessor or system macro handling`.
- GCC warning-only cases are treated as accept cases because cvm does not currently expose a warning diagnostics channel.

## File Structure

**Modify:**
- `sema/gcc_c99_test.go` - hard gate for manifest skip reasons, optional manifest refresh helper tests, and sharper failure messages.
- `sema/testdata/gcc-c99/manifest.tsv` - reclassify imported/skipped cases after each semantic slice.
- `sema/testdata/gcc-c99/README.md` - document the stricter skip policy.
- `sema/sema.go` - inject `__func__`, track translation-unit deferred diagnostics.
- `sema/init.go` - implement string-to-character-array initialization and stricter designator checks.
- `sema/cast.go` - implement pointer compatibility, qualifier preservation, and function-pointer/`void *` constraints.
- `sema/type.go` and `sema/type_table.go` - add helpers for compatible types, object types, variably modified types, and flexible array detection.
- `sema/spec.go` and `sema/decl.go` - validate flexible arrays, bit-fields, `restrict`, array element legality, `for` declarations, and tag redeclarations.
- `sema/stmt.go` - reject jumps into variably modified scopes and validate `return`.
- `sema/const.go` - distinguish C99 integer constant expressions from broader foldable constants.
- `sema/expr.go` - tighten inc/dec, comparison, conditional, compound literal, and assignment behavior.
- `entity/token.go` - add a `_Static_assert` token only if current tokenization treats it as a plain identifier.
- `lexer/identifier.go` - map `_Static_assert` to the new token only if `entity/token.go` needs that token.
- `parser/phrase_structure_grammar.txt` - add a declaration production for `_Static_assert ( constant_expression ) ;` only if current parser rejects the warning-only GCC case.
- `parser/glr_table.go` - regenerate from `parser/phrase_structure_grammar.txt` after grammar changes.

**Create:**
- `sema/c99_compliance_test.go` - focused regression tests for each mismatch category before importing the GCC case.
- `sema/type_compat.go` - type compatibility and qualifier helper functions if `cast.go` becomes too broad.
- `sema/vla_scope.go` - variably modified scope tracking if the logic is too large for `stmt.go`.

---

## Task 1: Turn GCC C99 Manifest Into A Hard Compliance Gate

**Files:**
- Modify: `sema/gcc_c99_test.go`
- Modify: `sema/testdata/gcc-c99/README.md`
- Test: `sema/gcc_c99_test.go`

- [ ] **Step 1: Add a failing manifest gate test**

Add this test to `sema/gcc_c99_test.go`:

```go
func TestGCCC99OnlyPreprocessorSkipsRemain(t *testing.T) {
	manifest := filepath.Join("testdata", "gcc-c99", "manifest.tsv")
	content, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read GCC C99 manifest: %v", err)
	}
	for lineNo, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" || lineNo == 0 {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("manifest line %d malformed: %q", lineNo+1, line)
		}
		status, reason := fields[1], fields[3]
		if status != "skipped" {
			continue
		}
		if reason != "requires preprocessor or system macro handling" {
			t.Fatalf("non-preprocessor GCC C99 skip remains at line %d: %s: %s", lineNo+1, fields[0], reason)
		}
	}
}
```

- [ ] **Step 2: Run the new gate and confirm it fails**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run TestGCCC99OnlyPreprocessorSkipsRemain -count=1 -v
```

Expected: FAIL on the first non-preprocessor skipped mismatch, currently a case such as `c99-array-lval-8.c`.

- [ ] **Step 3: Document the strict policy**

Update `sema/testdata/gcc-c99/README.md` to state:

```markdown
严格 skip 策略：

- 只有依赖预处理器、系统头、GCC predefined macro/type 的用例可以保留为 `skipped`。
- GCC warning-only 用例按 accept 处理，因为当前 cvm 没有 warning 通道。
- 没有 `dg-error` 的 GCC C99 用例必须被 `lexer -> parser -> sema.Analyze` 接受。
- 有 `dg-error` 的 GCC C99 用例必须被 `lexer -> parser -> sema.Analyze` 拒绝。
```

- [ ] **Step 4: Commit the failing gate**

Run:

```bash
gofmt -w sema/gcc_c99_test.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run TestGCCC99OnlyPreprocessorSkipsRemain -count=1 -v
```

Expected: FAIL for the gate test.

Commit:

```bash
git add sema/gcc_c99_test.go sema/testdata/gcc-c99/README.md
git commit -m "test(sema): require GCC C99 skips to be preprocessor-only"
```

## Task 2: Accept GCC C99 Warning-Only And Valid Accept Cases

**Files:**
- Create: `sema/c99_compliance_test.go`
- Modify: `sema/sema.go`
- Modify: `sema/init.go`
- Modify: parser/lexer files only if `_Static_assert` is not tokenized or parsed
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused accept regression tests**

Create `sema/c99_compliance_test.go`:

```go
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

func TestC99ForwardEnumWarningOnlyIsAccepted(t *testing.T) {
	mustAnalyze(t, `
		enum e1;
		enum e2 { E };
		enum e2;
	`)
}
```

Use the existing test helper pattern in `sema/test_util_test.go`; if there is no `mustAnalyze`, add this helper to `sema/c99_compliance_test.go`:

```go
func mustAnalyze(t *testing.T, src string) *Program {
	t.Helper()
	candidates := parseCandidates(t, src)
	prog, err := Analyze(candidates)
	if err != nil {
		t.Fatalf("Analyze rejected source: %v", err)
	}
	return prog
}
```

- [ ] **Step 2: Run focused tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99(FuncIdentifier|StringLiteral|ForwardEnum)' -count=1 -v
```

Expected: FAIL for `__func__`, string literal array initialization, and current forward enum handling.

- [ ] **Step 3: Implement `__func__` injection**

In `sema/sema.go`, inside `walkFunctionBody`, after inserting parameters and before typing the body, insert a function-local symbol:

```go
funcType := s.Types.ArrayConstant(s.Types.Qual(s.Types.Builtin(Char), true, false, false), int64(len(pf.def.Sym.Name)+1))
funcSym := &Symbol{
	Name:    "__func__",
	Kind:    SymVar,
	T:       funcType,
	Storage: StorageStatic,
	Linkage: LinkageNone,
	Pos:     pf.def.Range.SourceStart,
}
_ = bodyScope.InsertChecked("__func__", funcSym)
```

If `TypeTable` does not expose `Qual`, add or reuse the local qualifier constructor following existing `QualType` patterns. The comment for this block must be Chinese:

```go
// C99 要求每个函数体内隐式声明 __func__，类型等价于 static const char[N]。
```

- [ ] **Step 4: Implement string literal initialization for character arrays**

In `sema/init.go`, before `typeInitializer` applies normal expression assignment conversion, add a special path:

```go
func (s *Sema) tryStringArrayInitializer(node *entity.AstNode, target Type) Expr {
	at, ok := unqual(target).(*ArrayType)
	if !ok || !isCharacterType(unqual(at.Elem)) {
		return nil
	}
	if !node.ReducedBy(parser.Initializer, 2) && !node.ReducedBy(parser.Initializer, 3) {
		return nil
	}
	list := node.Children[1]
	if !list.ReducedBy(parser.InitializerList, 1) {
		return nil
	}
	elem := list.Children[0]
	if !elem.ReducedBy(parser.Initializer, 1) {
		return nil
	}
	expr := s.typeExpr(elem.Children[0], s.scope)
	if _, ok := expr.(*StringLit); !ok {
		return nil
	}
	return expr
}
```

Call it at the top of `typeInitializer`:

```go
if expr := s.tryStringArrayInitializer(node, target); expr != nil {
	return expr
}
```

Add `isCharacterType` near other type helpers:

```go
func isCharacterType(t Type) bool {
	bt, ok := unqual(t).(*BuiltinType)
	return ok && (bt.Kind == Char || bt.Kind == SChar || bt.Kind == UChar)
}
```

- [ ] **Step 5: Accept forward enum declarations**

In `sema/spec.go`, change `buildEnum` for `enum identifier ;` style declarations so it creates or returns an incomplete enum type instead of reporting an undeclared identifier. Preserve completion when an enumerator list appears later:

```go
case node.ReducedBy(parser.EnumSpecifier, 5):
	name := node.Children[1].Terminal.Lexeme
	if existing := s.scope.LookupTag(name); existing != nil {
		return existing.T
	}
	tag := s.Types.NewTagID()
	et := s.Types.Enum(tag)
	_ = s.scope.InsertTagChecked(name, &TagInfo{Tag: tag, T: et}, node.SourceStart)
	return et
```

Add a Chinese comment above this branch:

```go
// GCC 的 C99 warning-only 用例会接受 enum 前向声明；当前没有 warning 通道，因此按可继续分析处理。
```

- [ ] **Step 6: Add `_Static_assert` acceptance if parser still rejects warning-only case**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run TestGCCC99AcceptSuite/c99-static-assert-2.c -count=1 -v
```

If it fails in lexer/parser, add `_Static_assert ( constant_expression );` as an ignored declaration form by changing these exact files:

```text
entity/token.go
lexer/identifier.go
parser/phrase_structure_grammar.txt
parser/glr_table.go
```

The sema behavior is: evaluate the expression; if it is zero, reject; otherwise emit no declaration. Add a focused test:

```go
func TestC99StaticAssertWarningOnlySyntaxIsAccepted(t *testing.T) {
	mustAnalyze(t, `_Static_assert(1);`)
}
```

- [ ] **Step 7: Move accept cases into fixtures**

Move these files from skipped to `accept/` when they pass:

```text
c99-complit-1.c
c99-func-1.c
c99-func-3.c
c99-init-1.c
c99-static-assert-2.c
c99-tag-5.c
c99-tag-6.c
```

Keep `c99-init-1.c` skipped only if the failure is still caused by `__SIZE_TYPE__` or `__WCHAR_TYPE__` macro expansion. Do not skip it for initializer semantics.

- [ ] **Step 8: Verify and commit accept slice**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99|TestGCCC99AcceptSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS for focused accept tests and current accept suite.

Commit:

```bash
git add sema parser lexer
git commit -m "fix(sema): accept non-preprocessor GCC C99 valid cases"
```

## Task 3: Tighten Pointer, Qualifier, And Tag Compatibility

**Files:**
- Create: `sema/type_compat.go`
- Modify: `sema/cast.go`
- Modify: `sema/expr.go`
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused rejection tests**

Append:

```go
func TestC99RejectsDiscardedPointerQualifiers(t *testing.T) {
	mustReject(t, `
		struct s { const int a[1]; };
		void f(struct s s) {
			int *p = s.a;
			(void)p;
		}
	`)
}

func TestC99RejectsFunctionPointerVoidPointerAssignment(t *testing.T) {
	mustReject(t, `
		void f(void) {}
		void (*fp)(void);
		void g(void) {
			fp = (void *)0;
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
```

Add helper:

```go
func mustReject(t *testing.T, src string) {
	t.Helper()
	candidates := parseCandidates(t, src)
	if _, err := Analyze(candidates); err == nil {
		t.Fatalf("Analyze accepted invalid source")
	}
}
```

- [ ] **Step 2: Run tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(Discarded|FunctionPointer|DifferentStruct)' -count=1 -v
```

Expected: FAIL because current pointer conversion is too permissive.

- [ ] **Step 3: Add compatibility helpers**

Create `sema/type_compat.go`:

```go
package sema

func compatibleType(a, b Type) bool {
	a = unqual(a)
	b = unqual(b)
	if a == b {
		return true
	}
	switch ax := a.(type) {
	case *BuiltinType:
		bx, ok := b.(*BuiltinType)
		return ok && ax.Kind == bx.Kind
	case *PointerType:
		bx, ok := b.(*PointerType)
		return ok && compatibleType(ax.Pointee, bx.Pointee)
	case *ArrayType:
		bx, ok := b.(*ArrayType)
		return ok && compatibleType(ax.Elem, bx.Elem)
	case *FunctionType:
		bx, ok := b.(*FunctionType)
		return ok && compatibleFunctionType(ax, bx)
	case *StructType:
		bx, ok := b.(*StructType)
		return ok && ax.Tag == bx.Tag
	case *UnionType:
		bx, ok := b.(*UnionType)
		return ok && ax.Tag == bx.Tag
	case *EnumType:
		bx, ok := b.(*EnumType)
		return ok && ax.Tag == bx.Tag
	}
	return false
}

func compatibleFunctionType(a, b *FunctionType) bool {
	if !compatibleType(a.Ret, b.Ret) || a.Variadic != b.Variadic || a.HasProto != b.HasProto {
		return false
	}
	if len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if !compatibleType(a.Params[i], b.Params[i]) {
			return false
		}
	}
	return true
}

func losesQualifier(from, to Type) bool {
	fq, fok := from.(*QualType)
	tq, tok := to.(*QualType)
	if fok {
		if !tok {
			return fq.Const || fq.Volatile || fq.Restrict
		}
		if fq.Const && !tq.Const {
			return true
		}
		if fq.Volatile && !tq.Volatile {
			return true
		}
		if fq.Restrict && !tq.Restrict {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Tighten assignment and comparison conversion**

In `sema/cast.go`, replace the unconditional pointer conversion inside `assignmentConversion` with:

```go
if pf, ok := unqual(from).(*PointerType); ok {
	if pt, ok := unqual(target).(*PointerType); ok {
		if pointsToFunction(pf) || pointsToFunction(pt) {
			if compatibleType(pf.Pointee, pt.Pointee) {
				return s.castPointerConversion(e, target)
			}
			s.report(IncompatibleAssignment(pos, from.String(), target.String()))
			return e
		}
		if losesQualifier(pf.Pointee, pt.Pointee) {
			s.report(IncompatibleAssignment(pos, from.String(), target.String()))
			return e
		}
		if isVoidPointer(pf) || isVoidPointer(pt) || compatibleType(pf.Pointee, pt.Pointee) {
			return s.castPointerConversion(e, target)
		}
		s.report(IncompatibleAssignment(pos, from.String(), target.String()))
		return e
	}
}
```

Add:

```go
func pointsToFunction(p *PointerType) bool {
	_, ok := unqual(p.Pointee).(*FunctionType)
	return ok
}
```

In `sema/expr.go`, make relational pointer comparison reject incompatible object pointer types, while equality still allows null pointer constants.

- [ ] **Step 5: Move relevant GCC reject cases into fixtures**

Move passing reject cases into `sema/testdata/gcc-c99/reject/` and update manifest:

```text
c99-array-lval-8.c
c99-compare-incomplete-2.c
c99-const-expr-5.c
c99-tag-2.c
```

- [ ] **Step 6: Verify and commit pointer slice**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

Commit:

```bash
git add sema
git commit -m "fix(sema): enforce C99 pointer compatibility constraints"
```

## Task 4: Enforce Flexible Array And Object-Type Declaration Constraints

**Files:**
- Modify: `sema/spec.go`
- Modify: `sema/decl.go`
- Modify: `sema/type_compat.go`
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused rejection tests**

Append:

```go
func TestC99RejectsInvalidFlexibleArrayPlacement(t *testing.T) {
	mustReject(t, `struct s1 { int x[]; };`)
	mustReject(t, `struct s2 { int x[]; int y; };`)
	mustReject(t, `union u { int a; char b[]; };`)
}

func TestC99RejectsArrayOfNonObjectTypes(t *testing.T) {
	mustReject(t, `struct incomplete a[1];`)
	mustReject(t, `int f(void)[1];`)
}
```

- [ ] **Step 2: Run tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidFlexible|ArrayOf)' -count=1 -v
```

Expected: FAIL on current sema.

- [ ] **Step 3: Add object and flexible array helpers**

Add:

```go
func isObjectType(t Type) bool {
	switch x := unqual(t).(type) {
	case *BuiltinType:
		return x.Kind != Void
	case *FunctionType:
		return false
	case *StructType:
		return x.Complete
	case *UnionType:
		return x.Complete
	case *EnumType:
		return x.Complete
	default:
		return true
	}
}

func isFlexibleArrayMember(t Type) bool {
	at, ok := unqual(t).(*ArrayType)
	return ok && at.SizeKind == ArrayUnsized
}
```

- [ ] **Step 4: Validate arrays in `makeArray`**

In `sema/decl.go`, inside `makeArray`, report an error if `elem` is not an object type or is an incomplete array/function type:

```go
if !isObjectType(elem) {
	s.report(InvalidTypeSpec(sizeNode.SourceStart, "array element type must be a complete object type"))
}
```

- [ ] **Step 5: Validate struct and union flexible members**

In `sema/spec.go`, after `parseStructDeclList` returns fields and before `completeStructUnion`, call:

```go
s.validateFlexibleArrayMembers(fields, isUnion, node.SourceStart)
```

Implement:

```go
func (s *Sema) validateFlexibleArrayMembers(fields []*Field, isUnion bool, pos entity.SourcePos) {
	for i, f := range fields {
		if !isFlexibleArrayMember(f.T) {
			continue
		}
		if isUnion {
			s.report(InvalidTypeSpec(pos, "flexible array member cannot appear in union"))
			continue
		}
		if f.Name == "" {
			s.report(InvalidTypeSpec(pos, "flexible array member requires a named preceding member"))
		}
		if i != len(fields)-1 {
			s.report(InvalidTypeSpec(pos, "flexible array member must be last"))
		}
		namedBefore := false
		for j := 0; j < i; j++ {
			if fields[j].Name != "" && !fields[j].IsBitField {
				namedBefore = true
				break
			}
		}
		if !namedBefore {
			s.report(InvalidTypeSpec(pos, "flexible array member requires another named member"))
		}
	}
}
```

Use a Chinese comment above the function:

```go
// C99 flexible array member 只能出现在结构体最后，并且前面必须有其他具名成员。
```

- [ ] **Step 6: Move relevant GCC reject cases into fixtures**

Move:

```text
c99-array-nonobj-1.c
c99-flex-array-1.c
c99-flex-array-3.c
c99-flex-array-5.c
c99-flex-array-typedef-1.c
c99-flex-array-typedef-3.c
c99-flex-array-typedef-5.c
```

- [ ] **Step 7: Verify and commit declaration slice**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidFlexible|ArrayOf)|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

Commit:

```bash
git add sema
git commit -m "fix(sema): enforce C99 flexible array constraints"
```

## Task 5: Implement C99 Constant Expression Constraints

**Files:**
- Modify: `sema/const.go`
- Modify: `sema/decl.go`
- Modify: `sema/spec.go`
- Modify: `sema/init.go`
- Modify: `sema/stmt.go`
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused rejection tests**

Append:

```go
func TestC99RejectsInvalidIntegerConstantExpressions(t *testing.T) {
	mustReject(t, `int f(int n) { int a[-1]; return 0; }`)
	mustReject(t, `enum E { A = -1 << 0 };`)
	mustReject(t, `int f(int n) { switch (n) { case n: return 1; } return 0; }`)
}

func TestC99RejectsNonIntegerArrayDesignator(t *testing.T) {
	mustReject(t, `int a[] = { [(void *)0] = 1 };`)
}
```

- [ ] **Step 2: Run tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidInteger|NonIntegerArray)' -count=1 -v
```

Expected: FAIL on at least one case.

- [ ] **Step 3: Add a stricter ICE evaluator**

In `sema/const.go`, add:

```go
func (e *Evaluator) EvalC99IntegerConstantExpression(expr Expr) (ConstValue, bool) {
	switch x := expr.(type) {
	case *IntLit, *CharLit, *EnumRef, *SizeofExpr:
		return e.EvalIntegerConstant(expr)
	case *BinOp:
		if x.Op == OpShl || x.Op == OpShr {
			l, lok := e.EvalC99IntegerConstantExpression(x.L)
			r, rok := e.EvalC99IntegerConstantExpression(x.R)
			if !lok || !rok || l.Int < 0 || r.Int < 0 {
				return ConstValue{}, false
			}
		}
		return e.EvalIntegerConstant(expr)
	case *UnOp:
		if x.Op == UnIncPre || x.Op == UnDecPre || x.Op == UnIncPost || x.Op == UnDecPost {
			return ConstValue{}, false
		}
		return e.EvalIntegerConstant(expr)
	case *AssignExpr, *CompoundAssign, *CallExpr, *CommaExpr, *CompoundLit:
		return ConstValue{}, false
	case *ImplicitCast:
		return e.EvalC99IntegerConstantExpression(x.X)
	case *ExplicitCast:
		if _, ok := unqual(x.To).(*BuiltinType); !ok {
			return ConstValue{}, false
		}
		return e.EvalIntegerConstant(expr)
	case *CondExpr:
		c, ok := e.EvalC99IntegerConstantExpression(x.Cond)
		if !ok {
			return ConstValue{}, false
		}
		if c.Int != 0 {
			return e.EvalC99IntegerConstantExpression(x.Then)
		}
		return e.EvalC99IntegerConstantExpression(x.Else)
	default:
		return ConstValue{}, false
	}
}
```

Keep `EvalIntegerConstant` for existing folding behavior; use the stricter method only where C99 requires an ICE.

- [ ] **Step 4: Use strict ICE in required locations**

Replace `EvalIntegerConstant` with `EvalC99IntegerConstantExpression` in:

```text
sema/decl.go: evalArraySize
sema/spec.go: evalBitWidth
sema/spec.go: parseEnumerator
sema/stmt.go: typeLabeled for case values
sema/init.go: parseDesignator for array designators
```

In `parseDesignator`, report when the expression is not an integer constant:

```go
if cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr); ok && isInteger(expr.GetType()) {
	return Designator{Kind: DesigArrayIndex, Index: cv.Int}
}
s.report(InvalidTypeSpec(node.SourceStart, "array designator must be an integer constant expression"))
return Designator{Kind: DesigArrayIndex}
```

- [ ] **Step 5: Reject zero and negative fixed array sizes where C99 requires it**

In `makeArray`, after strict size evaluation:

```go
if size == 0 {
	s.report(InvalidTypeSpec(sizeNode.SourceStart, "array size must be greater than zero"))
}
if size < 0 && sizeNode != nil {
	return s.Types.ArrayVLA(elem, s.typeExpr(sizeNode, s.scope))
}
```

Use current VLA behavior for non-constant positive runtime sizes.

- [ ] **Step 6: Move relevant GCC reject cases into fixtures**

Move:

```text
c99-const-expr-11.c
c99-const-expr-12.c
c99-init-3.c
c99-left-shift-1.c
```

- [ ] **Step 7: Verify and commit constant-expression slice**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidInteger|NonIntegerArray)|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

Commit:

```bash
git add sema
git commit -m "fix(sema): enforce C99 integer constant expressions"
```

## Task 6: Enforce Restrict, Bit-Field, For-Declaration, And Static Function Constraints

**Files:**
- Modify: `sema/decl.go`
- Modify: `sema/spec.go`
- Modify: `sema/sema.go`
- Modify: `sema/stmt.go`
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused rejection tests**

Append:

```go
func TestC99RejectsInvalidRestrictTargets(t *testing.T) {
	mustReject(t, `int restrict x;`)
	mustReject(t, `typedef void (*fp)(void); fp restrict f;`)
}

func TestC99RejectsBoolBitFieldTooWide(t *testing.T) {
	mustReject(t, `struct s { _Bool b : 2; };`)
}

func TestC99RejectsInvalidForInitDeclarations(t *testing.T) {
	mustReject(t, `void f(void) { for (static int i = 0; ; ) { break; } }`)
	mustReject(t, `void f(void) { for (struct s { int x; } *p = 0; ; ) { break; } }`)
}

func TestC99RejectsUsedStaticFunctionWithoutDefinition(t *testing.T) {
	mustReject(t, `static void f(void); void g(void) { f(); }`)
}
```

- [ ] **Step 2: Run tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidRestrict|BoolBitField|InvalidFor|UsedStatic)' -count=1 -v
```

Expected: FAIL on current sema.

- [ ] **Step 3: Restrict validation**

Add:

```go
func (s *Sema) validateRestrictType(t Type, pos entity.SourcePos) {
	q, ok := t.(*QualType)
	if !ok || !q.Restrict {
		return
	}
	pt, ok := unqual(q.Base).(*PointerType)
	if !ok || !isObjectType(pt.Pointee) {
		s.report(InvalidTypeSpec(pos, "restrict qualifier requires pointer to object type"))
	}
}
```

Call it after declarator application for variables, parameters, and fields.

- [ ] **Step 4: Bit-field width validation**

In `parseStructDeclarator`, after `evalBitWidth`, reject `_Bool` width greater than 1:

```go
if bt, ok := unqual(base).(*BuiltinType); ok && bt.Kind == Bool && width > 1 {
	s.report(InvalidTypeSpec(node.SourceStart, "_Bool bit-field width must not exceed 1"))
}
```

- [ ] **Step 5: For-init declaration validation**

In `collectForParts`, after `walkBlockDecl`, walk the produced declarations:

```go
for _, d := range decls {
	switch x := d.(type) {
	case *FuncDecl:
		s.report(InvalidTypeSpec(x.Range.SourceStart, "function declaration is not allowed in for init declaration"))
	case *VarDecl:
		if x.Storage == StorageStatic || x.Storage == StorageExtern {
			s.report(InvalidTypeSpec(x.Range.SourceStart, "static or extern declaration is not allowed in for init declaration"))
		}
	case *TagDecl:
		s.report(InvalidTypeSpec(x.Range.SourceStart, "tag definition is not allowed in for init declaration"))
	}
}
```

- [ ] **Step 6: Static function used but never defined**

In `Sema`, add a usage marker when `lookupVar` returns `SymFunc` from a call context or when a function reference decays. The minimal implementation may add `Used bool` to `Symbol`.

At the end of `analyzeOne`, after `foldStaticInitializers`, scan file-scope functions:

```go
s.validateStaticFunctionDefinitions(prog)
```

Implement:

```go
func (s *Sema) validateStaticFunctionDefinitions(prog *Program) {
	for _, d := range prog.Globals {
		fd, ok := d.(*FuncDecl)
		if !ok || fd.Storage != StorageStatic || fd.Sym == nil || !fd.Sym.Used {
			continue
		}
		hasDefinition := false
		for _, def := range fd.Sym.Defs {
			if _, ok := def.(*FuncDef); ok {
				hasDefinition = true
				break
			}
		}
		if !hasDefinition {
			s.report(InvalidTypeSpec(fd.Range.SourceStart, "static function used but never defined"))
		}
	}
}
```

- [ ] **Step 7: Move relevant GCC reject cases into fixtures**

Move:

```text
c99-bool-2.c
c99-complex-2.c
c99-fordecl-2.c
c99-fordecl-3.c
c99-impl-int-2.c
c99-restrict-1.c
c99-restrict-3.c
c99-static-1.c
c99-tag-3.c
```

For `c99-complex-2.c`, reject `++`/`--` on complex types in `typeIncDec`.

- [ ] **Step 8: Verify and commit miscellaneous constraints**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(InvalidRestrict|BoolBitField|InvalidFor|UsedStatic)|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

Commit:

```bash
git add sema
git commit -m "fix(sema): enforce remaining C99 declaration constraints"
```

## Task 7: Reject Jumps Into Variably Modified Scopes

**Files:**
- Create: `sema/vla_scope.go`
- Modify: `sema/stmt.go`
- Modify: `sema/decl.go`
- Test: `sema/c99_compliance_test.go`, `sema/gcc_c99_test.go`

- [ ] **Step 1: Add focused rejection tests**

Append:

```go
func TestC99RejectsGotoIntoVLAScope(t *testing.T) {
	mustReject(t, `
		void f(int n) {
			goto label;
			{
				int a[n];
			label:
				a[0] = 1;
			}
		}
	`)
}

func TestC99RejectsSwitchIntoVLAScope(t *testing.T) {
	mustReject(t, `
		void f(int n, int x) {
			switch (x) {
				int a[n];
			case 1:
				a[0] = 1;
			}
		}
	`)
}
```

- [ ] **Step 2: Run tests and confirm failures**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(GotoIntoVLA|SwitchIntoVLA)' -count=1 -v
```

Expected: FAIL.

- [ ] **Step 3: Track variably modified declarations per block**

Create `sema/vla_scope.go`:

```go
package sema

type vlaScopeInfo struct {
	scope *Scope
	depth int
}

func hasVariablyModifiedType(t Type) bool {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayVLA || x.SizeKind == ArrayStarSize {
			return true
		}
		return hasVariablyModifiedType(x.Elem)
	case *PointerType:
		return hasVariablyModifiedType(x.Pointee)
	case *FunctionType:
		for _, p := range x.Params {
			if hasVariablyModifiedType(p) {
				return true
			}
		}
		return hasVariablyModifiedType(x.Ret)
	}
	return false
}
```

- [ ] **Step 4: Extend `funcCtx`**

In `sema/stmt.go`:

```go
type funcCtx struct {
	def          *FuncDef
	loopDepth    int
	switchStack  []*SwitchStmt
	pendingGotos []*GotoStmt
	vlaScopes    map[*Scope]int
	labelScopes  map[*LabeledStmt]*Scope
	gotoScopes   map[*GotoStmt]*Scope
}
```

Initialize the maps in `walkFunctionBody`.

- [ ] **Step 5: Record VLA scopes**

In `walkBlockInitDecl`, after type creation:

```go
if ctx != nil && hasVariablyModifiedType(t) {
	if ctx.vlaScopes == nil {
		ctx.vlaScopes = map[*Scope]int{}
	}
	ctx.vlaScopes[scope]++
}
```

- [ ] **Step 6: Record label and goto scopes**

In `typeJump`, set `ctx.gotoScopes[g] = scope`.

In `typeLabeled`, create the label, type the body, and set `ctx.labelScopes[label] = scope`.

- [ ] **Step 7: Validate jumps after label resolution**

After `resolveGotos`, call:

```go
validateVLAJumps(ctx, s)
```

Implement:

```go
func validateVLAJumps(ctx *funcCtx, s *Sema) {
	for _, g := range ctx.pendingGotos {
		if g.Target == nil {
			continue
		}
		from := ctx.gotoScopes[g]
		to := ctx.labelScopes[g.Target]
		for cur := to; cur != nil && cur != from; cur = cur.Parent {
			if ctx.vlaScopes[cur] > 0 {
				s.report(InvalidTypeSpec(g.Range.SourceStart, "goto jumps into scope of identifier with variably modified type"))
				break
			}
		}
	}
}
```

For switch/case, reuse the same principle by recording case/default scopes and comparing against the switch statement's entry scope.

- [ ] **Step 8: Move relevant GCC reject cases into fixtures**

Move:

```text
c99-vla-jump-1.c
c99-vla-jump-2.c
c99-vla-jump-3.c
c99-vla-jump-4.c
c99-vla-jump-5.c
```

- [ ] **Step 9: Verify and commit VLA jump constraints**

Run:

```bash
gofmt -w sema/*.go
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestC99Rejects(GotoIntoVLA|SwitchIntoVLA)|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

Commit:

```bash
git add sema
git commit -m "fix(sema): reject jumps into variably modified scopes"
```

## Task 8: Refresh Manifest And Enforce Preprocessor-Only Skips

**Files:**
- Modify: `sema/testdata/gcc-c99/accept/*.c`
- Modify: `sema/testdata/gcc-c99/reject/*.c`
- Modify: `sema/testdata/gcc-c99/manifest.tsv`
- Modify: `sema/testdata/gcc-c99/README.md`
- Test: `sema/gcc_c99_test.go`

- [ ] **Step 1: Re-run the classifier**

Use the existing local sparse GCC checkout:

```bash
test -d /tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg
```

Expected: exit 0. If it is missing, recreate the sparse checkout from GCC before continuing.

Run the current classifier or regenerate by scanning:

```bash
find /tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg -maxdepth 1 -name 'c99*.c' | sort
```

For each candidate:

- strip DejaGNU comments with the same logic as `stripGCCDirectives`;
- if the source requires preprocessing/system macro handling, keep `skipped`;
- if GCC expects accept, place it in `accept/`;
- if GCC expects reject, place it in `reject/`;
- update `manifest.tsv`.

- [ ] **Step 2: Check manifest status counts**

Run:

```bash
awk -F '\t' 'NR>1 {count[$2 FS $3 FS $4]++} END {for (k in count) print count[k] FS k}' sema/testdata/gcc-c99/manifest.tsv | sort -nr
```

Expected: all `skipped` rows have reason `requires preprocessor or system macro handling`.

- [ ] **Step 3: Run the hard gate**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99' -count=1 -v
```

Expected: PASS, including `TestGCCC99OnlyPreprocessorSkipsRemain`.

- [ ] **Step 4: Run full verification**

Run:

```bash
gofmt -w sema/*.go
git diff --check
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected:

```text
git diff --check
# no output

go test ./...
# all packages pass
```

- [ ] **Step 5: Commit the final compliance gate**

Commit:

```bash
git add sema parser lexer
git commit -m "test(sema): enforce preprocessor-only GCC C99 skips"
```

## Task 9: Final Review Before Starting IR Work

**Files:**
- Read: `sema/testdata/gcc-c99/manifest.tsv`
- Read: `sema/testdata/gcc-c99/README.md`
- Read: `sema/gcc_c99_test.go`

- [ ] **Step 1: Confirm no IR work was included**

Run:

```bash
git diff --name-only main...HEAD | rg '(^ir/|lower|codegen|backend|middle|ssa|tac)' || true
```

Expected: no output.

- [ ] **Step 2: Confirm remaining skips are preprocessor-only**

Run:

```bash
awk -F '\t' 'NR>1 && $2=="skipped" && $4!="requires preprocessor or system macro handling" {print}' sema/testdata/gcc-c99/manifest.tsv
```

Expected: no output.

- [ ] **Step 3: Confirm final test suite**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: both commands pass.

- [ ] **Step 4: Commit final documentation if counts changed**

If `README.md` count lines changed, commit:

```bash
git add sema/testdata/gcc-c99/README.md sema/testdata/gcc-c99/manifest.tsv
git commit -m "docs(sema): document GCC C99 compliance gate"
```

If no files changed, skip this commit.

---

## Execution Notes

- Use Chinese comments for any new non-obvious sema logic.
- Keep each task independently testable and commit after each green slice.
- Do not add IR files, lowering passes, or backend tests in this plan.
- If a GCC case fails only because of missing preprocessor support, keep it skipped and state the exact macro/header dependency in `manifest.tsv`.
- If a GCC case fails because cvm lacks warning support, accept it when GCC has no `dg-error`.
