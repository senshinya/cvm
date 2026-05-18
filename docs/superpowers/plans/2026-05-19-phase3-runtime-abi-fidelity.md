# Phase 3 Runtime ABI Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Phase 3 hosted-runtime and ABI-fidelity work through 14 calibrated milestones.

**Architecture:** Keep CVM deterministic: source-level behavior lowers into bytecode/runtime representations, hosted state is configured through runtime APIs, and C programs do not read ambient host filesystem or environment by default. Every milestone starts with a calibration scan before implementation.

**Tech Stack:** Go, CVM bytecode/codegen/runtime, builtin C headers in `preprocessor`, sema analysis, GCC-style runtime execution tests.

---

## Common Verification

Run after every coherent increment:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Milestone Rule

Before each milestone, write a short calibration note in this plan or the roadmap describing current support, missing behavior, and whether the milestone should proceed unchanged, shrink, split, or close as already complete.

## Milestone 1: Phase 3 Baseline And Gap Map

**Files:**
- Create: `docs/superpowers/specs/2026-05-19-phase3-runtime-abi-fidelity-design.md`
- Create: `docs/phase3-runtime-gap-map.md`
- Create: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Calibrate current state**

Scanned handoff, runtime/codegen/header support, and tests for `va_arg`, `va_list`, `scanf`, `FILE` state, environment, termination, long double, complex, and aggregate runtime coverage.

- [x] **Step 2: Record Phase 3 design**

Saved the Phase 3 scope, architecture, milestone list, and completion criteria.

- [x] **Step 3: Record gap map**

Saved the current high-value runtime gaps and selected source-level `va_arg` as the first implementation target.

- [x] **Step 4: Verify, commit, and push**

Run Common Verification, then:

```bash
git add docs/superpowers/specs/2026-05-19-phase3-runtime-abi-fidelity-design.md docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "docs: define phase 3 runtime abi roadmap"
git push -u origin codex/bytecode-runtime-phase-3
```

## Milestone 2: Source-Level `va_arg` And Memory `va_list` Unification

**Calibration:** Pending. Start by inspecting `<stdarg.h>` macro lowering, sema representation for `__builtin_va_arg` if any, bytecode `OpVa*`, and runtime memory-backed `va_list` layout.

**Initial target:** Replace the `va_arg(ap, type)` placeholder with a source-level path that executes real argument reads for integer and pointer values.

## Milestone 3: `va_copy`, Multiple `va_list`, And Nested Varargs

**Calibration:** Pending after Milestone 2.

## Milestone 4: Formatted Input Runtime

**Calibration:** Pending after varargs milestones.

## Milestone 5: FILE State Model Completion

**Calibration:** Pending after formatted input scan.

## Milestone 6: Hermetic File Mode Expansion

**Calibration:** Pending after FILE state model.

## Milestone 7: Environment Runtime

**Calibration:** Pending after file milestones.

## Milestone 8: Program Termination Semantics

**Calibration:** Pending after environment runtime.

## Milestone 9: Long Double And Complex Runtime Fidelity Sweep

**Calibration:** Pending after termination semantics.

## Milestone 10: Struct And Union ABI Edge Runtime Sweep

**Calibration:** Pending after long double/complex sweep.

## Milestone 11: GCC Runtime Fixture Expansion

**Calibration:** Pending after ABI sweeps.

## Milestone 12: Runtime Error And Diagnostics Stabilization

**Calibration:** Pending after fixture expansion.

## Milestone 13: CLI Runtime UX

**Calibration:** Pending after diagnostics stabilization.

## Milestone 14: Phase 3 Closure Docs

**Calibration:** Pending after all earlier milestones are implemented or closed.
