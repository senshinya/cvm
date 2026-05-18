# Phase 2B Nested Trampolines Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support GNU capturing nested-function pointers while their enclosing frame is alive.

**Architecture:** Add a `MakeClosure` bytecode instruction that creates VM-managed trampoline pointers. `CallIndirect` detects those pointers and appends stored static-chain capture arguments before invoking the target function.

**Tech Stack:** Go bytecode/codegen/runtime, existing sema GNU nested function model, GCC runtime execution tests.

---

### Task 1: Closure Bytecode And Runtime

**Files:**
- Modify: `bytecode/opcode.go`
- Modify: `bytecode/printer.go`
- Modify: `bytecode/printer_test.go`
- Modify: `bytecode/validator.go`
- Modify: `runtime/vm.go`
- Modify: `codegen/expr.go`
- Modify: `codegen/codegen_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase2b-nested-trampolines.md`

- [x] **Step 1: Add failing tests**

Replace the old rejection expectation with codegen coverage for `MakeClosure`, and add GCC runtime tests for local and passed capturing nested function pointers.

- [x] **Step 2: Implement bytecode instruction plumbing**

Add `OpMakeClosure`, printer output, validator global/signature/stack checks, and runtime closure records.

- [x] **Step 3: Lower capturing function addresses to closures**

Update `emitFunctionAddress` so capturing nested functions emit capture arguments and `MakeClosure` instead of returning a diagnostic.

- [x] **Step 4: Focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -run TestGenerateCapturingNestedFunctionPointerUsesClosure -count=1 -v
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestGCCNestedFunctionPointerCaptureExecutesThroughRuntime|TestGCCNestedFunctionPointerPassedToCalleeExecutesThroughRuntime' -count=1 -v
```

- [x] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add bytecode/opcode.go bytecode/printer.go bytecode/printer_test.go bytecode/validator.go runtime/vm.go codegen/expr.go codegen/codegen_test.go runtime/gcc_exec_test.go docs/superpowers/specs/2026-05-19-phase2b-nested-trampolines-design.md docs/superpowers/plans/2026-05-19-phase2b-nested-trampolines.md
git commit -m "feat(runtime): call capturing nested function pointers"
git push origin codex/bytecode-runtime-phase-1
```
