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

- [x] Add tests for whitespace, plus sign, minus sign, and trailing text.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover atoi family parsing`.

Findings:

- Direct extern coverage already included whitespace, plus sign, minus sign, trailing text, and no-digit input across `atoi`, `atol`, and `atoll`.
- Added source-level runtime coverage for the `atoi` plus-sign path.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibAtoi' -count=1` passed.

## Milestone 6: `strtod` Decimal And Exponent Boundaries

Calibration before execution: Re-read `strtoFloatExtern`.

- [x] Add tests for leading whitespace, plus sign, exponent, endptr, and no conversion.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): harden strtod decimals`.

Findings:

- Existing coverage already included leading whitespace, negative decimal exponent input, end-pointer checks, and no-conversion behavior.
- Added direct extern and source-level runtime coverage for plus-sign decimal exponent parsing with a negative exponent.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 7: `strtod` Hex Float Boundaries

Calibration before execution: Check hex-float parsing through Go parser.

- [x] Add tests for `0x1p0`, fractional hex, exponent sign, and endptr.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): harden strtod hex floats`.

Findings:

- Existing coverage already included a positive-exponent hex float.
- Added direct extern and source-level runtime coverage for `0x1p0` and fractional hex with a negative exponent sign.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 8: `strtof` And `strtold` Storage

Calibration before execution: Confirm storage widths and long double approximation.

- [x] Add direct/source tests for `strtof` f32 and `strtold` current flong storage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover strtof strtold`.

Findings:

- Direct extern coverage already asserted `strtof` returns `TypeF32` and `strtold` returns current `TypeFLong`.
- Updated source-level runtime coverage to store `strtof` into a `float` local and `strtold` into a `long double` local before comparison.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibMoreFloatParser' -count=1` passed.

## Milestone 9: `atof` Wrapper Coverage

Calibration before execution: Re-read `atofExtern`.

- [x] Add tests for decimal, exponent, and no-conversion behavior.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover atof parsing`.

Findings:

- Existing `atof` wrapper coverage included decimal parsing.
- Added direct extern and source-level runtime coverage for exponent parsing and no-conversion fallback to `0.0`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 10: `rand` Sequence Determinism

Calibration before execution: Re-read `randExtern` and `srandExtern`.

- [x] Add tests for default seed, reseed repeatability, and value range.
- [x] Add source-level runtime coverage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover rand determinism`.

Findings:

- Existing direct/source coverage already checked reseed repeatability, non-repeating sequential values, and `RAND_MAX` range.
- Added direct extern and source-level runtime coverage proving default state matches `srand(1)`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibRand' -count=1` passed.

## Milestone 11: `getenv` Static Storage

Calibration before execution: Re-read `getenvExtern` and `ExternRegistry.SetEnv`.

- [x] Add tests for missing variables, configured values, stable pointers, and per-memory strings.
- [x] Add source-level runtime coverage with load options env.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): harden getenv storage`.

Findings:

- Existing direct/source coverage already checked missing variables, configured values, and stable repeated pointers in one memory.
- Added direct extern coverage for per-memory static string materialization through the same registry and source-level coverage that repeated `getenv` returns the same pointer.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlib.*Getenv' -count=1` passed.

## Milestone 12: `system` Hermetic Semantics

Calibration before execution: Re-read `systemExtern`.

- [x] Add tests for NULL command availability and non-NULL command failure.
- [x] Add source-level runtime coverage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover hermetic system`.

Findings:

- Existing direct extern and source-level runtime coverage already checked `system(NULL) == 0` and non-NULL commands returning `-1` without invoking a host shell.
- No runtime code changes were needed.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibSystem' -count=1` passed.

## Milestone 13: `setlocale` C Locale Queries

Calibration before execution: Re-read `setlocaleExtern`.

- [x] Add tests for query, `C`, empty locale, and unsupported locale.
- [x] Fix deterministic C-locale behavior if needed.
- [x] Verify, commit, and push `feat(runtime): harden setlocale c locale`.

Findings:

- Existing direct/source coverage already checked query, explicit `C`, and unsupported locale rejection.
- Added direct extern and source-level runtime coverage for empty locale `""` mapping to the deterministic C locale.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestLocaleSetlocale' -count=1` passed.

## Milestone 14: `clock` And `time` Determinism

Calibration before execution: Re-read `clockExtern`, `timeExtern`, and storage width.

- [x] Add tests for NULL time pointer, non-NULL store, invalid pointer, and `clock`.
- [x] Add source-level runtime coverage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover deterministic time`.

Findings:

- Existing direct/source coverage already checked deterministic `time(NULL)`, `time(&slot)`, stored `time_t`, and `clock()`.
- Added direct extern coverage for invalid non-NULL time pointer memory errors.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestTimeExterns|TestTimeHeader' -count=1` passed.

## Milestone 15: `difftime` Arithmetic

Calibration before execution: Re-read `difftimeExtern`.

- [x] Add tests for positive, zero, and negative differences.
- [x] Add source-level runtime coverage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover difftime arithmetic`.

Findings:

- Existing direct/source coverage checked positive differences.
- Added direct extern and source-level runtime coverage for zero and negative differences.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestTimeExterns|TestTimeHeader' -count=1` passed.

## Milestone 16: `atexit` Ordering Recheck

Calibration before execution: Re-read atexit registration and runtime exit handling.

- [x] Add source-level test for multiple callbacks and reverse order.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): recheck atexit ordering`.

Findings:

- Existing source-level runtime coverage registers two callbacks and asserts reverse-order stdout output `21`.
- No runtime code changes were needed.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibAtexitRunsHandlersInReverseOrderThroughRuntime' -count=1` passed.

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
