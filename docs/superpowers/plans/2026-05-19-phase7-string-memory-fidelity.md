# Phase 7 String And Memory Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close deterministic hosted C string and memory helper edge cases in the bytecode runtime.

**Architecture:** Keep semantics inside `runtime/extern.go` and coverage in `runtime/extern_test.go` plus focused source-level tests in `runtime/gcc_exec_test.go`. Each milestone starts by re-reading the relevant extern and existing tests, then adds one small direct/source coverage increment and fixes behavior only where calibration exposes a real gap.

**Tech Stack:** Go runtime externs, bytecode memory helpers, GCC-style source execution tests, hermetic test cache via `GOCACHE=/private/tmp/cvm-go-build-cache`.

---

## Milestone 1: Baseline And Branch Setup

Calibration before execution: Confirm branch `codex/bytecode-runtime-phase-7` starts from clean Phase 6 closure and upstream is set.

- [x] Create Phase 7 and Phase 8 roadmap docs.
- [x] Run standard verification.
- [x] Commit and push `docs: define phase 7 and phase 8 roadmaps`.

## Milestone 2: `memchr` Zero-Length And Byte Masking

Calibration before execution: Re-read `memoryCharSearchExtern` and current `TestStringSearchAndSpanExterns`.

- [x] Add direct tests for zero-length no-read and `int` argument byte masking.
- [x] Fix `memchr` if the search byte is not masked to unsigned char.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden memchr byte semantics`.

## Milestone 3: `memcmp` Zero-Length And Byte Ordering

Calibration before execution: Re-read `memoryCompareExtern` and `TestMemoryCompareExtern`.

- [x] Add direct tests for zero-length equality and unsigned-byte ordering around `0xff`.
- [x] Fix comparison if it treats bytes as signed.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden memcmp byte semantics`.

## Milestone 4: `strncmp` Termination And Zero Count

Calibration before execution: Re-read `stringNCompareExtern`.

- [x] Add direct tests for count zero, terminator-before-count, and unsigned-byte ordering.
- [x] Fix `strncmp` if it reads past terminators or mishandles zero count.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden strncmp c locale`.

## Milestone 5: `strnlen` Boundary Semantics

Calibration before execution: Re-read `stringNLengthExtern` and existing no-NUL coverage.

- [x] Add direct tests for zero maxlen, early NUL, exact maxlen, and no-NUL buffer.
- [x] Fix `strnlen` if it reads beyond `maxlen`.
- [x] Verify, commit, and push `test(runtime): cover strnlen boundaries`.

## Milestone 6: `strchr` NUL And Byte Masking

Calibration before execution: Re-read `stringCharSearchExtern`.

- [ ] Add direct/source tests for searching `'\0'`, missing byte, and high `int` argument masking.
- [ ] Fix `strchr` if needed.
- [ ] Verify, commit, and push `feat(runtime): harden strchr byte semantics`.

## Milestone 7: `strrchr` Last Match Semantics

Calibration before execution: Re-read `stringReverseCharSearchExtern`.

- [ ] Add tests for repeated bytes, NUL terminator lookup, missing byte, and byte masking.
- [ ] Fix `strrchr` if needed.
- [ ] Verify, commit, and push `feat(runtime): harden strrchr byte semantics`.

## Milestone 8: `strstr` Empty Needle And Misses

Calibration before execution: Re-read `stringSearchExtern`.

- [ ] Add tests for empty needle returning haystack, full match, suffix match, and miss.
- [ ] Fix `strstr` if needed.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `test(runtime): cover strstr boundaries`.

## Milestone 9: `strpbrk` Empty Accept Set

Calibration before execution: Re-read `stringSetSearchExtern`.

- [ ] Add tests for empty accept set, first accept hit, later hit, and miss.
- [ ] Fix `strpbrk` if needed.
- [ ] Verify, commit, and push `test(runtime): cover strpbrk boundaries`.

## Milestone 10: `strspn` And `strcspn` Empty Sets

Calibration before execution: Re-read `stringSpanExtern`.

- [ ] Add direct/source tests for empty accept/reject sets and immediate hit/miss.
- [ ] Fix span behavior if needed.
- [ ] Verify, commit, and push `test(runtime): cover span edge cases`.

## Milestone 11: `strtok` Delimiter Mutation

Calibration before execution: Re-read `stringTokenExtern` and `strtokNext` memory-keyed state.

- [ ] Add direct tests where delimiter set changes between calls.
- [ ] Fix delimiter handling if it caches old delimiters.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `test(runtime): cover strtok delimiter changes`.

## Milestone 12: `strtok` Exhaustion And Restart

Calibration before execution: Re-check exhausted state and restart with a new string.

- [ ] Add tests for all-delimiter input, final null return, and restart with a new string.
- [ ] Fix restart state if needed.
- [ ] Verify, commit, and push `test(runtime): cover strtok restart`.

## Milestone 13: `strxfrm` Truncation And Length Query

Calibration before execution: Re-read `stringTransformExtern`.

- [ ] Add tests for `n == 0`, exact fit, truncation with terminator, and return length.
- [ ] Fix C-locale transform behavior if needed.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `feat(runtime): harden strxfrm c locale`.

## Milestone 14: `strcoll` C-Locale Ordering

Calibration before execution: Re-read `stringCollateExtern`.

- [ ] Add tests for equal strings, less-than, greater-than, and prefix ordering.
- [ ] Fix `strcoll` if needed.
- [ ] Verify, commit, and push `test(runtime): cover strcoll ordering`.

## Milestone 15: `strerror` Stable Static Strings

Calibration before execution: Re-read `stringErrorExtern` and static string allocation.

- [ ] Add tests for multiple errno values, repeated calls, and per-memory stable pointers.
- [ ] Fix static storage if needed.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `feat(runtime): harden strerror storage`.

## Milestone 16: `memccpy` Stop And Return Pointer

Calibration before execution: Re-read `memoryCharCopyExtern`.

- [ ] Add tests for hit returning byte-after-destination, miss returning NULL, and byte masking.
- [ ] Fix `memccpy` if needed.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `feat(runtime): harden memccpy semantics`.

## Milestone 17: `bcopy` Overlap Semantics

Calibration before execution: Re-read `memoryBcopyExtern`.

- [ ] Add direct/source tests for forward and backward overlap.
- [ ] Fix by using memmove-style temporary copy if needed.
- [ ] Verify, commit, and push `feat(runtime): harden bcopy overlap`.

## Milestone 18: `memset` And `bzero` Boundaries

Calibration before execution: Re-read `memorySetExtern`, checked memset, and `memoryBzeroExtern`.

- [ ] Add tests for zero length, high int byte masking, and exact bounds.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover memory fill boundaries`.

## Milestone 19: `strcpy` And `stpcpy` Return Values

Calibration before execution: Re-read `stringCopyExtern`.

- [ ] Add tests for copied terminator and `stpcpy` returning the terminator pointer.
- [ ] Add source-level runtime coverage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover strcpy returns`.

## Milestone 20: `strncpy` And `stpncpy` Padding

Calibration before execution: Re-read `stringNCopyExtern`.

- [ ] Add tests for padding, truncation without terminator, and `stpncpy` return pointer.
- [ ] Fix if needed.
- [ ] Add source-level runtime coverage.
- [ ] Verify, commit, and push `feat(runtime): harden strncpy padding`.

## Milestone 21: `strcat` And `strncat` Termination

Calibration before execution: Re-read `stringConcatExtern` and `stringNConcatExtern`.

- [ ] Add tests for return destination, appended NUL, zero count, and truncation count.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover strcat termination`.

## Milestone 22: Checked String And Memory Builtins

Calibration before execution: Re-read `*_chk` externs for copy, move, set, concat.

- [ ] Add tests for successful boundary-sized operations and failing object-size checks.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover checked string memory builtins`.

## Milestone 23: Header/Registry/GCC Recheck

Calibration before execution: Search touched `string.h`/memory surfaces across headers, registry, direct tests, and source tests.

- [ ] Run header/registry search for all Phase 7 surfaces.
- [ ] Run GCC runtime gap report.
- [ ] Scan imported GCC accept roots for newly unblocked string/memory candidates.
- [ ] Record findings.
- [ ] Verify, commit, and push `docs: record phase 7 fixture recheck`.

## Milestone 24: Phase 7 Closure Docs

Calibration before execution: Confirm all prior Phase 7 milestones are committed and pushed.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Create `docs/phase7-string-memory-fidelity-gap-map.md`.
- [ ] Mark this plan complete.
- [ ] Run standard verification, commit `docs: close phase 7 string memory work`, and push.
