# Frontend Skipped Reduction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the remaining front-end-solvable GCC C99/C90-as-C99 skipped cases into imported accept/reject fixtures, leaving only target/runtime, asm, and codegen-dependent skips.

**Architecture:** Keep GCC fixture import metadata in the existing `sema/testdata/*/manifest.tsv` files and make the existing GCC suite harness the source of truth. Implement small, front-end-only features in preprocessor, lexer, parser, and sema; do not add IR/codegen work. GNU features must be gated by existing `SemaOptions.GNUExtensions`; warning/pedwarn cases must be represented as sema/preprocessor diagnostics only where GCC expects an error.

**Tech Stack:** Go 1.x, existing cvm preprocessor/lexer/parser/sema packages, GCC DejaGNU fixture files under `/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg`, `go test`.

---

## Files And Responsibilities

- Modify `preprocessor/headers.go`: hosted C99 header model for `math.h`, `stdint.h`, and `tgmath.h`.
- Modify `preprocessor/expand.go`: attribute handling and extension-only warning hooks when needed.
- Modify `preprocessor/scanner.go`, `preprocessor/token.go`, `lexer/identifier.go`: UCN/Unicode identifiers and token conversion.
- Modify `sema/builtin.go`: GCC builtin registry signatures for string, checked string, complex, nonnull, and object-size related builtins.
- Modify `sema/sema.go`, `sema/spec.go`, `sema/expr.go`, `sema/init.go`, `sema/cast.go`, `sema/stmt.go`: GNU extension semantics, C99 warning/error diagnostics, and semantic constraints.
- Modify `parser/phrase_structure_grammar.txt` and regenerate `parser/glr_table.go`, `parser/parser.go` only if parser productions are required.
- Modify `sema/gcc_c99_test.go`, `sema/gcc_c99_extra_test.go`, `sema/gcc_c90_as_c99_test.go`: harness support for expected GCC warning/error mode if needed.
- Add fixture files under `sema/testdata/gcc-c99`, `sema/testdata/gcc-c99-extra`, and `sema/testdata/gcc-c90-as-c99`.
- Update `sema/testdata/*/manifest.tsv`: change imported cases from `skipped` to `imported-accept` or `imported-reject` with concrete reasons.

## Global Commands

- Baseline: `GOCACHE=/tmp/cvm-go-cache go test ./... -count=1 -timeout=300s`
- Parser regeneration: `go run ./parser/generator`
- Formatting: `gofmt -w <changed .go files>`
- Full verification: `git diff --check && GOCACHE=/tmp/cvm-go-cache go test ./... -count=1 -timeout=300s`

---

### Task 1: Import Header Model Red Tests

**Files:**
- Copy from `/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg`: `c99-math-double-1.c`, `c99-math-float-1.c`, `c99-math-long-double-1.c`, `c99-stdint-1.c` through `c99-stdint-8.c`, `c99-tgmath-1.c` through `c99-tgmath-4.c`.
- Modify: `sema/testdata/gcc-c99/manifest.tsv`

- [ ] **Step 1: Copy fixtures**

Run:

```bash
for f in c99-math-double-1.c c99-math-float-1.c c99-math-long-double-1.c c99-stdint-1.c c99-stdint-2.c c99-stdint-3.c c99-stdint-4.c c99-stdint-5.c c99-stdint-6.c c99-stdint-7.c c99-stdint-8.c c99-tgmath-1.c c99-tgmath-2.c c99-tgmath-3.c c99-tgmath-4.c; do
  cp "/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/$f" "sema/testdata/gcc-c99/accept/$f"
done
```

- [ ] **Step 2: Mark fixtures imported**

Change each corresponding `manifest.tsv` line from:

```text
skipped	accept	requires real system header
```

to:

```text
imported-accept	accept	hosted C99 header model
```

- [ ] **Step 3: Run red test**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema -run 'TestGCCC99AcceptSuite/(c99-math-double-1.c|c99-math-float-1.c|c99-math-long-double-1.c|c99-stdint-[1-8].c|c99-tgmath-[1-4].c)' -count=1 -timeout=180s
```

Expected: FAIL with missing header macro/type/function names.

### Task 2: Implement Hosted Header Model

**Files:**
- Modify: `preprocessor/headers.go`
- Test: imported header fixtures from Task 1

- [ ] **Step 1: Expand `stdint.h`**

Add missing C99 macros and typedefs used by the GCC fixtures:

```go
// in builtinHeader("stdint.h")
#define INT8_C(c) c
#define INT16_C(c) c
#define INT32_C(c) c
#define INT64_C(c) c ## L
#define UINT8_C(c) c ## U
#define UINT16_C(c) c ## U
#define UINT32_C(c) c ## U
#define UINT64_C(c) c ## UL
#define INTMAX_C(c) c ## L
#define UINTMAX_C(c) c ## UL
#define INT_LEAST8_MAX 127
#define INT_FAST8_MAX 127
```

Keep macro spellings simple and target-consistent with `DefaultTarget()`.

- [ ] **Step 2: Expand `math.h`**

Provide declarations and macros sufficient for preprocess/sema:

```c
#define HUGE_VAL (__builtin_huge_val())
#define HUGE_VALF (__builtin_huge_valf())
#define HUGE_VALL (__builtin_huge_vall())
#define NAN (__builtin_nan(""))
double sin(double);
float sinf(float);
long double sinl(long double);
double cos(double);
float cosf(float);
long double cosl(long double);
double fabs(double);
float fabsf(float);
long double fabsl(long double);
```

- [ ] **Step 3: Add `tgmath.h` model**

Use simple generic-like macro aliases acceptable to the imported GCC tests:

```c
#define sin(x) __cvm_tgmath_sin(x)
#define cos(x) __cvm_tgmath_cos(x)
#define fabs(x) __cvm_tgmath_fabs(x)
double __cvm_tgmath_sin(double);
double __cvm_tgmath_cos(double);
double __cvm_tgmath_fabs(double);
```

- [ ] **Step 4: Verify green**

Run the Task 1 targeted command. Expected: PASS.

- [ ] **Step 5: Run full tests**

Run: `GOCACHE=/tmp/cvm-go-cache go test ./... -count=1 -timeout=300s`

Expected: PASS.

### Task 3: Import Builtin Red Tests

**Files:**
- Copy listed builtin fixtures to `sema/testdata/gcc-c90-as-c99` or `sema/testdata/gcc-c99-extra`.
- Modify manifests.

- [ ] **Step 1: Copy accept/reject fixtures**

Run:

```bash
cp /tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/builtin-complex-err-2.c sema/testdata/gcc-c90-as-c99/accept/
cp /tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/builtins-16.c sema/testdata/gcc-c90-as-c99/accept/
cp /tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/builtins-35.c sema/testdata/gcc-c90-as-c99/reject/
for f in builtin-stringop-chk-1.c builtin-strncat-chk-1.c builtins-58.c builtins-nonnull.c pr102897.c pr113013.c pr19984.c pr68134.c; do
  cp "/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/$f" "sema/testdata/gcc-c99-extra/accept/$f"
done
```

- [ ] **Step 2: Update manifests**

Use reason:

```text
GCC builtin frontend model
```

Set `builtins-35.c` to `imported-reject`; set all others to `imported-accept`.

- [ ] **Step 3: Run red tests**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema -run 'TestGCCC90AsC99ExplicitStdSuite/(accept/(builtin-complex-err-2.c|builtins-16.c)|reject/builtins-35.c)|TestGCCC99ExtraExplicitStdSuite/accept/(builtin-stringop-chk-1.c|builtin-strncat-chk-1.c|builtins-58.c|builtins-nonnull.c|pr102897.c|pr113013.c|pr19984.c|pr68134.c)' -count=1 -timeout=180s
```

Expected: FAIL with missing builtins or namespace diagnostics.

### Task 4: Implement Builtin Frontend Registry

**Files:**
- Modify: `sema/builtin.go`, `sema/expr.go`, possibly `sema/const.go`

- [ ] **Step 1: Add builtin type helpers**

Add helpers in `sema/builtin.go`:

```go
func (s *Sema) builtinSizeT() Type { return s.Types.Builtin(ULong) }
func (s *Sema) builtinConstVoidPtr() Type {
	return s.Types.Pointer(s.Types.Qualified(s.Types.Builtin(Void), true, false, false))
}
func (s *Sema) builtinCharPtr() Type { return s.Types.Pointer(s.Types.Builtin(Char)) }
func (s *Sema) builtinConstCharPtr() Type {
	return s.Types.Pointer(s.Types.Qualified(s.Types.Builtin(Char), true, false, false))
}
```

- [ ] **Step 2: Add signatures**

Add minimal signatures for imported fixtures:

```go
"__builtin_memcpy": voidPtr, []Type{voidPtr, constVoidPtr, sizeT}
"__builtin_memmove": voidPtr, []Type{voidPtr, constVoidPtr, sizeT}
"__builtin_memset": voidPtr, []Type{voidPtr, intT, sizeT}
"__builtin_strcpy": charPtr, []Type{charPtr, constCharPtr}
"__builtin_strncpy": charPtr, []Type{charPtr, constCharPtr, sizeT}
"__builtin_strcat": charPtr, []Type{charPtr, constCharPtr}
"__builtin_strncat": charPtr, []Type{charPtr, constCharPtr, sizeT}
"__builtin_strlen": sizeT, []Type{constCharPtr}
"__builtin_object_size": sizeT, []Type{constVoidPtr, intT}
"__builtin___memcpy_chk": voidPtr, []Type{voidPtr, constVoidPtr, sizeT, sizeT}
"__builtin___strncat_chk": charPtr, []Type{charPtr, constCharPtr, sizeT, sizeT}
"__builtin_huge_val": double, nil
"__builtin_huge_valf": float, nil
"__builtin_huge_vall": longDouble, nil
"__builtin_nan": double, []Type{constCharPtr}
```

- [ ] **Step 3: Add builtin namespace diagnostics**

Reject user declarations that conflict with reserved `__builtin_` names when GCC expects an error. Implement in declaration merging path by reporting `InvalidTypeSpec(pos, "conflicting declaration of builtin function")` when an explicit declaration for an existing builtin has an incompatible type.

- [ ] **Step 4: Verify green**

Run the Task 3 targeted command. Expected: PASS.

### Task 5: Import Unicode/UCN Red Tests

**Files:**
- Copy `ucnid-10-utf8.c`, `ucnid-10.c`, `ucnid-16-utf8.c`
- Modify `sema/testdata/gcc-c99-extra/manifest.tsv`

- [ ] **Step 1: Copy fixtures**

```bash
for f in ucnid-10-utf8.c ucnid-10.c ucnid-16-utf8.c; do
  cp "/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/$f" "sema/testdata/gcc-c99-extra/accept/$f"
done
```

- [ ] **Step 2: Mark imported**

Use reason:

```text
Unicode and UCN lexer support
```

- [ ] **Step 3: Run red tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema -run 'TestGCCC99ExtraExplicitStdSuite/accept/(ucnid-10-utf8.c|ucnid-10.c|ucnid-16-utf8.c)' -count=1 -timeout=120s
```

Expected: FAIL with unsupported universal-character-name or identifier tokenization.

### Task 6: Implement Unicode/UCN Identifier Frontend

**Files:**
- Modify: `preprocessor/scanner.go`, `preprocessor/token.go`, `lexer/identifier.go`

- [ ] **Step 1: Decode UCN inside identifiers**

Change scanner behavior so `\uXXXX` and `\UXXXXXXXX` are accepted in identifier positions and emitted as an identifier token. Reject malformed UCNs with the existing `ppError`.

- [ ] **Step 2: Accept UTF-8 identifier bytes conservatively**

Permit non-ASCII UTF-8 bytes in identifiers by treating bytes with high bit set as identifier continuation/start. Keep ASCII rules unchanged.

- [ ] **Step 3: Verify green**

Run the Task 5 targeted command. Expected: PASS.

### Task 7: Import Non-Asm GNU Extension Red Tests

**Files:**
- Copy all GNU-extension skipped cases except `no-asm-3.c` and `no-asm-4.c`.
- Modify `sema/testdata/gcc-c90-as-c99/manifest.tsv` and `sema/testdata/gcc-c99-extra/manifest.tsv`.

- [ ] **Step 1: Copy fixtures**

```bash
for f in anon-struct-1.c c90-const-expr-3.c c90-const-expr-6.c c90-intprom-1.c; do
  dest=accept
  case "$f" in anon-struct-1.c|c90-const-expr-6.c) dest=reject;; esac
  cp "/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/$f" "sema/testdata/gcc-c90-as-c99/$dest/$f"
done
for f in array-const-1.c array-const-2.c complex-4.c const-elim-1.c fltconst-2.c fltconst-double-pedantic-1.c gnu99-if-decls-1.c gnu99-if-decls-2.c gnu99-named-loops-1.c init-desig-obs-2.c pr111914.c pr18502-1.c pr59011.c pr59351.c pr70418.c transparent-union-1.c visibility-8.c vla-5.c vla-9.c; do
  cp "/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg/$f" "sema/testdata/gcc-c99-extra/accept/$f"
done
```

- [ ] **Step 2: Mark imported**

Use reason:

```text
GNU frontend extension supported by cvm
```

- [ ] **Step 3: Run red tests**

Run targeted suite for the copied filenames. Expected: FAIL with unsupported grammar/semantics.

### Task 8: Implement Non-Asm GNU Extensions

**Files:**
- Modify parser/sema files as needed.

- [ ] **Step 1: Anonymous struct/union fields**

Allow unnamed struct/union members and expose nested fields for member lookup. Reject anonymous fields where GCC reject fixtures expect errors.

- [ ] **Step 2: GNU constant expression extensions**

Under `GNUExtensions`, permit GCC accepted extensions in constant expressions from imported fixtures while preserving C99 reject behavior.

- [ ] **Step 3: GNU if declaration and named loops**

Add parser/sema support for the imported `gnu99-if-decls-*` and `gnu99-named-loops-1.c` cases. Keep labels and goto/VLA barrier checks intact.

- [ ] **Step 4: Transparent union and visibility attributes**

Stop rejecting `transparent_union` and `visibility` attributes globally. Model transparent union compatibility only enough for imported accept/reject cases.

- [ ] **Step 5: Verify green**

Run targeted GNU-extension command. Expected: PASS.

### Task 9: Import Warning Diagnostic Red Tests

**Files:**
- Copy the 35 warning/diagnostic cases into accept/reject directories per manifest.
- Modify manifests.

- [ ] **Step 1: Copy fixtures**

Use manifest expected column for destination:

```bash
awk -F '\t' '$2=="skipped" && $4=="requires GCC warning/diagnostic behavior" {print $1 "\t" $3}' sema/testdata/*/manifest.tsv
```

Copy each listed GCC file to the matching testdata directory.

- [ ] **Step 2: Mark imported**

Use reason:

```text
GCC warning and pedwarn frontend diagnostics
```

- [ ] **Step 3: Run red tests**

Run the affected GCC suite tests. Expected: FAIL where diagnostics are missing or warning-only cases are escalated incorrectly.

### Task 10: Implement Warning/Pedwarn Diagnostics

**Files:**
- Modify: `sema/gcc_c99_test.go`, `sema/error.go`, preprocessor/sema diagnostic sites as needed.

- [ ] **Step 1: Extend harness options**

Parse `dg-options` for warning flags used by imported fixtures, at minimum:

```text
-pedantic
-pedantic-errors
-Werror
-Wdeclaration-after-statement
-Wimplicit-function-declaration
-Woverlength-strings
-Wvla
-Wunsuffixed-float-constants
```

- [ ] **Step 2: Add diagnostic mode to `SemaOptions`**

Represent enabled warnings as booleans or a small map. Keep current accept/reject tests unchanged unless imported GCC directives require stricter behavior.

- [ ] **Step 3: Implement targeted diagnostics**

Add only diagnostics used by imported fixtures:

```text
declaration after statement
implicit function declaration
overlength string
binary/octal constants pedwarn
missing field initializer
qualified function return type
VLA warning as error
unsuffixed float constants
integer overflow warning-as-error cases
```

- [ ] **Step 4: Verify green**

Run affected warning diagnostic tests. Expected: PASS.

### Task 11: Final Verification And Manifest Audit

**Files:**
- Modify: `sema/testdata/*/manifest.tsv` only if audit shows stale reasons.

- [ ] **Step 1: Audit remaining skips**

Run:

```bash
awk -F '\t' '$2=="skipped" {count[$4]++} END {for (r in count) print count[r] "\t" r}' sema/testdata/*/manifest.tsv | sort -nr
```

Expected: no remaining `real system header`, `GCC builtin support`, `GCC builtin namespace diagnostics`, `Unicode/UCN lexer support`, non-asm `GNU C extension support`, or `GCC warning/diagnostic behavior` entries.

- [ ] **Step 2: Run full verification**

Run:

```bash
git diff --check
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1 -timeout=300s
```

Expected: both commands exit 0.

- [ ] **Step 3: Commit**

Run:

```bash
git add -A
git commit -m "Reduce frontend-solvable GCC skipped cases"
```

Expected: commit succeeds on `codex/frontend-skipped-reduction`.

---

## Self-Review

- Spec coverage: all requested groups are represented: system headers 15, builtin 11, Unicode/UCN 3, non-asm GNU extensions 23, warning diagnostics 35.
- Exclusions: `no-asm-3.c` and `no-asm-4.c` remain skipped because the user explicitly asked to skip asm. Target/runtime and inline/codegen cases remain out of scope.
- Risk: warning diagnostic support is broad and may need staged implementation; if it blocks full completion, keep earlier batches committed and report the exact remaining red cases.
