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

## Plan 43: `stdio.h` Position Stubs - Completed

The pre-plan adjustment found `scanf` too broad for phase 1 because it needs real format parsing, input consumption, typed write-back, and conversion-count semantics. This plan instead adds narrower `stdio` position helpers: `SEEK_SET`, `SEEK_CUR`, `SEEK_END`, `fseek`, `ftell`, and `rewind`. Runtime validates host-backed stream handles; `fseek` and `ftell` conservatively report failure, while `rewind` clears EOF state.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run 'TestBuiltinStdioHeaderDeclaresFormattingSurface|TestBuiltinStdioHeaderDefinesBufferingMacros' -count=1 -v
go test ./runtime -run 'TestStdioPositionStubs|TestStdioPositionStubsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio position stubs`
  - `docs: record stdio position stubs`

## Plan 44: Plain `math.h` Unary Helpers - Completed

The pre-plan adjustment found tgmath real unary dispatch already covered, but plain `<math.h>` did not declare or register ordinary `fabs*` and `sqrt*` helpers. This plan adds direct header declarations and runtime externs for `fabs`, `fabsf`, `fabsl`, `sqrt`, `sqrtf`, and `sqrtl`, backed by the existing unary math runtime helper.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain math unary externs`
  - `docs: record plain math unary externs`

## Plan 45: `string.h` `memccpy` Helper - Completed

The pre-plan adjustment found `mempcpy`, `memcpy`, and `memmove` already covered, while `memccpy` was absent from builtin `<string.h>` and the runtime registry. This plan adds `memccpy` as a narrow memory helper: copy bytes from source to destination through the first matching byte, return the destination pointer just after the match, or return null if no match is copied.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
go test ./runtime -run 'TestMemoryCharCopyExtern|TestMemoryCharCopyExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string memccpy extern`
  - `docs: record string memccpy extern`

## Plan 46: `string.h` `strndup` Helper - Completed

The pre-plan adjustment found `strdup` already covered, while POSIX-style bounded duplication was absent. This plan adds builtin `<string.h>` declaration and runtime support for `strndup`, allocating a new writable NUL-terminated string containing at most the requested number of source bytes.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
go test ./runtime -run 'TestPlainAllocationExterns|TestStringBoundedDupExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add string strndup extern`
  - `docs: record string strndup extern`

## Plan 47: `stdio.h` File Position Stubs - Completed

The pre-plan adjustment kept avoiding broad formatted input work and selected the narrower file-position surface. This plan adds `fpos_t`, `fgetpos`, and `fsetpos`; runtime validates known stream handles and readable/writable position objects, then conservatively reports failure just like the existing `fseek`/`ftell` stubs.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
go test ./runtime -run 'TestStdioPositionStubs|TestStdioPositionStubsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio fpos stubs`
  - `docs: record stdio fpos stubs`

## Plan 48: Plain Trigonometric `math.h` Helpers - Completed

The pre-plan adjustment found plain `<math.h>` still much narrower than the existing tgmath-backed runtime helper set. This plan adds direct `sin*`, `cos*`, and `tan*` declarations and extern registrations, while guarding `<math.h>` plain prototypes when included by `<tgmath.h>` so tgmath function-like macros do not rewrite those prototypes.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestTgmathSinFloatExecutesThroughRuntime|TestTgmathCosExecuteThroughRuntime|TestTgmathTanExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./codegen -run 'TestGCCTgmathFloatSinUsesFloatExtern|TestGCCBytecodeCompileSuite/c99-tgmath' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain trig math externs`
  - `docs: record plain trig math externs`

## Plan 49: Plain Binary `math.h` Helpers - Completed

The pre-plan adjustment selected a small binary math subset already supported by the tgmath runtime helper shape. This plan adds direct `pow*`, `atan2*`, and `hypot*` declarations and extern registrations, while retaining the `<tgmath.h>` macro isolation introduced in Plan 48.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestPlainMathBinaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestTgmathExpPowExecuteThroughRuntime|TestTgmathAtan2ExecuteThroughRuntime|TestTgmathHypotExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./codegen -run 'TestGCCTgmathComplexPowFloatUsesComplexFloatExtern|TestGCCBytecodeCompileSuite/c99-tgmath' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain binary math externs`
  - `docs: record plain binary math externs`

## Plan 50: `stdlib.h` Multibyte Helpers - Completed

The pre-plan adjustment found the safe stdlib conversion/allocation/process surface already covered and avoided broader callback-based `qsort`/`bsearch` work. This plan adds a narrow C locale single-byte multibyte/wide-char surface: `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, and `wcstombs`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestStdlibMultibyteExterns|TestStdlibMultibyteExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib multibyte externs`
  - `docs: record stdlib multibyte externs`

## Plan 51: Plain Exponential/Log `math.h` Helpers - Completed

The pre-plan adjustment kept avoiding full locale work and selected another small plain `<math.h>` slice already supported by the tgmath unary helper shape. This plan adds direct `exp*`, `log*`, and `log10*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain exp log math externs`
  - `docs: record plain exp log math externs`

## Plan 52: GCC Fixture Gap Recheck - Completed

The pre-plan adjustment returned to the GCC fixture manifests before selecting more runtime/header work. Runtime execution has no current runnable accept gap, and bytecode compile coverage still covers every imported GCC accept `.c` fixture in the tracked roots.

- Files: `runtime/testdata/gcc-exec/gap-report.md`, `runtime/testdata/gcc-exec/manifest.tsv`, `codegen/testdata/gcc-bytecode-compile.tsv`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1 -v
go test ./codegen -run 'TestGCCBytecodeManifest|TestGCCBytecodeCompileManifest|TestGCCBytecodeCompileSuite' -count=1 -v
```

- Result:
  - Runtime gap report remains current with 18 runnable fixtures and no missing manifest entries.
  - Codegen bytecode compile manifest remains at 231 fixture entries plus header and has no uncovered imported accept `.c` fixtures.

- Commit message:
  - `docs: record gcc fixture gap recheck`

## Plan 53: Plain Rounding `math.h` Helpers - Completed

The pre-plan adjustment found no GCC fixture gap to add, so this plan continued the plain `<math.h>` surface using existing tgmath-compatible unary runtime helpers. It adds `ceil*`, `floor*`, `trunc*`, and `round*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain rounding math externs`
  - `docs: record plain rounding math externs`

## Plan 54: Plain Inverse Trig `math.h` Helpers - Completed

The pre-plan adjustment stayed with small plain `<math.h>` gaps that mirror existing tgmath helpers. This plan adds `asin*`, `acos*`, and `atan*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain inverse trig math externs`
  - `docs: record plain inverse trig math externs`

## Plan 55: Plain Hyperbolic `math.h` Helpers - Completed

The pre-plan adjustment continued the plain `<math.h>` sweep with another exact-value unary subset. This plan adds `sinh*`, `cosh*`, and `tanh*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain hyperbolic math externs`
  - `docs: record plain hyperbolic math externs`

## Plan 56: Plain Inverse Hyperbolic `math.h` Helpers - Completed

The pre-plan adjustment continued the plain `<math.h>` sweep with another exact-value unary subset. This plan adds `asinh*`, `acosh*`, and `atanh*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain inverse hyperbolic math externs`
  - `docs: record plain inverse hyperbolic math externs`

## Plan 57: Plain `cbrt`/`erf` `math.h` Helpers - Completed

The pre-plan adjustment found `cbrt*`, `erf*`, and `erfc*` already supported by tgmath runtime helpers but still absent from the plain `<math.h>` surface. This plan adds the plain declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain cbrt erf math externs`
  - `docs: record plain cbrt erf math externs`

## Plan 58: Plain `exp2`/`log2` `math.h` Helpers - Completed

The pre-plan adjustment found `exp2*`, `expm1*`, `log1p*`, and `log2*` already supported by tgmath runtime helpers but still absent from the plain `<math.h>` surface. This plan adds the plain declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain exp2 log2 math externs`
  - `docs: record plain exp2 log2 math externs`

## Plan 59: Plain `fdim`/Minmax `math.h` Helpers - Completed

The pre-plan adjustment moved to binary real math helpers and intentionally kept the group small. This plan adds plain `fdim*`, `fmax*`, and `fmin*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathBinaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain fdim minmax math externs`
  - `docs: record plain fdim minmax math externs`

## Plan 60: Plain Remainder/Sign Binary `math.h` Helpers - Completed

The pre-plan adjustment continued the binary real math sweep and kept `fma`/`nextafter` for later because they have distinct arity or edge-case behavior. This plan adds plain `fmod*`, `remainder*`, and `copysign*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathBinaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain fmod copysign math externs`
  - `docs: record plain fmod copysign math externs`

## Plan 61: Plain `fma` `math.h` Helpers - Completed

The pre-plan adjustment picked `fma*` as a standalone ternary helper because it has a different extern shape from the binary math families. This plan adds plain `fma*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathTernaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain fma math externs`
  - `docs: record plain fma math externs`

## Plan 62: Plain `nextafter`/`nexttoward` `math.h` Helpers - Completed

The pre-plan adjustment selected the stable endpoint-equality cases for next-direction helpers. This plan adds plain `nextafter*` and `nexttoward*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathBinaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain nextafter math externs`
  - `docs: record plain nextafter math externs`

## Plan 63: Plain Rint/Logb `math.h` Helpers - Completed

The pre-plan adjustment returned to unary real math helpers and selected stable exact cases. This plan adds plain `nearbyint*`, `rint*`, and `logb*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain rint logb math externs`
  - `docs: record plain rint logb math externs`

## Plan 64: Plain Gamma `math.h` Helpers - Completed

The pre-plan adjustment selected another unary real group with existing tgmath runtime support and stable exact checks. This plan adds plain `tgamma*` and `lgamma*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathUnaryExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain gamma math externs`
  - `docs: record plain gamma math externs`

## Plan 65: Plain `ilogb` `math.h` Helpers - Completed

The pre-plan adjustment split integer-returning math helpers into a small first slice and selected `ilogb*` before the long and long-long rounding families. This plan adds plain `ilogb*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestMathClassificationExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain ilogb math externs`
  - `docs: record plain ilogb math externs`

## Plan 66: Plain `lrint`/`lround` `math.h` Helpers - Completed

The pre-plan adjustment continued the integer-returning family and selected the `long` return helpers separately from the `long long` variants. This plan adds plain `lrint*` and `lround*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestTgmathLongExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain lrint lround math externs`
  - `docs: record plain lrint lround math externs`

## Plan 67: Plain `llrint`/`llround` `math.h` Helpers - Completed

The pre-plan adjustment completed the integer-returning rounding family with the `long long` variants. This plan adds plain `llrint*` and `llround*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestTgmathLongExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain llround math externs`
  - `docs: record plain llround math externs`

## Plan 68: Plain `scalbn`/`ldexp` `math.h` Helpers - Completed

The pre-plan adjustment selected the real-plus-integer binary helpers before the pointer-output helpers. This plan adds plain `scalbn*`, `scalbln*`, and `ldexp*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathRealIntExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain scalbn math externs`
  - `docs: record plain scalbn math externs`

## Plan 69: Plain `frexp` `math.h` Helpers - Completed

The pre-plan adjustment selected pointer-output `frexp*` separately from `remquo*` so the memory write-back path stays easy to isolate. This plan adds plain `frexp*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestTgmathFrexpExtern|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain frexp math externs`
  - `docs: record plain frexp math externs`

## Plan 70: Plain `remquo` `math.h` Helpers - Completed

The pre-plan adjustment confirmed plain `remquo*` remained the last pointer-output real math family missing from `<math.h>` declarations and default runtime registration. This plan adds plain `remquo*` declarations and extern registrations.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestTgmathRemquoExtern|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain remquo math externs`
  - `docs: record plain remquo math externs`

## Plan 71: Plain `<complex.h>` Projection Helpers - Completed

The pre-plan adjustment found the plain real `<math.h>` surface mostly filled and selected the smallest complex-surface increment: builtin `<complex.h>` declarations plus runtime externs for projection helpers only. The GCC runtime fixture intentionally stores complex literals in matching complex variables before calling `carg*`, avoiding expansion into the separate complex conversion boundary.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexProjectionExterns|TestComplexProjectionExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex projection externs`
  - `docs: record complex projection externs`

## Plan 72: Plain `<complex.h>` Absolute-Value Helpers - Completed

The pre-plan adjustment selected `cabs*` because the runtime already had `__builtin_cabs*` helper support while the plain `<complex.h>` declarations and default registry names were still missing.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexAbsExterns|TestComplexAbsExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex abs externs`
  - `docs: record complex abs externs`

## Plan 73: Plain `<complex.h>` Unary Helpers - Completed

The pre-plan adjustment confirmed the requested GCC bytecode accept fixtures were already covered, then selected the next smallest runtime/header gap: plain `<complex.h>` `conj*` and `cproj*` declarations plus default registry names.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexUnaryExterns|TestComplexUnaryExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex unary externs`
  - `docs: record complex unary externs`

## Plan 74: Plain `<complex.h>` Trigonometric Helpers - Completed

The pre-plan adjustment selected the basic trigonometric complex unary helpers and kept test inputs at zero to cover the plain extern plumbing and object-return path without introducing approximate comparison noise.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexTrigExterns|TestComplexTrigExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex trig externs`
  - `docs: record complex trig externs`

## Plan 75: Plain `<complex.h>` Hyperbolic Helpers - Completed

The pre-plan adjustment mirrored Plan 74 for the hyperbolic complex unary helpers and again used zero inputs for exact expected results.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexHyperbolicExterns|TestComplexHyperbolicExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex hyperbolic externs`
  - `docs: record complex hyperbolic externs`

## Plan 76: Plain `<complex.h>` Inverse Trigonometric Helpers - Completed

The pre-plan adjustment selected the circular inverse complex helpers only, leaving inverse hyperbolic helpers for a separate increment. Tests use exact `0+0i` and `1+0i` cases.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexInverseTrigExterns|TestComplexInverseTrigExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex inverse trig externs`
  - `docs: record complex inverse trig externs`

## Plan 77: Plain `<complex.h>` Inverse Hyperbolic Helpers - Completed

The pre-plan adjustment selected the remaining inverse hyperbolic complex unary helpers after the circular inverse helpers landed. Tests use exact `0+0i` and `1+0i` cases.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexInverseHyperbolicExterns|TestComplexInverseHyperbolicExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex inverse hyperbolic externs`
  - `docs: record complex inverse hyperbolic externs`

## Plan 78: Plain `<complex.h>` Exp/Log/Sqrt Helpers - Completed

The pre-plan adjustment picked `cexp*`, `clog*`, and `csqrt*` because tgmath already had runtime backends for them and exact inputs keep the tests deterministic: `cexp(0+0i) == 1+0i`, `clog(1+0i) == 0+0i`, and `csqrt(0+0i) == 0+0i`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexExpLogSqrtExterns|TestComplexExpLogSqrtExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex exp log sqrt externs`
  - `docs: record complex exp log sqrt externs`

## Plan 79: Plain `<complex.h>` Pow Helpers - Completed

The pre-plan adjustment selected `cpow*` as the only remaining known complex helper already backed by the existing tgmath binary-complex runtime path. Tests use `cpow(1+0i, 2+0i) == 1+0i` to avoid approximate comparisons.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexPowExterns|TestComplexPowExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add complex pow externs`
  - `docs: record complex pow externs`

## Plan 80: Plain `<complex.h>` Gap Scan - Completed

The pre-plan adjustment switched from implementation to fact-finding after `cpow*` landed. The scan compared the C99 plain `<complex.h>` function set against builtin header declarations and runtime registry bases.

- Result: no remaining plain `<complex.h>` header gap.
- Result: no remaining plain runtime registry base gap for the real/unary/binary helper-backed complex functions; `cabs*` remains intentionally registered through `complexAbsExtern`.
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinComplexHeaderDeclaresProjectionSurface -count=1 -v
go test ./runtime -run 'TestComplexProjectionExterns|TestComplexAbsExterns|TestComplexUnaryExterns|TestComplexTrigExterns|TestComplexHyperbolicExterns|TestComplexInverseTrigExterns|TestComplexInverseHyperbolicExterns|TestComplexExpLogSqrtExterns|TestComplexPowExterns|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit message:
  - `docs: record complex surface gap scan`

## Plan 81: Complex Plain Extern Followup - Completed

The pre-plan adjustment consumed Plan 80's result: no low-risk plain complex extern remained to implement. This plan completed without code changes.

## Plan 82: GCC Runtime Manifest Gap Recheck - Completed

The pre-plan adjustment returned to the GCC runtime execution manifest after the complex surface closed. The report remains current: 18 runnable fixtures, all 18 represented in `runtime/testdata/gcc-exec/manifest.tsv`, with no preprocess, parse, sema, codegen, encode-load, runtime, or exit-mismatch failures.

- Focused test:

```bash
go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1 -v
```

- Commit message:
  - `docs: record GCC runtime gap recheck`

## Plan 83: GCC Runtime Fixture Addition - Completed

The pre-plan adjustment consumed Plan 82's result: no new runnable GCC runtime fixture exists to add. This plan completed without code changes.

## Plan 84: GCC Bytecode Compile Manifest Recheck - Completed

The pre-plan adjustment rechecked bytecode compile coverage after runtime manifest closure. `codegen/testdata/gcc-bytecode-compile.tsv` still has 232 lines including the header, covering 231 fixture entries. The direct `comm` gap check returned zero missing `.c` files across `sema/testdata/gcc-c99/accept`, `sema/testdata/gcc-c99-extra/accept`, and `sema/testdata/gcc-c90-as-c99/accept`.

- Focused tests:

```bash
go test ./codegen -run 'TestGCCBytecodeManifestCoversImportedAcceptFixtures|TestGCCBytecodeCompileSuite' -count=1 -v
```

- Commit message:
  - `docs: record GCC bytecode manifest recheck`

## Plan 85: Small GCC Accept Probe - Completed

The pre-plan adjustment consumed Plan 82 and Plan 84's result: both runtime and bytecode GCC manifests were closed, so probing another small accept fixture would duplicate existing manifest coverage. This plan completed without code changes.

## Plan 86: `<math.h>` Plain Surface Scan - Completed

The pre-plan adjustment scanned the standard C99 plain `<math.h>` function set against builtin header declarations and runtime registry names. The scan found two remaining groups:

- `nanf`, `nan`, and `nanl`: low-risk header/runtime gap.
- `modff`, `modf`, and `modfl`: pointer-output helper gap, deferred to a later increment.

## Plan 87: `<math.h>` `nan*` Variants - Completed

The pre-plan adjustment selected the low-risk `nan*` group before `modf*` because it only needed return-type specialization of the existing `nan` helper.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathNanExterns|TestBuiltinFloatingConstantsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add math nan variants`
  - `docs: record math nan variants`

## Plan 88: `<math.h>` `modf*` Pointer-Output Helpers - Completed

The pre-plan adjustment stayed on the remaining math scan gap after `nan*` landed. `modf*` uses the same pointer-output shape as `frexp*`, but stores the integer part using the same floating rank and size as the return value.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathModfExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `feat(runtime): add math modf externs`
  - `docs: record math modf externs`

## Plan 89: `<math.h>` Plain Surface Recheck - Completed

The pre-plan adjustment switched from implementation to closure verification after `modf*` landed. The rescan compared the expected C99 math surface from `TestBuiltinMathHeaderDeclaresRuntimeSurface` against builtin header declarations and `TestDefaultExternRegistryHasExitAndAbort` registry coverage.

- Result: no remaining expected `<math.h>` header gaps.
- Result: no remaining expected `<math.h>` runtime registry gaps.
- Focused tests:

```bash
go test ./preprocessor -run TestBuiltinMathHeaderDeclaresRuntimeSurface -count=1 -v
go test ./runtime -run 'TestPlainMathModfExterns|TestMathPlainUnaryExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit message:
  - `docs: record math modf externs`

## Plan 90: GCC Runtime Manifest Gap Recheck - Completed

The pre-plan adjustment returned to GCC runtime execution after the math surface closed. `TestGCCExecutionGapReportIsCurrent` still reports 18 runnable GCC accept fixtures, all represented in `runtime/testdata/gcc-exec/manifest.tsv`, with no failures by stage.

- Focused test:

```bash
go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1 -v
```

## Plan 91: Compile-Only GCC Runtime Candidate Scan - Completed

The pre-plan adjustment consumed Plan 90's result: since the directive-based runtime manifest was closed, the next useful direction was scanning compile-only GCC accept fixtures that still contain a `main` and can provide direct runtime coverage. The scan selected `sema/testdata/gcc-c90-as-c99/accept/Wdeclaration-after-statement-4.c` as the first low-risk candidate because it has self-contained `abort`/`exit` assertions and no unresolved helper functions.

## Plan 92: C90-As-C99 Declaration-After-Statement Runtime Coverage - Completed

The pre-plan adjustment added the selected compile-only fixture as a direct runtime test rather than changing the manifest rule that requires `{ dg-do run }` or `c99_runtime` directives.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCC90DeclarationAfterStatementExecutesThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCC90DeclarationAfterStatementExecutesThroughRuntime|TestGCCExecutionGapReportIsCurrent' -count=1 -v
```

- Commit messages:
  - `test(runtime): execute GCC declaration-after-statement fixture`
  - `docs: record GCC declaration-after-statement runtime coverage`

## Plan 93: Header/Extern Failure Triage - Completed

The Plan 92 focused probe passed immediately, so there was no header or runtime extern failure to fix.

## Plan 94: Codegen/Interpreter Failure Triage - Completed

The Plan 92 focused probe passed immediately, so there was no codegen or interpreter opcode failure to fix.

## Plan 95: Runtime Coverage Increment Recording - Completed

The coverage-only increment added a reusable helper for running compile-only GCC accept fixtures through the runtime and recorded `Wdeclaration-after-statement-4.c` as direct C90-as-C99 runtime coverage.

## Plan 96: GCC Runtime Gap Stability Check - Completed

The pre-plan adjustment re-ran the directive-based GCC runtime gap report after the direct compile-only coverage test was added. The manifest remains stable and closed because the new test intentionally does not alter the directive-based manifest scope.

- Focused test:

```bash
go test ./runtime -run 'TestGCCC90DeclarationAfterStatementExecutesThroughRuntime|TestGCCExecutionGapReportIsCurrent' -count=1 -v
```

- Commit message:
  - `docs: record GCC declaration-after-statement runtime coverage`

## Plan 97: GCC Bytecode Compile Manifest Recheck - Completed

The pre-plan adjustment returned to compile coverage after the direct runtime fixture landed. `codegen/testdata/gcc-bytecode-compile.tsv` still has 232 lines including the header, and the direct `comm` gap check across `sema/testdata/gcc-c99/accept`, `sema/testdata/gcc-c99-extra/accept`, and `sema/testdata/gcc-c90-as-c99/accept` returned no missing `.c` fixtures.

- Focused tests:

```bash
go test ./codegen -run 'TestGCCBytecodeManifestCoversImportedAcceptFixtures|TestGCCBytecodeCompileSuite' -count=1 -v
```

## Plan 98: VLA Runtime Candidate Selection - Completed

The pre-plan adjustment revisited the originally requested VLA fixtures. `vla-2.c` and `vla-26.c` already map to existing direct runtime coverage for VLA struct/union members and VLA parameter dynamic strides. The next low-risk VLA-shaped candidate was `Wstrict-aliasing-bogus-vla-1.c`, adapted into a self-contained runtime probe so the current runtime does not need to inject `argc`/`argv` entry arguments.

## Plan 99: VLA Dynamic-Size Memcpy Runtime Coverage - Completed

The pre-plan adjustment added a focused runtime probe for local VLAs copied through `__builtin_memcpy` using a dynamic `n * sizeof(*x)` byte count.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCVLAMemcpyDynamicSizeExecutesThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCVLAMemcpyDynamicSizeExecutesThroughRuntime|TestGCCVLAParameterDynamicStrideExecutesThroughRuntime|TestGCCVLAStructAndUnionMembersExecuteThroughRuntime' -count=1 -v
```

- Commit messages:
  - `test(runtime): execute VLA memcpy dynamic size`
  - `docs: record VLA memcpy dynamic-size runtime coverage`

## Plan 100: VLA Failure Triage - Completed

The Plan 99 focused probe passed immediately, so no sema, codegen, or runtime fix was required.

## Plan 101: VLA Regression Coverage - Completed

The added focused runtime probe is the regression coverage for dynamic VLA byte counts flowing into builtin memory operations.

## Plan 102: VLA Coverage Verification - Completed

The coverage-only VLA increment passed focused runtime tests plus full verification.

- Commit message:
  - `docs: record VLA memcpy dynamic-size runtime coverage`

## Plan 103: Long Double Runtime Candidate Scan - Completed

The pre-plan adjustment scanned current long double runtime coverage and GCC accept files containing long double or long-double-like constants. Existing runtime coverage is already broad across plain math, tgmath, complex long double externs, direct long double arithmetic, by-value calls/returns, struct fields, compound assignment, and inc/dec.

- Result: no new low-risk long-double GCC accept entry-point fixture remained. The only long-double accept file with `main` is `c99-math-long-double-1.c`, already covered by the runtime manifest.

## Plan 104: Long Double Followup - Completed

The Plan 103 scan produced no low-risk implementation or coverage target, so this plan completed without code changes.

## Plan 105: Complex Runtime Candidate Scan - Completed

The pre-plan adjustment scanned complex-related GCC accept fixtures and the direct runtime complex suite. Complex runtime coverage already includes direct complex arithmetic, casts, argument/return passing, struct and array storage, pointer dereference, compound assignment, conditional/comma expressions, static initializers, and tgmath/extern paths.

- Result: no new low-risk complex GCC accept entry-point fixture remained. The remaining complex accept files are no-entry compile diagnostics/assembler-shape tests or are already represented by direct runtime coverage.

## Plan 106: Complex Followup - Completed

The Plan 105 scan produced no low-risk implementation or coverage target, so this plan completed without code changes.

## Plan 107: Documentation Timing Adjustment - Completed

The pre-plan adjustment deferred documentation until the next concrete coverage increment so the no-code scan results could be recorded together with actionable work.

## Plan 108: Rolling Plan Reorder - Completed

The plan list was reordered toward compile-only GCC accept fixtures with self-contained `main` functions after long double and complex scans produced no new low-risk target.

## Plan 109: Compile-Only Runtime Candidate Scan - Completed

The scan reused the compile-only `main` candidate list and selected `overflow-2.c` first because it was self-contained and small enough to test quickly.

## Plan 110: `overflow-2.c` Runtime Candidate Rejection - Completed

The focused probe returned exit 1. Root-cause confirmation with system `cc -std=c99 -pedantic-errors` produced the same exit status, so this is not a cvm runtime bug. `overflow-2.c` is compile-only warning coverage for an overflowing constant expression and does not promise exit-0 runtime semantics.

## Plan 118: Large Static Array Loop Runtime Coverage - Completed

The pre-plan adjustment selected `sema/testdata/gcc-c99-extra/accept/pr27639.c` after `overflow-2.c` was rejected. The fixture is self-contained and exercises a large static array write loop. The first probe exceeded the default 100000-step GCC runtime helper budget, so the helper was extended with a fixture-specific step-limit path while preserving the default budget for other tests.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCLargeStaticArrayLoopExecutesThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCLargeStaticArrayLoopExecutesThroughRuntime|TestGCCC90DeclarationAfterStatementExecutesThroughRuntime' -count=1 -v
```

- Commit messages:
  - `test(runtime): execute GCC large static array loop`
  - `docs: record GCC large static array runtime coverage`

## Plan 119: Large Static Array Failure Triage - Completed

The only failure was the helper step budget. Raising the budget only for `pr27639.c` resolved the focused probe without changing runtime semantics.

## Plan 120: Large Static Array Coverage Verification - Completed

The coverage increment passed focused runtime tests plus full verification.

- Commit message:
  - `docs: record GCC large static array runtime coverage`

## Plan 111: Stdlib/String/Stdio Surface Recheck - Completed

The pre-plan adjustment compared the builtin header smoke surfaces for stdlib, string, strings, and stdio against the default extern registry smoke list. Typedefs and stdio stream variables were intentionally ignored. The actionable result was not an implementation gap: plain `sprintf` and `snprintf` were already registered, but the registry smoke list and GCC runtime execution coverage only named builtin/chk formatter paths.

## Plan 112: Plain Stdio Formatter Coverage - Completed

The pre-plan adjustment added coverage for the existing plain `sprintf` and `snprintf` extern registrations. The original focused C test included `<stdio.h>` for the plain prototypes and declared `strcmp` manually, matching existing formatter tests and avoiding the then-unresolved duplicate-`size_t` limitation when combining `<stdio.h>` and `<string.h>`. Plan 147 later updated this test to include `<string.h>` now that shared `size_t` header guards exist.

- Files: `runtime/extern_test.go`, `runtime/gcc_exec_test.go`
- Focused test:

```bash
go test ./runtime -run 'TestPlainSprintfExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
```

- Commit messages:
  - `test(runtime): cover plain stdio sprintf externs`
  - `docs: record plain stdio formatter coverage`

## Plan 117: Remaining Compile-Only Main Candidate Scan - Completed

The pre-plan adjustment returned to self-contained compile-only GCC accept fixtures after stdio formatter coverage landed. `overflow-2.c` had already been rejected as not runtime-stable, and `pr27639.c` was covered, so the next low-risk candidate was `pr71969-1.c`.

## Plan 118: Inline Volatile Calls Runtime Coverage - Completed

The pre-plan adjustment added direct runtime coverage for `sema/testdata/gcc-c99-extra/accept/pr71969-1.c`. The fixture repeatedly calls inline functions that increment a volatile global, then returns zero. It runs cleanly with the fixture-specific step-limit helper.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCInlineVolatileCallsExecuteThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCInlineVolatileCallsExecuteThroughRuntime|TestGCCLargeStaticArrayLoopExecutesThroughRuntime' -count=1 -v
```

- Commit messages:
  - `test(runtime): execute GCC inline volatile calls`
  - `docs: record GCC inline volatile runtime coverage`

## Plan 119: Inline Candidate Triage - Completed

The Plan 118 focused probe passed immediately, so no sema, codegen, or runtime fix was required.

## Plan 120: Inline Coverage Verification - Completed

The inline/volatile coverage increment passed focused runtime tests plus full verification.

- Commit message:
  - `docs: record GCC inline volatile runtime coverage`

## Plan 123: GNU89 Inline Volatile Runtime Coverage - Completed

The pre-plan adjustment selected `sema/testdata/gcc-c99-extra/accept/pr71969-3.c`, the GNU89 inline variant of the prior fixture. It uses the same repeated volatile side-effect shape and also runs cleanly through runtime with the fixture-specific step-limit helper.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCGNU89InlineVolatileCallsExecuteThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCGNU89InlineVolatileCallsExecuteThroughRuntime|TestGCCInlineVolatileCallsExecuteThroughRuntime' -count=1 -v
```

- Commit messages:
  - `test(runtime): execute GCC GNU89 inline volatile calls`
  - `docs: record GCC GNU89 inline volatile runtime coverage`

## Plan 124: GNU89 Inline Candidate Triage - Completed

The Plan 123 focused probe passed immediately, so no sema, codegen, or runtime fix was required.

## Plan 125: GNU89 Inline Coverage Verification - Completed

The GNU89 inline coverage increment passed focused runtime tests plus full verification.

- Commit message:
  - `docs: record GCC GNU89 inline volatile runtime coverage`

## Plan 141: Header Typedef Collision Triage - Completed

The pre-plan adjustment revisited the duplicate `size_t` issue discovered while adding plain `sprintf` runtime coverage. A sema-level experiment that allowed compatible typedef redeclarations made the focused header test pass, but it incorrectly accepted GCC reject fixtures `c99-typedef-1.c` and `c90-typedef-1.c`. The root cause was therefore narrowed to builtin header expansion rather than ordinary C99 typedef rules.

## Plan 142: Header Typedef Regression Coverage - Completed

Added focused sema coverage proving both sides of the boundary:

- ordinary same-scope typedef redeclarations are still rejected
- `<stdio.h>` plus `<string.h>` can be preprocessed, parsed, and analyzed together

## Plan 143: Shared `size_t` Header Guard - Completed

Builtin headers that expose `size_t` now wrap the typedef in `__CVM_SIZE_T`, covering `<stddef.h>`, `<stdio.h>`, `<stdlib.h>`, `<string.h>`, `<strings.h>`, and the internal chk header. This fixes standard-header composition without weakening sema's ordinary typedef redeclaration diagnostics.

- Files: `preprocessor/headers.go`, `sema/c99_compliance_test.go`
- Focused tests:

```bash
go test ./sema -run 'TestC99(RejectsTypedefRedeclaration|BuiltinHeadersMayRedeclareSizeT)|TestGCCC99RejectSuite/c99-typedef-1.c|TestGCCC90AsC99ExplicitStdSuite/reject/c90-typedef-1.c' -count=1
go test ./sema -count=1
go test ./preprocessor -count=1
```

- Commit message:
  - `fix(headers): guard shared size_t typedef`

## Plan 144: Header Typedef Verification - Completed

The header typedef guard increment passed full verification:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Plan 146: Remaining Runtime Fixture Candidate Scan - Completed

The pre-plan adjustment rescanned GCC accept fixtures with `main`. The C99 `{ dg-do run }` fixtures are already represented in `runtime/testdata/gcc-exec/manifest.tsv`. The remaining compile-only `main` candidates were not better low-risk direct runtime targets: `inline-10.c` intentionally returns 1, `transparent-union-1.c` calls unresolved extern declarations, `pr70418.c` relies on GNU nested-function/VLA-in-struct extension behavior, and the originally requested VLA/conversion surfaces already have direct runtime-derived coverage. The next coherent increment was therefore adjusted toward exercising the newly fixed header composition through an existing runtime test.

## Plan 147: Plain Formatter Header Composition Runtime Coverage - Completed

Updated `TestPlainSprintfExecutesThroughRuntime` to include both `<stdio.h>` and `<string.h>` instead of manually declaring `strcmp`. This preserves plain `sprintf`/`snprintf` execution coverage and adds runtime compile-and-execute coverage for the shared guarded `size_t` typedef path.

- Files: `runtime/gcc_exec_test.go`
- Focused test:

```bash
go test ./runtime -run TestPlainSprintfExecutesThroughRuntime -count=1 -v
```

- Commit message:
  - `test(runtime): include string header in plain sprintf coverage`

## Plan 151: Header Composition Runtime Gap Scan - Completed

The pre-plan adjustment scanned runtime tests that include builtin headers exposing `size_t`. After Plan 147, `<stdio.h>` plus `<string.h>` was covered, but `<stdlib.h>` plus `<string.h>` still had no shared execution path.

## Plan 152: Stdlib/String Header Composition Runtime Coverage - Completed

Updated `TestPlainAllocationExecuteThroughRuntime` to include both `<stdlib.h>` and `<string.h>`, then validate the `strdup` result with `strlen` and `strcmp`. This keeps the existing allocation semantics coverage and adds another runtime path through the shared guarded `size_t` typedef.

- Files: `runtime/gcc_exec_test.go`
- Focused test:

```bash
go test ./runtime -run TestPlainAllocationExecuteThroughRuntime -count=1 -v
```

- Commit message:
  - `test(runtime): combine stdlib and string headers`

## Plan 154: VLA Candidate Rescan - Completed

The pre-plan adjustment rescanned VLA-related GCC accept fixtures. Existing runtime coverage already maps to `vla-2.c`, `vla-26.c`, and `Wstrict-aliasing-bogus-vla-1.c`. The clean remaining low-risk shape was `c99-vla-1.c`: a typedef naming a VLA, a local object of that typedef type, and a pointer to that VLA.

## Plan 155: VLA Typedef Pointer Runtime Coverage - Completed

Added a focused runtime probe derived from `c99-vla-1.c`. It fills a VLA through `A *p = &a` and reads back dynamic-indexed elements through the pointer.

- Files: `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./runtime -run TestGCCVLATypedefPointerExecutesThroughRuntime -count=1 -v
go test ./runtime -run 'TestGCCVLA(StructAndUnionMembers|ParameterDynamicStride|MemcpyDynamicSize|TypedefPointer)ExecutesThroughRuntime' -count=1 -v
go test ./runtime -run TestGCCVLAStructAndUnionMembersExecuteThroughRuntime -count=1 -v
```

- Commit message:
  - `test(runtime): execute VLA typedef pointer access`

## Plan 156: Phase 1 Maths Surface Rescan - Completed

The pre-plan adjustment rescanned builtin header prototypes, runtime registry entries, and runtime/direct test coverage for the Phase 1 maths surface. The scan found no remaining gaps:

- `<math.h>`: 171 plain prototypes, no missing registry entries, no untested names.
- `<complex.h>`: 66 plain prototypes, no missing registry entries, no untested names.
- `<tgmath.h>`: 61 pseudo functions, no missing declarations or registry entries.

## Plan 157: Type-Generic `signbit` Completion - Completed

The scan found one semantic cleanup rather than a missing extern: `<math.h>` `signbit` still forced its argument through `(double)` while the runtime already had `__cvm_signbitf`, `__cvm_signbit`, and `__cvm_signbitl`. Updating `signbit` to use `__cvm_math_select1` exposed a codegen bug for `sizeof`-based conditional macro dispatch nested inside a larger expression. Codegen now folds C99 integer-constant conditional expressions before emitting runtime branch labels.

- Files: `preprocessor/headers.go`, `codegen/expr.go`, `codegen/codegen_test.go`, `runtime/gcc_exec_test.go`
- Focused tests:

```bash
go test ./codegen -run 'TestGenerateConstantConditionalExpressionWithOuterStackValue|TestGCCBytecodeCompileSuite/signbit-sa.c' -count=1 -v
go test ./runtime -run 'TestGCCExecutionFixtures/signbit-sa.c|TestBuiltinFloatingConstantsExecuteThroughRuntime' -count=1 -v
```

- Commit message:
  - `fix(codegen): fold constant conditional expressions`

## Plan 158: Phase 1 Maths Closure Verification - Completed

The `signbit` increment passed full verification and was pushed as `e60581b`. A final mechanical scan confirmed the Phase 1 maths surface remains closed after the codegen fix:

```text
mathCount=171 missingMath=[] untestedMath=[]
complexCount=66 missingComplex=[] untestedComplex=[]
tgmathPseudoCount=61 missingPseudoDecl=[] missingPseudoReg=[]
```

Phase 1 maths is complete. Further runtime work should move to non-math execution gaps unless a new imported fixture reveals a regression in this surface.

## Plan 159: Phase 1 Non-Math Runtime Candidate Scan - Completed

The pre-plan adjustment scanned all imported GCC accept fixtures with `main` across the three tracked roots and removed paths already represented by the `{ dg-do run }` manifest or direct runtime tests. After the VLA, inline, stdio/header, and maths increments, the remaining candidates were:

```text
sema/testdata/gcc-c99-extra/accept/Wstrict-aliasing-bogus-vla-1.c
sema/testdata/gcc-c99-extra/accept/inline-10.c
sema/testdata/gcc-c99-extra/accept/inline-8.c
sema/testdata/gcc-c99-extra/accept/overflow-2.c
sema/testdata/gcc-c99-extra/accept/pr70418.c
sema/testdata/gcc-c99-extra/accept/transparent-union-1.c
```

Only `Wstrict-aliasing-bogus-vla-1.c` was a low-risk exit-0 runtime candidate.

## Plan 160: Strict Aliasing VLA Runtime Coverage - Completed

Added direct runtime coverage for `Wstrict-aliasing-bogus-vla-1.c`. The focused red test exposed a runtime loader gap rather than a VLA/memcpy bug: the fixture uses `int main(int argc, char *argv[])`, while Phase 1 loading previously accepted only no-argument `main`. Runtime loading now accepts integer-returning `main(int, char **)`, initializes default entry arguments as `argc = 1` and `argv = {"cvm", NULL}`, and passes those arguments into the initial VM frame.

- Files: `runtime/gcc_exec_test.go`, `runtime/program.go`, `runtime/vm.go`, `runtime/vm_test.go`
- Focused test:

```bash
go test ./runtime -run 'TestRunPassesDefaultArgcToMain|TestGCCStrictAliasingVLADereferenceExecutesThroughRuntime|TestGCCVLAMemcpyDynamicSizeExecutesThroughRuntime' -count=1 -v
```

- Commit message:
  - `test(runtime): execute GCC strict aliasing VLA fixture`

## Plan 161: Phase 1 Non-Math Runtime Closure - Completed

After Plan 160, the remaining compile-only `main` fixtures are intentionally not Phase 1 runtime targets:

- `inline-8.c`: inline declaration of `main` only; compile validation target, no runtime entry body.
- `inline-10.c`: inline `main` returns 1 and is not an exit-0 runtime candidate.
- `overflow-2.c`: compile-only overflow diagnostic fixture; system run exits 1.
- `pr70418.c`: GNU nested function with VLA-in-struct parameter extension; no stable Phase 1 runtime assertion.
- `transparent-union-1.c`: compile-only warning fixture with unresolved extern calls; not runnable without artificial extern definitions.

The Phase 1 non-math runtime fixture milestone is complete. Future work should be scoped as Phase 2 runtime expansion rather than continuing to mine the already-closed low-risk candidate set.

## Plan 162: Phase 2 Runtime Environment Design - Completed

The pre-plan adjustment split Phase 2 into a first bounded runtime environment milestone rather than jumping directly to larger ABI work. The milestone covers configured `argv`, configured stdin reads, and closed host stream handles. Larger items are reserved for Phase 2B: memory-backed `va_list`, escaping nested-function trampolines, and filesystem-backed streams.

- Files: `docs/superpowers/specs/2026-05-19-phase2-runtime-environment-design.md`, `docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md`
- Commit message:
  - `feat(runtime): pass configured argv to main`

## Plan 163: Configured Argv Runtime Support - Completed

Runtime loading now accepts `LoadOptions.Args`, allocates an `argv` pointer vector, and passes configured `argc`/`argv` into integer-returning `main(int, char **)`. `cvm run file.cvmbc [args...]` forwards the bytecode path as `argv[0]` and trailing command-line arguments as the remaining entries.

- Files: `runtime/program.go`, `runtime/vm.go`, `runtime/vm_test.go`, `runtime/gcc_exec_test.go`, `compiler_test.go`
- Focused tests:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestRunPassesCustomArgvToMain|TestGCCMainArgvExecutesThroughRuntime' -count=1 -v
env GOCACHE=/private/tmp/cvm-go-build-cache go test . -run TestRunBytecodeForwardsArgs -count=1 -v
```

- Commit message:
  - `feat(runtime): pass configured argv to main`

## Plan 164: Configured Stdin Runtime Support - Completed

The default extern registry now has `NewExternRegistryWithIO` and `DefaultExternRegistryWithIO` constructors. Stdio input externs consume configured stdin bytes after any `ungetc` pushback, covering `getchar`, `fgetc`, `fgets`, and `fread`.

- Files: `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`
- Focused test:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdioReadsConfiguredStdin|TestGCCStdioConfiguredStdinExecutesThroughRuntime' -count=1 -v
```

- Commit message:
  - `feat(runtime): read configured stdin streams`

## Plan 165: Closed Host Stream Handles - Completed

`fclose` now marks known host stream handles closed, clears their writer/fd/pushback/EOF state, and rejects later use through the same stream handle. Existing GCC execution coverage for successful `fclose(stdout)` remains green.

- Files: `runtime/extern.go`, `runtime/extern_test.go`
- Focused test:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestFcloseRejectsLaterStreamUse|TestStdioFcloseExecutesThroughRuntime' -count=1 -v
```

- Commit message:
  - `feat(runtime): reject closed host streams`

## Plan 166: Phase 2 Environment Closure - Completed

The Phase 2 runtime environment milestone is complete. Handoff and roadmap docs now separate the finished environment work from larger Phase 2B candidates: memory-backed `va_list`, escaping nested-function trampolines, filesystem-backed streams, and extern-backed warning fixtures.

- Files: `docs/bytecode-runtime-handoff.md`, `docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md`, `docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md`
- Commit message:
  - `docs: record phase 2 runtime environment closure`

## Continuous Execution Rule

After each plan is committed and pushed, immediately start the Common Pre-Plan Adjustment for the next plan. Keep at least twenty rolling followup plans visible, adjust the next plan against current repository state before executing it, and continue until a stop condition is reached.
