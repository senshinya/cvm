# Phase 3 Runtime ABI Fidelity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Phase 3 hosted-runtime and ABI-fidelity work through 14 calibrated milestones.

**Architecture:** Keep CVM deterministic: source-level behavior lowers into bytecode/runtime representations, hosted state is configured through runtime APIs, and C programs do not read ambient host filesystem or environment by default. Every milestone starts with a calibration scan before implementation.

**Tech Stack:** Go, CVM bytecode/codegen/runtime, builtin C headers in `preprocessor`, sema analysis, GCC-style runtime execution tests.

---

## Common Verification

Run after every coherent increment:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Milestone Rule

Before each milestone, write a short calibration note in this plan or the roadmap describing current support, missing behavior, and whether the milestone should proceed unchanged, shrink, split, or close as already complete.

## Milestone 1: Phase 3 Baseline And Gap Map

**Files:**
- Create: `docs/superpowers/specs/2026-05-19-phase3-runtime-abi-fidelity-design.md`
- Create: `docs/phase3-runtime-gap-map.md`
- Create: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Calibrate current state**

Scanned handoff, runtime/codegen/header support, and tests for `va_arg`, `va_list`, `scanf`, `FILE` state, environment, termination, long double, complex, and aggregate runtime coverage.

- [x] **Step 2: Record Phase 3 design**

Saved the Phase 3 scope, architecture, milestone list, and completion criteria.

- [x] **Step 3: Record gap map**

Saved the current high-value runtime gaps and selected source-level `va_arg` as the first implementation target.

- [x] **Step 4: Verify, commit, and push**

Run Common Verification, then:

```bash
git add docs/superpowers/specs/2026-05-19-phase3-runtime-abi-fidelity-design.md docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "docs: define phase 3 runtime abi roadmap"
git push -u origin codex/bytecode-runtime-phase-3
```

## Milestone 2: Source-Level `va_arg` And Memory `va_list` Unification

**Calibration:** Completed. `<stdarg.h>` still lowered `va_arg(ap, type)` to `((type)0)`, while VM bytecode already supported `OpVaStart`, `OpVaArg`, and `OpVaEnd`. Source-level `va_start` and `va_end` called no-op externs, so source code never activated the VM variadic cursor.

**Initial target:** Replace the `va_arg(ap, type)` placeholder with a source-level path that executes real argument reads for integer and pointer values.

**Files:**
- Modify: `preprocessor/headers.go`
- Modify: `sema/builtin.go`
- Modify: `codegen/expr.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing source-level integer `va_arg` runtime test**

Added `TestGCCSourceVaArgIntegerExecutesThroughRuntime`, which failed when `va_arg` still returned `0`.

- [x] **Step 2: Lower source-level stdarg builtins to VM vararg ops**

Updated `<stdarg.h>` so `va_arg(ap, type)` expands through `__builtin_va_arg`, added the builtin signature, and taught codegen to emit `OpVaStart`, `OpVaArg`, and `OpVaEnd` for the recognized source-level shapes.

- [x] **Step 3: Add pointer `va_arg` runtime coverage**

Added `TestGCCSourceVaArgPointerExecutesThroughRuntime`.

- [x] **Step 4: Verify, commit, and push**

Run Common Verification, then:

```bash
git add preprocessor/headers.go sema/builtin.go codegen/expr.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(codegen): lower source va_arg to runtime varargs"
git push
```

## Milestone 3: `va_copy`, Multiple `va_list`, And Nested Varargs

**Calibration:** Completed after Milestone 2. `va_copy` had no header, sema builtin, codegen, or runtime support. VM frames already kept a cursor per va-list slot, but `OpVaArg` read `activeVaList`, so multiple live source-level `va_list` variables were not independent.

**Files:**
- Modify: `bytecode/opcode.go`
- Modify: `bytecode/printer.go`
- Modify: `bytecode/printer_test.go`
- Modify: `bytecode/validator.go`
- Modify: `runtime/vm.go`
- Modify: `preprocessor/headers.go`
- Modify: `sema/builtin.go`
- Modify: `codegen/expr.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing multiple-list and `va_copy` tests**

Added `TestGCCSourceVaArgMultipleListsExecuteIndependently` and `TestGCCSourceVaCopyExecutesThroughRuntime`.

- [x] **Step 2: Make source-level `va_arg` slot-aware**

Changed codegen to emit the `va_list` slot on `OpVaArg`, and changed the VM to advance that slot instead of always using the last active `va_list`.

- [x] **Step 3: Add `OpVaCopy`**

Added bytecode, printer, validator, VM, header, sema builtin, and codegen support for `va_copy(dst, src)`.

- [x] **Step 4: Verify, commit, and push**

Run Common Verification, then:

```bash
git add bytecode/opcode.go bytecode/printer.go bytecode/printer_test.go bytecode/validator.go runtime/vm.go preprocessor/headers.go sema/builtin.go codegen/expr.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): support source va_copy"
git push
```

## Milestone 4: Formatted Input Runtime

**Calibration:** In progress after Milestone 3. `printf`/`vprintf` output formatting is broad, and runtime already has memory string reads plus stream input primitives (`fgetc`, `ungetc`, `fgets`, `fread`). Formatted input was absent from `<stdio.h>`, sema, and extern registration. The milestone is split so the first coherent increment builds a deterministic `sscanf` core before stream-backed `scanf`/`fscanf`.

**Files:**
- Modify: `preprocessor/headers.go`
- Modify: `sema/builtin.go`
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing `sscanf` registry, direct extern, and GCC runtime tests**

Added coverage for `<stdio.h>` declaration/sema visibility and runtime execution of `%d %u %5s %c`.

- [x] **Step 2: Implement bounded `sscanf` core**

Added `<stdio.h>` declaration, sema builtin signature, extern registration, and runtime scanning for whitespace/literal matching plus `%d`, `%i`, `%u`, `%s`, `%c`, and `%%`.

- [x] **Step 3: Verify, commit, and push**

Run Common Verification, then:

```bash
git add preprocessor/headers.go sema/builtin.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): add bounded sscanf extern"
git push
```

- [x] **Step 4: Calibrate next formatted-input slice**

Chose stream-backed `scanf`/`fscanf` over wider `sscanf` verbs because existing `readHostChar` and pushback support made a small deterministic increment possible.

- [x] **Step 5: Add stream-backed formatted input**

Added `<stdio.h>` declarations, sema signatures, extern registration, runtime scanning over stdin and hermetic file handles, and tests that confirm unread input remains available to `getchar`/`fgetc`.

- [x] **Step 6: Verify, commit, and push stream scanning**

Run Common Verification, then:

```bash
git add preprocessor/headers.go sema/builtin.go runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): add stream scanf externs"
git push
```

- [x] **Step 7: Calibrate and extend integer scan fidelity**

After stream scanning landed, the next low-risk gap was common integer-format fidelity rather than scansets or floats. Added `%x`, `%X`, `%o`, assignment suppression coverage, and `%n` count writes.

- [x] **Step 8: Verify, commit, and push integer scan fidelity**

Run Common Verification, then:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): extend scanf integer formats"
git push
```

- [x] **Step 9: Close bounded formatted input milestone**

Calibration found no GCC accept fixture pressure requiring scansets, floating input, or pointer input. Milestone 4 is closed as a bounded hosted-runtime surface with `sscanf`, `scanf`, `fscanf`, integer/string/char conversions, assignment suppression, `%n`, and stream pushback preservation. Residual conversions stay documented as deferred surface area.

## Milestone 5: FILE State Model Completion

**Calibration:** In progress after formatted input. Existing coverage already exercises configured files, read/write/append modes, `fseek`, `ftell`, `rewind`, EOF status, close invalidation, and unlocked aliases. The most useful missing state bit is the error indicator: write failures on a read-only `FILE *` trapped the VM instead of returning C-level failure and making `ferror` observable.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing stream error indicator tests**

Added direct and GCC-runtime coverage for writing to a read-only file, checking EOF/failure return, `ferror`, `clearerr`, and subsequent reads.

- [x] **Step 2: Track host stream error indicators**

Added runtime host error state and mapped write failures through `fputc`, `fputs`, and `fwrite` to C-level return values instead of VM traps.

- [x] **Step 3: Verify, commit, and push FILE error state**

Run Common Verification, then:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): track stream error indicator"
git push
```

- [x] **Step 4: Calibrate and cover read-side stream errors**

After write-side error indicators landed, read-side mode violations still looked like EOF. Added coverage for reading from a write-only stream and updated `fgetc`, `fgets`, `fread`, and formatted input to set the error indicator without marking EOF.

- [x] **Step 5: Verify, commit, and push read-side FILE errors**

Run Common Verification, then:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): mark read-side stream errors"
git push
```

- [x] **Step 6: Close FILE state milestone**

Calibration after EOF/error changes found the remaining high-value work is mode semantics (`a+`, append write positioning, update-mode transitions) rather than stream status bits. Milestone 5 is closed for EOF/error/clear/close state and mode expansion moves to Milestone 6.

## Milestone 6: Hermetic File Mode Expansion

**Calibration:** In progress after FILE state model. `r`, `w`, `a`, and `+` are recognized, but append mode only initialized the position at EOF. After an explicit `fseek`, writes in `a`/`a+` still overwrote earlier bytes instead of appending.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing append-position tests**

Added direct and GCC-runtime coverage for opening `a+`, seeking to the beginning, writing, closing, and reading back appended content.

- [x] **Step 2: Track append mode on host files**

Added append-mode state to hermetic file handles and forced writes to the current file end.

- [x] **Step 3: Verify, commit, and push append mode positioning**

Run Common Verification, then:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): honor append-mode write positioning"
git push
```

- [x] **Step 4: Close hermetic file mode milestone**

Calibration found C99 `r`, `w`, `a`, and `+` behavior covered after append write positioning. Strict update-stream read/write sequencing remains a residual limit because it has no current fixture pressure and would require more intrusive per-stream direction state.

## Milestone 7: Environment Runtime

**Calibration:** Completed after file milestones. `getenv` was present as a hermetic stub that always returned null after validating the name pointer. The runtime already keeps deterministic configured state in `ExternRegistry`, so environment variables should be explicit registry entries and should not read the ambient host process environment.

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing configured environment tests**

Added direct and GCC-runtime coverage for configured `getenv`, while preserving the default null result for missing values.

- [x] **Step 2: Add explicit registry environment state**

Added `ExternRegistry.SetEnv` and changed `getenv` to return stable runtime C strings for configured values.

- [x] **Step 3: Verify, commit, and push environment runtime**

Run Common Verification, then:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): add configured getenv"
git push
```

## Milestone 8: Program Termination Semantics

**Calibration:** Completed after environment runtime. `exit`, `_Exit`, and `atexit` were all declared and executable, but `atexit` only accepted callback pointers and returned success. The VM treated every single-frame return as the program entry return, so normal top-level callback execution also needed a small frame-mode distinction before void handlers could be scheduled after program termination.

**Files:**
- Modify: `runtime/errors.go`
- Modify: `runtime/extern.go`
- Modify: `runtime/program.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing termination tests**

Added GCC-runtime coverage showing `atexit` handlers must run in reverse registration order after `main` returns, must run when `exit(code)` terminates the program, and must not run for `_Exit(code)`.

- [x] **Step 2: Execute registered `atexit` handlers**

Added deterministic handler storage on the extern registry, retained the registry on loaded programs, marked `_Exit` statuses as skipping cleanup, and taught `Run` to drain handlers after normal termination. The VM now distinguishes the true entry frame from top-level callback frames so void cleanup callbacks can return normally.

- [x] **Step 3: Verify, commit, and push termination semantics**

Run Common Verification, then:

```bash
git add runtime/errors.go runtime/extern.go runtime/program.go runtime/vm.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "feat(runtime): execute atexit handlers"
git push
```

## Milestone 9: Long Double And Complex Runtime Fidelity Sweep

**Calibration:** Completed after termination semantics. Existing runtime coverage already exercised long double scalar operations, complex object-address arithmetic, complex tgmath, complex by-value arguments/returns, and complex aggregate fields. The highest-value Phase 3 seam was the new source-level varargs path: earlier tests covered integer and pointer `va_arg`, but not long double or complex ABI values. Probing that path exposed that ordinary `4.0L` literals were typed as `double`, so default argument promotion placed `f64` in the variadic argument list and `va_arg(ap, long double)` trapped expecting `flong`.

**Files:**
- Modify: `sema/expr.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing long double and complex vararg ABI tests**

Added runtime coverage for source-level `va_arg(ap, long double)` and `va_arg(ap, __complex__ double)`. The complex case already passed; the long double case failed with `va_arg has type f64, want flong`.

- [x] **Step 2: Preserve floating literal suffix types**

Changed sema floating literal typing so `f` suffixes produce `float`, `l` suffixes produce `long double`, and unsuffixed literals remain `double`. This keeps long double variadic arguments as `flong` unless an explicit conversion changes them.

- [x] **Step 3: Verify, commit, and push ABI sweep**

Run Common Verification, then:

```bash
git add sema/expr.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "fix(sema): preserve long double literal varargs"
git push
```

## Milestone 10: Struct And Union ABI Edge Runtime Sweep

**Calibration:** Completed after long double/complex sweep. Existing runtime coverage already exercises ordinary struct by-value arguments, struct returns, conditional/comma aggregate rvalues, nested structs, union storage, and complex fields inside aggregates. The remaining Phase 3 risk was whether the source-level `va_arg` path preserved aggregate values through the object-address ABI.

**Files:**
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add aggregate vararg ABI coverage**

Added runtime coverage for `va_arg(ap, struct pair)` and `va_arg(ap, union number)`.

- [x] **Step 2: Close sweep as already supported**

Focused tests passed without production changes, confirming the existing object-address call/vararg path preserves struct and union values for this bounded ABI surface.

- [x] **Step 3: Verify, commit, and push aggregate sweep**

Run Common Verification, then:

```bash
git add runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "test(runtime): cover aggregate va_arg values"
git push
```

## Milestone 11: GCC Runtime Fixture Expansion

**Calibration:** Completed after ABI sweeps. A fresh `TestGCCExecutionGapReportIsCurrent` run reports the runtime manifest gap closed: 18 runnable `{ dg-do run }` or `c99_runtime` fixtures, all represented, with no preprocess/parse/sema/codegen/load/runtime/exit mismatches. The broader compile-only `main` scan has 28 entry-point fixtures: 18 manifest fixtures, 5 direct runtime tests already present (`Wdeclaration-after-statement-4.c`, `pr27639.c`, `pr71969-1.c`, `pr71969-3.c`, `Wstrict-aliasing-bogus-vla-1.c`), and 5 documented non-runtime targets (`inline-8.c`, `inline-10.c`, `overflow-2.c`, `pr70418.c`, `transparent-union-1.c`).

**Files:**
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Re-run GCC runtime gap report**

Ran:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1 -v
```

It passed without regenerating `runtime/testdata/gcc-exec/gap-report.md`.

- [x] **Step 2: Re-scan compile-only `main` fixtures**

Confirmed no new low-risk GCC accept fixture needs adding after Phase 3 ABI work.

- [x] **Step 3: Verify, commit, and push fixture expansion closure**

Run Common Verification, then:

```bash
git add docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "docs: close phase 3 gcc fixture sweep"
git push
```

## Milestone 12: Runtime Error And Diagnostics Stabilization

**Calibration:** Completed after fixture expansion. Recent hosted-runtime work added an internal cleanup-control bit to `ExitStatus` so `_Exit` can skip atexit handlers. That bit leaked through the public `Run` return value, making `_Exit(7)` produce an `ExitStatus` that carried internal VM control state in addition to the observable exit code.

**Files:**
- Modify: `runtime/errors.go`
- Modify: `runtime/extern.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/phase3-runtime-gap-map.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md`

- [x] **Step 1: Add failing public-status test**

Added runtime coverage showing `_Exit(7)` should return code `7` without exposing the internal cleanup-control flag.

- [x] **Step 2: Keep cleanup control internal**

Made the atexit-skip flag unexported and cleared it before returning from `Run`, while preserving `_Exit` handler-skipping behavior inside the VM.

- [x] **Step 3: Verify, commit, and push diagnostics stabilization**

Run Common Verification, then:

```bash
git add runtime/errors.go runtime/extern.go runtime/vm.go runtime/gcc_exec_test.go docs/phase3-runtime-gap-map.md docs/superpowers/plans/2026-05-19-phase3-runtime-abi-fidelity.md
git commit -m "fix(runtime): keep exit cleanup state internal"
git push
```

## Milestone 13: CLI Runtime UX

**Calibration:** Pending after diagnostics stabilization.

## Milestone 14: Phase 3 Closure Docs

**Calibration:** Pending after all earlier milestones are implemented or closed.
