# Phase 12 Wide Stdio And Format Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a deterministic C-locale subset of wide-character stdio and wide formatted input/output after Phase 11 wide string support is complete.

**Architecture:** Phase 12 layers wide stdio externs on top of existing hermetic stream state, configured stdin/stdout, Phase 10 multibyte conversion, and Phase 11 wide string helpers. It intentionally keeps host locale and stateful encodings out of scope.

**Tech Stack:** Go runtime externs, builtin C headers, hermetic stream runtime, source-level GCC execution fixtures.

---

## Milestone 1: Phase 12 Baseline

Calibration before execution: Start from merged Phase 11 and confirm wide string helpers are available.

- [x] Create or update this plan after Phase 11 closure.
- [x] Run baseline verification.
- [x] Commit and push `docs: start phase 12 wide stdio work`.

Baseline notes:
- Phase 12 starts from `main` after `merge: bytecode runtime phase 11`.
- `git diff --check`, `go test ./codegen -count=1`, and `go test ./... -count=1` passed with `GOCACHE=/private/tmp/cvm-go-build-cache`.

## Milestone 2: Header Declarations For Wide Stdio

Calibration before execution: Re-read `stdioHeader`, `wcharHeader`, and C99 declaration placement.

- [x] Declare `fwide`, `fputwc`, `putwc`, `putwchar`, `fgetwc`, `getwc`, `getwchar`, `ungetwc`, `fputws`, and `fgetws`.
- [x] Add header smoke tests.
- [x] Verify, commit `feat(headers): declare wide stdio functions`, and push.

Calibration adjustment: runtime loading requires every declared extern to have a registry entry, even when a source file only includes `<wchar.h>` and does not call the new function. Milestone 2 therefore adds conservative placeholder registrations for the new wide stdio names; Milestones 3-10 replace them with real deterministic implementations.

## Milestone 3: Stream Orientation Model

Calibration before execution: Inspect current stream state for a minimal orientation field.

- [x] Add deterministic stream orientation tracking for unoriented, byte, and wide.
- [x] Implement `fwide` query/set semantics.
- [x] Cover direct extern behavior.
- [x] Verify, commit `feat(runtime): track stream orientation`, and push.

## Milestone 4: `fputwc`

Calibration before execution: Reuse C-locale single-byte `wcrtomb` semantics.

- [x] Register and implement `fputwc`.
- [x] Cover ASCII, NUL, invalid high wide char, closed stream, and orientation.
- [x] Verify, commit `feat(runtime): add fputwc`, and push.

## Milestone 5: `putwc` And `putwchar`

Calibration before execution: Confirm aliases should route through `fputwc` and stdout.

- [x] Register and implement `putwc` and `putwchar`.
- [x] Cover direct extern and source-level stdout behavior.
- [x] Verify, commit `feat(runtime): add putwc aliases`, and push.

## Milestone 6: `fgetwc`

Calibration before execution: Reuse configured stdin/file read path and C-locale byte-to-wide conversion.

- [x] Register and implement `fgetwc`.
- [x] Cover ASCII, EOF, invalid high byte, closed stream, and orientation.
- [x] Verify, commit `feat(runtime): add fgetwc`, and push.

## Milestone 7: `getwc` And `getwchar`

Calibration before execution: Confirm aliases should route through `fgetwc` and stdin.

- [x] Register and implement `getwc` and `getwchar`.
- [x] Cover direct extern and source-level configured stdin behavior.
- [x] Verify, commit `feat(runtime): add getwc aliases`, and push.

## Milestone 8: `ungetwc`

Calibration before execution: Re-read narrow `ungetc` pushback storage.

- [x] Register and implement `ungetwc`.
- [x] Cover one wide pushback, EOF rejection, invalid high value rejection, and readback.
- [x] Verify, commit `feat(runtime): add ungetwc`, and push.

## Milestone 9: `fputws`

Calibration before execution: Reuse wide string scan and `fputwc` write behavior.

- [x] Register and implement `fputws`.
- [x] Cover ASCII string, empty string, invalid high wide value, closed stream, and return convention.
- [x] Verify, commit `feat(runtime): add fputws`, and push.

## Milestone 10: `fgetws`

Calibration before execution: Mirror `fgets` line/width semantics in wide units.

- [x] Register and implement `fgetws`.
- [x] Cover newline, width limit, EOF before data, EOF after data, invalid high byte, and terminator.
- [x] Verify, commit `feat(runtime): add fgetws`, and push.

## Milestone 11: Wide Output Format Header Surface

Calibration before execution: Re-read current printf/vprintf declaration pattern.

- [x] Declare `wprintf`, `fwprintf`, `swprintf`, `vwprintf`, `vfwprintf`, and `vswprintf`.
- [x] Add header smoke tests.
- [x] Verify, commit `feat(headers): declare wide printf functions`, and push.

## Milestone 12: Wide Format Parser Bridge

Calibration before execution: Inspect current formatter and identify reusable byte-format pieces.

- [x] Add helper to convert wide format strings to deterministic C-locale byte formats.
- [x] Reject high wide format characters with a deterministic error result.
- [x] Cover direct helper through extern tests.
- [x] Verify, commit `feat(runtime): add wide format bridge`, and push.

## Milestone 13: `swprintf`

Calibration before execution: Reuse buffer formatter output, then widen ASCII bytes into destination.

- [x] Register and implement `swprintf`.
- [x] Cover integer, string, char, percent, width/precision, truncation, and return value.
- [x] Verify, commit `feat(runtime): add swprintf`, and push.

## Milestone 14: `wprintf`

Calibration before execution: Route formatted bytes to stdout through wide output semantics.

- [x] Register and implement `wprintf`.
- [x] Cover source-level stdout output and return count.
- [x] Verify, commit `feat(runtime): add wprintf`, and push.

## Milestone 15: `fwprintf`

Calibration before execution: Reuse `fprintf` stream routing and wide format bridge.

- [x] Register and implement `fwprintf`.
- [x] Cover configured output file/stdout, closed stream, and return count.
- [x] Verify, commit `feat(runtime): add fwprintf`, and push.

## Milestone 16: Wide Variadic Output Aliases

Calibration before execution: Re-read memory-backed `va_list` support for v-format externs.

- [x] Register and implement `vwprintf`, `vfwprintf`, and `vswprintf`.
- [x] Cover memory-backed va-list integer/string arguments.
- [x] Verify, commit `feat(runtime): add wide vprintf functions`, and push.

## Milestone 17: Wide Input Format Header Surface

Calibration before execution: Re-read current scanf declaration pattern.

- [x] Declare `wscanf`, `fwscanf`, and `swscanf`.
- [x] Add header smoke tests.
- [x] Verify, commit `feat(headers): declare wide scanf functions`, and push.

## Milestone 18: Wide Scanner Bridge

Calibration before execution: Inspect existing scanner support and keep the first subset bounded.

- [x] Add helper to read wide format strings and wide input buffers as C-locale byte strings.
- [x] Cover integer, string, char, width, assignment suppression, and failure cases.
- [x] Verify, commit `feat(runtime): add wide scanner bridge`, and push.

## Milestone 19: `swscanf`

Calibration before execution: Use wide input and wide format buffers without stream side effects.

- [ ] Register and implement `swscanf`.
- [ ] Cover integer, `%ls`, `%lc`, `%n`, mismatch, and EOF-like empty input.
- [ ] Verify, commit `feat(runtime): add swscanf`, and push.

## Milestone 20: `wscanf`

Calibration before execution: Reuse configured stdin scanner behavior and orientation rules.

- [ ] Register and implement `wscanf`.
- [ ] Cover configured stdin source-level input and unread preservation.
- [ ] Verify, commit `feat(runtime): add wscanf`, and push.

## Milestone 21: `fwscanf`

Calibration before execution: Reuse hermetic file scanner behavior.

- [ ] Register and implement `fwscanf`.
- [ ] Cover configured file input, matching failure, EOF, and stream state.
- [ ] Verify, commit `feat(runtime): add fwscanf`, and push.

## Milestone 22: Wide Stdio Source Workflow

Calibration before execution: Search for overlapping source runtime tests before adding workflows.

- [ ] Add source-level workflow covering `fwide`, `fputwc`, `fgetwc`, `ungetwc`, `fputws`, and `fgetws`.
- [ ] Verify, commit `test(runtime): cover wide stdio workflow`, and push.

## Milestone 23: Wide Format Source Workflow

Calibration before execution: Confirm source-level wide printf/scanf subset remains deterministic.

- [ ] Add source-level workflow covering wide formatted output and input.
- [ ] Verify, commit `test(runtime): cover wide format workflow`, and push.

## Milestone 24: GCC Fixture Recheck

Calibration before execution: Run GCC runtime gap report and scan imported accept roots for wide stdio candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan imported GCC accept fixtures for `wprintf`, `wscanf`, `fwide`, `fgetwc`, `fputwc`, `fgetws`, and `fputws`.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit `docs: record phase 12 gcc fixture recheck`, and push.

## Milestone 25: Header And Registry Recheck

Calibration before execution: Search all touched declarations, registrations, and tests.

- [ ] Recheck declarations, registry entries, direct tests, and source tests.
- [ ] Add missing smoke coverage if found.
- [ ] Verify, commit `docs: record phase 12 header registry recheck`, and push.

## Milestone 26: Phase 12 Gap Map

Calibration before execution: Confirm all Phase 12 implementation increments are committed and pushed.

- [ ] Create `docs/phase12-wide-stdio-format-fidelity-gap-map.md`.
- [ ] Record closed surfaces and residual deterministic limits.
- [ ] Verify, commit `docs: map phase 12 wide stdio gaps`, and push.

## Milestone 27: Phase 12 Closure

Calibration before execution: Confirm branch, remote, and final verification status.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Mark this plan complete.
- [ ] Verify, commit `docs: close phase 12 wide stdio work`, and push.
