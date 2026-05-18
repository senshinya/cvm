# Phase 3 Runtime Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-3`
Base: `main` after merge commit `34f4e1e`

## Calibration Summary

Phase 2 closed the runtime environment, memory-backed v-format `va_list`, hermetic file streams, and GNU nested-function closure/trampoline surfaces. Phase 3 should not reopen those as broad rewrites. The next work should connect existing isolated runtime capabilities into more faithful source-level C behavior.

## Current High-Value Gaps

1. Hermetic file mode handling should expand append/update behavior now that `FILE *` state bits are modeled.
2. `getenv` currently returns null from a hermetic stub. Phase 3 can add configured environment variables without reading ambient host environment.
3. `atexit` currently accepts callbacks without executing them. Phase 3 can add deterministic callback execution at normal program termination.
4. Long double, complex, and aggregate ABI behavior has broad coverage, but Phase 3 should rescan for runtime gaps after varargs and formatted I/O improvements.
5. Runtime diagnostics should be audited after the larger hosted surfaces settle.
6. `cvm run` should expose only stable deterministic runtime knobs.

## Closed During Phase 3

- `va_arg(ap, type)` now lowers through `__builtin_va_arg` and executes VM `OpVaArg` for source-level integer and pointer reads.
- Multiple live source-level `va_list` cursors are independent because codegen emits slot-aware `OpVaArg`.
- `va_copy(dst, src)` now lowers through `__builtin_va_copy` and executes VM `OpVaCopy`.
- `sscanf` is declared by `<stdio.h>`, typed by sema, registered as a runtime extern, and handles bounded `%d`, `%i`, `%u`, `%s`, `%c`, and `%%` string input for the first formatted-input slice.
- `scanf` and `fscanf` now reuse the same scanner over stdin and hermetic `FILE *` handles, preserving unread stream input via runtime pushback.
- Formatted input now supports `%x`, `%X`, `%o`, assignment suppression, and `%n` count writes for integer scanning coverage.
- `FILE *` write failures now set a stream error indicator instead of trapping; `ferror` observes it and `clearerr` clears it.
- `FILE *` read failures on write-only streams now set the same error indicator without falsely setting EOF.
- The Phase 3 `FILE *` state milestone is closed for EOF/error/clear/close state; remaining mode semantics move to the hermetic file mode milestone.

## Residual Bounded Runtime Surface

- Formatted input intentionally remains bounded to integer/string/char conversions. Scansets, floating input, pointer input, and exact EOF corner semantics are deferred until a fixture or user workflow needs them.

## First Milestone Decision

Milestone 2 remains the best first implementation target after baseline mapping: source-level `va_arg` is an obvious placeholder, and the runtime already has adjacent machinery. Work should start with integer and pointer reads through source-level `va_arg`, then expand only after focused tests pass.
