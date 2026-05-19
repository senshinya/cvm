# Phase 6 Hosted Libc Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete a 20+ milestone hosted libc fidelity sweep for hermetic C-standard runtime surfaces already present in headers and registry.

**Architecture:** Keep all behavior deterministic and hermetic inside `runtime/extern.go` and `ExternRegistry`. Extend configured `FILE *` state, temporary-name generation, buffer-control validation, and C-locale multibyte helpers without introducing ambient host filesystem or native descriptor dependencies.

**Tech Stack:** Go runtime extern registry, memory model, GCC-style source execution harness, builtin C headers.

---

## Milestone 1: Baseline And Branch Setup

Calibration: Phase 5 is closed at `a0cbf4a`; Phase 6 will target hosted libc surfaces that are already declared/registered but still stubbed or narrow.

**Files:**
- Create: `docs/superpowers/specs/2026-05-19-phase6-hosted-libc-fidelity-design.md`
- Create: `docs/superpowers/plans/2026-05-19-phase6-hosted-libc-fidelity.md`

- [x] **Step 1: Create Phase 6 branch**
- [x] **Step 2: Write Phase 6 design and 20+ milestone plan**
- [x] **Step 3: Verify, commit, and push planning docs**

Run standard verification, then commit `docs: define phase 6 hosted libc roadmap`.

## Milestone 2: `fgetpos` For Configured Files

Calibration before execution: Re-read `fgetposExtern`, `ftellExtern`, and tests around stdio position helpers; confirm configured files still return `-1` from `fgetpos`.

- [x] Add failing direct extern test for `fgetpos` storing the current configured-file offset.
- [x] Implement `fgetpos` for `hostFile` by storing `file.pos` as `TypeI64`.
- [x] Add source-level runtime test using `fgetpos`.
- [x] Verify, commit, and push `feat(runtime): get configured file positions`.

## Milestone 3: `fsetpos` For Configured Files

Calibration before execution: Re-read `fsetposExtern`, `fseekExtern`, and update-mode direction reset behavior.

- [x] Add failing direct extern test for restoring a configured-file position.
- [x] Implement `fsetpos` for `hostFile` by loading `TypeI64`, rejecting negative offsets, clearing pushback/EOF, and resetting update-mode direction.
- [x] Add source-level runtime test using `fgetpos` plus `fsetpos`.
- [x] Verify, commit, and push `feat(runtime): set configured file positions`.

## Milestone 4: `fgetpos`/`fsetpos` Error Edges

Calibration before execution: Check standard-stream stub tests and invalid position-pointer coverage.

- [x] Add direct tests for standard streams, negative positions, and invalid memory.
- [x] Tighten `fgetpos`/`fsetpos` error returns without trapping where C-level failure is appropriate.
- [x] Verify, commit, and push `test(runtime): cover file position errors`.

## Milestone 5: `freopen` Opens Configured Read Files

Calibration before execution: Re-read `freopenExtern` and `fopenExtern`; confirm `freopen` still returns null for configured existing files.

- [x] Add failing direct extern test reopening an existing stream as configured read file.
- [x] Implement `freopen` replacement state for `r` mode over configured files.
- [x] Add source-level runtime test for `freopen(..., "r", stdout_like_handle)` using a configured file handle.
- [x] Verify, commit, and push `feat(runtime): reopen configured read files`.

## Milestone 6: `freopen` Write And Truncate Modes

Calibration before execution: Check `w`/`w+` handling in `fopen` and file persistence on close.

- [x] Add direct tests for `freopen(path, "w", stream)` truncating and writing a configured file.
- [x] Implement write/truncate reopen state.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): reopen configured write files`.

## Milestone 7: `freopen` Append Modes

Calibration before execution: Check append positioning and update-mode sequencing after reopen.

- [x] Add direct tests for `a` and `a+` reopen append positioning.
- [x] Implement append reopen state.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): reopen configured append files`.

## Milestone 8: `freopen` Failure Semantics

Calibration before execution: Check C failure expectations and current close behavior.

- [x] Add tests that failed `freopen` leaves the original stream usable.
- [x] Adjust implementation to build replacement state before mutating the old handle.
- [x] Verify, commit, and push `fix(runtime): preserve stream on freopen failure`.

## Milestone 9: Deterministic `tmpnam(NULL)`

Calibration before execution: Re-read `tmpnamExtern` and existing stub tests; confirm `tmpnam(NULL)` currently returns null.

- [x] Add failing direct test expecting a stable non-null static C string for `tmpnam(NULL)`.
- [x] Implement deterministic generated names such as `/tmp/cvm-tmp-0`.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): generate hermetic tmp names`.

## Milestone 10: Deterministic `tmpnam(buffer)`

Calibration before execution: Check writable buffer helpers and `L_tmpnam`.

- [x] Add failing direct test expecting `tmpnam(buf)` to write and return `buf`.
- [x] Implement caller-buffer write with null termination.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): write tmpnam buffers`.

## Milestone 11: `tmpnam` Sequence Uniqueness

Calibration before execution: Confirm registry state can hold a temporary-name counter without affecting existing tests.

- [x] Add tests for multiple distinct `tmpnam` calls.
- [x] Add counter state to `ExternRegistry`.
- [x] Verify, commit, and push `test(runtime): cover tmpnam sequence`.

## Milestone 12: `tmpnam` With Hermetic File APIs

Calibration before execution: Check `fopen("w")` can create registry files whose names come from runtime memory.

- [x] Add source-level test using `tmpnam`, `fopen(name, "w")`, write, close, reopen, read.
- [x] Fix any integration gap.
- [x] Verify, commit, and push `test(runtime): use tmpnam with configured files`.

## Milestone 13: `setvbuf` Mode Validation

Calibration before execution: Re-read `setvbufExtern`, header macros, and existing no-op tests.

- [x] Add direct tests for `_IOFBF`, `_IOLBF`, `_IONBF`, and invalid mode.
- [x] Make invalid modes return nonzero while valid modes remain no-op success.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): validate setvbuf modes`.

## Milestone 14: `setvbuf` Size And Stream Validation

Calibration before execution: Check size argument handling and closed stream behavior.

- [x] Add tests for closed streams, null buffer with nonzero size, and ordinary buffer pointer validation.
- [x] Tighten validation while preserving no-op buffering.
- [x] Verify, commit, and push `test(runtime): cover setvbuf validation`.

## Milestone 15: `setbuf` Delegation Coverage

Calibration before execution: Confirm `setbuf` currently validates stream and returns void.

- [x] Add direct/source-level tests for `setbuf(stream, buf)` and `setbuf(stream, NULL)`.
- [x] Adjust implementation only if validation differs from `setvbuf`.
- [x] Verify, commit, and push `test(runtime): cover setbuf controls`.

## Milestone 16: `mblen` C-Locale Edge Cases

Calibration before execution: Re-read `mblenExtern` and existing multibyte tests.

- [x] Add tests for null state query, zero `n`, NUL character, ASCII character, and high-bit byte rejection.
- [x] Tighten C-locale behavior.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden mblen c locale`.

## Milestone 17: `mbtowc` C-Locale Edge Cases

Calibration before execution: Check wchar storage width and null state behavior.

- [x] Add tests for null destination, null source, zero `n`, NUL, ASCII, and high-bit rejection.
- [x] Tighten `mbtowc` behavior.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden mbtowc c locale`.

## Milestone 18: `wctomb` C-Locale Edge Cases

Calibration before execution: Check existing `wctomb` range behavior.

- [x] Add tests for null destination state query, ASCII, NUL, and out-of-range wchar.
- [x] Tighten `wctomb` behavior.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden wctomb c locale`.

## Milestone 19: `mbstowcs` Length And Null Destination

Calibration before execution: Re-read `mbstowcsExtern` and C-locale count semantics.

- [x] Add tests for null destination length query, truncation, NUL termination when space allows, and high-bit rejection.
- [x] Tighten `mbstowcs` behavior.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden mbstowcs c locale`.

## Milestone 20: `wcstombs` Length And Null Destination

Calibration before execution: Re-read `wcstombsExtern` and wide string helpers.

- [x] Add tests for null destination length query, truncation, NUL termination when space allows, and high wchar rejection.
- [x] Tighten `wcstombs` behavior.
- [x] Add source-level runtime coverage.
- [x] Verify, commit, and push `feat(runtime): harden wcstombs c locale`.

## Milestone 21: Multibyte Cross-Function Runtime Coverage

Calibration before execution: Re-scan direct tests and source-level tests from milestones 16-20 for gaps.

- [x] Add one source-level workflow using all five helpers together.
- [x] Verify, commit, and push `test(runtime): cover c locale multibyte workflow`.

## Milestone 22: Header And Registry Surface Recheck

Calibration before execution: Compare builtin stdio/stdlib/locale declarations against runtime registry and direct smoke tests.

- [x] Run/search header-vs-registry checks for touched surfaces.
- [x] Add missing smoke coverage if found.
- [x] Verify, commit, and push `docs: record phase 6 header registry recheck`.

Recheck notes: `rg "fgetpos|fsetpos|freopen|tmpnam|setbuf|setvbuf|mblen|mbtowc|wctomb|mbstowcs|wcstombs" preprocessor/headers.go preprocessor/headers_test.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go` shows each touched surface has a builtin declaration, header smoke entry where applicable, runtime registration, direct registry smoke coverage, and source-level runtime coverage. No additional smoke test was needed.

## Milestone 23: GCC Runtime Fixture Recheck

Calibration before execution: Re-run GCC runtime gap report and scan accept fixtures for fpos, freopen, tmpnam, setvbuf, and multibyte candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan imported GCC accept roots for newly unblocked candidates.
- [ ] Add a fixture only if it is low-risk and stable; otherwise record no candidate.
- [ ] Verify, commit, and push `docs: record phase 6 gcc fixture recheck`.

## Milestone 24: Phase 6 Closure Docs

Calibration before execution: Confirm all prior Phase 6 milestones are committed and pushed.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Create `docs/phase6-hosted-libc-fidelity-gap-map.md`.
- [ ] Mark this plan complete.
- [ ] Run standard verification, commit `docs: close phase 6 hosted libc work`, and push.
