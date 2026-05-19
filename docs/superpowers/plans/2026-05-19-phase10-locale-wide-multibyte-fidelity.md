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

- [x] Register and implement `iswdigit`, `iswalpha`, `iswalnum`, `iswspace`, `iswlower`, `iswupper`, `iswxdigit`, and `iswprint`.
- [x] Add direct extern tests for EOF, ASCII, high values.
- [x] Verify, commit, and push `feat(runtime): add wide ctype classification`.

Findings:

- Added wide classification externs for the core C-locale ASCII classes.
- Wide classification uses the full `wint_t` value and does not mask high values down to unsigned byte.
- Added direct extern tests for ASCII positives/negatives, `WEOF`/`-1`, and high non-ASCII values.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestDefaultExternRegistryHasExitAndAbort|TestWideCtypeClassificationExterns' -count=1` passed.

## Milestone 9: Wide Extra Classification

Calibration before execution: Continue wide classification surface.

- [x] Register and implement `iswblank`, `iswcntrl`, `iswgraph`, and `iswpunct`.
- [x] Add direct extern and source-level tests.
- [x] Verify, commit, and push `feat(runtime): add extra wide ctype classification`.

Findings:

- Added wide extra classification externs for blank, control, graph, and punctuation classes.
- Added direct extern and source-level runtime coverage for those classes, including high-value non-ASCII rejection.
- Source coverage exposed that including `<wctype.h>` requires all declared externs to be registered at load time, so wide case and descriptor externs were registered with deterministic implementations ahead of their dedicated test milestones.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestWideCtype.*Classification|TestDefaultExternRegistryHasExitAndAbort' -count=1` passed.

## Milestone 10: Wide Case Conversion

Calibration before execution: Mirror `tolower`/`toupper` EOF and ASCII behavior for `wint_t`.

- [x] Implement `towlower` and `towupper`.
- [x] Add direct extern tests for EOF, ASCII, punctuation, and high values.
- [x] Add source-level tests.
- [x] Verify, commit, and push `feat(runtime): add wide case conversion`.

Findings:

- `towlower` and `towupper` use deterministic C-locale ASCII mappings over full `wint_t` values.
- `WEOF` is preserved and high non-ASCII values are returned unchanged.
- Added direct extern and source-level runtime coverage for ASCII conversion, punctuation, `WEOF`, and high values.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestWideCtypeCaseConversion' -count=1` passed.

## Milestone 11: `wctype` And `iswctype`

Calibration before execution: Decide deterministic descriptor IDs for named wide classes.

- [x] Implement `wctype` for supported class names.
- [x] Implement `iswctype`.
- [x] Add direct/source tests for valid and invalid class names.
- [x] Verify, commit, and push `feat(runtime): add wctype descriptors`.

Findings:

- `wctype` returns deterministic nonzero descriptor IDs for supported C-locale class names and zero for unknown names.
- `iswctype` applies descriptors over full `wint_t` values, preserving high-value non-ASCII rejection and `WEOF` false behavior.
- Added direct extern and source-level runtime coverage for valid descriptors, invalid names, invalid descriptors, ASCII positives/negatives, high values, and `WEOF`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestWideCtypeDescriptor' -count=1` passed.

## Milestone 12: `wctrans` And `towctrans`

Calibration before execution: Decide deterministic descriptor IDs for case mappings.

- [x] Implement `wctrans` for `tolower` and `toupper`.
- [x] Implement `towctrans`.
- [x] Add direct/source tests for valid and invalid mappings.
- [x] Verify, commit, and push `feat(runtime): add wctrans descriptors`.

Findings:

- `wctrans` returns deterministic nonzero descriptor IDs for `tolower` and `toupper`, and zero for unknown mappings.
- `towctrans` applies ASCII-only C-locale mappings, preserves `WEOF`, keeps high values unchanged, and leaves input unchanged for invalid descriptors.
- Added direct extern and source-level runtime coverage for valid and invalid mappings.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestWideCtypeTransDescriptor' -count=1` passed.

## Milestone 13: Restartable State Type Baseline

Calibration before execution: Check whether `mbstate_t` exists in headers.

- [x] Add deterministic `mbstate_t` declaration if missing.
- [x] Add source compile/runtime smoke using zeroed state.
- [x] Verify, commit, and push `feat(headers): add mbstate_t baseline`.

Findings:

- Added guarded `mbstate_t` to builtin `<wchar.h>` as a deterministic opaque-compatible state struct.
- Extended wide header parse coverage to include `mbstate_t`.
- Added source-level runtime smoke coverage for a zero-initialized `mbstate_t`.
- Focused preprocessor and runtime tests passed.

## Milestone 14: `mbrlen` C-Locale Behavior

Calibration before execution: Re-read existing `mblen` behavior and state reset policy.

- [x] Implement/register `mbrlen`.
- [x] Cover NULL/reset, zero-length, NUL, ASCII, and high-bit rejection.
- [x] Verify, commit, and push `feat(runtime): add mbrlen c locale`.

Findings:

- Added `mbrlen` declaration to builtin `<wchar.h>` and registered the extern.
- C-locale `mbrlen` is stateless: NULL/reset returns `0`, NUL returns `0`, ASCII returns `1`, zero-length non-NULL input returns `(size_t)-2`, and high-bit input returns `(size_t)-1`.
- Added header, registry, and direct extern coverage.
- Focused preprocessor and runtime tests passed.

## Milestone 15: `mbrtowc` C-Locale Behavior

Calibration before execution: Mirror `mbtowc` while accepting `mbstate_t *`.

- [x] Implement/register `mbrtowc`.
- [x] Cover NULL/reset, destination NULL, NUL, ASCII, high-bit rejection, and zero-length.
- [x] Verify, commit, and push `feat(runtime): add mbrtowc c locale`.

Findings:

- Added `mbrtowc` declaration to builtin `<wchar.h>` and registered the extern.
- C-locale `mbrtowc` is stateless: NULL source reset returns `0`, NUL returns `0`, ASCII stores the wide char and returns `1`, destination NULL skips storage, zero-length non-NULL input returns `(size_t)-2`, and high-bit input returns `(size_t)-1` without modifying the destination.
- Focused preprocessor and runtime tests passed.

## Milestone 16: `wcrtomb` C-Locale Behavior

Calibration before execution: Mirror `wctomb` while accepting `mbstate_t *`.

- [x] Implement/register `wcrtomb`.
- [x] Cover reset, ASCII, NUL, high wide char rejection, and invalid memory.
- [x] Verify, commit, and push `feat(runtime): add wcrtomb c locale`.

Findings:

- Added `wcrtomb` declaration to builtin `<wchar.h>` and registered the extern.
- C-locale `wcrtomb` returns `1` for reset, ASCII, and NUL output, writes one byte when a destination is present, and returns `(size_t)-1` for invalid high wide characters without modifying the destination.
- Focused preprocessor and runtime tests passed.

## Milestone 17: `mbsrtowcs` C-Locale Behavior

Calibration before execution: Re-read `mbstowcs` and pointer-to-source update semantics.

- [x] Implement/register `mbsrtowcs`.
- [x] Cover length query, conversion, truncation, source pointer update, NUL completion, and high-bit rejection.
- [x] Verify, commit, and push `feat(runtime): add mbsrtowcs c locale`.

Findings:

- Added `mbsrtowcs` declaration to builtin `<wchar.h>` and registered the extern.
- C-locale implementation converts ASCII bytes to 32-bit wide chars, sets `*src = NULL` on NUL completion, updates `*src` after truncation, leaves `*src` unchanged for length queries, and returns `(size_t)-1` for high-bit input.
- Focused preprocessor and runtime tests passed.

## Milestone 18: `wcsrtombs` C-Locale Behavior

Calibration before execution: Re-read `wcstombs` and pointer-to-source update semantics.

- [x] Implement/register `wcsrtombs`.
- [x] Cover length query, conversion, truncation, source pointer update, NUL completion, and high wide char rejection.
- [x] Verify, commit, and push `feat(runtime): add wcsrtombs c locale`.

Findings:

- Added `wcsrtombs` declaration to builtin `<wchar.h>` and registered the extern.
- C-locale implementation converts 32-bit ASCII wide chars to bytes, sets `*src = NULL` on NUL completion, updates `*src` after truncation, leaves `*src` unchanged for length queries, and returns `(size_t)-1` for high wide chars.
- Focused preprocessor and runtime tests passed.

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
