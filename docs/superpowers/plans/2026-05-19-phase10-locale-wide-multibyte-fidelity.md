# Phase 10 Locale Wide Multibyte Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend deterministic hosted locale, wide-character classification, and restartable multibyte conversion fidelity without depending on the host locale.

**Architecture:** Keep the runtime deterministic and C-locale oriented unless an explicit locale state is modeled. Add declarations in `preprocessor/headers.go`, semantic builtin support only where signatures need it, runtime externs in `runtime/extern.go`, direct tests in `runtime/extern_test.go`, source tests in `runtime/gcc_exec_test.go`, and closure docs.

**Tech Stack:** Go runtime externs, bytecode memory helpers, deterministic hosted registry state, GCC-style source execution tests.

---

## Milestone 1: Baseline And Branch Setup

Calibration before execution: Start from Phase 9 closure on `codex/bytecode-runtime-phase-10`.

- [x] Create/switch Phase 10 branch.
- [x] Re-read Phase 10 roadmap.
- [x] Run baseline verification.
- [x] Commit branch setup docs if needed.

Baseline:

- Created `codex/bytecode-runtime-phase-10` from Phase 9 closure commit `00fcdc6`.
- Phase 10 contains 23 milestones covering deterministic locale state, `localeconv`, wide ctype/wctrans surfaces, restartable multibyte conversion, rechecks, gap map, and closure docs.
- Baseline `git diff --check`, `go test ./codegen`, and `go test ./...` passed with `GOCACHE=/private/tmp/cvm-go-build-cache`.

## Milestone 2: Deterministic Locale Model

Calibration before execution: Re-read `setlocaleExtern` and Phase 8 locale residuals.

- [x] Document category state policy for deterministic C locale.
- [x] Decide storage and per-memory string reuse behavior.
- [x] Verify, commit, and push `docs: define phase 10 locale model`.

Findings:

- Current `setlocale` accepts supported categories, `NULL`, `"C"`, and `""`, returning deterministic `"C"` through per-memory static string storage.
- Phase 10 keeps all categories effectively in the C locale; unsupported categories or locale strings return `NULL`.
- Added `docs/phase10-locale-model.md` to anchor category behavior, per-memory storage, `localeconv`, wide ctype, and restartable multibyte follow-ons.

## Milestone 3: Category-Specific `setlocale` State

Calibration before execution: Check current `setlocale` always returns static `"C"`.

- [x] Add tests for each supported category query after setting `C` and `""`.
- [x] Store deterministic category state if needed.
- [x] Verify, commit, and push `feat(runtime): track deterministic locale categories`.

Findings:

- The deterministic model only has an effective C locale, so explicit per-category mutable storage is not required yet.
- Added direct extern coverage for all supported category values after setting `"C"` and `""`, including stable pointer reuse and unsupported-locale non-mutation.
- Added source-level runtime coverage for all named category macros.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestLocaleSetlocale' -count=1` passed.

## Milestone 4: `localeconv` Header And Registry

Calibration before execution: Check whether `struct lconv` and `localeconv` are declared.

- [x] Add deterministic C-locale `struct lconv` declaration.
- [x] Register `localeconv` runtime extern.
- [x] Add registry smoke.
- [x] Verify, commit, and push `feat(runtime): declare localeconv`.

Findings:

- Added a C99-style `struct lconv` declaration and `struct lconv *localeconv(void)` to the builtin `<locale.h>`.
- Registered `localeconv` in the default extern registry and added registry smoke coverage.
- Added locale header coverage for `struct lconv`, core string fields, `setlocale`, and `localeconv`.
- Focused preprocessor and registry tests passed.

## Milestone 5: `localeconv` Static Storage

Calibration before execution: Design the deterministic `struct lconv` memory layout for current target ABI.

- [x] Add direct tests for decimal point, thousands separator, and grouping fields.
- [x] Implement per-memory static `localeconv` storage.
- [x] Verify, commit, and push `feat(runtime): implement c localeconv`.

Findings:

- Implemented per-memory static `struct lconv` storage with 10 pointer fields followed by C-locale `char` fields, aligned to target pointer alignment.
- C-locale string fields use `"."` for `decimal_point` and `""` for grouping/separator/currency/sign strings.
- C-locale unavailable `char` fields are initialized to `CHAR_MAX` (`127`).
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestLocaleconvExtern -count=1` passed.

## Milestone 6: `localeconv` Source Runtime Coverage

Calibration before execution: Re-read source test helpers and C header layout.

- [x] Add source-level runtime test for `localeconv()->decimal_point`.
- [x] Add source-level runtime test for stable pointer behavior.
- [x] Verify, commit, and push `test(runtime): cover localeconv source`.

Findings:

- Added source-level runtime coverage for `localeconv` returning a non-null stable pointer.
- The source test validates `decimal_point`, empty `thousands_sep`, empty `grouping`, and `CHAR_MAX` monetary digit sentinel behavior through struct field access.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestLocaleconvExecuteThroughRuntime -count=1` passed.

## Milestone 7: Wide Ctype Header Surface

Calibration before execution: Check existing `<wctype.h>` and `<wchar.h>` declarations.

- [x] Add deterministic declarations for `wint_t`, `wctrans_t`, `wctype_t`, `isw*`, `towlower`, `towupper`, `wctype`, `iswctype`, `wctrans`, and `towctrans`.
- [x] Add header parse tests.
- [x] Verify, commit, and push `feat(headers): add wide ctype declarations`.

Findings:

- Added builtin `<wchar.h>` with guarded `wchar_t`, `wint_t`, and `WEOF`.
- Added builtin `<wctype.h>` with deterministic descriptor typedefs and wide classification/conversion declarations.
- Added header parse coverage for the full Phase 10 wide ctype declaration surface.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./preprocessor -run TestBuiltinWideHeadersDeclareRuntimeSurface -count=1` passed.

## Milestone 8: Wide Classification Core

Calibration before execution: Mirror byte-oriented ctype behavior for C-locale ASCII `wint_t`.

- [ ] Register and implement `iswdigit`, `iswalpha`, `iswalnum`, `iswspace`, `iswlower`, `iswupper`, `iswxdigit`, and `iswprint`.
- [ ] Add direct extern tests for EOF, ASCII, high values.
- [ ] Verify, commit, and push `feat(runtime): add wide ctype classification`.

## Milestone 9: Wide Extra Classification

Calibration before execution: Continue wide classification surface.

- [ ] Register and implement `iswblank`, `iswcntrl`, `iswgraph`, and `iswpunct`.
- [ ] Add direct extern and source-level tests.
- [ ] Verify, commit, and push `feat(runtime): add extra wide ctype classification`.

## Milestone 10: Wide Case Conversion

Calibration before execution: Mirror `tolower`/`toupper` EOF and ASCII behavior for `wint_t`.

- [ ] Implement `towlower` and `towupper`.
- [ ] Add direct extern tests for EOF, ASCII, punctuation, and high values.
- [ ] Add source-level tests.
- [ ] Verify, commit, and push `feat(runtime): add wide case conversion`.

## Milestone 11: `wctype` And `iswctype`

Calibration before execution: Decide deterministic descriptor IDs for named wide classes.

- [ ] Implement `wctype` for supported class names.
- [ ] Implement `iswctype`.
- [ ] Add direct/source tests for valid and invalid class names.
- [ ] Verify, commit, and push `feat(runtime): add wctype descriptors`.

## Milestone 12: `wctrans` And `towctrans`

Calibration before execution: Decide deterministic descriptor IDs for case mappings.

- [ ] Implement `wctrans` for `tolower` and `toupper`.
- [ ] Implement `towctrans`.
- [ ] Add direct/source tests for valid and invalid mappings.
- [ ] Verify, commit, and push `feat(runtime): add wctrans descriptors`.

## Milestone 13: Restartable State Type Baseline

Calibration before execution: Check whether `mbstate_t` exists in headers.

- [ ] Add deterministic `mbstate_t` declaration if missing.
- [ ] Add source compile/runtime smoke using zeroed state.
- [ ] Verify, commit, and push `feat(headers): add mbstate_t baseline`.

## Milestone 14: `mbrlen` C-Locale Behavior

Calibration before execution: Re-read existing `mblen` behavior and state reset policy.

- [ ] Implement/register `mbrlen`.
- [ ] Cover NULL/reset, zero-length, NUL, ASCII, and high-bit rejection.
- [ ] Verify, commit, and push `feat(runtime): add mbrlen c locale`.

## Milestone 15: `mbrtowc` C-Locale Behavior

Calibration before execution: Mirror `mbtowc` while accepting `mbstate_t *`.

- [ ] Implement/register `mbrtowc`.
- [ ] Cover NULL/reset, destination NULL, NUL, ASCII, high-bit rejection, and zero-length.
- [ ] Verify, commit, and push `feat(runtime): add mbrtowc c locale`.

## Milestone 16: `wcrtomb` C-Locale Behavior

Calibration before execution: Mirror `wctomb` while accepting `mbstate_t *`.

- [ ] Implement/register `wcrtomb`.
- [ ] Cover reset, ASCII, NUL, high wide char rejection, and invalid memory.
- [ ] Verify, commit, and push `feat(runtime): add wcrtomb c locale`.

## Milestone 17: `mbsrtowcs` C-Locale Behavior

Calibration before execution: Re-read `mbstowcs` and pointer-to-source update semantics.

- [ ] Implement/register `mbsrtowcs`.
- [ ] Cover length query, conversion, truncation, source pointer update, NUL completion, and high-bit rejection.
- [ ] Verify, commit, and push `feat(runtime): add mbsrtowcs c locale`.

## Milestone 18: `wcsrtombs` C-Locale Behavior

Calibration before execution: Re-read `wcstombs` and pointer-to-source update semantics.

- [ ] Implement/register `wcsrtombs`.
- [ ] Cover length query, conversion, truncation, source pointer update, NUL completion, and high wide char rejection.
- [ ] Verify, commit, and push `feat(runtime): add wcsrtombs c locale`.

## Milestone 19: Restartable Source Runtime Sweep

Calibration before execution: Search source tests for restartable multibyte coverage.

- [ ] Add a compact source-level runtime workflow covering `mbrlen`, `mbrtowc`, `wcrtomb`, `mbsrtowcs`, and `wcsrtombs`.
- [ ] Verify, commit, and push `test(runtime): cover restartable multibyte workflow`.

## Milestone 20: Header/Registry Recheck

Calibration before execution: Search locale/wide/multibyte touched surfaces across headers, registry, direct tests, and source tests.

- [ ] Record declaration/registration/test surface status.
- [ ] Add smoke coverage if missing.
- [ ] Verify, commit, and push `docs: record phase 10 header registry recheck`.

## Milestone 21: GCC Runtime Fixture Recheck

Calibration before execution: Run gap report and scan imported GCC accept roots for wide/multibyte/locale candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan fixture roots for newly unblocked candidates.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit, and push `docs: record phase 10 gcc fixture recheck`.

## Milestone 22: Phase 10 Residual Gap Map

Calibration before execution: Confirm all Phase 10 implementation milestones are committed and pushed.

- [ ] Create `docs/phase10-locale-wide-multibyte-fidelity-gap-map.md`.
- [ ] Record closed surfaces and residual deterministic limits.
- [ ] Verify, commit, and push `docs: map phase 10 locale wide gaps`.

## Milestone 23: Phase 10 Closure Docs

Calibration before execution: Confirm all prior Phase 10 milestones are committed and pushed.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Mark this plan complete.
- [ ] Run standard verification, commit `docs: close phase 10 locale wide work`, and push.
