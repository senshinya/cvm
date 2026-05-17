# Stdio Runtime Followups Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue bytecode/runtime phase 1 by adding small stdio runtime externs that reuse the current host-stream model.

**Architecture:** Keep each increment narrow: add builtin `<stdio.h>` declarations, register one alias/function group in `runtime.DefaultExternRegistry`, add direct extern tests, add one GCC runtime execution test, then update the handoff. Avoid functions that require real filesystem streams or seekable stream state.

**Tech Stack:** Go runtime/interpreter tests, builtin C header strings, GCC-style runtime fixture execution through `runtime/gcc_exec_test.go`.

---

## File Map

- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers.go`: extend builtin `<stdio.h>` declarations and simple stdio macros.
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers_test.go`: require the new builtin header identifiers.
- Modify `/Users/shinya/Downloads/cvm/runtime/extern.go`: register and implement externs using existing helpers.
- Modify `/Users/shinya/Downloads/cvm/runtime/extern_test.go`: add direct extern coverage for registry presence and behavior.
- Modify `/Users/shinya/Downloads/cvm/runtime/gcc_exec_test.go`: add GCC execution coverage for each new surface.
- Modify `/Users/shinya/Downloads/cvm/docs/bytecode-runtime-handoff.md`: record each completed increment after implementation.

## Execution Rules

- For every task, first write tests and run a focused command that fails for the expected missing declaration/extern.
- Keep one task per implementation commit and one handoff docs commit after it.
- Before each implementation commit and each docs commit, run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Task 1: `puts_unlocked` Alias

**Files:**
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers.go`
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/gcc_exec_test.go`
- Modify `/Users/shinya/Downloads/cvm/docs/bytecode-runtime-handoff.md`

- [ ] **Step 1: Write failing tests**

Add `puts_unlocked` to the builtin header identifier list and extern registry list. Extend the direct puts test to run both `puts` and `puts_unlocked`, expecting newline output and i32 `len + 1`. Add a GCC runtime test that calls `puts_unlocked("unlocked")`.

- [ ] **Step 2: Verify red**

Run:

```bash
go test ./runtime -run 'TestPutsWritesCString|TestStdioPutsUnlockedExecutesThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

Expected: FAIL mentioning missing `puts_unlocked` extern and/or undeclared identifier.

- [ ] **Step 3: Implement minimal alias**

In `preprocessor/headers.go`, add:

```c
int puts_unlocked(const char *);
```

In `runtime/extern.go`, replace the single `puts` registration with a loop registering `puts` and `puts_unlocked` over a shared `putsExtern(name, r)` helper.

- [ ] **Step 4: Verify green and full suite**

Run the focused commands from Step 2, then run the full verification commands from Execution Rules.

- [ ] **Step 5: Commit and push**

```bash
git add preprocessor/headers.go preprocessor/headers_test.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go
git commit -m "feat(runtime): add stdio puts unlocked alias"
```

Update `docs/bytecode-runtime-handoff.md`, rerun full verification, commit:

```bash
git add docs/bytecode-runtime-handoff.md
git commit -m "docs: record stdio puts unlocked alias"
git push origin codex/bytecode-runtime-phase-1
```

## Task 2: `vprintf_unlocked` And `vfprintf_unlocked` Literal Aliases

**Files:**
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers.go`
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/gcc_exec_test.go`
- Modify `/Users/shinya/Downloads/cvm/docs/bytecode-runtime-handoff.md`

- [ ] **Step 1: Write failing tests**

Add `vprintf_unlocked` and `vfprintf_unlocked` to header and registry expectations. Add direct extern coverage for literal formats that do not consume `va_list` values, mirroring existing no-arg vformat tests. Add a GCC runtime test with `va_start`/`va_end` wrappers that call literal-only `vprintf_unlocked("vu", ap)` and `vfprintf_unlocked(stderr, "vf", ap)`.

- [ ] **Step 2: Verify red**

Run:

```bash
go test ./runtime -run 'TestNoArgVFormatExterns|TestStdioVFormatUnlockedAliasesExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

Expected: FAIL mentioning missing unlocked vformat externs or undeclared identifiers.

- [ ] **Step 3: Implement minimal aliases**

In `preprocessor/headers.go`, add:

```c
int vprintf_unlocked(const char *, void *);
int vfprintf_unlocked(FILE *, const char *, void *);
```

In `runtime/extern.go`, include `vprintf_unlocked` in the vprintf registration group and `vfprintf_unlocked` in the vfprintf registration group, using the existing no-arg vformat path.

- [ ] **Step 4: Verify green and full suite**

Run the focused commands from Step 2, then run the full verification commands from Execution Rules.

- [ ] **Step 5: Commit and push**

```bash
git add preprocessor/headers.go preprocessor/headers_test.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go
git commit -m "feat(runtime): add stdio unlocked vformat aliases"
```

Update `docs/bytecode-runtime-handoff.md`, rerun full verification, commit and push:

```bash
git add docs/bytecode-runtime-handoff.md
git commit -m "docs: record stdio unlocked vformat aliases"
git push origin codex/bytecode-runtime-phase-1
```

## Task 3: `setbuf` And `setvbuf` No-Op Buffer Controls

**Files:**
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers.go`
- Modify `/Users/shinya/Downloads/cvm/preprocessor/headers_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/extern_test.go`
- Modify `/Users/shinya/Downloads/cvm/runtime/gcc_exec_test.go`
- Modify `/Users/shinya/Downloads/cvm/docs/bytecode-runtime-handoff.md`

- [ ] **Step 1: Write failing tests**

Add `_IOFBF`, `_IOLBF`, `_IONBF`, `BUFSIZ`, `setbuf`, and `setvbuf` to header coverage. Add registry coverage for `setbuf` and `setvbuf`. Add direct extern coverage that `setvbuf(stdout, 0, _IONBF, 0)` returns i32 0 and `setbuf(stdout, 0)` validates the stream and returns void/no trap. Add a GCC runtime test that calls both before writing to stdout.

- [ ] **Step 2: Verify red**

Run:

```bash
go test ./runtime -run 'TestStdioBufferControls|TestStdioBufferControlsExecuteThroughRuntime|TestDefaultExternRegistryHasExitAndAbort' -count=1 -v
go test ./preprocessor -run TestBuiltinStdioHeaderDeclaresFormattingSurface -count=1 -v
```

Expected: FAIL mentioning missing `setbuf`/`setvbuf` externs or undeclared identifiers.

- [ ] **Step 3: Implement no-op buffering controls**

In `preprocessor/headers.go`, add:

```c
#define _IOFBF 0
#define _IOLBF 1
#define _IONBF 2
#define BUFSIZ 8192
void setbuf(FILE * restrict, char * restrict);
int setvbuf(FILE * restrict, char * restrict, int, size_t);
```

In `runtime/extern.go`, register `setbuf` and `setvbuf`. Validate the stream pointer is a known host handle. Return zero from `setvbuf`; return the zero `Value{}` from `setbuf`.

- [ ] **Step 4: Verify green and full suite**

Run the focused commands from Step 2, then run the full verification commands from Execution Rules.

- [ ] **Step 5: Commit and push**

```bash
git add preprocessor/headers.go preprocessor/headers_test.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go
git commit -m "feat(runtime): add stdio buffer control externs"
```

Update `docs/bytecode-runtime-handoff.md`, rerun full verification, commit and push:

```bash
git add docs/bytecode-runtime-handoff.md
git commit -m "docs: record stdio buffer control externs"
git push origin codex/bytecode-runtime-phase-1
```

## Stop Conditions

- Stop if a task requires real host input, filesystem stream state, errno semantics, or seekable stream offsets.
- Stop if focused tests expose a sema/codegen limitation unrelated to the targeted stdio surface.
- Stop if full verification fails and the failure is not traceable to the current increment.
