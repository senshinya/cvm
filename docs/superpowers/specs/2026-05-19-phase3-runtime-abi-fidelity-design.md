# Phase 3 Runtime ABI Fidelity Design

## Goal

Phase 3 extends the completed bytecode runtime work from bounded Phase 2 hosted-environment surfaces into broader hosted-runtime and ABI-fidelity behavior. The work remains deterministic and test-first: every milestone starts with a current-state calibration scan, then lands as small coherent increments with focused tests, full verification, commit, and push.

## Scope

Phase 3 is in scope for:

- unifying source-level varargs behavior with the memory-backed `va_list` runtime model
- improving multiple `va_list` and `va_copy` behavior
- adding bounded formatted input support
- hardening hermetic `FILE *` stream state and mode behavior
- adding deterministic configured environment support
- improving hosted program termination semantics
- sweeping long double, complex, aggregate ABI, and GCC runtime fixture gaps unlocked by the earlier milestones
- stabilizing runtime diagnostics and CLI knobs for supported deterministic runtime configuration

Phase 3 is not in scope for:

- direct host filesystem access from interpreted C programs
- native host ABI trampoline serialization
- fully conforming libc behavior in one pass
- target-specific ABI claims beyond the CVM bytecode/runtime model

## Architecture

The runtime remains a bytecode interpreter over `bytecode.Module`, `runtime.Program`, and `runtime.VM`. Hosted behavior continues to flow through `ExternRegistry` and deterministic test-configured state rather than ambient host state. New ABI-like features should prefer explicit bytecode/runtime representations and fixture-local helpers over opaque host behavior.

Each milestone must begin with a calibration note in the plan or roadmap. The note records whether the milestone is still missing, already complete, partially complete, or should be split. Implementation then follows the smallest viable slice that can be verified with focused tests and the common full verification commands.

## Milestones

1. Phase 3 Baseline and Gap Map
2. Source-Level `va_arg` and Memory `va_list` Unification
3. `va_copy`, Multiple `va_list`, and Nested Varargs
4. Formatted Input Runtime
5. FILE State Model Completion
6. Hermetic File Mode Expansion
7. Environment Runtime
8. Program Termination Semantics
9. Long Double and Complex Runtime Fidelity Sweep
10. Struct and Union ABI Edge Runtime Sweep
11. GCC Runtime Fixture Expansion
12. Runtime Error and Diagnostics Stabilization
13. CLI Runtime UX
14. Phase 3 Closure Docs

## Completion Criteria

Phase 3 is complete when all 14 milestones are either implemented or explicitly closed by calibration with a documented reason, the handoff and roadmap describe the final state, and these commands pass on the Phase 3 branch:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```
