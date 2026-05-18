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

## Continuous Execution Rule

After each plan is committed and pushed, immediately start the Common Pre-Plan Adjustment for the next plan. Continue until a stop condition is reached or all ten followup plans are complete.
