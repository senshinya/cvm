# Runtime Phase 1 Followup Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue bytecode/runtime phase 1 through ten more small libc/runtime increments after the completed Plan 1-12 roadmap.

**Architecture:** Each plan starts with a fresh state adjustment pass before tests are written: inspect current registry/header/sema/runtime support, keep the plan if still missing, shrink it if some pieces already landed, or skip it if complete. Each increment follows red-green-full-verification, lands as one implementation commit, then updates handoff/roadmap docs in a separate docs commit and pushes.

**Tech Stack:** Go runtime/interpreter, builtin C header strings in `preprocessor`, sema builtin signatures, direct extern tests, GCC runtime execution tests, existing bytecode/codegen verification.

---

## Common Pre-Plan Adjustment

Before executing each numbered plan:

```bash
git status --short --branch
rg -n "<planned-symbols>" preprocessor runtime sema codegen docs/bytecode-runtime-handoff.md docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md
```

Then adjust:

- If all symbols already have header declarations, sema support when needed, registry entries, direct tests, GCC runtime tests, and handoff docs, mark the plan skipped and move to the next.
- If some symbols exist, remove those from the current increment and execute only the missing coherent subset.
- If focused red tests reveal a sema/codegen/runtime limitation outside the plan, stop that plan and choose the next safe plan.

## Common Verification

Every implementation commit and every handoff/docs commit must be preceded by:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Plan 13: `stdlib.h` Process Termination Surface - Completed

Add builtin `<stdlib.h>` declarations and runtime-through-header coverage for `exit` and `abort`. `exit` already exists in the registry; `abort` already exists as a trap extern, so this plan should shrink to declarations and coverage unless the pre-plan adjustment finds missing runtime pieces.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibProcessTerminationExterns|TestStdlibExitExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib process termination surface`
  - `docs: record stdlib process termination surface`

## Plan 14: Extra `ctype.h` Classification Helpers - Completed

Add ASCII classification externs and declarations for `isblank`, `iscntrl`, `isgraph`, and `ispunct`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestCtypeExtraClassificationExterns|TestCtypeExtraClassificationExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinCtypeHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add extra ctype classification externs`
  - `docs: record extra ctype classification externs`

## Plan 15: `string.h` Bounded Compare/Search Helpers - Completed

Add `strncmp` and `memchr` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringBoundedCompareSearchExterns|TestStringBoundedCompareSearchExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string bounded compare search externs`
  - `docs: record string bounded compare search externs`

## Plan 16: `string.h` Reverse/Set Search Helpers - Completed

Add `strrchr` and `strpbrk` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringReverseAndSetSearchExterns|TestStringReverseAndSetSearchExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string reverse set search externs`
  - `docs: record string reverse set search externs`

## Plan 17: `string.h` Span Helpers - Completed

Add `strspn` and `strcspn` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringSpanExterns|TestStringSpanExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string span externs`
  - `docs: record string span externs`

## Plan 18: `stdlib.h` `strtol` And `strtoul` - Completed

Add `strtol` and `strtoul` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage. Parsing should skip ASCII whitespace, accept an optional sign, support bases `0`, `8`, `10`, and `16`, stop at the first invalid digit, and write `endptr` when non-null.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibStrtolExterns|TestStdlibStrtolExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib strtol externs`
  - `docs: record stdlib strtol externs`

## Plan 19: `stdlib.h` `strtoll` And `strtoull` - Completed

Add `strtoll` and `strtoull` by reusing the Plan 18 parser with `long long`/`unsigned long long` return types.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibStrtollExterns|TestStdlibStrtollExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib strtoll externs`
  - `docs: record stdlib strtoll externs`

## Plan 20: `stdlib.h` Floating String Parsers - Completed

Add `atof` and `strtod` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage. Parsing should skip ASCII whitespace, use Go decimal/hex float parsing for the consumed prefix, return `0.0` when no conversion is possible, and write `endptr` when non-null.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibFloatParserExterns|TestStdlibFloatParserExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib floating parser externs`
  - `docs: record stdlib floating parser externs`

## Plan 21: `strings.h` BSD Memory Aliases - Completed

Add builtin `<strings.h>` declarations, sema signatures, runtime aliases, direct coverage, and GCC runtime execution coverage for `bcmp`, `bcopy`, and `bzero`. Reuse existing `memcmp`, `memmove`, and `bzero` behavior.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringsBSDMemoryExterns|TestStringsBSDMemoryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringsHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add strings bsd memory aliases`
  - `docs: record strings bsd memory aliases`

## Plan 22: `stdio.h` `perror`

Add a simple `perror` declaration and runtime extern. The runtime extern should validate the optional string pointer, write `<prefix>: error\n` to stderr when the prefix is non-empty, write `error\n` otherwise, and return void.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdioPerrorExtern|TestStdioPerrorExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio perror extern`
  - `docs: record stdio perror extern`

## Continuous Execution Rule

After each plan is committed and pushed, immediately start the Common Pre-Plan Adjustment for the next plan. Continue until a stop condition is reached or all ten followup plans are complete.
