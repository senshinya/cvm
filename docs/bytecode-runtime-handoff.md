# Bytecode Runtime Phase 1 Handoff

Date: 2026-05-17

This document records the current state of the bytecode/runtime work so the branch can be resumed from another device.

## Repository State

- Workspace: `/Users/shinya/Downloads/cvm/.worktrees/bytecode-runtime-phase-1`
- Branch: `codex/bytecode-runtime-phase-1`
- Latest implementation commit before this handoff document: `3cb8bc9 support complex scalar initialization`
- Remote: `origin git@github.com:senshinya/cvm.git`
- Upstream: none configured for this branch
- Working tree at handoff time: clean
- Base remote branch used for comparison: `origin/main`

To make this work available on another device, push this branch first:

```bash
git push -u origin codex/bytecode-runtime-phase-1
```

Then on the other device:

```bash
git fetch origin
git switch -c codex/bytecode-runtime-phase-1 origin/codex/bytecode-runtime-phase-1
```

## Verification Commands

Use a writable Go build cache when running under Codex sandboxed environments:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./sema -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Most recent verification run passed:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Completed Work

### Runtime Phase 1

The branch includes a first runtime implementation able to load bytecode modules and execute a growing subset of the VM:

- constants, locals, returns, integer control flow
- static memory operations
- local object addresses
- function calls and variadic local calls
- built-in extern registry and basic host externs
- run command and GCC execution fixture runner
- scoped object cleanup and dynamic memory safety checks

Important commits in this part include:

- `5ce0ce0 feat(runtime): add cvm run command`
- `b04f604 feat(runtime): execute function calls`
- `8847270 feat(runtime): execute local object addresses`
- `3128954 feat(runtime): execute static memory opcodes`
- `33637f7 feat(runtime): execute constants locals and returns`
- `f6351f6 feat(runtime): add built-in extern registry`

### GCC Runtime Execution Coverage

GCC execution fixture support has been added and gradually expanded.

Important files:

- `runtime/`
- runtime GCC execution test files and manifests
- `docs/runtime` planning/design documents from earlier commits

Recent related commits:

- `235a461 test(runtime): require GCC run fixtures`
- `d16c502 advance GCC runtime execution coverage`
- `2937e3f support GCC math runtime fixtures`
- `67c4f41 include c99 runtime fixture in GCC execution suite`
- `de11b2a support real tgmath runtime externs`

### Bytecode GCC Compile Coverage

`codegen/testdata/gcc-bytecode-compile.tsv` currently has 174 lines including the header, so 173 fixture entries.

The bytecode compile suite now covers a large subset of imported GCC accept fixtures from:

- `sema/testdata/gcc-c99/accept`
- `sema/testdata/gcc-c99-extra/accept`
- `sema/testdata/gcc-c90-as-c99/accept`

Notable recent coverage additions:

- C99 array lvalue decay: `c99-array-lval-1/2/3/4/6/7.c`
- C99 bool: `c99-bool-1/3/4.c`
- C99 compound literal and initializers: `c99-complit-1.c`, `c99-init-1.c`, `c99-init-4.c`
- C99 const expr and scope/VLA: `c99-const-expr-1/2/3/4/13/15.c`, `c99-scope-1/2.c`, `c99-vla-1/2.c`
- GNU statement expressions: `stmt-expr-1/2/4.c`
- GNU condition expression and initializers: `gnu-cond-expr-1/2.c`, `init-empty-1/2.c`, `array-const-1/2.c`
- Complex conversion and arithmetic coverage: `Wconversion-complex-c99.c`, `Wconversion-complex-gnu.c`, `complex-4.c`

### Codegen/Sema Fixes Landed

Recent commits at the tip of this branch:

- `3cb8bc9 support complex scalar initialization`
  - Supports real scalar to complex object initialization by zeroing the object and writing the real component.
  - Supports complex-to-complex copy with per-component cast.
  - Allows `ObjectAddr` local slots, such as complex parameters, to be used as addresses.
  - Adds complex/GNU fixture coverage.

- `d2d7291 support GNU statement expression bytecode`
  - Adds `StmtExpr` lowering in codegen.
  - Preserves the last expression value of GNU statement expressions.
  - Fixes sema so statement-expression locals are registered in the current function layout context.
  - Makes address-taken analysis recurse into statement expressions.

- `5e9d890 expand c99 bytecode fixture coverage`
  - Adds 12 C99 fixture entries that already compiled with current implementation.

- `c2eee6d support rvalue struct member array decay`
  - Fixes C99 non-lvalue struct/union member array decay for bytecode generation.

- `e93e9b9 support complex tgmath bytecode lowering`
  - Adds generic tgmath pseudo-call lowering for selected functions.
  - Handles complex `tgmath` extern dispatch such as `__cvm_tgmath_cexp` and `__cvm_tgmath_cpowf`.

- `d11b734 fix(sema): isolate C99 statement tag scopes`
  - Fixes C99 tag scope isolation needed by `c99-scope-1.c`.

## Current Design Notes

### Bytecode File Format

Earlier in this work, the direction was set to generate a pure binary single-file bytecode artifact that contains all data required by the runtime:

- instruction stream
- signatures
- globals and extern references
- static data
- object layouts
- string/static data
- entry metadata

The design intent is that the runtime can start directly from this single binary file later.

### Complex Values

Complex values are currently lowered as object-address values (`bytecode.TypeObjectAddr`) instead of scalar VM stack pairs.

The current codegen support includes:

- complex locals and parameters represented by object storage or `ObjectAddr` slots
- real scalar to complex initialization
- complex to complex assignment/copy with real/imag component casts
- complex tgmath extern dispatch

Runtime execution of complex arithmetic is still incomplete. Current work primarily ensures the compiler can produce validated bytecode for more GCC fixtures.

### Tgmath

`<tgmath.h>` is represented through pseudo functions such as:

- `__cvm_tgmath_sin`
- `__cvm_tgmath_exp`
- `__cvm_tgmath_pow`

Sema preserves argument types for these pseudo calls, and codegen dispatches to concrete synthetic externs:

- real: `__cvm_tgmath_sinf`, `__cvm_tgmath_exp`, `__cvm_tgmath_powl`, etc.
- complex: `__cvm_tgmath_cexp`, `__cvm_tgmath_cpowf`, etc.

Runtime support exists for real math externs; complex extern runtime behavior remains a later phase.

## Known Limits

- The interpreter/runtime is intentionally not complete yet.
- Bytecode design is intended to be complete enough for the compiler artifact, but execution support is still catching up.
- Complex runtime execution is not complete.
- Long double runtime memory/operations are still limited in places.
- Some GCC accept fixtures are still outside the bytecode compile suite.
- The branch has no upstream; push it before moving to another device.

## Suggested Next Work

Continue expanding GCC fixture coverage in small batches. Good next candidates from the current uncovered set:

- `sema/testdata/gcc-c99-extra/accept/Wconversion-integer.c`
- `sema/testdata/gcc-c99-extra/accept/Wconversion-integer-no-sign.c`
- `sema/testdata/gcc-c99-extra/accept/Wsign-conversion.c`
- `sema/testdata/gcc-c99-extra/accept/vla-2.c`
- `sema/testdata/gcc-c99-extra/accept/vla-26.c`
- `sema/testdata/gcc-c99-extra/accept/gnu99-named-loops-1.c`
- `sema/testdata/gcc-c99-extra/accept/const-elim-1.c`

Recommended workflow:

1. Add one or a small related group of GCC accept fixtures to `codegen/testdata/gcc-bytecode-compile.tsv`.
2. Run a focused test:

   ```bash
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -run 'TestGCCBytecodeCompileSuite/<fixture-name>' -count=1
   ```

3. If it fails, find the exact sema/codegen root cause before changing code.
4. After green, run:

   ```bash
   git diff --check
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
   ```

5. Commit each coherent increment separately.

## Useful Commands

Show current branch state:

```bash
git status --short --branch
git log --oneline --decorate -12
```

List still-uncovered GCC accept fixtures:

```bash
comm -23 \
  <(rg --files sema/testdata/gcc-c99/accept sema/testdata/gcc-c99-extra/accept sema/testdata/gcc-c90-as-c99/accept | sort) \
  <(tail -n +2 codegen/testdata/gcc-bytecode-compile.tsv | cut -f1 | sort)
```

Run the full test suite:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```
