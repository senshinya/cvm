# Phase 3 Runtime Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-3`
Base: `main` after merge commit `34f4e1e`

## Calibration Summary

Phase 2 closed the runtime environment, memory-backed v-format `va_list`, hermetic file streams, and GNU nested-function closure/trampoline surfaces. Phase 3 should not reopen those as broad rewrites. The next work should connect existing isolated runtime capabilities into more faithful source-level C behavior.

## Current High-Value Gaps

1. Long double, complex, and aggregate ABI behavior has broad coverage, but Phase 3 should rescan for runtime gaps after varargs and formatted I/O improvements.
2. Runtime diagnostics should be audited after the larger hosted surfaces settle.
3. `cvm run` should expose only stable deterministic runtime knobs.

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
- Append-mode files now force writes to the current end of the hermetic file even after `fseek`, matching `a`/`a+` write positioning.
- The hermetic file mode milestone is closed for C99 `r`, `w`, `a`, and `+` behavior. Strict read/write sequencing rules for update streams remain a documented residual limit.
- `getenv` now reads only explicit `ExternRegistry.SetEnv` values and remains isolated from the ambient host environment by default.
- `atexit` now records registered callbacks and the VM runs them in reverse registration order after normal `main` return or `exit`. `_Exit` remains immediate and skips registered callbacks.
- The long double/complex ABI sweep found and fixed unsuffixed variadic coverage around floating literal suffixes: `long double` literals now keep `LongDouble` type before default argument promotion, so `va_arg(ap, long double)` receives `flong`. Source-level `va_arg` coverage now includes `long double` and `_Complex double`.
- The struct/union ABI sweep found existing object-address vararg handling sufficient for `va_arg(ap, struct S)` and `va_arg(ap, union U)`, and added runtime coverage for both cases.

## Residual Bounded Runtime Surface

- Formatted input intentionally remains bounded to integer/string/char conversions. Scansets, floating input, pointer input, and exact EOF corner semantics are deferred until a fixture or user workflow needs them.
- Update-mode streams do not enforce the C sequencing rule requiring a flush or positioning operation between certain read/write direction changes.

## First Milestone Decision

Milestone 2 remains the best first implementation target after baseline mapping: source-level `va_arg` is an obvious placeholder, and the runtime already has adjacent machinery. Work should start with integer and pointer reads through source-level `va_arg`, then expand only after focused tests pass.
