# Phase 4 Hosted Runtime Stdio Design

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-4`

## Goal

Phase 4 closes the remaining bounded hosted-runtime stdio/input surfaces left after Phase 3: broader formatted input coverage and C update-mode stream sequencing for hermetic streams.

## Scope

This phase stays deliberately narrow. It extends the existing deterministic runtime model instead of introducing host filesystem access, native libc binding, locale-dependent parsing, or a new floating-point representation.

Included:

- `scanf`/`fscanf`/`sscanf` scansets (`%[...]` and `%[^...]`) with width and assignment suppression.
- Floating formatted input for `%f`, `%e`, `%g`, and `%a` families, including length modifiers for `float *`, `double *`, and `long double *` storage under the current binary64-backed runtime model.
- Pointer formatted input for `%p`.
- More faithful formatted-input failure and EOF return behavior for empty input and first-conversion failure.
- Update-mode stream direction tracking for hermetic `+` streams, requiring a positioning or flush operation between read-after-write and write-after-read transitions, with the EOF exception for write-after-read.
- GCC-style runtime snippets that exercise the new behavior through compiled C, not only direct extern calls.
- Updated handoff/gap docs that mark Phase 4 closed and leave any remaining work as Phase 5 or review follow-up.

Excluded:

- Locale-specific scans, multibyte/wide-character input, and native file descriptors.
- Exact long-double precision beyond the runtime's current binary64-backed `long double` representation.
- Capturing nested-function lifetime changes.

## Architecture

The formatted-input scanner remains a single shared implementation behind `scanf`, `fscanf`, and `sscanf`. New conversion families should be added by small scanner helpers rather than by duplicating per-extern logic. Stream-backed scans continue to read into a bounded string, run the scanner, then push unread bytes back to the stream.

Hermetic streams gain a small direction state. Read and write helpers consult this state before touching stream contents. `fseek`, `rewind`, and `fflush` reset the transition barrier. Invalid update-mode transitions return normal C-style short counts or EOF-like failures and set the stream error indicator, matching the runtime's existing error-indicator approach.

## Testing

Each milestone starts with a calibration scan of current code and tests. Each implementation increment adds a direct runtime test and, when source-level syntax is relevant, a GCC-style runtime execution test. Every coherent increment runs:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Each coherent increment is committed and pushed before moving to the next milestone.
