# Phase 3 Runtime Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-3`
Base: `main` after merge commit `34f4e1e`

## Calibration Summary

Phase 2 closed the runtime environment, memory-backed v-format `va_list`, hermetic file streams, and GNU nested-function closure/trampoline surfaces. Phase 3 should not reopen those as broad rewrites. The next work should connect existing isolated runtime capabilities into more faithful source-level C behavior.

## Current High-Value Gaps

1. Formatted input functions have initial `sscanf`, `scanf`, and `fscanf` coverage, but scansets, floating conversions, pointer conversions, and EOF-edge semantics still need expansion.
2. Hermetic `FILE *` handles have useful read/write/seek behavior, but Phase 3 should calibrate remaining mode and state transitions before expanding them.
3. `getenv` currently returns null from a hermetic stub. Phase 3 can add configured environment variables without reading ambient host environment.
4. `atexit` currently accepts callbacks without executing them. Phase 3 can add deterministic callback execution at normal program termination.
5. Long double, complex, and aggregate ABI behavior has broad coverage, but Phase 3 should rescan for runtime gaps after varargs and formatted I/O improvements.
6. Runtime diagnostics should be audited after the larger hosted surfaces settle.
7. `cvm run` should expose only stable deterministic runtime knobs.

## Closed During Phase 3

- `va_arg(ap, type)` now lowers through `__builtin_va_arg` and executes VM `OpVaArg` for source-level integer and pointer reads.
- Multiple live source-level `va_list` cursors are independent because codegen emits slot-aware `OpVaArg`.
- `va_copy(dst, src)` now lowers through `__builtin_va_copy` and executes VM `OpVaCopy`.
- `sscanf` is declared by `<stdio.h>`, typed by sema, registered as a runtime extern, and handles bounded `%d`, `%i`, `%u`, `%s`, `%c`, and `%%` string input for the first formatted-input slice.
- `scanf` and `fscanf` now reuse the same scanner over stdin and hermetic `FILE *` handles, preserving unread stream input via runtime pushback.
- Formatted input now supports `%x`, `%X`, `%o`, assignment suppression, and `%n` count writes for integer scanning coverage.

## First Milestone Decision

Milestone 2 remains the best first implementation target after baseline mapping: source-level `va_arg` is an obvious placeholder, and the runtime already has adjacent machinery. Work should start with integer and pointer reads through source-level `va_arg`, then expand only after focused tests pass.
