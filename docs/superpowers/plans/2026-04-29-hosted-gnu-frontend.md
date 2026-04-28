# Hosted C99 And GNU Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import and pass GCC tests that can be satisfied by preprocessor, lexer, parser, and sema changes for hosted C99 headers, GCC builtin frontend behavior, and GNU C frontend extensions, while keeping asm/codegen/runtime cases skipped.

**Architecture:** Extend the existing preprocessor builtin-header table and macro evaluator for hosted headers, add a small sema builtin registry for GCC builtins that need only type checking or constant folding, and add narrowly-scoped GNU grammar productions for source forms present in imported GCC fixtures. Assembly, linker, runtime execution, and scan-assembler tests remain outside this plan.

**Tech Stack:** Go, generated GLR parser table, cvm preprocessor/lexer/parser/sema, GCC `gcc.dg` fixture manifests.

---

### Task 1: Import Frontend-Reachable GCC Cases

**Files:**
- Modify: `sema/testdata/gcc-c99/manifest.tsv`
- Modify: `sema/testdata/gcc-c99-extra/manifest.tsv`
- Modify: `sema/testdata/gcc-c90-as-c99/manifest.tsv`
- Create: selected `.c` fixtures under `sema/testdata/**/accept` or `reject`

- [ ] **Step 1: Import a small red batch**

Copy hosted header cases such as `c99-float-1.c`, GNU syntax cases such as `gnu-cond-expr-1.c`, `stmt-expr-1.c`, `binary-constants-1.c`, and builtin frontend cases such as `builtins-12.c`.

- [ ] **Step 2: Run red tests**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema -run 'TestGCCC99(AcceptSuite|ExtraExplicitStdSuite)' -count=1 -timeout=180s
```

Expected: failures mention missing headers/macros, unexpected GNU syntax, or unknown builtin calls.

### Task 2: Hosted Header Model

**Files:**
- Modify: `preprocessor/headers.go`
- Test: `preprocessor/headers_test.go`

- [ ] **Step 1: Add failing header macro tests**

Add tests that include `<float.h>`, `<stdint.h>`, `<stdio.h>`, `<signal.h>`, and `<tgmath.h>` and assert the macros/types used by imported GCC fixtures are present.

- [ ] **Step 2: Implement minimal hosted headers**

Extend `builtinHeader` with C99-required float macros, stdint least/fast typedefs and limit macros, `stdio.h` placeholder declarations, `signal.h` `sig_atomic_t`, and `tgmath.h` macro stubs sufficient for frontend typing.

- [ ] **Step 3: Verify**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor ./sema -run 'TestBuiltinHeaders|TestGCCC99AcceptSuite/c99-' -count=1
```

### Task 3: GNU Syntax Frontend

**Files:**
- Modify: `entity/token.go`
- Modify: `lexer/identifier.go`
- Modify: `preprocessor/scanner.go`
- Modify: `parser/phrase_structure_grammar.txt`
- Regenerate: `parser/glr_table.go`
- Modify: `sema/expr.go`
- Modify: `sema/stmt.go`
- Modify: `sema/init.go`
- Modify: `sema/spec.go`

- [ ] **Step 1: Add red syntax fixtures**

Use imported GCC cases for binary constants, omitted conditional operands, statement expressions, empty initializer braces, obsolete designators, and parse-only attributes.

- [ ] **Step 2: Implement syntax**

Add binary integer literal scanning, `a ?: b`, `({ ... })`, empty initializer `{}`, GNU old designator forms, and `__attribute__((...))` parse-and-ignore attachment points. Do not implement `asm` or inline assembly.

- [ ] **Step 3: Verify**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./parser ./sema -run 'TestGCCC99ExtraExplicitStdSuite' -count=1 -timeout=180s
```

### Task 4: GCC Builtin Frontend Registry

**Files:**
- Create: `sema/builtin.go`
- Modify: `sema/expr.go`
- Modify: `sema/sema.go`
- Test: `sema/c99_compliance_test.go`

- [ ] **Step 1: Add red builtin tests**

Add source snippets for `__builtin_malloc`, `__builtin_calloc`, `__builtin_strdup`, `__builtin_complex`, object-size style builtins used by imported fixtures, and namespace diagnostics such as C90 `nan` identifiers.

- [ ] **Step 2: Implement builtin declarations and folds**

Create builtin symbols lazily when unresolved names beginning with `__builtin_` are called. Type-check their argument count and return type. Fold only front-end constants needed by fixtures; leave execution semantics to IR.

- [ ] **Step 3: Verify**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema -run 'TestGCCC99ExtraExplicitStdSuite|TestGCCC90AsC99' -count=1 -timeout=180s
```

### Task 5: Manifest Cleanup And Full Verification

**Files:**
- Modify: `sema/testdata/**/manifest.tsv`
- Modify: `sema/gcc_c99_extra_test.go` if a new skip category is needed for asm/codegen

- [ ] **Step 1: Reclassify asm/codegen cases**

Keep `asm`, `__asm__`, hard register, symbol alias, `scan-assembler`, link-only, and `dg-do run` cases skipped with explicit non-frontend reasons.

- [ ] **Step 2: Run complete verification**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1 -timeout=240s
rg -n 'requires real system header|requires GCC builtin support|requires GNU C extension support' sema/testdata
```

Expected: full tests pass; remaining matches are only cases intentionally outside the frontend scope or not yet selected for this plan.
