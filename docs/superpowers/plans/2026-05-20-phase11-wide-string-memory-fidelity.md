# Phase 11 Wide String And Memory Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add deterministic C-locale runtime support for the standard wide string and wide memory helpers declared by `<wchar.h>`.

**Architecture:** Phase 11 builds on Phase 10's `wchar_t`, `wint_t`, C-locale wide ctype, and restartable multibyte support. Implement narrow, deterministic helpers in `runtime/extern.go`, declare them in `preprocessor/headers.go`, and cover each coherent group through direct extern tests plus compact source-level runtime tests.

**Tech Stack:** Go runtime externs, builtin C headers, CVM bytecode runtime tests, GCC-style source execution fixtures.

---

## Milestone 1: Phase 11 Baseline Plan

Calibration before execution: Confirm `main` includes Phase 10 closure and no local branch drift.

- [ ] Create this plan and a Phase 12 plan with at least 20 milestones each.
- [ ] Run standard verification.
- [ ] Commit and push `docs: start phase 11 wide string work`.

## Milestone 2: `<wchar.h>` Wide String Declarations

Calibration before execution: Re-read current `wcharHeader` and avoid declarations that are not implemented in the same phase.

- [x] Probe whether declarations can be added ahead of runtime registration.
- [x] Record loader behavior and adjust Phase 11 execution to declaration-plus-implementation increments.
- [x] Run focused preprocessor test and full verification before continuing.

Findings:

- Adding all `<wchar.h>` wide string declarations ahead of runtime registration made existing source runtime tests fail during load with unresolved `wcslen`.
- Phase 11 must add declarations together with each matching runtime extern registration and implementation.
- Milestones 4 through 24 are now the declaration-plus-implementation source of truth; this broad declaration milestone remains a calibration checkpoint only.

## Milestone 3: Wide Helper Primitives

Calibration before execution: Re-read existing narrow string helper functions and memory load/store helpers.

- [x] Add private helpers for reading 32-bit `wchar_t` values and computing bounded wide-element addresses.
- [x] Cover helpers indirectly through first extern tests.
- [x] Run focused runtime test, standard verification, commit with first consumer, and push.

Findings:

- The first primitive slice is deliberately small: `wideElementAddr` and `loadWideChar` cover bounded wide-array reads for `wmemchr`.
- Write, scan, compare, and copy helpers will be introduced with the first extern that needs each behavior, keeping declarations and implementations synchronized.

## Milestone 4: `wmemchr`

Calibration before execution: Compare `memchr` byte masking with full wide-value matching.

- [x] Register and implement `wmemchr`.
- [x] Cover hit, miss, zero-count invalid pointer, NUL, and high wide value.
- [x] Verify, commit `feat(runtime): add wmemchr`, and push.

## Milestone 5: `wmemcmp`

Calibration before execution: Confirm return sign conventions match existing `memcmp`.

- [x] Register and implement `wmemcmp`.
- [x] Cover equal, less, greater, zero count, and high-value ordering.
- [x] Verify, commit `feat(runtime): add wmemcmp`, and push.

## Milestone 6: `wmemcpy`

Calibration before execution: Re-read `memcpy` behavior and pointer return typing.

- [x] Register and implement `wmemcpy`.
- [x] Cover copy, zero-count invalid pointers, and destination return value.
- [x] Verify, commit `feat(runtime): add wmemcpy`, and push.

## Milestone 7: `wmemmove`

Calibration before execution: Re-read overlap handling in `memmove`.

- [x] Register and implement `wmemmove`.
- [x] Cover forward overlap, backward overlap, zero-count invalid pointers, and return value.
- [x] Verify, commit `feat(runtime): add wmemmove`, and push.

## Milestone 8: `wmemset`

Calibration before execution: Confirm high wide values should be stored unmasked as `wchar_t`.

- [x] Register and implement `wmemset`.
- [x] Cover fill, zero count, high wide value, and return value.
- [x] Verify, commit `feat(runtime): add wmemset`, and push.

## Milestone 9: `wcslen`

Calibration before execution: Confirm C runtime uses 32-bit `wchar_t` in the target model.

- [x] Register and implement `wcslen`.
- [x] Cover empty, non-empty, and source-level `L"..."` execution.
- [x] Verify, commit `feat(runtime): add wcslen`, and push.

## Milestone 10: `wcscmp`

Calibration before execution: Reuse wide compare primitives and keep result sign stable.

- [x] Register and implement `wcscmp`.
- [x] Cover equal, less, greater, prefix, empty, and high wide values.
- [x] Verify, commit `feat(runtime): add wcscmp`, and push.

## Milestone 11: `wcsncmp`

Calibration before execution: Re-read `strncmp` zero-count and terminator-before-count behavior.

- [x] Register and implement `wcsncmp`.
- [x] Cover zero-count invalid pointers, equal prefix, less, greater, early terminator, and high wide values.
- [x] Verify, commit `feat(runtime): add wcsncmp`, and push.

## Milestone 12: `wcschr`

Calibration before execution: Keep `WEOF` and full wide values unmasked.

- [x] Register and implement `wcschr`.
- [x] Cover hit, miss, NUL terminator, and high wide value.
- [x] Verify, commit `feat(runtime): add wcschr`, and push.

## Milestone 13: `wcsrchr`

Calibration before execution: Mirror `strrchr` but advance in 4-byte `wchar_t` units.

- [x] Register and implement `wcsrchr`.
- [x] Cover last hit, miss, NUL terminator, repeated values, and high wide value.
- [x] Verify, commit `feat(runtime): add wcsrchr`, and push.

## Milestone 14: `wcsstr`

Calibration before execution: Re-read `strstr` empty needle behavior.

- [ ] Register and implement `wcsstr`.
- [ ] Cover empty needle, full match, suffix, middle, miss, and high wide value match.
- [ ] Verify, commit `feat(runtime): add wcsstr`, and push.

## Milestone 15: `wcspbrk`

Calibration before execution: Re-read `strpbrk` set-scan behavior and wide terminators.

- [ ] Register and implement `wcspbrk`.
- [ ] Cover hit, miss, empty accept set, NUL handling, and high wide values.
- [ ] Verify, commit `feat(runtime): add wcspbrk`, and push.

## Milestone 16: `wcsspn`

Calibration before execution: Re-read `strspn` accept-span semantics.

- [ ] Register and implement `wcsspn`.
- [ ] Cover full span, partial span, empty accept, empty string, and high wide values.
- [ ] Verify, commit `feat(runtime): add wcsspn`, and push.

## Milestone 17: `wcscspn`

Calibration before execution: Re-read `strcspn` reject-span semantics.

- [ ] Register and implement `wcscspn`.
- [ ] Cover full span, early reject, empty reject, empty string, and high wide values.
- [ ] Verify, commit `feat(runtime): add wcscspn`, and push.

## Milestone 18: `wcscpy`

Calibration before execution: Mirror `strcpy` including terminator copy and destination return.

- [ ] Register and implement `wcscpy`.
- [ ] Cover copy, empty string, high wide value, terminator, and return pointer.
- [ ] Verify, commit `feat(runtime): add wcscpy`, and push.

## Milestone 19: `wcsncpy`

Calibration before execution: Re-read `strncpy` padding/truncation behavior.

- [ ] Register and implement `wcsncpy`.
- [ ] Cover truncation, zero padding, zero count, high wide value, and return pointer.
- [ ] Verify, commit `feat(runtime): add wcsncpy`, and push.

## Milestone 20: `wcscat`

Calibration before execution: Mirror `strcat` destination scan and terminator write.

- [ ] Register and implement `wcscat`.
- [ ] Cover append, empty source, high wide value, terminator, and return pointer.
- [ ] Verify, commit `feat(runtime): add wcscat`, and push.

## Milestone 21: `wcsncat`

Calibration before execution: Re-read `strncat` terminator and zero-count behavior.

- [ ] Register and implement `wcsncat`.
- [ ] Cover bounded append, zero count, truncation terminator, high wide value, and return pointer.
- [ ] Verify, commit `feat(runtime): add wcsncat`, and push.

## Milestone 22: `wcscoll`

Calibration before execution: C locale collation is deterministic code-point ordering.

- [ ] Register and implement `wcscoll` as C-locale `wcscmp`.
- [ ] Cover source-level execution and direct extern comparisons.
- [ ] Verify, commit `feat(runtime): add wcscoll`, and push.

## Milestone 23: `wcsxfrm`

Calibration before execution: Mirror C-locale `strxfrm` with wide-character destination semantics.

- [ ] Register and implement `wcsxfrm`.
- [ ] Cover length query, bounded copy with terminator, truncation, zero count, and return length.
- [ ] Verify, commit `feat(runtime): add wcsxfrm`, and push.

## Milestone 24: `wcstok`

Calibration before execution: Re-read `strtok` and C99 `wcstok` save-pointer semantics.

- [ ] Register and implement `wcstok`.
- [ ] Cover first token, repeated delimiters, delimiter changes, exhaustion, empty delimiter set, and save-pointer updates.
- [ ] Verify, commit `feat(runtime): add wcstok`, and push.

## Milestone 25: Wide Source Runtime Sweep

Calibration before execution: Search existing source runtime tests for `<wchar.h>` workflows before adding more.

- [ ] Add compact source-level workflows covering wide memory, search/span, copy/concat, collation/transform, and tokenizer behavior.
- [ ] Run focused runtime source tests.
- [ ] Verify, commit `test(runtime): cover wide string workflows`, and push.

## Milestone 26: GCC Fixture Recheck

Calibration before execution: Run GCC runtime gap report and scan imported accept roots for wide-string candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan imported GCC accept fixtures for `wcs`, `wmem`, `wcstok`, `wcscoll`, and `wcsxfrm`.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit `docs: record phase 11 gcc fixture recheck`, and push.

## Milestone 27: Phase 11 Gap Map

Calibration before execution: Confirm all implementation milestones are committed and pushed.

- [ ] Create `docs/phase11-wide-string-memory-fidelity-gap-map.md`.
- [ ] Record closed wide string/memory surfaces and residual deterministic limits.
- [ ] Verify, commit `docs: map phase 11 wide string gaps`, and push.

## Milestone 28: Phase 11 Closure

Calibration before execution: Confirm branch, remote, and final verification status.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Mark this plan complete.
- [ ] Verify, commit `docs: close phase 11 wide string work`, and push.
