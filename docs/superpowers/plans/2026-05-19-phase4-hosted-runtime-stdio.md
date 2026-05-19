# Phase 4 Hosted Runtime Stdio Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the bounded Phase 4 hosted-runtime stdio/input surface: scansets, floating and pointer formatted input, formatted-input failure returns, update-mode stream sequencing, GCC-style runtime coverage, and closure docs.

**Architecture:** Extend the existing `runtime/extern.go` scanner and hermetic stream helpers in small increments. Keep the shared `scanf`/`fscanf`/`sscanf` scanner as the single formatted-input path, and add stream direction state only to hermetic stream handles.

**Tech Stack:** Go runtime package, CVM bytecode runtime tests, GCC-style C snippets compiled through the existing runtime test harness.

---

## Milestone 1: Baseline And Branch Setup

Calibration: Phase 3 is closed at `fef3f3b`; Phase 4 is scoped to residual stdio/input behavior from the closure docs.

**Files:**
- Create: `docs/superpowers/specs/2026-05-19-phase4-hosted-runtime-stdio-design.md`
- Create: `docs/superpowers/plans/2026-05-19-phase4-hosted-runtime-stdio.md`

- [x] **Step 1: Create Phase 4 branch**

Run: `git switch -c codex/bytecode-runtime-phase-4`
Expected: new local branch from the closed Phase 3 tip.

- [x] **Step 2: Write Phase 4 design and plan**

Record scope, residual surfaces, test policy, and milestone order.

- [x] **Step 3: Verify, commit, and push planning docs**

Run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
git add docs/superpowers/specs/2026-05-19-phase4-hosted-runtime-stdio-design.md docs/superpowers/plans/2026-05-19-phase4-hosted-runtime-stdio.md
git commit -m "docs: define phase 4 hosted stdio roadmap"
git push -u origin codex/bytecode-runtime-phase-4
```

## Milestone 2: Scanset Formatted Input

Calibration before execution: Re-read `scanString` and existing `TestSscanfExternScans...` coverage. Confirm `%[` still reports unsupported format before changing implementation.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add failing direct scanset tests**

Add direct `sscanf` tests for `%[a-z]`, `%[^0-9]`, width, and assignment suppression.

- [x] **Step 2: Implement scanset parsing**

Add scanner helpers for bracket expressions, negation, leading `]`, ranges, width, and no-match failure.

- [x] **Step 3: Add source-level runtime scanset coverage**

Add a GCC-style runtime test that compiles C using `sscanf` scansets and validates output buffers and return counts.

- [x] **Step 4: Verify, commit, and push scansets**

Run the standard verification chain, then commit `feat(runtime): scan formatted input sets`.

## Milestone 3: Floating Formatted Input

Calibration before execution: Re-check existing `strtof`/`strtod`/`strtold` parsing helpers and scanner length-modifier handling. Prefer reuse over new ad hoc float parsing.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add failing direct float scanf tests**

Cover `%f` into `float *`, `%lf` into `double *`, `%Lf` into long-double storage, plus `%a` hex-float input.

- [x] **Step 2: Implement float scan storage**

Parse a consumed float prefix, store using the target value type, and keep assignment counts and consumed offsets correct.

- [x] **Step 3: Add source-level runtime float scanf coverage**

Compile a C snippet using `sscanf` and `fscanf` for decimal and hex floats.

- [x] **Step 4: Verify, commit, and push float scanning**

Run the standard verification chain, then commit `feat(runtime): scan floating input formats`.

## Milestone 4: Pointer Formatted Input

Calibration before execution: Confirm pointer values are stored as runtime pointer-sized integers and decide accepted textual forms before writing implementation.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add failing direct pointer scanf tests**

Cover `%p` with `0x...`, null-like zero input, width, and first-conversion failure.

- [x] **Step 2: Implement pointer scan storage**

Parse hexadecimal pointer input and store into `void **`/pointer-sized memory consistently with the runtime target.

- [x] **Step 3: Add source-level pointer scanf coverage**

Compile a C snippet that scans a pointer and compares the resulting integer representation.

- [x] **Step 4: Verify, commit, and push pointer scanning**

Run the standard verification chain, then commit `feat(runtime): scan pointer input formats`.

## Milestone 5: Formatted Input Failure And EOF Semantics

Calibration before execution: Compare current scanner returns with C expectations for empty input, whitespace-only input, literal mismatch, and matching failure after prior assignment.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add focused failure-return tests**

Cover empty input returning EOF when no conversion occurs, first matching failure returning `0`, and later matching failure returning the number of completed assignments.

- [x] **Step 2: Adjust scanner result model**

Return assignment count, consumed bytes, and an EOF/input-failure classification instead of flattening all no-match cases.

- [x] **Step 3: Add source-level failure coverage**

Compile C snippets for empty `scanf` stdin and failed `sscanf` conversions.

- [x] **Step 4: Verify, commit, and push failure semantics**

Run the standard verification chain, then commit `fix(runtime): distinguish scanf input failures`.

## Milestone 6: Update-Mode Stream Direction State

Calibration before execution: Re-read hermetic stream mode parsing, `fread`, `fwrite`, `fseek`, `rewind`, and `fflush`; confirm update streams currently allow invalid read/write transitions.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [x] **Step 1: Add failing update-mode direct tests**

Cover read-after-write without `fflush`/`fseek`, write-after-read before EOF without positioning, and allowed transitions after `fflush`, `fseek`, `rewind`, or read EOF.

- [x] **Step 2: Add stream direction bookkeeping**

Track last operation on hermetic update streams and gate invalid transitions through the existing error indicator path.

- [x] **Step 3: Add source-level update-mode coverage**

Compile a C snippet using `fopen("r+")` or `fopen("w+")` over configured files and validate allowed transition cases.

- [x] **Step 4: Verify, commit, and push update-mode sequencing**

Run the standard verification chain, then commit `feat(runtime): enforce update stream sequencing`.

## Milestone 7: Cross-Extern Scan Coverage

Calibration before execution: After scansets/floats/pointers/failures land, re-run targeted searches to find scanner paths covered only through `sscanf`, then add `scanf`/`fscanf` coverage where needed.

**Files:**
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: this plan file

- [ ] **Step 1: Add stdin-backed cross-extern tests**

Cover at least one scanset or float through `scanf` with configured stdin.

- [ ] **Step 2: Add file-backed cross-extern tests**

Cover at least one scanset or float through `fscanf` over a configured hermetic file, including unread-byte preservation.

- [ ] **Step 3: Verify, commit, and push cross-extern coverage**

Run the standard verification chain, then commit `test(runtime): cover scanf scanner variants`.

## Milestone 8: GCC Runtime Fixture Recheck

Calibration before execution: Re-run the existing GCC runtime gap report and scan compile-only `main` candidates for newly unblocked stdio/input fixtures.

**Files:**
- Modify: `runtime/testdata/gcc-exec/gap-report.md` if regenerated content changes
- Modify: `runtime/gcc_exec_test.go` only if a low-risk newly unblocked fixture exists
- Modify: this plan file

- [ ] **Step 1: Re-run GCC runtime gap report**

Run: `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1`

- [ ] **Step 2: Scan for newly unblocked stdio fixtures**

Use `rg` over imported GCC accept roots for `scanf`, `fscanf`, `sscanf`, `%[`, `%f`, and `%p`.

- [ ] **Step 3: Add one fixture only if it is low-risk and exit-0**

If no candidate exists, record that explicitly in this plan and docs.

- [ ] **Step 4: Verify, commit, and push fixture recheck**

Run the standard verification chain, then commit either fixture coverage or docs-only recheck.

## Milestone 9: Phase 4 Closure Docs

Calibration before execution: Confirm all prior Phase 4 milestones are committed and pushed, and that residual docs no longer list scansets, float input, pointer input, EOF scan returns, or update-mode sequencing as Phase 4 gaps.

**Files:**
- Modify: `docs/bytecode-runtime-handoff.md`
- Create or modify: `docs/phase4-hosted-runtime-stdio-gap-map.md`
- Modify: this plan file

- [ ] **Step 1: Update handoff**

Record branch, latest commit, closed Phase 4 milestones, verification commands, and residual non-Phase-4 limits.

- [ ] **Step 2: Add Phase 4 gap map**

Record calibration, closed items, and any deferred Phase 5 candidates.

- [ ] **Step 3: Final verify, commit, and push closure docs**

Run the standard verification chain, then commit `docs: close phase 4 hosted stdio work`.
