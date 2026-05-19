# Phase 13 Formatted I/O Spec Fidelity Implementation Plan

Date: 2026-05-20
Branch: `codex/bytecode-runtime-phase-13`

**Goal:** Expand the deterministic C-locale formatted I/O subset shared by narrow and wide stdio after Phase 12. Keep native-locale behavior, host filesystem side effects, and non-deterministic libc quirks out of scope.

**Architecture:** Continue using the existing `formatCString`, `scanString`, hermetic stream state, and Phase 12 wide ASCII bridges. Each milestone starts with a focused calibration of current behavior, adds the smallest stable coverage increment, fixes root cause if needed, then runs focused tests and the standard gates.

## Milestone 1: Phase 13 Baseline

Calibration before execution: Start from closed Phase 12 and confirm the branch/remote state.

- [x] Create this Phase 13 plan with 20+ milestones.
- [x] Verify, commit `docs: start phase 13 formatted io work`, and push.

## Milestone 2: Format Parser Recheck

Calibration before execution: Re-read `formatCString`, `scanString`, and direct/source coverage before changing behavior.

- [x] Map supported printf flags, width, precision, length modifiers, and verbs.
- [x] Map supported scanf width, assignment suppression, length modifiers, and verbs.
- [x] Record the next low-risk implementation order in this plan.
- [ ] Verify, commit `docs: map phase 13 format parser baseline`, and push.

Baseline result:
- `formatCString` parses flags, numeric and dynamic width, numeric and dynamic precision, and `hh`/`h`/`l`/`ll`/`j`/`z`/`t`/`L` length modifiers. It supports narrow string, integer, pointer, character, count, floating, and literal percent verbs, but length modifiers are only semantically honored for `%n` today.
- `scanString` parses assignment suppression, numeric width, the same length modifiers, integer, float, pointer, count, narrow/wide text, scansets, whitespace, literals, and literal percent. Its integer, float, count, and text stores already route through length-aware helpers.
- The next low-risk order remains printf signed and unsigned integer length semantics first, then count/text/character edge hardening, then pointer, dynamic width/precision, alternate form, floating, scanf edge cases, and fixture rechecks.

## Milestone 3: Printf Signed Integer Length Modifiers

Calibration before execution: Compare current `%d`/`%i` handling for `h`, `hh`, `l`, `ll`, `j`, `z`, and `t`.

- [x] Add direct and source coverage for signed integer length modifiers.
- [x] Fix formatting where argument narrowing/sign extension is wrong.
- [x] Verify, commit `feat(runtime): harden printf signed lengths`, and push.

Calibration result:
- Focused direct and source tests showed `%hhd` and `%hd` still printed the promoted `int` values rather than applying the signed-char and signed-short conversions required by the length modifiers.
- `%d` now formats through the default `int` view, while `l`/`ll`/`j`/`z`/`t` keep the target 64-bit signed representation used by the current ABI model.

## Milestone 4: Printf Unsigned Integer Length Modifiers

Calibration before execution: Reuse signed-length findings for `%u`, `%x`, `%X`, and `%o`.

- [x] Add direct and source coverage for unsigned length modifiers.
- [x] Fix masking/narrowing behavior where needed.
- [x] Verify, commit `feat(runtime): harden printf unsigned lengths`, and push.

Calibration result:
- Focused coverage showed `%hhu`, `%hu`, default `%u`, `%hhx`, `%hx`, default `%x`, `%hho`, `%ho`, and default `%o` were formatting the full incoming runtime value instead of the C length-selected unsigned view.
- `%u`/`%x`/`%X`/`%o` now share an unsigned length helper: `hh` and `h` mask to 8 and 16 bits, the default masks to 32 bits, and `l`/`ll`/`j`/`z`/`t` use the current 64-bit ABI view.

## Milestone 5: Printf Count Length Modifiers

Calibration before execution: Re-read existing `%n` storage and scanf count handling.

- [x] Cover `%hhn`, `%hn`, `%n`, `%ln`, and `%lln` for printf.
- [x] Fix count storage widths and alignment behavior if needed.
- [x] Verify, commit `test(runtime): cover printf count lengths`, and push.

Calibration result:
- Existing `%n` storage already routed through `writeCountType`, which covers `hh`, `h`, `l`, `ll`, `j`, `z`, and `t`.
- The milestone added direct coverage for `l`/`j`/`z`/`t` count writes and source coverage for the full `hh`/`h`/default/`l`/`ll`/`j`/`z`/`t` set; no runtime code change was needed.

## Milestone 6: Printf Character Length Modifiers

Calibration before execution: Confirm current `%c` and wide bridge behavior.

- [x] Cover narrow `%c` and wide `%lc` for narrow and wide formatted output.
- [x] Fix deterministic C-locale wide-char output behavior if needed.
- [x] Verify, commit `feat(runtime): support printf wide chars`, and push.

Calibration result:
- ASCII `%lc` already worked by falling through the narrow `%c` byte path, including through the wide formatted-output bridge.
- The missing deterministic C-locale behavior was rejecting non-ASCII `%lc` for narrow formatting; that now errors instead of silently writing a high byte.

## Milestone 7: Printf String Length Modifiers

Calibration before execution: Confirm current `%s` behavior and Phase 12 wide string storage.

- [x] Cover narrow `%s` and wide `%ls` for narrow and wide formatted output.
- [x] Add precision coverage for wide strings.
- [x] Verify, commit `feat(runtime): support printf wide strings`, and push.

Calibration result:
- `%ls` was parsed but still read the argument as a narrow C string, so wide buffers formatted as only the first low byte and non-ASCII wide strings were not rejected.
- `%ls` now reads a wide C string through the deterministic ASCII bridge shared with wide format strings; precision applies to the resulting C-locale byte output.

## Milestone 8: Printf Pointer Formatting Width

Calibration before execution: Re-read current `%p` formatting and source expectations.

- [x] Cover `%p` null/non-null with width and left alignment.
- [x] Fix deterministic pointer string formatting if needed.
- [x] Verify, commit `test(runtime): cover printf pointer width`, and push.

Calibration result:
- `%p` already formats as deterministic `0x` plus lowercase hex and participates in the common width/left-align writer.
- The milestone added direct null/non-null width coverage and source coverage for null pointer right/left alignment; no runtime code change was needed.

## Milestone 9: Printf Dynamic Width And Precision

Calibration before execution: Recheck `*` width/precision handling across integer, string, and float paths.

- [x] Add coverage for negative dynamic width, negative precision, and combined `*.*`.
- [x] Fix behavior where it diverges from deterministic C-locale expectations.
- [x] Verify, commit `test(runtime): cover printf dynamic sizing`, and push.

Calibration result:
- Existing dynamic width and precision logic already treats negative width as left alignment with positive width and negative precision as omitted precision.
- The milestone added direct and source coverage for negative width, negative precision, and combined `*.*` across string, integer, and floating output; no runtime code change was needed.

## Milestone 10: Printf Zero Precision Edge Cases

Calibration before execution: Compare integer zero-value precision behavior for `%d`, `%u`, `%x`, and `%o`.

- [x] Cover `%.0d`, `%.0u`, `%.0x`, `%.0o`, and alternate forms.
- [x] Fix omitted digits and prefix behavior if needed.
- [x] Verify, commit `feat(runtime): harden printf zero precision`, and push.

Calibration result:
- Zero precision already omitted zero digits for decimal, unsigned decimal, hex, and plain octal.
- `%#.0o` incorrectly lost the required single `0` after integer precision was applied; alternate octal now restores that digit for the zero-value precision edge case.

## Milestone 11: Printf Alternate Form Edge Cases

Calibration before execution: Re-read current `#` handling.

- [ ] Cover `%#x`, `%#X`, `%#o`, `%#a`, `%#f`, `%#g`.
- [ ] Fix deterministic alternate-form output where supported.
- [ ] Verify, commit `feat(runtime): harden printf alternate forms`, and push.

## Milestone 12: Printf Floating Width Precision

Calibration before execution: Recheck current float formatting and binary64-backed long double limits.

- [ ] Cover `%f`, `%e`, `%E`, `%g`, `%G`, `%a`, and `%A` width/precision.
- [ ] Fix stable C-locale float formatting gaps that do not require host-locale behavior.
- [ ] Verify, commit `feat(runtime): harden printf float sizing`, and push.

## Milestone 13: Printf Special Floating Values

Calibration before execution: Confirm formatting for infinities, NaNs, signed zero, and case variants.

- [ ] Cover `inf`, `nan`, signed zero, uppercase verbs, and signs.
- [ ] Fix deterministic spelling/case behavior if needed.
- [ ] Verify, commit `feat(runtime): harden printf special floats`, and push.

## Milestone 14: Scanf Signed Integer Length Modifiers

Calibration before execution: Re-read `scanStoreInteger` and current source tests.

- [ ] Cover `%hhd`, `%hd`, `%d`, `%ld`, `%lld`, `%jd`, `%zd`, and `%td`.
- [ ] Fix storage widths, signedness, and alignment behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf signed lengths`, and push.

## Milestone 15: Scanf Unsigned Integer Length Modifiers

Calibration before execution: Extend signed-length findings to unsigned conversions.

- [ ] Cover `%hhu`, `%hu`, `%u`, `%lu`, `%llu`, `%x`, and `%o`.
- [ ] Fix unsigned storage and negative input behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf unsigned lengths`, and push.

## Milestone 16: Scanf Count Length Modifiers

Calibration before execution: Compare scanf `%n` with printf `%n` storage rules.

- [ ] Cover `%hhn`, `%hn`, `%n`, `%ln`, and `%lln`.
- [ ] Fix storage widths and no-assignment count behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf count lengths`, and push.

## Milestone 17: Scanf Character Length Modifiers

Calibration before execution: Recheck `%c` and `%lc` direct and source coverage.

- [ ] Cover narrow and wide char scanning with widths.
- [ ] Fix deterministic C-locale wide-char storage behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf wide chars`, and push.

## Milestone 18: Scanf String Length Modifiers

Calibration before execution: Recheck `%s`, `%ls`, and width/null termination.

- [ ] Cover narrow and wide strings with width limits.
- [ ] Fix wide string termination and unread preservation if needed.
- [ ] Verify, commit `feat(runtime): harden scanf wide strings`, and push.

## Milestone 19: Scanf Scanset Length Modifiers

Calibration before execution: Re-read `parseScanSet` and wide scanner bridge behavior.

- [ ] Cover `%[`, `%l[`, negated sets, ranges, and widths.
- [ ] Fix wide scanset output behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf wide scansets`, and push.

## Milestone 20: Scanf Floating Length Modifiers

Calibration before execution: Recheck `%f`, `%lf`, and `%Lf` scanner behavior.

- [ ] Cover float, double, and current long-double storage for decimal and hex input.
- [ ] Fix deterministic float scanner gaps that do not require native extended precision.
- [ ] Verify, commit `feat(runtime): harden scanf float lengths`, and push.

## Milestone 21: Scanf Literal And Whitespace Edges

Calibration before execution: Re-read input failure and matching failure handling.

- [ ] Cover literal percent, trailing whitespace, leading whitespace, and literal mismatch.
- [ ] Fix consumed-byte and EOF behavior if needed.
- [ ] Verify, commit `feat(runtime): harden scanf literal edges`, and push.

## Milestone 22: Stream Formatted I/O Preservation

Calibration before execution: Confirm `scanf`, `fscanf`, `wscanf`, and `fwscanf` push back unread bytes consistently.

- [ ] Add source workflows for unread preservation after narrow and wide scans.
- [ ] Fix stream pushback or EOF state if needed.
- [ ] Verify, commit `test(runtime): cover formatted stream preservation`, and push.

## Milestone 23: GCC Fixture Recheck

Calibration before execution: Run GCC runtime gap report and scan imported accept roots for newly unblocked formatted I/O fixtures.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan imported accept roots for printf/scanf candidates.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit `docs: record phase 13 gcc fixture recheck`, and push.

## Milestone 24: Header Registry Source Recheck

Calibration before execution: Search declarations, sema/builtin typing where relevant, registry entries, direct tests, and source workflows.

- [ ] Recheck all touched formatted I/O surfaces.
- [ ] Add missing smoke coverage if found.
- [ ] Verify, commit `docs: record phase 13 formatted io recheck`, and push.

## Milestone 25: Phase 13 Gap Map

Calibration before execution: Confirm all implementation increments are committed and pushed.

- [ ] Create `docs/phase13-formatted-io-spec-fidelity-gap-map.md`.
- [ ] Record closed surfaces and residual deterministic limits.
- [ ] Verify, commit `docs: map phase 13 formatted io gaps`, and push.

## Milestone 26: Phase 13 Closure

Calibration before execution: Confirm branch, remote, and final verification status.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Mark this plan complete.
- [ ] Verify, commit `docs: close phase 13 formatted io work`, and push.
