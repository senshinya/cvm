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

## Plan 22: `stdio.h` `perror` - Completed

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

## Plan 23: `stdlib.h` `strtof` And `strtold` - Completed

Add `strtof` and `strtold` declarations, sema signatures, runtime externs, direct coverage, and GCC runtime execution coverage. Reuse the existing `strtod` consumed-prefix parser; return `float` for `strtof` and long-double storage semantics for `strtold`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibMoreFloatParserExterns|TestStdlibMoreFloatParserExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib strtof externs`
  - `docs: record stdlib strtof externs`

## Plan 24: `string.h` `strnlen` - Completed

Add `strnlen` declaration, sema signature, runtime extern, direct coverage, and GCC runtime execution coverage. The extern should scan no more than the supplied maximum byte count and should not require the string to contain a null terminator within that bound.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringNLengthExtern|TestStringNLengthExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string strnlen extern`
  - `docs: record string strnlen extern`

## Plan 25: `stdlib.h` `realloc` - Completed

Add `realloc` declaration, runtime extern, direct coverage, and GCC runtime execution coverage. The extern should support `realloc(NULL, size)` as allocation, `realloc(ptr, 0)` as free-and-null, and copying the common prefix from the old allocation into a new runtime block for normal resizing.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibReallocExtern|TestStdlibReallocExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib realloc extern`
  - `docs: record stdlib realloc extern`

## Plan 26: `string.h` `strerror` - Completed

Add `strerror` declaration, sema signature, runtime extern, direct coverage, and GCC runtime execution coverage. Phase 1 behavior should return a stable runtime C string for any input error number, using `"error"` as the deterministic message.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringStrerrorExtern|TestStringStrerrorExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string strerror extern`
  - `docs: record string strerror extern`

## Plan 27: `stdlib.h` Deterministic Random Helpers - Completed

Add `rand` and `srand` declarations, runtime externs, direct coverage, and GCC runtime execution coverage. Keep phase 1 deterministic by storing a simple registry-local unsigned seed and producing repeatable non-negative `int` values.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibRandExterns|TestStdlibRandExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib random externs`
  - `docs: record stdlib random externs`

## Plan 28: `stdlib.h` Environment Stubs - Completed

Add `getenv` declaration, runtime extern, direct coverage, and GCC runtime execution coverage. Phase 1 behavior should be hermetic: validate the name pointer and return null for every variable.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibGetenvExtern|TestStdlibGetenvExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib getenv extern`
  - `docs: record stdlib getenv extern`

## Plan 29: `stdlib.h` System Command Stub - Completed

Add `system` declaration, runtime extern, direct coverage, and GCC runtime execution coverage. Phase 1 behavior should remain hermetic: `system(NULL)` returns `0` and any non-null command returns `-1` after validating that the command pointer names a readable C string.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibSystemExtern|TestStdlibSystemExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib system extern`
  - `docs: record stdlib system extern`

## Plan 30: `string.h` `strtok` - Completed

Add `strtok` declaration, sema signature, runtime extern, direct coverage, and GCC runtime execution coverage. Store tokenizer continuation state on the registry and mutate delimiters in-place to match C token splitting behavior for single-threaded phase 1 execution.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringStrtokExtern|TestStringStrtokExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string strtok extern`
  - `docs: record string strtok extern`

## Plan 31: `stdlib.h` Absolute Division Types - Completed

Investigate and, if ABI support is already sufficient, add `div`, `ldiv`, and `lldiv` declarations plus runtime-through-header coverage. If struct-return lowering is not ready, mark this plan blocked in the roadmap and move to Plan 32 instead of forcing a broad ABI change.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibDivExterns|TestStdlibDivExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib div externs`
  - `docs: record stdlib div externs`

## Plan 32: `stdio.h` File-Operation Stubs - Completed

Add hermetic stubs for a small stdio file-operation set only if they unblock GCC accept fixtures without host filesystem effects. Candidate subset: `remove` and `rename` returning `-1` after pointer validation; defer `fopen`, `freopen`, `tmpfile`, and `tmpnam` unless the pre-plan adjustment finds a narrow fixture requiring them.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdioFileOperationStubs|TestStdioFileOperationStubsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio file operation stubs`
  - `docs: record stdio file operation stubs`

## Plan 33: `stdio.h` Open Stubs - Completed

Add hermetic null-return stubs for `fopen`, `freopen`, and `tmpfile`. `fopen` and `freopen` validate path and mode strings; `freopen` also validates the stream handle. None of these externs touches the host filesystem.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdioOpenStubs|TestStdioOpenStubsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio open stubs`
  - `docs: record stdio open stubs`

## Plan 34: `string.h` Collation Helpers - Completed

Add deterministic C-locale `strcoll` and `strxfrm` support. `strcoll` reuses bytewise string comparison; `strxfrm` copies the source string into the destination buffer with normal C truncation/null-termination behavior and returns the full source length.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `sema/builtin.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringCollateAndTransformExterns|TestStringCollateTransformExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string collation externs`
  - `docs: record string collation externs`

## Plan 35: GCC Runtime Manifest Expansion - Skipped

The pre-plan adjustment found no uncovered GCC runtime accept candidates. `runtime/testdata/gcc-exec/gap-report.md` reports 18 runnable `{ dg-do run }` or `c99_runtime` fixtures, and all 18 are already present in `runtime/testdata/gcc-exec/manifest.tsv`.

- Verification used for the adjustment:

```bash
git status --short --branch
sed -n '1,220p' runtime/testdata/gcc-exec/manifest.tsv
rg -l "dg-do run|dg-require-effective-target c99_runtime" sema/testdata/gcc-c99/accept sema/testdata/gcc-c99-extra/accept sema/testdata/gcc-c90-as-c99/accept | sort
sed -n '1,120p' runtime/testdata/gcc-exec/gap-report.md
```

- Commit messages:
  - `docs: record gcc runtime manifest adjustment`

## Plan 36: `stdlib.h` `atexit` Stub - Completed

Add a phase 1 `atexit` declaration and runtime stub. The stub accepts a callback pointer and returns `0` to model successful registration, but it does not execute callbacks during process shutdown.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestStdlibAtexitExtern|TestStdlibAtexitExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib atexit extern`
  - `docs: record stdlib atexit extern`

## Plan 37: `stdio.h` `tmpnam` Stub - Completed

Add `L_tmpnam`, `TMP_MAX`, `tmpnam` declaration, runtime registration, direct extern coverage, and GCC runtime execution coverage. Phase 1 behavior remains hermetic: `tmpnam(NULL)` returns null, and a non-null buffer is validated as writable and nul-terminated before returning null.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run 'TestBuiltinStdioHeaderDeclaresFormattingSurface|TestBuiltinStdioHeaderDefinesBufferingMacros' -count=1 -v
go test ./runtime -run 'TestStdioTmpnamStub|TestStdioTmpnamExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio tmpnam stub`
  - `docs: record stdio tmpnam stub`

## Plan 38: `stdlib.h` `_Exit` Immediate Exit - Completed

Add the C99 `_Exit` declaration, register it in the default extern registry, and share the existing exit-code termination path. Since phase 1 does not maintain an `atexit` callback queue, `_Exit` is currently equivalent to `exit` at runtime.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestStdlibProcessTerminationExterns|TestStdlibImmediateExitExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib immediate exit extern`
  - `docs: record stdlib immediate exit extern`

## Plan 39: `locale.h` `setlocale` C-Locale Stub - Completed

Add a minimal builtin `<locale.h>` and a deterministic C-locale `setlocale` extern. The header exposes `LC_ALL`, `LC_COLLATE`, `LC_CTYPE`, `LC_MONETARY`, `LC_NUMERIC`, `LC_TIME`, and `setlocale`; runtime supports querying the current locale and selecting `"C"` or `""`, while unsupported locale names return null.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinLocaleHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestLocaleSetlocaleExtern|TestLocaleSetlocaleExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add locale setlocale extern`
  - `docs: record locale setlocale extern`

## Plan 40: `time.h` Deterministic Time Helpers - Completed

Add a minimal builtin `<time.h>` plus deterministic runtime externs for `clock`, `difftime`, and `time`. `clock()` and `time(NULL)` return `0`, `time(&slot)` stores `0`, and `difftime` returns the numeric difference between its two `time_t` arguments.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinTimeHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestTimeExterns|TestTimeHeaderExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add time header externs`
  - `docs: record time header externs`

## Plan 41: `errno.h` Extern Variable - Completed

Add a minimal builtin `<errno.h>` and resolve `errno` as a stable writable extern variable for each runtime memory instance. The header exposes `EDOM`, `ERANGE`, `EILSEQ`, and `extern int errno`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinErrnoHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestErrnoExternVariable|TestErrnoHeaderExecuteThroughRuntime' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add errno extern variable`
  - `docs: record errno extern variable`

## Plan 42: `assert.h` Header Surface - Completed

Add builtin `<assert.h>` macro coverage. In normal mode `assert(expr)` expands to an expression that calls `abort()` on failure; with `NDEBUG` it expands to `((void)0)`. Runtime coverage executes the non-failing path through the header and reuses the existing `abort` extern for the failure path.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinAssertHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run TestAssertHeaderExecuteThroughRuntime -count=1 -v
```

- Commit messages:
  - `feat(runtime): add assert header surface`
  - `docs: record assert header surface`

## Continuous Execution Rule

After each plan is committed and pushed, immediately start the Common Pre-Plan Adjustment for the next plan. Continue until a stop condition is reached or all ten followup plans are complete.
