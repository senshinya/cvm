# Phase 4 Hosted Runtime Stdio Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-4`
Base: closed Phase 3 tip `fef3f3b`

## Calibration Summary

Phase 3 closed runtime ABI fidelity and left two bounded hosted-runtime stdio/input surfaces: formatted input beyond integer/string/char conversions, and update-mode stream read/write sequencing. Phase 4 scoped those surfaces without reopening native libc, locale, multibyte, descriptor, or long-double precision work.

## Closed During Phase 4

- `scanf`, `fscanf`, and `sscanf` now share scanset support for `%[...]` and `%[^...]`, including ranges, leading `]`, width, assignment suppression, no-match behavior, and source-level runtime coverage.
- Formatted input now scans floating conversions for `%f`, `%F`, `%e`, `%E`, `%g`, `%G`, `%a`, and `%A`. Storage follows C `scanf` pointer rules: no length modifier stores `float`, `l` stores `double`, and `L` stores the runtime's current `long double` representation.
- Formatted input now scans `%p` as deterministic hexadecimal pointer text and stores pointer-sized runtime values.
- The scanner now distinguishes input failure from matching failure. Empty or whitespace-only first-conversion input returns `EOF`; nonmatching input returns `0`; later failures return the number of completed assignments.
- Hermetic update-mode streams now track read/write direction. Invalid read-after-write and write-after-read transitions set the stream error indicator and return short counts or EOF-like failures through existing stdio extern behavior.
- `fflush`, `fseek`, and `rewind` reset the update-mode direction barrier. Write-after-read is also allowed after a read reaches EOF.
- Cross-extern coverage confirms scansets and floating input work through stdin-backed `scanf`, file-backed `fscanf`, and string-backed `sscanf`, with unread input preserved.
- GCC runtime fixture recheck remains closed: 18 directive-based runnable manifest candidates pass, and imported GCC accept roots contain no newly unblocked scanf/scanset/float/pointer fixture to add.

## Current High-Value Gaps

No current Phase 4 implementation gap remains.

Further runtime work should be scoped as a new phase or review follow-up. Candidate Phase 5 surfaces, if needed, are:

- Native-like descriptor-backed file behavior rather than only hermetic configured streams.
- Locale-specific and multibyte/wide-character formatted input.
- A different long-double storage/arithmetic model if future fixtures require precision beyond the current binary64-backed representation.
- Explicit diagnostics or lifecycle checks for calling GNU nested-function closure pointers after their enclosing frame has returned.
