# Phase 9 Floating Conversion Errno Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close deterministic C-locale floating conversion range, errno, NaN/Inf, and end-pointer fidelity gaps left after Phase 8.

**Architecture:** Preserve the deterministic hosted-runtime model. Strengthen parsing and range handling in `runtime/extern.go`, direct coverage in `runtime/extern_test.go`, source-level coverage in `runtime/gcc_exec_test.go`, and closure docs/gap maps.

**Tech Stack:** Go runtime externs, bytecode memory helpers, deterministic hosted registry state, GCC-style source execution tests.

---

## Milestone 1: Baseline And Branch Setup

Calibration before execution: Start from Phase 8 closure on `codex/bytecode-runtime-phase-9`.

- [x] Create/switch Phase 9 branch.
- [x] Re-read Phase 9 roadmap.
- [x] Run baseline verification.
- [x] Commit branch setup docs if needed.

Baseline:

- Created `codex/bytecode-runtime-phase-9` from Phase 8 closure commit `b1b0f01`.
- Phase 9 contains 22 milestones covering floating conversion errno/range behavior, NaN/Inf parsing, header/registry recheck, GCC fixture recheck, gap map, and closure docs.
- Phase 10 roadmap was created alongside Phase 9 so the next phase can start immediately after Phase 9 closure.
- No runtime code changes were needed for branch setup.

## Milestone 2: Errno Static Variable Baseline

Calibration before execution: Re-read `LookupVariable("errno")`, static var storage, and existing errno tests.

- [x] Add direct tests proving `errno` is per-memory stable and initially zero.
- [x] Add source-level smoke for reading/writing `errno`.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): recheck errno storage`.

Findings:

- Existing direct coverage already checked one-memory stable `errno` address and read/write behavior.
- Existing source-level runtime coverage already checked `<errno.h>` constants and source read/write.
- Added direct coverage proving a second `Memory` gets its own stable, initially zero `errno` under the same registry.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestErrno' -count=1` passed.

## Milestone 3: Floating Conversion Errno Preservation

Calibration before execution: Re-read `strtoFloatExtern`, `atofExtern`, and current errno behavior.

- [x] Add tests proving successful `strtod` does not alter nonzero `errno`.
- [x] Add tests proving no-conversion `strtod` does not alter nonzero `errno`.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): preserve errno in float parsing`.

Findings:

- `strtoFloatExtern` did not write `errno`; the missing piece was regression coverage.
- Added direct extern and source-level runtime tests proving successful `strtod` and no-conversion `strtod` preserve a nonzero `errno`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 4: `strtod` Infinity Spelling

Calibration before execution: Check `strconv.ParseFloat` behavior for `inf`, `infinity`, sign, and end index.

- [x] Add direct/source tests for `inf`, `+infinity`, `-INF`, trailing text, and endptr.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): cover strtod infinity parsing`.

Findings:

- Existing parser behavior already accepted infinity spellings through `strconv.ParseFloat`.
- Added direct extern and source-level runtime coverage for `inf`, `+infinity`, `-INF`, trailing text, and end pointers.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 5: `strtod` NaN Spelling

Calibration before execution: Check accepted NaN spellings and deterministic payload policy.

- [x] Add direct/source tests for `nan`, `NAN`, `nan(payload)`, trailing text, and endptr.
- [x] Fix if needed.
- [x] Verify, commit, and push `feat(runtime): cover strtod nan parsing`.

Findings:

- Tests exposed that `nan(payload)` previously parsed as `nan` but left the payload in `endptr`.
- Added a deterministic NaN special-token parser that consumes a balanced parenthesized payload when present.
- Added direct extern and source-level runtime coverage for `nan`, `NAN`, `nan(payload)`, trailing text, and end pointers.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed after the fix.

## Milestone 6: `nan` Family Payload Inputs

Calibration before execution: Re-read `mathNanExtern` and existing `nan/nanf/nanl` tests.

- [x] Add tests for empty payload, numeric payload, and nonnumeric payload for `nan`, `nanf`, and `nanl`.
- [x] Add source-level runtime coverage.
- [x] Fix if needed.
- [x] Verify, commit, and push `test(runtime): cover nan payload inputs`.

Findings:

- `mathNanExtern` intentionally ignores payload text and returns a deterministic NaN of the requested return type.
- Added direct extern coverage for empty, numeric, and nonnumeric payloads across `nan`, `nanf`, and `nanl`.
- Added source-level runtime coverage for payload inputs.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestPlainMathNanExterns|TestBuiltinFloatingConstantsExecuteThroughRuntime' -count=1` passed.

## Milestone 7: Decimal Overflow Range Trapdoor

Calibration before execution: Determine deterministic range policy for `strtod("1e309")` and errno.

- [x] Add direct/source tests for positive and negative decimal overflow returning infinities.
- [x] Set `errno` to `ERANGE` on overflow.
- [x] Verify, commit, and push `feat(runtime): mark strtod decimal overflow`.

Findings:

- `strconv.ParseFloat` reports `ErrRange` for overflowing tokens while still returning the correctly signed infinity.
- The previous backtracking parser could ignore `ErrRange` and accept a shorter prefix, so `1e309!` could lose the intended end pointer.
- `ErrRange` is now treated as a successful conversion with range status, preserving the full token end pointer and setting `errno` to `ERANGE`.
- Added direct extern and source-level runtime coverage for positive and negative decimal overflow.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 8: Decimal Underflow Range Trapdoor

Calibration before execution: Check small decimal parsing and deterministic underflow threshold.

- [x] Add direct/source tests for decimal underflow to zero.
- [x] Set `errno` to `ERANGE` on underflow-to-zero.
- [x] Preserve successful subnormal parsing behavior.
- [x] Verify, commit, and push `feat(runtime): mark strtod decimal underflow`.

Findings:

- Go's `strconv.ParseFloat` returns nil error for `1e-400`/`-1e-400`, so underflow-to-zero needed an explicit significand check.
- Added detection for parsed zero values whose significand contains a nonzero digit, while preserving exact zero inputs and representable subnormal values.
- Added direct extern and source-level runtime coverage for positive underflow, negative underflow, and the smallest positive subnormal `5e-324`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed after the fix.

## Milestone 9: Hex Float Overflow Range Trapdoor

Calibration before execution: Check `0x1p+2048` parsing behavior and endptr behavior.

- [x] Add direct/source tests for positive and negative hex overflow.
- [x] Set `errno` to `ERANGE` on overflow.
- [x] Verify, commit, and push `feat(runtime): mark strtod hex overflow`.

Findings:

- `strconv.ParseFloat` reports `ErrRange` for `0x1p+2048` and `-0x1p+2048`, so the range path added for decimal overflow also covers hex overflow.
- Added direct extern and source-level runtime coverage for positive and negative hex overflow, including end pointer checks after the full hex token.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 10: Hex Float Underflow Range Trapdoor

Calibration before execution: Check tiny hex float parsing and deterministic zero/subnormal boundary.

- [x] Add direct/source tests for hex underflow to zero.
- [x] Set `errno` to `ERANGE` on underflow-to-zero.
- [x] Verify, commit, and push `feat(runtime): mark strtod hex underflow`.

Findings:

- `strconv.ParseFloat` also returns nil error for `0x1p-20000` and `-0x1p-20000`, so the explicit nonzero-significand underflow detector is required for hex tokens too.
- The detector correctly stops at `p`/`P` for hex exponents, preserving exact zero hex inputs and representable hex subnormals.
- Added direct extern and source-level runtime coverage for positive hex underflow, negative hex underflow, and `0x1p-1074`.
- Focused `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestStdlibFloatParser' -count=1` passed.

## Milestone 11: `strtof` Float32 Overflow

Calibration before execution: Confirm `strtof` currently parses at float64 precision before returning TypeF32.

- [ ] Add direct/source tests for float32 overflow.
- [ ] Return signed float32 infinities and set `errno` to `ERANGE`.
- [ ] Verify, commit, and push `feat(runtime): mark strtof overflow`.

## Milestone 12: `strtof` Float32 Underflow

Calibration before execution: Confirm float32 underflow/subnormal behavior in runtime values.

- [ ] Add direct/source tests for float32 underflow to zero.
- [ ] Set `errno` to `ERANGE` on underflow-to-zero.
- [ ] Preserve representable float32 subnormals where possible.
- [ ] Verify, commit, and push `feat(runtime): mark strtof underflow`.

## Milestone 13: `strtold` Current Flong Overflow

Calibration before execution: Reconfirm current `TypeFLong` is binary64-backed.

- [ ] Add direct/source tests documenting current flong overflow behavior.
- [ ] Set `errno` to `ERANGE` consistently with current representation.
- [ ] Verify, commit, and push `feat(runtime): mark strtold overflow`.

## Milestone 14: `strtold` Current Flong Underflow

Calibration before execution: Reconfirm current flong underflow behavior.

- [ ] Add direct/source tests documenting current flong underflow-to-zero behavior.
- [ ] Set `errno` to `ERANGE` consistently with current representation.
- [ ] Verify, commit, and push `feat(runtime): mark strtold underflow`.

## Milestone 15: `atof` Wrapper Range Behavior

Calibration before execution: Re-read `atofExtern` and decide whether wrapper writes errno through shared parser.

- [ ] Add direct/source tests for `atof` overflow, underflow, and no-conversion.
- [ ] Reuse shared float parser range handling.
- [ ] Verify, commit, and push `feat(runtime): align atof range behavior`.

## Milestone 16: Endptr With Range Errors

Calibration before execution: Check `strtoFloatExtern` endptr calculation after range errors.

- [ ] Add direct/source tests proving overflow and underflow still set `endptr` after the consumed token.
- [ ] Fix if needed.
- [ ] Verify, commit, and push `test(runtime): cover float range endptr`.

## Milestone 17: Errno Direct Extern Sweep

Calibration before execution: Search direct extern tests for all floating conversion errno cases.

- [ ] Add helper coverage if duplicated errno setup/assertions are brittle.
- [ ] Ensure no-conversion, success, overflow, and underflow are all represented.
- [ ] Verify, commit, and push `test(runtime): sweep float errno externs`.

## Milestone 18: Source Runtime Sweep

Calibration before execution: Search `runtime/gcc_exec_test.go` for Phase 9 source-level surfaces.

- [ ] Ensure source tests cover `errno`, `ERANGE`, `strtod`, `strtof`, `strtold`, `atof`, `nan`, and infinity paths.
- [ ] Add one compact source fixture if missing.
- [ ] Verify, commit, and push `test(runtime): sweep float errno source coverage`.

## Milestone 19: Header/Registry Recheck

Calibration before execution: Search errno/math/stdlib headers and extern registry for touched Phase 9 surfaces.

- [ ] Record declaration/registration/test surface status.
- [ ] Add smoke coverage if missing.
- [ ] Verify, commit, and push `docs: record phase 9 header registry recheck`.

## Milestone 20: GCC Runtime Fixture Recheck

Calibration before execution: Run gap report and scan GCC accept roots for newly stable floating conversion candidates.

- [ ] Run `TestGCCExecutionGapReportIsCurrent`.
- [ ] Scan fixture roots for float conversion, errno, nan, inf, and range candidates.
- [ ] Add a low-risk fixture only if stable.
- [ ] Verify, commit, and push `docs: record phase 9 gcc fixture recheck`.

## Milestone 21: Phase 9 Residual Gap Map

Calibration before execution: Confirm all Phase 9 implementation milestones are committed and pushed.

- [ ] Create `docs/phase9-floating-conversion-errno-fidelity-gap-map.md`.
- [ ] Record closed surfaces and residual deterministic limits.
- [ ] Verify, commit, and push `docs: map phase 9 float conversion gaps`.

## Milestone 22: Phase 9 Closure Docs

Calibration before execution: Confirm all prior Phase 9 milestones are committed and pushed.

- [ ] Update `docs/bytecode-runtime-handoff.md`.
- [ ] Mark this plan complete.
- [ ] Run standard verification, commit `docs: close phase 9 float conversion work`, and push.
