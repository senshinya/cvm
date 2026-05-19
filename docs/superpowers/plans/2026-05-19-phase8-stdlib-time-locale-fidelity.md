# Phase 8 Stdlib Time Locale Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close deterministic hosted stdlib, conversion, time, environment, locale, and process-control runtime fidelity gaps.

**Architecture:** Preserve hermetic deterministic behavior: no host shell execution, no host clock dependency, no ambient environment. Strengthen parsing and return semantics in `runtime/extern.go`, direct coverage in `runtime/extern_test.go`, source-level coverage in `runtime/gcc_exec_test.go`, and closure docs.

**Tech Stack:** Go runtime externs, bytecode memory helpers, deterministic hosted registry state, GCC-style source execution tests.

---

## Milestone 1: Baseline And Branch Setup

Calibration before execution: Start from Phase 7 closure on `codex/bytecode-runtime-phase-8`.

- [x] Create/switch Phase 8 branch.
- [x] Re-read Phase 8 roadmap.
- [x] Run baseline verification.
- [x] Commit branch setup docs if needed.

Baseline:

- Created `codex/bytecode-runtime-phase-8` from Phase 7 closure commit `ea28f54`.
- Phase 8 contains 22 milestones covering stdlib conversion, random/env/process, locale/time, ctype, registry recheck, GCC fixture recheck, and closure docs.
- No runtime code changes were needed for branch setup.

## Milestone 2: `strtol` Base Auto-Detection

Calibration before execution: Re-read `strtoIntegerExtern` and existing `strtol` tests.

- [x] Add direct/source tests for base 0 decimal, octal, hex, and sign handling.
- [x] Fix parsing if needed.
- [x] Verify, commit, and push `feat(runtime): harden strtol base detection`.

Findings:

- Existing parsing already handled base 0 decimal, octal, hex, whitespace, and sign routing.
- Added direct extern and source-level runtime coverage for signed decimal plus-sign input and signed octal auto-detection.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibStrtol' -count=1` passed.

## Milestone 3: `strtoul` Unsigned Boundaries

Calibration before execution: Check unsigned parsing and end-pointer behavior.

- [x] Add tests for hex, octal, no conversion, and negative input conversion.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): harden strtoul semantics`.

Findings:

- Existing `strtoul` coverage already included octal auto-detection and explicit-base hex conversion.
- Added direct extern and source-level runtime coverage for negative unsigned conversion and no-conversion `endptr` behavior.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibStrtol' -count=1` passed.

## Milestone 4: `strtoll` And `strtoull` 64-Bit Edges

Calibration before execution: Check 64-bit signed/unsigned wrapping behavior.

- [x] Add direct/source tests for max signed, high unsigned, and endptr.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): harden strtoll variants`.

Findings:

- Existing coverage included negative signed 64-bit parsing and unsigned values above 32-bit range.
- Added direct extern and source-level runtime coverage for `LLONG_MAX`-sized decimal input and `ULLONG_MAX`-sized decimal input with end-pointer checks.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibStrtoll' -count=1` passed.

## Milestone 5: `atoi` Family Whitespace And Sign

Calibration before execution: Re-read `atoiExtern`.

- [ ] Add tests for whitespace, plus sign, minus sign, and trailing text.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover atoi family parsing`.

## Milestone 6: `strtod` Decimal And Exponent Boundaries

Calibration before execution: Re-read `strtoFloatExtern`.

- [ ] Add tests for leading whitespace, plus sign, exponent, endptr, and no conversion.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `feat(runtime): harden strtod decimals`.

## Milestone 7: `strtod` Hex Float Boundaries

Calibration before execution: Check hex-float parsing through Go parser.

- [ ] Add tests for `0x1p0`, fractional hex, exponent sign, and endptr.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `feat(runtime): harden strtod hex floats`.

## Milestone 8: `strtof` And `strtold` Storage

Calibration before execution: Confirm storage widths and long double approximation.

- [ ] Add direct/source tests for `strtof` f32 and `strtold` current flong storage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover strtof strtold`.

## Milestone 9: `atof` Wrapper Coverage

Calibration before execution: Re-read `atofExtern`.

- [ ] Add tests for decimal, exponent, and no-conversion behavior.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover atof parsing`.

## Milestone 10: `rand` Sequence Determinism

Calibration before execution: Re-read `randExtern` and `srandExtern`.

- [ ] Add tests for default seed, reseed repeatability, and value range.
- [ ] Add source-level runtime coverage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover rand determinism`.

## Milestone 11: `getenv` Static Storage

Calibration before execution: Re-read `getenvExtern` and `ExternRegistry.SetEnv`.

- [ ] Add tests for missing variables, configured values, stable pointers, and per-memory strings.
- [ ] Add source-level runtime coverage with load options env.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `feat(runtime): harden getenv storage`.

## Milestone 12: `system` Hermetic Semantics

Calibration before execution: Re-read `systemExtern`.

- [ ] Add tests for NULL command availability and non-NULL command failure.
- [ ] Add source-level runtime coverage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover hermetic system`.

## Milestone 13: `setlocale` C Locale Queries

Calibration before execution: Re-read `setlocaleExtern`.

- [ ] Add tests for query, `C`, empty locale, and unsupported locale.
- [ ] Fix deterministic C-locale behavior if needed.
- [ ] Verify, commit, and push `feat(runtime): harden setlocale c locale`.

## Milestone 14: `clock` And `time` Determinism

Calibration before execution: Re-read `clockExtern`, `timeExtern`, and storage width.

- [ ] Add tests for NULL time pointer, non-NULL store, invalid pointer, and `clock`.
- [ ] Add source-level runtime coverage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover deterministic time`.

## Milestone 15: `difftime` Arithmetic

Calibration before execution: Re-read `difftimeExtern`.

- [ ] Add tests for positive, zero, and negative differences.
- [ ] Add source-level runtime coverage.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover difftime arithmetic`.

## Milestone 16: `atexit` Ordering Recheck

Calibration before execution: Re-read atexit registration and runtime exit handling.

- [ ] Add source-level test for multiple callbacks and reverse order.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): recheck atexit ordering`.

## Milestone 17: `exit` And `_Exit` Divergence

Calibration before execution: Confirm `_Exit` skips handlers while `exit` runs them.

- [ ] Add source-level coverage for `_Exit` skip and `exit` run behavior.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover exit handler split`.

## Milestone 18: Ctype Classification Bounds

Calibration before execution: Re-read ctype classification externs.

- [ ] Add tests for EOF `-1`, ASCII letters/digits/space, and high int masking.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `feat(runtime): harden ctype classification`.

## Milestone 19: Ctype Case Conversion Bounds

Calibration before execution: Re-read `tolower`/`toupper` externs.

- [ ] Add tests for EOF passthrough, lower-to-upper, upper-to-lower, punctuation, and high int masking.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `feat(runtime): harden ctype case conversion`.

## Milestone 20: Header/Registry Recheck

Calibration before execution: Search stdlib/time/locale/ctype touched surfaces across headers, registry, and tests.

- [ ] Record declaration/registration/test surface status.
- [ ] Add smoke coverage if missing.
- [ ] Verify, commit, and push `docs: record phase 8 header registry recheck`.

## Milestone 21: GCC Runtime Fixture Recheck

Calibration before execution: Re-run runtime gap report and scan imported GCC accept roots for conversion/time/locale/env candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan fixture roots for newly unblocked candidates.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit, and push `docs: record phase 8 gcc fixture recheck`.

## Milestone 22: Phase 8 Closure Docs

Calibration before execution: Confirm all prior Phase 8 milestones are committed and pushed.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Create `docs/phase8-stdlib-time-locale-fidelity-gap-map.md`.
- [ ] Mark this plan complete.
- [ ] Run standard verification, commit `docs: close phase 8 stdlib time locale work`, and push.
