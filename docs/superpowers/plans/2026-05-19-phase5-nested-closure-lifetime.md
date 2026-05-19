# Phase 5 Nested Closure Lifetime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close Phase 5 by making expired GNU nested-function closure pointer calls trap with an explicit lifecycle diagnostic while preserving valid in-frame closure behavior.

**Architecture:** Keep closure lifetime enforcement inside `runtime/vm.go`. Move closure addresses from the live closure map to an expired-closure map when their owning frame exits, and check that map during indirect calls before falling back to ordinary function-address lookup.

**Tech Stack:** Go runtime VM, existing C frontend/codegen pipeline, GNU nested-function runtime tests.

---

## Milestone 1: Baseline And Branch Setup

Calibration: Phase 4 is closed at `76dd4f8`; Phase 5 is scoped to the explicit nested-closure lifecycle diagnostic candidate recorded in `docs/phase4-hosted-runtime-stdio-gap-map.md`.

**Files:**
- Create: `docs/superpowers/specs/2026-05-19-phase5-nested-closure-lifetime-design.md`
- Create: `docs/superpowers/plans/2026-05-19-phase5-nested-closure-lifetime.md`

- [x] **Step 1: Create Phase 5 branch**

Run: `git switch -c codex/bytecode-runtime-phase-5`
Expected: new local branch from the closed Phase 4 tip.

- [x] **Step 2: Write Phase 5 design and plan**

Record scope, excluded work, runtime-only architecture, testing policy, and milestone order.

- [x] **Step 3: Verify, commit, and push planning docs**

Run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
git add docs/superpowers/specs/2026-05-19-phase5-nested-closure-lifetime-design.md docs/superpowers/plans/2026-05-19-phase5-nested-closure-lifetime.md
git commit -m "docs: define phase 5 closure lifetime roadmap"
git push -u origin codex/bytecode-runtime-phase-5
```

## Milestone 2: Expired Closure Trap

Calibration before execution: Re-read `runtime/vm.go` around `OpCallIndirect`, `popFrame`, and `makeClosure`. Confirm escaped capturing nested-function pointers currently trap through the generic invalid indirect-call target path.

**Files:**
- Modify: `runtime/vm.go`
- Modify: `runtime/integration_test.go`
- Modify: this plan file

- [x] **Step 1: Add failing source-level trap test**

Add a GNU nested-function integration test that returns a capturing nested function pointer from its enclosing function, calls it in `main`, and expects an error containing `expired closure pointer`.

- [x] **Step 2: Track expired closure addresses**

Add an expired-closure map to `VM`, move frame-owned closures there during `popFrame`, and check it before normal function-address lookup in `OpCallIndirect`.

- [x] **Step 3: Verify focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestCompileAndRunEscapedNestedFunctionPointerTraps|TestGCCNestedFunctionPointer' -count=1
```

- [x] **Step 4: Verify, commit, and push expired closure trap**

Run the standard verification chain, then commit `fix(runtime): trap expired nested closures`.

## Milestone 3: Nested Closure Regression Sweep

Calibration before execution: Re-scan existing nested-function pointer tests and source-level closure coverage. Add coverage only where the expired closure map could accidentally affect valid live closure calls.

Result: Existing coverage is sufficient. `runtime/integration_test.go` covers direct nested calls, VLA capture, outer-capture forwarding, and the new escaped closure trap. `runtime/gcc_exec_test.go` covers local closure pointer calls, passing closure pointers to another callee, VLA capture through closure pointers, and transitive capture through closure pointers. The focused nested closure suite passed after the expired-closure map change.

**Files:**
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add source-level nonescaping closure regression if needed**

If current coverage already exercises local, passed, VLA, and transitive live closure calls, record that no new test is needed. Otherwise add one focused GCC-style runtime test.

- [x] **Step 2: Run nested closure focused suite**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'NestedFunctionPointer|NestedFunctionDirectCall|NestedFunctionVLA|NestedFunctionForwardsOuterCapture' -count=1
```

- [x] **Step 3: Verify, commit, and push regression sweep**

If no code/test change is needed, commit this plan update as docs-only. Otherwise run the standard verification chain and commit the coverage increment.

## Milestone 4: GCC Fixture Recheck

Calibration before execution: Re-run GCC runtime gap reporting and scan imported GCC accept fixtures for nested-function pointer cases that are newly better classified by the explicit expired-closure diagnostic.

**Files:**
- Modify: `runtime/testdata/gcc-exec/gap-report.md` if regenerated content changes
- Modify: `runtime/gcc_exec_test.go` only if a low-risk fixture exists
- Modify: this plan file

- [ ] **Step 1: Re-run GCC runtime gap report**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1
```

- [ ] **Step 2: Scan nested-function accept fixtures**

Use `rg` over imported GCC accept roots for nested function pointer patterns.

- [ ] **Step 3: Add fixture only if it is low-risk and has stable expected behavior**

If no candidate exists, record that explicitly.

- [ ] **Step 4: Verify, commit, and push fixture recheck**

Run the standard verification chain, then commit fixture coverage or docs-only recheck.

## Milestone 5: Phase 5 Closure Docs

Calibration before execution: Confirm all prior Phase 5 milestones are committed and pushed, and that handoff/gap docs no longer list expired closure diagnostics as an open Phase 5 gap.

**Files:**
- Modify: `docs/bytecode-runtime-handoff.md`
- Create: `docs/phase5-nested-closure-lifetime-gap-map.md`
- Modify: this plan file

- [ ] **Step 1: Update handoff**

Record branch, latest commit, closed Phase 5 milestones, verification commands, and residual non-Phase-5 limits.

- [ ] **Step 2: Add Phase 5 gap map**

Record calibration, closed items, and candidate follow-up work.

- [ ] **Step 3: Final verify, commit, and push closure docs**

Run the standard verification chain, then commit `docs: close phase 5 closure lifetime work`.
