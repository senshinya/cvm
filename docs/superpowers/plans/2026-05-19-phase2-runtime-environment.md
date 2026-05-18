# Phase 2 Runtime Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Phase 2 hosted runtime environment support for configurable `argv`, stdin-backed reads, and closed stream handles.

**Architecture:** Extend existing runtime options and extern registry APIs without breaking existing tests. Keep each increment small: first arguments, then stdin, then stream lifetime, then documentation closure. Existing pushback and host writer maps remain the foundation for stdio behavior.

**Tech Stack:** Go runtime/interpreter, bytecode loader tests, CLI tests in `main.go` package, runtime extern tests, GCC runtime execution snippets, docs under `docs/superpowers` and `docs/bytecode-runtime-handoff.md`.

---

## Common Verification

Before every implementation and docs commit:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Task 1: Configurable Program Arguments

**Files:**
- Modify: `runtime/program.go`
- Modify: `runtime/vm_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `main.go`
- Modify: `compiler_test.go` or `compiler_dump_test.go` only if existing CLI tests require updates

- [x] **Step 1: Add failing runtime tests for custom `argv`**

Add a runtime or GCC execution test that loads a program with `LoadOptions{Args: []string{"prog", "alpha", "beta"}}` and verifies `argc == 3`, `argv[0] == "prog"`, `argv[1] == "alpha"`, and `argv[2] == "beta"`.

- [x] **Step 2: Add failing CLI test for `cvm run file.cvmbc arg...`**

Extend the CLI tests to compile a small program to bytecode, run it with extra args, and expect the program to observe them.

- [x] **Step 3: Implement `LoadOptions.Args`**

Add `Args []string` to `runtime.LoadOptions`. Change entry argument setup so nil args default to `[]string{"cvm"}`, while non-nil args are allocated exactly as provided. Allocate each string as a C string and allocate an `(argc + 1)` pointer vector terminated with null.

- [x] **Step 4: Forward CLI args**

Change `cvm run` usage to `cvm run file.cvmbc [args...]` and pass the trailing arguments to `runtime.Load`.

- [x] **Step 5: Run focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestRunPassesCustomArgvToMain|TestGCCMainArgvExecutesThroughRuntime' -count=1 -v
env GOCACHE=/private/tmp/cvm-go-build-cache go test . -run TestRunBytecodeForwardsArgs -count=1 -v
```

- [x] **Step 6: Full verification, commit, push**

Run Common Verification, then:

```bash
git add runtime/program.go runtime/vm_test.go runtime/gcc_exec_test.go main.go compiler_test.go compiler_dump_test.go docs/superpowers/specs/2026-05-19-phase2-runtime-environment-design.md docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md
git commit -m "feat(runtime): pass configured argv to main"
git push origin codex/bytecode-runtime-phase-1
```

## Task 2: Stdin-Backed Reads

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `runtime/integration_test.go` if compile-and-run helper needs stdin support

- [ ] **Step 1: Add failing direct extern tests**

Add tests proving `getchar`, `fgetc`, `fgets`, and `fread` consume bytes from configured stdin after existing `ungetc` bytes are exhausted.

- [ ] **Step 2: Add failing GCC runtime stdin snippet**

Add a helper or test path that runs a C snippet with stdin set to a small string and verifies character, line, and block reads.

- [ ] **Step 3: Implement stdin reader plumbing**

Add a backward-compatible constructor such as `NewExternRegistryWithIO(stdin io.Reader, stdout, stderr io.Writer)` and `DefaultExternRegistryWithIO(stdin io.Reader, stdout, stderr io.Writer)`. Store stdin as a reader in the registry and have `readHostChar` fall back to that reader after pushback.

- [ ] **Step 4: Run focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdioReadsConfiguredStdin|TestGCCStdioConfiguredStdinExecutesThroughRuntime' -count=1 -v
```

- [ ] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go runtime/integration_test.go docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md
git commit -m "feat(runtime): read configured stdin streams"
git push origin codex/bytecode-runtime-phase-1
```

## Task 3: Closed Stream Handles

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`

- [ ] **Step 1: Add failing closed-handle tests**

Add direct extern coverage where `fclose(stdout)` succeeds and a later `fputc('x', stdout)` returns an unknown/closed handle error.

- [ ] **Step 2: Implement closed host stream tracking**

Add a closed-handle set to `ExternRegistry`. `fclose` marks the handle closed and removes or masks writer/reader/fd state. Shared lookup helpers must reject closed handles.

- [ ] **Step 3: Run focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestFcloseRejectsLaterStreamUse|TestStdioFcloseExecutesThroughRuntime' -count=1 -v
```

- [ ] **Step 4: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md
git commit -m "feat(runtime): reject closed host streams"
git push origin codex/bytecode-runtime-phase-1
```

## Task 4: Phase 2 Environment Closure Docs

**Files:**
- Modify: `docs/bytecode-runtime-handoff.md`
- Modify: `docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md`

- [ ] **Step 1: Update handoff known limits**

Remove or narrow the fixed-argv, no-input stdin, and no-op fclose limits. Add next Phase 2B candidates: memory-backed `va_list`, escaping nested-function trampolines, and filesystem-backed streams.

- [ ] **Step 2: Mark plan complete and verify**

Run Common Verification, then commit:

```bash
git add docs/bytecode-runtime-handoff.md docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md docs/superpowers/plans/2026-05-19-phase2-runtime-environment.md
git commit -m "docs: record phase 2 runtime environment closure"
git push origin codex/bytecode-runtime-phase-1
```
