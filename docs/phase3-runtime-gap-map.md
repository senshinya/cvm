# Phase 3 Runtime Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-3`
Base: `main` after merge commit `34f4e1e`

## Calibration Summary

Phase 2 closed the runtime environment, memory-backed v-format `va_list`, hermetic file streams, and GNU nested-function closure/trampoline surfaces. Phase 3 should not reopen those as broad rewrites. The next work should connect existing isolated runtime capabilities into more faithful source-level C behavior.

## Current High-Value Gaps

1. Source-level `<stdarg.h>` still defines `va_arg(ap, type)` as `((type)0)`, even though the bytecode VM can execute `OpVaArg` and v-format externs can consume a CVM memory-backed `va_list` layout.
2. `va_copy` and multiple live `va_list` traversal semantics need coverage once source-level `va_arg` is no longer a placeholder.
3. Formatted input functions are not yet modeled as a bounded runtime surface.
4. Hermetic `FILE *` handles have useful read/write/seek behavior, but Phase 3 should calibrate remaining mode and state transitions before expanding them.
5. `getenv` currently returns null from a hermetic stub. Phase 3 can add configured environment variables without reading ambient host environment.
6. `atexit` currently accepts callbacks without executing them. Phase 3 can add deterministic callback execution at normal program termination.
7. Long double, complex, and aggregate ABI behavior has broad coverage, but Phase 3 should rescan for runtime gaps after varargs and formatted I/O improvements.
8. Runtime diagnostics should be audited after the larger hosted surfaces settle.
9. `cvm run` should expose only stable deterministic runtime knobs.

## First Milestone Decision

Milestone 2 remains the best first implementation target after baseline mapping: source-level `va_arg` is an obvious placeholder, and the runtime already has adjacent machinery. Work should start with integer and pointer reads through source-level `va_arg`, then expand only after focused tests pass.
