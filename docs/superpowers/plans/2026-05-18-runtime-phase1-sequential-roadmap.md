# Runtime Phase 1 Sequential Roadmap

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue bytecode/runtime phase 1 through at least twelve small, sequential libc/runtime increments.

**Architecture:** Each plan starts with a short state adjustment pass before tests are written: inspect current registry/header support, keep the plan if still missing, shrink it if some pieces already landed, or skip it if it no longer applies. Each increment follows red-green-full-verification, then lands as one implementation commit plus one handoff/docs commit when appropriate.

**Tech Stack:** Go runtime/interpreter, builtin C header strings in `preprocessor`, direct extern tests, GCC runtime execution tests, existing bytecode/codegen verification.

---

## Common Pre-Plan Adjustment

Before executing each numbered plan:

```bash
git status --short --branch
rg -n "<planned-symbols>" preprocessor runtime sema codegen docs/bytecode-runtime-handoff.md
```

Then adjust:

- If all symbols already have header declarations, registry entries, direct tests, GCC runtime tests, and handoff docs, mark the plan skipped and move to the next.
- If some symbols exist, remove those from the current increment and execute only the missing coherent subset.
- If focused red tests reveal a sema/codegen/runtime limitation outside the plan, stop that plan and choose the next safe plan.

## Common Verification

Every implementation commit and every handoff/docs commit must be preceded by:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Plan 1: `puts_unlocked` - Completed

Add `puts_unlocked` as an alias over the existing `puts` behavior.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestPutsWritesCString|TestStdioPutsUnlockedExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio puts unlocked alias`
  - `docs: record stdio puts unlocked alias`

## Plan 2: `vprintf_unlocked` And `vfprintf_unlocked` - Completed

Add literal-only unlocked vformat aliases using the existing vformat path that does not consume `va_list` values.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestNoArgVFormatExterns|TestStdioVFormatUnlockedAliasesExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio unlocked vformat aliases`
  - `docs: record stdio unlocked vformat aliases`

## Plan 3: `setbuf` And `setvbuf`

Add no-op buffering controls for known host stream handles.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Header symbols/macros: `_IOFBF`, `_IOLBF`, `_IONBF`, `BUFSIZ`, `setbuf`, `setvbuf`
- Focused tests:

```bash
go test ./runtime -run 'TestStdioBufferControls|TestStdioBufferControlsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio buffer control externs`
  - `docs: record stdio buffer control externs`

## Plan 4: Stdio Stream Lock No-Ops

Add `flockfile`, `ftrylockfile`, and `funlockfile` as no-op validation externs for known host stream handles.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdioStreamLockControls|TestStdioStreamLockControlsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdio stream lock externs`
  - `docs: record stdio stream lock externs`

## Plan 5: `stdlib.h` Absolute Value Helpers

Add builtin `<stdlib.h>` with `abs`, `labs`, and `llabs` runtime externs.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibAbsExterns|TestStdlibAbsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib abs externs`
  - `docs: record stdlib abs externs`

## Plan 6: `stdlib.h` Integer String Parsers

Add `atoi`, `atol`, and `atoll` over simple C decimal parsing: skip ASCII whitespace, accept optional sign, consume leading digits, stop at first non-digit.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStdlibAtoiExterns|TestStdlibAtoiExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add stdlib atoi externs`
  - `docs: record stdlib atoi externs`

## Plan 7: `ctype.h` Classification Helpers

Add builtin `<ctype.h>` and classification externs: `isdigit`, `isalpha`, `isalnum`, `isspace`, `islower`, `isupper`, `isxdigit`, `isprint`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestCtypeClassificationExterns|TestCtypeClassificationExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinCtypeHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add ctype classification externs`
  - `docs: record ctype classification externs`

## Plan 8: `ctype.h` Case Conversion Helpers

Add `tolower` and `toupper`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestCtypeCaseConversionExterns|TestCtypeCaseConversionExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinCtypeHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add ctype case conversion externs`
  - `docs: record ctype case conversion externs`

## Plan 9: `string.h` Read-Only Helpers

Add builtin `<string.h>` declarations for already modeled read-only helpers: `strcmp`, `memcmp`, `strlen`, `strchr`, and `strstr`, plus GCC runtime coverage through the header.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestStringHeaderReadOnlyHelpersExecuteThroughRuntime' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(preprocessor): declare string read-only helpers`
  - `docs: record string read-only helper header`

## Plan 10: Plain Memory Operation Aliases

Add plain runtime aliases and header declarations for `memcpy`, `memmove`, `mempcpy`, `memset`, and `bzero`, reusing existing builtin memory extern helpers.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestPlainMemoryOperationExterns|TestPlainMemoryOperationsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain memory operation aliases`
  - `docs: record plain memory operation aliases`

## Plan 11: Plain String Write Aliases

Add plain runtime aliases and header declarations for `strcpy`, `stpcpy`, `strcat`, `strncpy`, `stpncpy`, and `strncat`, reusing existing builtin string-writing extern helpers.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestPlainStringWriteExterns|TestPlainStringWritesExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStringHeaderDeclaresReadOnlySurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain string write aliases`
  - `docs: record plain string write aliases`

## Plan 12: Plain Allocation Aliases

Add builtin `<stdlib.h>` declarations and runtime aliases for `malloc`, `calloc`, `free`, and `strdup`. Reuse existing allocation paths for allocation/duplication and use runtime memory free for `free`.

- Files: `preprocessor/headers.go`, `preprocessor/headers_test.go`, `runtime/extern.go`, `runtime/extern_test.go`, `runtime/gcc_exec_test.go`, `docs/bytecode-runtime-handoff.md`
- Focused tests:

```bash
go test ./runtime -run 'TestPlainAllocationExterns|TestPlainAllocationExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdlibHeaderDeclaresRuntimeSurface -count=1 -v
```

- Commit messages:
  - `feat(runtime): add plain allocation externs`
  - `docs: record plain allocation externs`

## Continuous Execution Rule

After each plan is committed and pushed, immediately start the Common Pre-Plan Adjustment for the next plan. Continue until a stop condition is reached or all twelve plans are complete.
