# Bytecode Runtime Phase 1 Handoff

Date: 2026-05-18
Updated: 2026-05-19

This document records the current state of the bytecode/runtime work so the branch can be resumed from another device.

## Repository State

- Workspace: `/Users/shinya/Downloads/cvm`
- Branch: `codex/bytecode-runtime-phase-8`
- Latest implementation/coverage commit before this handoff document update: `59cb5d2 docs: record phase 8 gcc fixture recheck`
- Remote: `origin git@github.com:senshinya/cvm.git`
- Upstream: `origin/codex/bytecode-runtime-phase-8`
- Working tree at handoff time: clean
- Base remote branch used for comparison: `origin/main`

To update this work on another device:

```bash
git fetch origin
git switch -c codex/bytecode-runtime-phase-8 origin/codex/bytecode-runtime-phase-8
```

If the local branch already exists, use `git switch codex/bytecode-runtime-phase-8` instead.

## Verification Commands

Use a writable Go build cache when running under Codex sandboxed environments:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./sema -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Most recent verification run passed:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Phase 3 Closure

Phase 3 runtime ABI fidelity is closed on `codex/bytecode-runtime-phase-3`.

Closed Phase 3 milestones:

- Baseline roadmap and gap map.
- Source-level `va_arg` lowering for integer, pointer, long double, complex, struct, and union values.
- Multiple live source-level `va_list` cursors and `va_copy`.
- Bounded formatted input: `sscanf`, `scanf`, `fscanf`, integer/string/char conversions, assignment suppression, and `%n`.
- FILE state: EOF/error/clear/close behavior plus read/write mode error indicators.
- Hermetic file mode expansion for `r`, `w`, `a`, `+`, and append write positioning.
- Explicit configured environment via `ExternRegistry.SetEnv` and `cvm run --env NAME=VALUE`.
- Program termination semantics: `atexit` callbacks run in reverse order for normal termination and `exit`, while `_Exit` skips them.
- Long double/complex ABI sweep, including preservation of `f`/`l` floating literal suffix types before default argument promotion.
- Struct/union ABI sweep through aggregate `va_arg` coverage.
- GCC runtime fixture sweep: manifest gap report remains closed.
- Runtime status/error stabilization: internal atexit cleanup state no longer leaks through `ExitStatus`.
- CLI runtime UX: `cvm run --stdin text` and `--env NAME=VALUE` configure deterministic hosted state.

Residual bounded surfaces after Phase 3:

- Formatted input remains intentionally scoped to integer/string/char conversions; scansets, floating input, pointer input, and exact EOF corner semantics are deferred until a fixture or workflow requires them.
- Update-mode streams do not enforce the C sequencing rule requiring a flush or positioning operation between certain read/write direction changes.
- Capturing GNU nested-function closure pointers follow stack-trampoline lifetime rules; calling one after the enclosing frame returns remains invalid.
- Long double storage and arithmetic continue to use the current binary64-backed approximation inside the runtime model.

## Phase 4 Closure

Phase 4 hosted runtime stdio/input fidelity is closed on `codex/bytecode-runtime-phase-4`.

Closed Phase 4 milestones:

- Baseline Phase 4 roadmap and hosted stdio/input design.
- Formatted input scansets: `%[...]`, `%[^...]`, ranges, width, assignment suppression, and source-level `sscanf` execution coverage.
- Floating formatted input: `%f`, `%e`, `%g`, and `%a` families with `float *`, `double *`, and current runtime `long double *` storage.
- Pointer formatted input: `%p` parses deterministic hexadecimal pointer input and writes pointer-sized runtime values.
- Formatted-input failure returns: empty input or whitespace-only first-conversion input failure returns `EOF`, while matching failures return `0` or the number of prior assignments.
- Hermetic update-mode stream sequencing: invalid read-after-write and write-after-read transitions set the stream error indicator; `fflush`, `fseek`, `rewind`, and read EOF allow the documented transitions.
- Cross-extern scanner coverage: new scanner formats are covered through `scanf`, `fscanf`, and `sscanf`, including unread-byte preservation for stdin and configured files.
- GCC runtime fixture recheck: the directive-based runtime gap report remains closed with 18 runnable manifest candidates and no failures; no newly unblocked imported scanf fixtures were found.

Residual bounded surfaces after Phase 4:

- Capturing GNU nested-function closure pointers follow stack-trampoline lifetime rules; calling one after the enclosing frame returns remains invalid.
- Long double storage and arithmetic continue to use the current binary64-backed approximation inside the runtime model.
- Locale-specific formatted input, multibyte/wide-character formatted input, native file descriptors, and exact native libc compatibility remain outside the deterministic hosted-runtime model.

## Phase 5 Closure

Phase 5 nested closure lifetime safety is closed on `codex/bytecode-runtime-phase-5`.

Closed Phase 5 milestones:

- Baseline Phase 5 roadmap and nested closure lifetime design.
- Expired GNU nested-function closure pointers are now tracked after the creating frame exits.
- Indirect calls through expired closure pointers now trap with an explicit `expired closure pointer` lifecycle diagnostic instead of falling through to a generic invalid indirect-call target.
- Valid live nested closure calls still pass through existing local, passed-to-callee, VLA capture, and transitive capture runtime coverage.
- GCC runtime fixture recheck remains closed with 18 runnable manifest candidates and no failures; no newly suitable nested-function pointer fixture was found.

Residual bounded surfaces after Phase 5:

- Escaped GNU nested-function pointers remain invalid after the enclosing frame returns, matching stack-trampoline lifetime constraints. Phase 5 improves diagnostics but does not make those pointers callable.
- Long double storage and arithmetic continue to use the current binary64-backed approximation inside the runtime model.
- Locale-specific formatted input, multibyte/wide-character formatted input, native file descriptors, and exact native libc compatibility remain outside the deterministic hosted-runtime model.

## Phase 6 Closure

Phase 6 hosted libc fidelity is closed on `codex/bytecode-runtime-phase-6`.

Closed Phase 6 milestones:

- Configured-file `fgetpos`/`fsetpos` support, including negative position, invalid memory, and standard-stream failure edges.
- `freopen` over hermetic configured files for read, write/truncate, append, update-mode append, and failure-preservation semantics.
- Deterministic `tmpnam` behavior for `tmpnam(NULL)`, caller-provided buffers, sequence uniqueness, and integration with hermetic `fopen` read/write flows.
- `setvbuf` mode validation for `_IOFBF`, `_IOLBF`, `_IONBF`, invalid modes, buffer range checks, NULL buffers, and closed-stream validation.
- `setbuf` source/direct coverage for explicit buffer and NULL buffer usage.
- C-locale multibyte helper hardening for `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, and `wcstombs`, including null/reset paths, truncation, terminators, and high-bit rejection.
- Header/registry recheck for touched stdio/stdlib surfaces.
- GCC runtime gap report recheck; no new low-risk imported GCC fixture was found.

Residual bounded surfaces after Phase 6:

- Multibyte handling intentionally models the C locale only. Stateful encodings, locale-specific behavior, and wide-character classification remain out of scope.
- `tmpnam` uses deterministic hermetic names and does not model native filesystem races or host temporary directories.
- File streams remain hermetic registry-backed streams, not native file descriptors.
- Long double storage and arithmetic continue to use the current binary64-backed approximation inside the runtime model.

## Phase 7 Closure

Phase 7 string and memory fidelity is closed on `codex/bytecode-runtime-phase-7`.

Closed Phase 7 milestones:

- Hardened `memchr`, `memcmp`, `strncmp`, and `strnlen` zero-length, bounded-read, and unsigned-byte behavior.
- Covered `strchr`, `strrchr`, `strstr`, `strpbrk`, `strspn`, and `strcspn` search/span boundaries.
- Covered `strtok` delimiter changes, exhaustion, all-delimiter input, and restart behavior.
- Covered `strxfrm`, `strcoll`, and deterministic per-memory `strerror` storage.
- Hardened `memccpy`, `bcopy`, `memset`, and `bzero` edge cases.
- Covered string write helpers and checked string/memory builtin object-size success/failure behavior.

Phase 7 recheck:

- `runtime/testdata/gcc-exec/gap-report.md` remains closed with 18 runnable manifest candidates and no failures.
- Header, sema builtin, registry, direct extern, and source-level runtime coverage were rechecked for the Phase 7 surface.
- Imported GCC accept scans found only warning-only/non-runtime candidates or fixtures already represented in the compile manifest.

Residual bounded surfaces after Phase 7:

- `strerror` remains a deterministic hosted stub returning `"error"` rather than host-specific errno text.
- Checked builtin overflow diagnostics remain split between compile-time GCC accept coverage and runtime direct trap coverage.
- Broader stdlib, time, locale, and ctype edge fidelity is next-phase work.

## Phase 8 Closure

Phase 8 stdlib, time, locale, and ctype fidelity is closed on `codex/bytecode-runtime-phase-8`.

Closed Phase 8 milestones:

- Hardened integer conversion coverage for `strtol`, `strtoul`, `strtoll`, and `strtoull`, including base auto-detection, unsigned negative input, max signed, max unsigned, no-conversion, and `endptr` behavior.
- Covered `atoi`, `atol`, `atoll`, `atof`, `strtod`, `strtof`, and `strtold` parsing, including decimal, exponent, hex float, no-conversion, and current `long double` storage behavior.
- Covered deterministic `rand`/`srand`, `getenv`, and `system` hosted semantics.
- Rechecked `atexit`, `exit`, and `_Exit` handler ordering and cleanup-control behavior.
- Covered deterministic C-locale `setlocale`, deterministic `clock`/`time`, and `difftime` arithmetic.
- Hardened ctype classification and case conversion bounds, preserving EOF while masking non-EOF inputs through unsigned-byte values.

Phase 8 recheck:

- Header declarations, sema builtin typing, extern registry entries, direct extern tests, and source-level runtime tests were rechecked for touched Phase 8 surfaces.
- `runtime/testdata/gcc-exec/gap-report.md` remains closed with 18 runnable manifest candidates and no failures.
- Imported GCC accept scans did not identify a new stable low-risk runtime fixture for the Phase 8 surface.

Residual bounded surfaces after Phase 8:

- Floating conversion continues to use Go `strconv.ParseFloat` over the deterministic C-locale subset; host locale, errno/range diagnostics, and exact native libc overflow behavior remain out of scope.
- `time` and `clock` intentionally return deterministic zero values rather than host time.
- `system` intentionally never invokes a host shell; non-NULL commands fail deterministically.
- Ctype behavior is deterministic byte-oriented C-locale behavior; broad locale-specific classification remains out of scope.

## Completed Work

### Runtime Phase 1

The branch includes a first runtime implementation able to load bytecode modules and execute a growing subset of the VM:

- constants, locals, returns, integer control flow
- static memory operations
- local object addresses
- function calls and variadic local calls
- built-in extern registry and basic host externs
- run command and GCC execution fixture runner
- scoped object cleanup and dynamic memory safety checks

Important commits in this part include:

- `5ce0ce0 feat(runtime): add cvm run command`
- `b04f604 feat(runtime): execute function calls`
- `8847270 feat(runtime): execute local object addresses`
- `3128954 feat(runtime): execute static memory opcodes`
- `33637f7 feat(runtime): execute constants locals and returns`
- `f6351f6 feat(runtime): add built-in extern registry`

### Runtime Phase 2 Environment

The first Phase 2 runtime environment milestone is now complete:

- `7cd8df6 feat(runtime): pass configured argv to main`
  - Adds `runtime.LoadOptions.Args`.
  - Passes configured `argc`/`argv` into integer-returning `main(int, char **)`.
  - Updates `cvm run file.cvmbc [args...]` so `argv[0]` is the bytecode path and trailing CLI arguments are visible to the program.
- `82425df feat(runtime): read configured stdin streams`
  - Adds `DefaultExternRegistryWithIO`/`NewExternRegistryWithIO`.
  - Lets `getchar`, `fgetc`, `fgets`, and `fread` consume configured stdin bytes after any `ungetc` pushback.
  - Adds direct extern and GCC runtime execution coverage for configured stdin.
- `2971ab4 feat(runtime): reject closed host streams`
  - Tracks closed host stream handles.
  - Makes `fclose(stdout)` invalidate the handle and reject later writes through the same stream.
  - Preserves existing GCC runtime execution coverage for successful `fclose(stdout)`.

### Runtime Phase 2B `va_list`, File Streams, And Nested Trampolines

The larger Phase 2B runtime surfaces have been advanced with bounded, hermetic semantics:

- `53138f0 feat(runtime): consume memory backed va lists`
  - Adds a CVM-owned memory `va_list` layout consumed by v-format externs.
  - Covers integer and pointer payloads through direct extern tests and GCC runtime execution.
- `f1a1a12 test(runtime): cover memory va lists in vformat aliases`
  - Covers checked buffer v-format and FILE/stdout v-format entry points over the same memory layout.
- `4650aec test(runtime): cover floating memory va list formats`
  - Covers double payloads and `%n` count pointer writes.
- `e940095 feat(runtime): open configured read only files`
  - Adds hermetic registry-backed file contents via `ExternRegistry.AddFile`.
  - Supports `fopen(path, "r")` plus `fgetc`/`fread` reads from configured file handles.
- `ee24658 feat(runtime): seek configured read only files`
  - Implements `fseek`, `ftell`, and `rewind` for configured file handles.
- `190dc75 feat(runtime): write configured files`
  - Supports `fopen(path, "w")`, in-memory writes through stdio output helpers, close/reopen, and readback.
- `27d70ef feat(runtime): remove and rename configured files`
  - Implements hermetic `remove`/`rename` over configured registry files.
- `d4c5545 feat(runtime): support hermetic tmpfile streams`
  - Implements anonymous in-memory `tmpfile` streams with write/rewind/read behavior.
- `0d547b8 test(runtime): cover configured file append mode`
  - Covers `fopen(path, "a")` appending to configured registry files.
- `06a90b9 feat(runtime): call capturing nested function pointers`
  - Adds bytecode/runtime closure pointers for GNU nested functions with captured static-chain state.
  - Lets indirect calls through those pointers append hidden captures and invoke the lowered nested-function signature while the creating frame is alive.
  - Covers local closure calls and passing a capturing nested-function pointer to another callee.
- `e363b6d fix(codegen): propagate captures through nested function pointers`
  - Propagates nested callee capture requirements through function address references, not only through direct nested calls.
  - Covers VLA captures and transitive captures through nested-function pointers.

### GCC Runtime Execution Coverage

GCC execution fixture support has been added and gradually expanded.

Important files:

- `runtime/`
- runtime GCC execution test files and manifests
- `docs/runtime` planning/design documents from earlier commits

Recent related commits:

- `235a461 test(runtime): require GCC run fixtures`
- `d16c502 advance GCC runtime execution coverage`
- `2937e3f support GCC math runtime fixtures`
- `67c4f41 include c99 runtime fixture in GCC execution suite`
- `de11b2a support real tgmath runtime externs`

The latest Plan 82 adjustment found no uncovered GCC runtime accept candidates: `runtime/testdata/gcc-exec/gap-report.md` reports 18 runnable fixtures, all currently represented in `runtime/testdata/gcc-exec/manifest.tsv`.

The latest Phase 1 maths completion scan found no remaining expected C99 `<math.h>`, `<complex.h>`, or `<tgmath.h>` gaps: 171 math prototypes, 66 complex prototypes, and 61 tgmath pseudo functions are declared, registered where runtime execution needs them, and represented in direct/runtime coverage.

Plan 90 rechecked the GCC runtime execution gap report and it remains closed. Plan 91 then scanned compile-only GCC accept fixtures with a `main` and selected `sema/testdata/gcc-c90-as-c99/accept/Wdeclaration-after-statement-4.c` as a low-risk direct runtime coverage candidate. The fixture now executes through a focused runtime test without joining `runtime/testdata/gcc-exec/manifest.tsv`, because that manifest is intentionally limited to fixtures with `{ dg-do run }` or `c99_runtime` directives.

Plan 97 rechecked `codegen/testdata/gcc-bytecode-compile.tsv`: it still has 232 lines including the header, and the direct gap check across the three imported GCC accept roots returned no missing `.c` fixtures.

Plans 103-106 scanned long double and complex GCC accept candidates after the VLA increment. No new low-risk entry-point fixture remained: the only long-double `main` fixture is already in the runtime manifest, and complex accept files are either no-entry compile diagnostics or already covered by the direct complex runtime suite. Plan 110 rejected `overflow-2.c` as a runtime candidate after system `cc` also returned exit 1; the fixture is compile-only warning coverage, not an exit-0 runtime test. Plan 118 then added direct runtime coverage for `pr27639.c`, using a fixture-specific higher step limit for its 50000-iteration static array loop.

Plan 111 rescanned stdlib/string/stdio header-vs-registry coverage. The only actionable stdio mismatch was coverage bookkeeping for plain `sprintf` and `snprintf`: the externs were already registered, but the default registry smoke list and GCC runtime execution coverage only named builtin/chk variants. Plan 112 added plain stdio formatting coverage. A later header guard increment fixed the duplicate `size_t` typedef limitation, and the plain formatter runtime test now includes both `<stdio.h>` and `<string.h>`.

Plan 117 rescanned the remaining self-contained compile-only `main` candidates and selected `pr71969-1.c`. Plan 118 added direct runtime coverage for its inline function calls and volatile global increments using the fixture-specific high step-limit helper.

Plan 123 followed up with `pr71969-3.c`, the matching GNU89 inline variant. It also executes cleanly through runtime with the fixture-specific high step-limit helper.

The Phase 1 non-math runtime fixture closure scan then found one remaining low-risk exit-0 candidate: `sema/testdata/gcc-c99-extra/accept/Wstrict-aliasing-bogus-vla-1.c`. It now executes directly through runtime. The same increment added default entry arguments for integer-returning `main(int, char **)`, using `argc = 1` and `argv = {"cvm", NULL}`. The remaining compile-only `main` candidates are classified as non-runtime Phase 1 targets: `inline-8.c` has no runtime entry body, `inline-10.c` is not an exit-0 candidate, `overflow-2.c` intentionally exits 1 when run, `pr70418.c` depends on GNU nested-function/VLA-in-struct extension behavior without a stable runtime assertion, and `transparent-union-1.c` calls unresolved extern declarations.

### Bytecode GCC Compile Coverage

`codegen/testdata/gcc-bytecode-compile.tsv` currently has 232 lines including the header, so 231 fixture entries.

The bytecode compile suite now covers every imported `.c` GCC accept fixture from:

- `sema/testdata/gcc-c99/accept`
- `sema/testdata/gcc-c99-extra/accept`
- `sema/testdata/gcc-c90-as-c99/accept`

The Plan 84 coverage check returned no uncovered `.c` fixtures for those roots.

Notable recent coverage additions:

- C99 array lvalue decay: `c99-array-lval-1/2/3/4/6/7.c`
- C99 bool: `c99-bool-1/3/4.c`
- C99 compound literal and initializers: `c99-complit-1.c`, `c99-init-1.c`, `c99-init-4.c`
- C99 const expr and scope/VLA: `c99-const-expr-1/2/3/4/13/15.c`, `c99-scope-1/2.c`, `c99-vla-1/2.c`
- GNU statement expressions: `stmt-expr-1/2/4.c`
- GNU condition expression and initializers: `gnu-cond-expr-1/2.c`, `init-empty-1/2.c`, `array-const-1/2.c`
- Complex conversion and arithmetic coverage: `Wconversion-complex-c99.c`, `Wconversion-complex-gnu.c`, `complex-4.c`

### Codegen/Sema Fixes Landed

Recent commits at the tip of this branch:

- `0d547b8 test(runtime): cover configured file append mode`
  - Adds direct and GCC runtime coverage for append-mode configured files.

- `d4c5545 feat(runtime): support hermetic tmpfile streams`
  - Adds anonymous in-memory `tmpfile` handles that support write, rewind, and readback.

- `27d70ef feat(runtime): remove and rename configured files`
  - Implements hermetic `remove` and `rename` over registry-backed files.

- `190dc75 feat(runtime): write configured files`
  - Supports `fopen(path, "w")` and registry-backed stdio writes.

- `ee24658 feat(runtime): seek configured read only files`
  - Implements `fseek`, `ftell`, and `rewind` for configured file handles.

- `e940095 feat(runtime): open configured read only files`
  - Adds `ExternRegistry.AddFile` and read-only `fopen` support for hermetic file contents.

- `4650aec test(runtime): cover floating memory va list formats`
  - Covers floating payloads and `%n` writes through memory-backed v-format `va_list` arguments.

- `f1a1a12 test(runtime): cover memory va lists in vformat aliases`
  - Covers checked and FILE/stdout v-format entry points over memory-backed `va_list` arguments.

- `53138f0 feat(runtime): consume memory backed va lists`
  - Adds CVM memory `va_list` decoding for v-format externs.

- `2971ab4 feat(runtime): reject closed host streams`
  - Tracks closed standard stream handles after `fclose`.
  - Rejects later use of the closed stream handle instead of silently writing.

- `82425df feat(runtime): read configured stdin streams`
  - Adds configured stdin reader plumbing to the default extern registry constructors.
  - Covers `getchar`, `fgetc`, `fgets`, and `fread` over configured stdin.

- `7cd8df6 feat(runtime): pass configured argv to main`
  - Adds runtime loader support for configured `argc`/`argv`.
  - Forwards `cvm run` trailing arguments to bytecode programs.

- `36d63f0 test(runtime): execute GCC strict aliasing VLA fixture`
  - Adds direct runtime coverage for `sema/testdata/gcc-c99-extra/accept/Wstrict-aliasing-bogus-vla-1.c`.
  - Extends runtime loading and VM startup to support integer-returning `main(int, char **)` with default `argc`/`argv` entry arguments.
  - Records the Phase 1 non-math runtime fixture closure plan and focused verification.

- `e60581b fix(codegen): fold constant conditional expressions`
  - Folds C99 integer-constant conditional expressions during bytecode generation, avoiding expression-internal branch labels when `sizeof`-based macro dispatch is compile-time decidable.
  - Updates `<math.h>` `signbit` to use the same type-generic `__cvm_math_select1` path as the other classification macros.
  - Adds focused codegen regression coverage plus runtime coverage for `signbit(-0.0f)`, `signbit(-0.0)`, and `signbit(-0.0L)`.

- `8396282 test(runtime): execute GCC GNU89 inline volatile calls`
  - Adds direct runtime coverage for `sema/testdata/gcc-c99-extra/accept/pr71969-3.c`.
  - Exercises repeated volatile side effects under `-fgnu89-inline` fixture semantics.
  - Keeps this inline mode separate from the prior C99 inline coverage increment.

- `f27d944 test(runtime): execute GCC inline volatile calls`
  - Adds direct runtime coverage for `sema/testdata/gcc-c99-extra/accept/pr71969-1.c`.
  - Exercises repeated inline function calls with volatile global side effects.
  - Reuses the fixture-specific step-limit helper introduced for large compile-only runtime candidates.

- `d009ae6 test(runtime): cover plain stdio sprintf externs`
  - Adds plain `sprintf` and `snprintf` to the default extern registry smoke list.
  - Adds GCC runtime execution coverage through `<stdio.h>` for plain `sprintf`/`snprintf`.
  - Confirms plain stdio formatter externs share the existing builtin formatter implementation path.
- `55b3876 test(runtime): include string header in plain sprintf coverage`
  - Updates the plain formatter runtime source to include both `<stdio.h>` and `<string.h>`.
  - Confirms the shared guarded `size_t` typedef path is exercised during runtime compile-and-execute coverage.
- `506b766 test(runtime): combine stdlib and string headers`
  - Updates plain allocation runtime coverage to include both `<stdlib.h>` and `<string.h>`.
  - Uses `strlen`/`strcmp` on `strdup` results, adding a second shared `size_t` header composition path through runtime execution.

- `43f0676 test(runtime): execute GCC large static array loop`
  - Adds direct runtime coverage for `sema/testdata/gcc-c99-extra/accept/pr27639.c`.
  - Extends the GCC runtime test helper with fixture-specific step limits while keeping the default budget unchanged.
  - Covers a large static array write loop through bytecode/runtime execution.

- `dac4a68 test(runtime): execute VLA memcpy dynamic size`
  - Adds direct runtime coverage inspired by `Wstrict-aliasing-bogus-vla-1.c`.
  - Exercises local VLA allocation, dynamic `sizeof(*x)` sizing, and `__builtin_memcpy` over the resulting dynamic byte count.
  - Complements existing runtime coverage for `vla-2.c`-style VLA struct/union members and `vla-26.c`-style VLA parameter dynamic strides.
- `d8b98e6 test(runtime): execute VLA typedef pointer access`
  - Adds direct runtime coverage inspired by `c99-vla-1.c`.
  - Exercises a typedef naming a VLA, taking its address through a pointer-to-VLA, and indexing through that pointer with dynamic bounds.

- `eaf0dc4 test(runtime): execute GCC declaration-after-statement fixture`
  - Adds a direct runtime coverage test for `sema/testdata/gcc-c90-as-c99/accept/Wdeclaration-after-statement-4.c`.
  - Adds a helper for running compile-only GCC accept fixtures through the runtime without relaxing the `{ dg-do run }` manifest invariant.
  - Covers declaration-after-statement execution with `abort`/`exit` fixture assertions.

- `4a230d9 feat(runtime): add math modf externs`
  - Extends builtin `<math.h>` with plain pointer-output `modff`, `modf`, and `modfl` declarations.
  - Registers `modff`, `modf`, and `modfl` runtime externs backed by a pointer-output math helper, plus `<tgmath.h>` `modf` pseudo-function plumbing.
  - Adds direct extern coverage for integer-part write-back and GCC runtime execution coverage through `<math.h>`.

- `def1cec feat(runtime): add math nan variants`
  - Extends builtin `<math.h>` with plain `nanf`, `nan`, and `nanl` declarations.
  - Registers `nanf` and `nanl` alongside the existing `nan` runtime extern, returning NaN with the correct float rank.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `a2129a6 feat(runtime): add complex pow externs`
  - Extends builtin `<complex.h>` with plain `cpow*` declarations.
  - Registers `cpow`, `cpowf`, and `cpowl` runtime externs backed by the existing complex-binary helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `97a5e7e feat(runtime): add complex exp log sqrt externs`
  - Extends builtin `<complex.h>` with plain `cexp*`, `clog*`, and `csqrt*` declarations.
  - Registers `cexp`, `cexpf`, `cexpl`, `clog`, `clogf`, `clogl`, `csqrt`, `csqrtf`, and `csqrtl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `503b1ff feat(runtime): add complex inverse hyperbolic externs`
  - Extends builtin `<complex.h>` with plain `casinh*`, `cacosh*`, and `catanh*` declarations.
  - Registers `casinh`, `casinhf`, `casinhl`, `cacosh`, `cacoshf`, `cacoshl`, `catanh`, `catanhf`, and `catanhl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `a07454b feat(runtime): add complex inverse trig externs`
  - Extends builtin `<complex.h>` with plain `casin*`, `cacos*`, and `catan*` declarations.
  - Registers `casin`, `casinf`, `casinl`, `cacos`, `cacosf`, `cacosl`, `catan`, `catanf`, and `catanl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `74de548 feat(runtime): add complex hyperbolic externs`
  - Extends builtin `<complex.h>` with plain `csinh*`, `ccosh*`, and `ctanh*` declarations.
  - Registers `csinh`, `csinhf`, `csinhl`, `ccosh`, `ccoshf`, `ccoshl`, `ctanh`, `ctanhf`, and `ctanhl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `0332113 feat(runtime): add complex trig externs`
  - Extends builtin `<complex.h>` with plain `csin*`, `ccos*`, and `ctan*` declarations.
  - Registers `csin`, `csinf`, `csinl`, `ccos`, `ccosf`, `ccosl`, `ctan`, `ctanf`, and `ctanl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `4b685c2 feat(runtime): add complex unary externs`
  - Extends builtin `<complex.h>` with plain `conj*` and `cproj*` declarations.
  - Registers `conj`, `conjf`, `conjl`, `cproj`, `cprojf`, and `cprojl` runtime externs backed by existing complex-unary helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `d126d12 feat(runtime): add complex abs externs`
  - Extends builtin `<complex.h>` with plain `cabs*` declarations.
  - Registers `cabs`, `cabsf`, and `cabsl` runtime externs backed by the existing complex-absolute-value helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `e4a68ee feat(runtime): add complex projection externs`
  - Adds a builtin `<complex.h>` projection surface for `creal*`, `cimag*`, and `carg*`.
  - Registers plain `creal`, `crealf`, `creall`, `cimag`, `cimagf`, `cimagl`, `carg`, `cargf`, and `cargl` externs using the existing complex-real helper path.
  - Adds direct extern coverage and GCC runtime execution coverage through `<complex.h>`.

- `10d93c6 feat(runtime): add plain remquo math externs`
  - Extends builtin `<math.h>` with plain pointer-output `remquo*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `remquo`, `remquof`, and `remquol` runtime externs backed by the existing pointer-output remquo helper.
  - Adds direct extern coverage for quotient write-back and GCC runtime execution coverage through `<math.h>`.

- `6c16d88 feat(runtime): add plain frexp math externs`
  - Extends builtin `<math.h>` with plain pointer-output `frexp*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `frexp`, `frexpf`, and `frexpl` runtime externs backed by the existing pointer-output frexp helper.
  - Adds direct extern coverage for exponent write-back and GCC runtime execution coverage through `<math.h>`.

- `7c9691b feat(runtime): add plain scalbn math externs`
  - Extends builtin `<math.h>` with plain `scalbn*`, `scalbln*`, and `ldexp*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `scalbn`, `scalbnf`, `scalbnl`, `scalbln`, `scalblnf`, `scalblnl`, `ldexp`, `ldexpf`, and `ldexpl` runtime externs backed by the existing real-plus-integer math helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `782e82b feat(runtime): add plain llround math externs`
  - Extends builtin `<math.h>` with plain `llrint*` and `llround*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `llrint`, `llrintf`, `llrintl`, `llround`, `llroundf`, and `llroundl` runtime externs backed by the existing long-returning unary math helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `5cdd27f feat(runtime): add plain lrint lround math externs`
  - Extends builtin `<math.h>` with plain `lrint*` and `lround*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `lrint`, `lrintf`, `lrintl`, `lround`, `lroundf`, and `lroundl` runtime externs backed by the existing long-returning unary math helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `1c96326 feat(runtime): add plain ilogb math externs`
  - Extends builtin `<math.h>` with plain integer-returning `ilogb*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `ilogb`, `ilogbf`, and `ilogbl` runtime externs backed by the existing integer unary math helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `acb3bd0 feat(runtime): add plain gamma math externs`
  - Extends builtin `<math.h>` with plain `tgamma*` and `lgamma*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `tgamma`, `tgammaf`, `tgammal`, `lgamma`, `lgammaf`, and `lgammal` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `2fab716 feat(runtime): add plain rint logb math externs`
  - Extends builtin `<math.h>` with plain `nearbyint*`, `rint*`, and `logb*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `nearbyint`, `nearbyintf`, `nearbyintl`, `rint`, `rintf`, `rintl`, `logb`, `logbf`, and `logbl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `50b6d67 feat(runtime): add plain nextafter math externs`
  - Extends builtin `<math.h>` with plain `nextafter*` and `nexttoward*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `nextafter`, `nextafterf`, `nextafterl`, `nexttoward`, `nexttowardf`, and `nexttowardl` runtime externs backed by existing binary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `c7d93f3 feat(runtime): add plain fma math externs`
  - Extends builtin `<math.h>` with plain `fma*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `fma`, `fmaf`, and `fmal` runtime externs backed by the existing ternary math helper.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `f7569fc feat(runtime): add plain fmod copysign math externs`
  - Extends builtin `<math.h>` with plain `fmod*`, `remainder*`, and `copysign*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `fmod`, `fmodf`, `fmodl`, `remainder`, `remainderf`, `remainderl`, `copysign`, `copysignf`, and `copysignl` runtime externs backed by existing binary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `abf2332 feat(runtime): add plain fdim minmax math externs`
  - Extends builtin `<math.h>` with plain `fdim*`, `fmax*`, and `fmin*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `fdim`, `fdimf`, `fdiml`, `fmax`, `fmaxf`, `fmaxl`, `fmin`, `fminf`, and `fminl` runtime externs backed by existing binary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `0b8daa6 feat(runtime): add plain exp2 log2 math externs`
  - Extends builtin `<math.h>` with plain `exp2*`, `expm1*`, `log1p*`, and `log2*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `exp2`, `exp2f`, `exp2l`, `expm1`, `expm1f`, `expm1l`, `log1p`, `log1pf`, `log1pl`, `log2`, `log2f`, and `log2l` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `8fa882c feat(runtime): add plain cbrt erf math externs`
  - Extends builtin `<math.h>` with plain `cbrt*`, `erf*`, and `erfc*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `cbrt`, `cbrtf`, `cbrtl`, `erf`, `erff`, `erfl`, `erfc`, `erfcf`, and `erfcl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `e5fea5e feat(runtime): add plain inverse hyperbolic math externs`
  - Extends builtin `<math.h>` with plain `asinh*`, `acosh*`, and `atanh*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `asinh`, `asinhf`, `asinhl`, `acosh`, `acoshf`, `acoshl`, `atanh`, `atanhf`, and `atanhl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `d502463 feat(runtime): add plain hyperbolic math externs`
  - Extends builtin `<math.h>` with plain `sinh*`, `cosh*`, and `tanh*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `sinh`, `sinhf`, `sinhl`, `cosh`, `coshf`, `coshl`, `tanh`, `tanhf`, and `tanhl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `d642e49 feat(runtime): add plain inverse trig math externs`
  - Extends builtin `<math.h>` with plain `asin*`, `acos*`, and `atan*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `asin`, `asinf`, `asinl`, `acos`, `acosf`, `acosl`, `atan`, `atanf`, and `atanl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `fcf7c5a feat(runtime): add plain rounding math externs`
  - Extends builtin `<math.h>` with plain `ceil*`, `floor*`, `trunc*`, and `round*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `ceil`, `ceilf`, `ceill`, `floor`, `floorf`, `floorl`, `trunc`, `truncf`, `truncl`, `round`, `roundf`, and `roundl` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `62a3d4e feat(runtime): add plain exp log math externs`
  - Extends builtin `<math.h>` with plain `exp*`, `log*`, and `log10*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `exp`, `expf`, `expl`, `log`, `logf`, `logl`, `log10`, `log10f`, and `log10l` runtime externs backed by existing unary math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `0b723bc feat(runtime): add stdlib multibyte externs`
  - Extends builtin `<stdlib.h>` with `wchar_t` plus C locale `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, and `wcstombs` declarations.
  - Registers single-byte runtime externs for ASCII-oriented multibyte/wide-char conversion, including null reset and length-query cases.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `e01d2ef feat(runtime): add plain binary math externs`
  - Extends builtin `<math.h>` with plain `pow*`, `atan2*`, and `hypot*` declarations while preserving `<tgmath.h>` macro isolation.
  - Registers `pow`, `powf`, `powl`, `atan2`, `atan2f`, `atan2l`, `hypot`, `hypotf`, and `hypotl` runtime externs backed by existing binary math helpers.
  - Adds direct extern coverage, GCC runtime execution coverage through `<math.h>`, and focused tgmath regression coverage.

- `88758d4 feat(runtime): add plain trig math externs`
  - Extends builtin `<math.h>` with plain `sin*`, `cos*`, and `tan*` declarations while keeping `<tgmath.h>` macro expansion isolated from those prototypes.
  - Registers `sin`, `sinf`, `sinl`, `cos`, `cosf`, `cosl`, `tan`, `tanf`, and `tanl` runtime externs backed by existing math helpers.
  - Adds direct extern coverage, GCC runtime execution coverage through `<math.h>`, and focused tgmath regression coverage.

- `555992e feat(runtime): add stdio fpos stubs`
  - Extends builtin `<stdio.h>` with `fpos_t`, `fgetpos`, and `fsetpos`.
  - Registers conservative runtime position stubs that validate stream and position pointers before returning failure.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `82a36df feat(runtime): add string strndup extern`
  - Extends builtin `<string.h>` with `strndup`.
  - Registers a runtime `strndup` extern that copies at most the requested byte count into a newly allocated NUL-terminated string.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `eaa7fac feat(runtime): add string memccpy extern`
  - Extends builtin `<string.h>` with `memccpy`.
  - Registers a runtime `memccpy` extern that copies through the first matching byte and returns the destination position after the match, or null when no match is copied.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `75a63ec feat(runtime): add plain math unary externs`
  - Extends builtin `<math.h>` with `fabs*` and `sqrt*` declarations.
  - Registers plain `fabs`, `fabsf`, `fabsl`, `sqrt`, `sqrtf`, and `sqrtl` runtime externs backed by existing math helpers.
  - Adds direct extern coverage and GCC runtime execution coverage through `<math.h>`.

- `3cd5f06 feat(runtime): add stdio position stubs`
  - Extends builtin `<stdio.h>` with `SEEK_SET`, `SEEK_CUR`, `SEEK_END`, `fseek`, `ftell`, and `rewind`.
  - Registers conservative stream-position stubs: `fseek` and `ftell` report failure, while `rewind` validates the stream and clears EOF state.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `ac75953 feat(runtime): add assert header surface`
  - Adds a builtin `<assert.h>` with `assert` macro behavior for normal and `NDEBUG` modes.
  - Reuses the existing runtime `abort` extern for the assertion failure path.
  - Adds preprocessor coverage and GCC runtime execution coverage through `<assert.h>`.

- `cf901f6 feat(runtime): add errno extern variable`
  - Adds a minimal builtin `<errno.h>` with `EDOM`, `ERANGE`, `EILSEQ`, and `errno`.
  - Resolves `errno` as a stable writable runtime extern variable per memory instance.
  - Adds direct extern-variable coverage and GCC runtime execution coverage through `<errno.h>`.

- `0812e35 feat(runtime): add time header externs`
  - Adds a minimal builtin `<time.h>` with `clock_t`, `time_t`, `CLOCKS_PER_SEC`, `clock`, `difftime`, and `time`.
  - Registers deterministic hermetic runtime externs for `clock`, `difftime`, and `time`.
  - Adds direct extern coverage and GCC runtime execution coverage through `<time.h>`.

- `aa2396e feat(runtime): add locale setlocale extern`
  - Adds a minimal builtin `<locale.h>` with `LC_*` category macros and `setlocale`.
  - Registers a deterministic C-locale `setlocale` extern that returns `"C"` for queries and supported C-locale requests.
  - Adds direct extern coverage and GCC runtime execution coverage through `<locale.h>`.

- `1c5be6d feat(runtime): add stdlib immediate exit extern`
  - Extends builtin `<stdlib.h>` declarations with `_Exit`.
  - Registers `_Exit` as an immediate process-termination extern sharing the existing exit-code path.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `9942e59 feat(runtime): add stdio tmpnam stub`
  - Extends builtin `<stdio.h>` with `L_tmpnam`, `TMP_MAX`, and `tmpnam`.
  - Registers a hermetic runtime `tmpnam` stub that validates an optional writable buffer and returns null.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `15d3b09 feat(runtime): add stdlib atexit extern`
  - Extends builtin `<stdlib.h>` declarations with `atexit`.
  - Registers a phase 1 runtime `atexit` extern that accepts a callback pointer and returns success without executing callbacks.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `65520aa feat(runtime): add string collation externs`
  - Extends builtin `<string.h>` and sema signatures with `strcoll` and `strxfrm`.
  - Registers deterministic C-locale runtime externs for bytewise collation and transform-copy behavior.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `614b50e feat(runtime): add stdio open stubs`
  - Extends builtin `<stdio.h>` declarations with `fopen`, `freopen`, and `tmpfile`.
  - Registers hermetic runtime stubs that validate path/mode strings and return null without host filesystem effects.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `6d84a19 feat(runtime): add stdio file operation stubs`
  - Extends builtin `<stdio.h>` declarations with `remove` and `rename`.
  - Registers hermetic runtime stubs that validate path strings and return `-1`.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `5bed951 feat(runtime): add stdlib div externs`
  - Extends builtin `<stdlib.h>` with `div_t`, `ldiv_t`, `lldiv_t`, `div`, `ldiv`, and `lldiv`.
  - Registers runtime externs that return object-address division result structs for `quot` and `rem`.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `2b25d5b feat(runtime): add string strtok extern`
  - Extends builtin `<string.h>` and sema signatures with `strtok`.
  - Registers a single-threaded runtime `strtok` extern with per-memory continuation state and in-place delimiter splitting.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `e1b4d33 feat(runtime): add stdlib system extern`
  - Extends builtin `<stdlib.h>` and sema signatures with `system`.
  - Registers a hermetic runtime `system` extern that returns `0` for null and `-1` for non-null commands after validation.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `361f7dd feat(runtime): add stdlib getenv extern`
  - Extends builtin `<stdlib.h>` and sema signatures with `getenv`.
  - Registers a hermetic runtime `getenv` extern that validates the variable name and returns null.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `9cf4e9f feat(runtime): add stdlib random externs`
  - Extends builtin `<stdlib.h>` and sema signatures with `RAND_MAX`, `rand`, and `srand`.
  - Registers deterministic registry-local random externs backed by a simple LCG seed.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `6066fc4 feat(runtime): add string strerror extern`
  - Extends builtin `<string.h>` and sema signatures with `strerror`.
  - Registers a deterministic runtime `strerror` extern backed by a per-memory cached `"error"` C string.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `b13063e feat(runtime): add stdlib realloc extern`
  - Extends builtin `<stdlib.h>` and sema signatures with `realloc`.
  - Registers a runtime `realloc` extern that handles null allocation, zero-size free-and-null, and prefix-copy resizing.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `b9b41a4 feat(runtime): add string strnlen extern`
  - Extends builtin `<string.h>` and sema signatures with `strnlen`.
  - Registers a runtime `strnlen` extern that scans byte-by-byte up to the supplied bound and stops early at NUL.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `e572b0b feat(runtime): add stdlib strtof externs`
  - Extends builtin `<stdlib.h>` and sema signatures with `strtof` and `strtold`.
  - Generalizes the existing `strtod` runtime parser over float, double, and long-double return types.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `09acfa0 feat(runtime): add stdio perror extern`
  - Extends builtin `<stdio.h>` declarations with `perror`.
  - Registers a runtime `perror` extern that writes a fixed error message to stderr with an optional prefix.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdio.h>`.

- `91a303b feat(runtime): add strings bsd memory aliases`
  - Adds builtin `<strings.h>` declarations for `bcmp`, `bcopy`, and `bzero`.
  - Registers `bcmp` over the memory compare path and `bcopy` over the memmove-style copy path.
  - Adds direct extern coverage and GCC runtime execution coverage through `<strings.h>`.

- `c6e593f feat(runtime): add stdlib floating parser externs`
  - Extends builtin `<stdlib.h>` and sema signatures with `atof` and `strtod`.
  - Adds decimal/hex floating parsing via longest consumed prefix and writes non-null `endptr`.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `0a79c39 feat(runtime): add stdlib strtoll externs`
  - Extends builtin `<stdlib.h>` and sema signatures with `strtoll` and `strtoull`.
  - Reuses the integer parser added for `strtol`/`strtoul` with long long return types.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `4a4257f feat(runtime): add stdlib strtol externs`
  - Extends builtin `<stdlib.h>` and sema signatures with `strtol` and `strtoul`.
  - Adds runtime integer parsing with base 0/8/10/16, optional sign, invalid-digit stopping, and non-null `endptr` writes.
  - Adds direct extern coverage and GCC runtime execution coverage through `<stdlib.h>`.

- `be20570 feat(runtime): add string span externs`
  - Extends builtin `<string.h>` and sema signatures with `strspn` and `strcspn`.
  - Registers runtime externs for accept-set and reject-set span counting.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `ddf3556 feat(runtime): add string reverse set search externs`
  - Extends builtin `<string.h>` and sema signatures with `strrchr` and `strpbrk`.
  - Registers runtime externs for reverse character search and accept-set string search.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `fb214fd feat(runtime): add string bounded compare search externs`
  - Extends builtin `<string.h>` and sema signatures with `strncmp` and `memchr`.
  - Registers runtime externs for bounded string compare and byte search.
  - Adds direct extern coverage and GCC runtime execution coverage through `<string.h>`.

- `60c319a feat(runtime): add extra ctype classification externs`
  - Extends builtin `<ctype.h>` declarations with `isblank`, `iscntrl`, `isgraph`, and `ispunct`.
  - Registers ASCII runtime externs for the extra ctype classification helpers.
  - Adds direct extern coverage and GCC runtime execution coverage for those helpers.

- `6b7aa8d feat(runtime): add stdlib process termination surface`
  - Extends builtin `<stdlib.h>` declarations with `exit` and `abort`.
  - Adds direct extern coverage for `exit` and `abort` behavior.
  - Adds GCC runtime execution coverage for `<stdlib.h>` `exit`.

- `945da07 feat(runtime): add plain allocation aliases`
  - Extends builtin `<stdlib.h>` declarations with `malloc`, `calloc`, `free`, and `strdup`.
  - Registers plain allocation aliases over the existing allocation paths and adds `free` support for runtime-allocated blocks.
  - Adds direct extern coverage and GCC runtime execution coverage for plain allocation helpers through `<stdlib.h>`.

- `8f08167 feat(runtime): add plain string write aliases`
  - Extends builtin `<string.h>` declarations with `strcpy`, `stpcpy`, `strcat`, `strncpy`, `stpncpy`, and `strncat`.
  - Locks existing plain string-writing aliases into registry smoke coverage.
  - Adds GCC runtime execution coverage for plain string writes through `<string.h>`.

- `646dc4a feat(runtime): add plain memory operation aliases`
  - Extends builtin `<string.h>` declarations with `memcpy`, `memmove`, `mempcpy`, `memset`, and `bzero`.
  - Registers plain `bzero` over the existing memory-zeroing runtime path and exposes its sema builtin signature.
  - Adds direct extern coverage and GCC runtime execution coverage for plain memory operations through `<string.h>`.

- `6c98dab feat(runtime): add string read-only header surface`
  - Adds builtin `<string.h>` declarations for `strcmp`, `memcmp`, `strlen`, `strchr`, and `strstr`.
  - Locks the existing plain read-only string extern aliases into registry smoke coverage.
  - Adds GCC runtime execution coverage through the `<string.h>` header.

- `ab0cbe3 feat(runtime): add ctype case conversion externs`
  - Extends builtin `<ctype.h>` declarations with `tolower` and `toupper`.
  - Registers ASCII case-conversion externs that leave non-convertible inputs unchanged.
  - Adds direct extern coverage and GCC runtime execution coverage for ctype case conversion helpers.

- `9c42f64 feat(runtime): add ctype classification externs`
  - Adds builtin `<ctype.h>` declarations for ASCII classification helpers.
  - Registers `isdigit`, `isalpha`, `isalnum`, `isspace`, `islower`, `isupper`, `isxdigit`, and `isprint` externs.
  - Adds direct extern coverage and GCC runtime execution coverage for ctype classification helpers.

- `c87fe45 feat(runtime): add stdlib atoi externs`
  - Extends builtin `<stdlib.h>` declarations with `atoi`, `atol`, and `atoll`.
  - Registers simple ASCII decimal parsing externs that skip leading whitespace, accept an optional sign, and stop at the first non-digit.
  - Adds direct extern coverage and GCC runtime execution coverage for stdlib string-to-integer helpers.

- `d5590e9 feat(runtime): add stdlib abs externs`
  - Adds builtin `<stdlib.h>` declarations for `abs`, `labs`, and `llabs`.
  - Registers signed absolute-value externs returning target-shaped `int`, `long`, and `long long` values.
  - Adds direct extern coverage and GCC runtime execution coverage for the absolute-value helpers.

- `a731fd5 feat(runtime): add stdio stream lock externs`
  - Declares `flockfile`, `ftrylockfile`, and `funlockfile` in builtin `<stdio.h>`.
  - Registers no-op stream-lock externs that validate known host stream handles.
  - Adds direct extern coverage and GCC runtime execution coverage for stdio stream lock controls.

- `093bfe8 feat(runtime): add stdio buffer control externs`
  - Declares `_IOFBF`, `_IOLBF`, `_IONBF`, `BUFSIZ`, `setbuf`, and `setvbuf` in builtin `<stdio.h>`.
  - Registers no-op `setbuf` and `setvbuf` externs that validate known host stream handles.
  - Adds direct extern coverage and GCC runtime execution coverage for stdio buffering controls.

- `8cebd7e feat(runtime): add stdio unlocked vformat aliases`
  - Registers `vprintf_unlocked` and `vfprintf_unlocked` over the existing literal-only vformat runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for unlocked vformat calls that do not consume `va_list` values.

- `44395f4 feat(runtime): add stdio puts unlocked alias`
  - Registers `puts_unlocked` as an alias over the existing stdout `puts` runtime path.
  - Declares `puts_unlocked` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for newline-terminated unlocked string output.

- `dea2273 feat(runtime): add stdio fileno externs`
  - Tracks standard-stream file descriptors for host-backed `stdin`, `stdout`, and `stderr` handles.
  - Registers `fileno` and `fileno_unlocked` runtime externs returning 0, 1, and 2 for the standard streams.
  - Declares those names in builtin `<stdio.h>` and adds direct extern plus GCC runtime execution coverage.

- `94f5caf feat(runtime): add remaining stdio unlocked aliases`
  - Registers `fgetc_unlocked` and `fflush_unlocked` as aliases over existing stdio input/flush runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Extends direct extern coverage and GCC runtime execution coverage for unlocked character input and flushing.

- `8361666 feat(runtime): add stdio unlocked status aliases`
  - Registers `ferror_unlocked`, `clearerr_unlocked`, and `feof_unlocked` as aliases over existing stream-status stdio runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for EOF/error status aliases.

- `1697ef2 feat(runtime): add stdio unlocked block io aliases`
  - Registers `fread_unlocked`, `fwrite_unlocked`, and `fgets_unlocked` as aliases over existing block/string stdio runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for unlocked block input/output over host handles and pushback.

- `f6468e7 feat(runtime): read stdio pushback with fread`
  - Teaches `fread` to consume bytes from the interpreter-local `ungetc` pushback stack.
  - Writes read bytes into runtime memory and returns the number of complete elements read.
  - Preserves the empty-input behavior when no pushed-back bytes are available and updates EOF status on exhaustion.
  - Adds direct extern coverage and GCC runtime execution coverage for `fread(stdin)` over pushed-back bytes.

- `b514780 feat(runtime): add stdio character output aliases`
  - Registers `putc`, `putc_unlocked`, and `putchar_unlocked` as aliases over the existing character-output runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for stdout/FILE byte writes through the alias functions.

- `a55215e feat(runtime): add stdio character input aliases`
  - Registers `getc`, `getc_unlocked`, and `getchar_unlocked` as aliases over the existing character-input runtime paths.
  - Declares those aliases in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for pushback consumed through the alias functions.

- `ced51f2 feat(runtime): track stdio EOF status`
  - Tracks per-host-handle EOF status for empty `fgetc`, `fgets`, `getchar`, and nonzero-sized empty `fread`.
  - Teaches `feof` to report the tracked EOF bit.
  - Teaches `clearerr` and successful `ungetc` to clear the EOF bit.
  - Adds direct extern coverage and GCC runtime execution coverage for EOF status transitions.

- `6f433ef feat(runtime): support stdio fgets pushback`
  - Adds first-stage `fgets` runtime extern support over interpreter-local `ungetc` pushback.
  - Writes a NUL-terminated buffer when pushed-back bytes are available and returns null when no bytes are read.
  - Declares `fgets` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for `fgets(stdin)` reading pushed-back bytes.

- `247e777 feat(runtime): support stdio ungetc pushback`
  - Adds `ungetc` runtime extern support for known host handles with an interpreter-local pushback stack.
  - Teaches `fgetc` and `getchar` to consume pushed-back bytes before returning `EOF`.
  - Declares `ungetc` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for `ungetc(stdin)` feeding `fgetc` and `getchar`.

- `8993e78 feat(runtime): execute empty stdio character input`
  - Adds first-stage empty-input `getchar` and `fgetc` runtime extern support.
  - Returns `EOF` for the current no-input model and validates known stream handles for `fgetc`.
  - Defines `EOF` and declares `getchar`/`fgetc` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for empty `<stdio.h>` character input.

- `fe7671f feat(runtime): execute stdio fclose`
  - Adds no-op `fclose` runtime extern support for known host handles.
  - Returns zero for the current host-backed standard streams without modeling stream lifetime.
  - Declares `fclose` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for `<stdio.h>` `fclose(stdout)`.

- `4fd8f48 feat(runtime): execute empty stdio fread`
  - Adds first-stage `fread` runtime extern support for known host handles under the current no-input model.
  - Returns zero and leaves the destination buffer untouched.
  - Declares `fread` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for empty `<stdio.h>` `fread(stdin)`.

- `1bca212 feat(runtime): execute stdio fwrite`
  - Adds `fwrite` runtime extern support for writing contiguous runtime memory bytes to known host handles.
  - Returns the requested element count on full host write success and returns zero for zero size/count without touching memory.
  - Declares `fwrite` in the builtin `<stdio.h>` header.
  - Adds direct extern coverage and GCC runtime execution coverage for `<stdio.h>` `fwrite(stdout)`.

- `c02f791 feat(runtime): execute stdio status functions`
  - Adds `ferror`, `clearerr`, and `feof` runtime extern support for known host handles.
  - Models `ferror`/`feof` as returning `0` and `clearerr` as no-op for the current host-writer streams.
  - Declares `ferror`/`clearerr`/`feof` in the builtin `<stdio.h>` header.
  - Adds GCC runtime execution coverage for `<stdio.h>` stream status calls.

- `b1578a4 feat(runtime): execute stdio fflush`
  - Adds no-op `fflush` runtime extern support for known host handles and null.
  - Declares `fflush` in the builtin `<stdio.h>` header.
  - Adds GCC runtime execution coverage for `<stdio.h>` `fflush(stdout)`.

- `e9be3ca feat(runtime): execute stdio fputc`
  - Adds `fputc` and `fputc_unlocked` runtime extern support and direct `fputc` coverage.
  - Declares `fputc`/`fputc_unlocked` in the builtin `<stdio.h>` header.
  - Adds GCC runtime execution coverage for `<stdio.h>` `fputc(stdout)`.

- `a6c2613 feat(runtime): execute stdio putchar`
  - Adds `putchar` runtime extern support and direct extern coverage.
  - Declares `puts` and `putchar` in the builtin `<stdio.h>` header.
  - Adds GCC runtime execution coverage for `<stdio.h>` `putchar`.

- `83c7e02 feat(preprocessor): declare stdio format functions`
  - Expands the builtin `<stdio.h>` header to declare the standard formatting surface currently modeled by sema/runtime.
  - Adds header coverage for `FILE`, `size_t`, standard streams, `fputs`/`fputs_unlocked`, and printf/vprintf family declarations.

- `d31031d test(runtime): cover no-arg vformat externs`
  - Adds direct runtime extern coverage for `vprintf`/`vfprintf` literal output.
  - Adds checked v-format extern coverage for stdout/FILE and buffer-writing variants.

- `732ea7e feat(runtime): support no-arg vformat externs`
  - Registers `vprintf`/`vfprintf`/`vsprintf`/`vsnprintf` plain and builtin externs, plus checked builtin variants.
  - Executes v-format calls when the format string does not consume `va_list` values, using the shared formatter with an empty argument list.
  - Adds runtime GCC exec coverage for `<stdio.h>` + `<stdarg.h>` `vsprintf`/`vsnprintf` literal formatting.
  - Adds `stdin` extern variable allocation and `fputs_unlocked` output alias support for the builtin `<stdio.h>` header surface.

- `1b95703 feat(runtime): execute va_start externs`
  - Registers no-op runtime externs for `__builtin_va_start` and `__builtin_va_end`.
  - Adds runtime GCC exec coverage for `<stdarg.h>` `va_start`/`va_end` expansion.
  - This does not change the current header-level `va_arg(ap, type)` placeholder.

- `ea84e79 fix(bytecode): reject va_start outside variadic functions`
  - Threads the current function's variadic flag into bytecode stack validation.
  - Keeps `OpVaStart` legal only in variadic functions.
  - Adds validator coverage for accepted variadic va opcodes and rejected non-variadic `va_start`.

- `2a0c35a feat(runtime): execute va_arg opcodes`
  - Allows `OpVaStart`, `OpVaArg`, and `OpVaEnd` through bytecode validation.
  - Preserves variadic extra arguments on runtime frames for local variadic calls.
  - Adds runtime coverage for reading the first extra argument through a bytecode `va_list` cursor.

- `dd45ac7 feat(runtime): support sized printf counts`
  - Preserves printf length modifiers for `%n`.
  - Supports `hh`, `h`, `l`, `ll`, `j`, `z`, and `t` write-count sizes.
  - Adds coverage for byte, short, and long-long count writes.

- `a24af80 feat(runtime): support printf hex floats`
  - Adds `%a` and `%A` hexadecimal floating output to the shared runtime formatter.
  - Covers explicit precision and uppercase/lowercase exponent spellings through sprintf.

- `47aaead feat(runtime): support printf write count`
  - Adds `%n` support to the shared runtime formatter.
  - Writes the current formatted byte count through an `int *` argument.
  - Keeps flags, width, and precision on `%n` explicitly unsupported.

- `7f0665c feat(runtime): support printf dynamic width`
  - Adds `*` dynamic field width and dynamic precision parsing to the shared runtime formatter.
  - Consumes dynamic width/precision arguments before the formatted value.
  - Handles negative dynamic width as left alignment and negative dynamic precision as omitted precision.

- `b51911d feat(runtime): support printf float formats`
  - Adds formatter support for `%f`, `%F`, `%e`, `%E`, `%g`, and `%G`.
  - Uses default precision 6 and honors literal precision for floating conversions.
  - Reuses existing width and sign flag padding for floating output.

- `a9ffcf0 feat(runtime): support printf sign flags`
  - Adds formatter support for `+`, space, and `#` flags.
  - Covers signed decimal prefixes plus alternate hex and octal output.
  - Preserves numeric prefixes before zero padding.

- `35bd436 feat(runtime): support printf precision`
  - Adds literal precision parsing to the shared runtime formatter.
  - Supports string truncation and minimum digit counts for integer conversions.
  - Keeps unsupported conversion precision explicit.

- `e4f3a55 feat(runtime): support printf width flags`
  - Adds basic literal field width parsing for the shared runtime formatter.
  - Supports left alignment with `-` and zero padding with `0` for the currently supported conversions.
  - Keeps unsupported format features explicit rather than silently approximating them.

- `d42200b feat(runtime): support printf pointer format`
  - Extends the shared runtime formatter with `%p` pointer conversion.
  - Uses stable `0x` plus lowercase hexadecimal output and covers return length through the sprintf path.

- `def405f feat(runtime): support printf radix formats`
  - Extends the shared runtime formatter with `%x`, `%X`, and `%o` integer conversions.
  - Adds sprintf-path coverage for lowercase hex, uppercase hex, octal output, and return length.

- `75c3316 feat(runtime): support printf length modifiers`
  - Teaches the shared runtime formatter to skip integer length modifiers such as `l`, `ll`, and `z`.
  - Covers formatted output for `%ld`, `%llu`, and `%zu` through the existing sprintf extern path.
  - Keeps unsupported width/precision handling explicit instead of silently formatting partial semantics.

- `6dce437 feat(runtime): execute checked printf builtins`
  - Registers `__builtin___printf_chk` and `__builtin___fprintf_chk`.
  - Skips the fortify flag argument and reuses the runtime formatter for stdout/FILE output.
  - Adds extern-registry coverage for checked formatted stdout/stderr writes.

- `ebff816 feat(runtime): execute printf builtins`
  - Registers `printf`/`printf_unlocked` and `fprintf`/`fprintf_unlocked` builtin/plain externs.
  - Reuses the runtime formatter for stdout and FILE-like host writer output.
  - Adds extern-registry coverage for formatted stdout/stderr writes.

- `c5f6296 feat(runtime): execute checked sprintf builtins`
  - Registers `__builtin___sprintf_chk` and `__builtin___snprintf_chk`.
  - Reuses the runtime formatter and checks object-size arguments before writing.
  - Adds runtime coverage for checked formatted buffer writes and checked truncating writes.

- `a44cbc8 feat(runtime): execute sprintf builtins`
  - Registers `__builtin_sprintf`/`sprintf` and `__builtin_snprintf`/`snprintf`.
  - Adds a small runtime formatter for `%s`, `%d`, `%i`, `%u`, `%c`, and `%%`.
  - Adds runtime coverage for formatted buffer writes, return lengths, and `snprintf` truncation with NUL termination.

- `6671b32 feat(runtime): execute object size builtins`
  - Registers conservative runtime externs for `__builtin_object_size` and `__builtin_dynamic_object_size`.
  - Returns unknown-size GCC-compatible values: `(size_t)-1` for type 0/1 and `0` for type 2/3.
  - Adds runtime coverage for both object-size builtins across the supported type modes.

- `062b070 feat(runtime): execute allocation builtins`
  - Registers allocation externs for `__builtin_malloc`, `__builtin_calloc`, and `__builtin_strdup`.
  - Allocates writable runtime memory blocks for host allocation calls and copies duplicated strings into VM memory.
  - Adds runtime coverage for writable malloc memory, zero-initialized calloc memory, and duplicated C strings.

- `3e3664c feat(runtime): execute checked string builtins`
  - Registers checked string externs for `__builtin___strcpy_chk`, `__builtin___stpcpy_chk`, `__builtin___strcat_chk`, `__builtin___strncpy_chk`, `__builtin___stpncpy_chk`, and `__builtin___strncat_chk`.
  - Checks object-size arguments before writing and reuses the existing string write paths.
  - Adds runtime coverage for successful checked string copy, concatenation, bounded copy, and bounded concatenation operations.

- `0c2ba8d feat(runtime): execute checked memory builtins`
  - Registers checked memory externs for `__builtin___memcpy_chk`, `__builtin___memmove_chk`, `__builtin___mempcpy_chk`, and `__builtin___memset_chk`.
  - Reuses the memory copy/set runtime helpers and checks the supplied object-size argument before writing.
  - Adds runtime coverage for successful checked copy, move, end-pointer copy, and fill operations.

- `7461104 feat(runtime): execute builtin string writes`
  - Registers string-writing externs for `__builtin_strcpy`, `__builtin_stpcpy`, `__builtin_strcat`, `__builtin_strncpy`, `__builtin_stpncpy`, and `__builtin_strncat`.
  - Adds shared runtime helpers for writing byte slices into VM memory.
  - Adds runtime coverage for copy, copy-end pointer returns, concatenation, bounded copy padding, bounded copy return pointers, and bounded concatenation.

- `30760f0 feat(runtime): execute builtin string search`
  - Registers read-only string externs for `__builtin_strlen`, `__builtin_strchr`, and `__builtin_strstr`.
  - Adds runtime coverage for string length, character search including the terminating NUL, substring search, and null results.

- `b0eb819 feat(runtime): execute builtin memory ops`
  - Registers memory externs for direct `__builtin_memcpy`, `__builtin_memmove`, `__builtin_mempcpy`, `__builtin_memset`, and `__builtin_bzero`.
  - Reuses runtime memory copy/set primitives and preserves C return-pointer behavior for copy/set functions.
  - Adds runtime coverage for copy, overlapping move, end-pointer return, fill, and zeroing behavior.

- `b295b18 feat(runtime): alias builtin abort`
  - Registers `__builtin_abort` as the same runtime trap behavior as `abort`.
  - Extends extern-registry coverage so both abort entry points are checked.

- `b066200 feat(runtime): execute builtin floating constants`
  - Registers direct `__builtin_huge_valf`, `__builtin_huge_val`, and `__builtin_huge_vall` runtime externs.
  - Registers direct `__builtin_nan` plus the C library `nan` entry point.
  - Adds runtime coverage for huge-value and NaN-producing calls.

- `5048fc3 feat(runtime): execute builtin pow`
  - Registers the direct `__builtin_pow` runtime extern using the existing binary floating helper.
  - Adds runtime coverage for a GCC-style direct `__builtin_pow(2.0, 3.0)` call.
  - Keeps tgmath `pow` dispatch unchanged while covering the sema-modeled direct builtin.

- `44ef76c feat(runtime): execute tgmath cproj`
  - Adds `<tgmath.h>` pseudo-function plumbing for complex `cproj`.
  - Registers complex runtime externs with finite identity and infinite projection behavior.
  - Adds runtime and codegen coverage for complex float `cproj`.

- `736b7f3 feat(runtime): execute tgmath carg`
  - Adds `<tgmath.h>` pseudo-function plumbing for complex-to-real `carg`.
  - Registers runtime externs using `atan2(imag, real)` across complex suffix variants.
  - Adds runtime and codegen coverage for complex float `carg`.

- `eddd3e7 feat(runtime): execute tgmath conj`
  - Adds `<tgmath.h>` pseudo-function plumbing for complex `conj`.
  - Registers complex runtime externs for `conj` across float/double/long-double suffix variants.
  - Adds runtime and codegen coverage for complex float `conj`.

- `bd60213 feat(runtime): execute tgmath complex projections`
  - Adds `<tgmath.h>` pseudo-function plumbing for `creal` and `cimag`.
  - Adds complex-to-real projection extern dispatch using the complex argument rank.
  - Adds runtime and codegen coverage for complex float projections.

- `18804b3 feat(runtime): execute tgmath fabs`
  - Adds `<tgmath.h>` pseudo-function plumbing for `fabs`.
  - Dispatches real arguments to `__cvm_tgmath_fabs*` and complex arguments to existing `__builtin_cabs*` externs.
  - Adds runtime/direct extern/codegen coverage for real and complex `fabs`.

- `6b73d87 feat(runtime): execute tgmath remquo`
  - Adds `<tgmath.h>` pseudo-function plumbing for pointer-output `remquo`.
  - Uses rank from the first two floating arguments while ignoring the quotient pointer.
  - Registers runtime externs that write the quotient through `int *` and adds runtime/direct extern/codegen coverage.

- `7cabed4 feat(runtime): execute tgmath ldexp`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-plus-integer `ldexp`.
  - Reuses first-argument rank handling and float-plus-integer runtime extern support.
  - Adds runtime/direct extern/codegen coverage.

- `dabecb1 feat(runtime): execute tgmath frexp`
  - Adds `<tgmath.h>` pseudo-function plumbing for pointer-output `frexp`.
  - Reuses first-argument rank handling for return type and extern suffix selection.
  - Registers runtime externs that write the exponent through `int *` and adds runtime/direct extern/codegen coverage.

- `ce095e4 feat(runtime): execute tgmath long rounding helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for integer-returning `lrint`, `lround`, `llrint`, and `llround`.
  - Adds special sema return typing for long and long-long tgmath pseudo calls.
  - Registers i64 runtime externs and adds runtime/direct extern/codegen coverage with integer-valued inputs.

- `e8ca093 feat(runtime): execute tgmath ilogb`
  - Adds `<tgmath.h>` pseudo-function plumbing for integer-returning `ilogb`.
  - Adds special sema return typing for the `ilogb` pseudo call.
  - Registers integer runtime externs and adds runtime/direct extern/codegen coverage.

- `cb056f1 feat(runtime): execute tgmath scalbn helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `scalbn` and `scalbln`.
  - Reuses first-argument rank handling for return type and extern suffix selection.
  - Registers real float-plus-integer runtime externs and adds runtime/direct extern/codegen coverage.

- `c7d11c6 feat(runtime): execute tgmath nexttoward`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `nexttoward`.
  - Uses first-argument rank for `nexttoward` return type and extern suffix selection.
  - Registers real runtime externs and adds runtime/direct extern/codegen coverage for `nexttoward`.

- `e6d208d feat(runtime): execute tgmath logb`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `logb`.
  - Registers real runtime externs for `logb` across float/double/long-double suffix variants.
  - Adds runtime and direct extern-registry coverage for a stable binary-exponent call.

- `5d8c145 feat(runtime): execute tgmath rint helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `nearbyint` and `rint`.
  - Registers real runtime externs for the rint-family helper functions across float/double/long-double suffix variants.
  - Adds runtime and direct extern-registry coverage for stable integer-valued calls.

- `2cd9508 feat(runtime): execute tgmath special unary helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `erf`, `erfc`, `tgamma`, and `lgamma`.
  - Registers real runtime externs for the special unary helper functions across float/double/long-double suffix variants.
  - Adds runtime and direct extern-registry coverage for stable special unary helper calls.

- `33d86fb feat(runtime): execute tgmath nextafter`
  - Adds `<tgmath.h>` pseudo-function plumbing for the real-only binary `nextafter` macro.
  - Registers real runtime externs for `nextafter`.
  - Adds runtime and direct extern-registry coverage for stable `nextafter` calls.

- `e60231c feat(runtime): execute tgmath fma`
  - Adds `<tgmath.h>` pseudo-function plumbing for the real-only ternary `fma` macro.
  - Extends tgmath pseudo-call arity handling and runtime extern helpers to three floating arguments.
  - Registers real runtime externs and adds runtime/extern-registry coverage for `fma`.

- `5bf83a7 feat(runtime): execute tgmath copysign`
  - Adds `<tgmath.h>` pseudo-function plumbing for the real-only binary `copysign` macro.
  - Registers real runtime externs for `copysign`.
  - Extends binary real helper runtime/extern-registry coverage.

- `b26bd6d feat(runtime): execute tgmath binary real helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only binary helpers `fdim`, `fmax`, `fmin`, `fmod`, and `remainder`.
  - Extends binary tgmath arity handling for this helper family.
  - Registers real runtime externs and adds runtime/extern-registry coverage for the helper calls.

- `22c8c7a feat(runtime): execute tgmath exp log helpers`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `exp2`, `expm1`, `log10`, `log1p`, and `log2`.
  - Registers real runtime externs for the exponential/log helper functions across float/double/long-double suffix variants.
  - Adds runtime coverage for stable helper calls plus direct extern-registry coverage.

- `485fecb feat(runtime): execute tgmath rounding functions`
  - Adds `<tgmath.h>` pseudo-function plumbing for real-only `ceil`, `floor`, `trunc`, and `round`.
  - Registers real runtime externs for the rounding functions across float/double/long-double suffix variants.
  - Adds runtime coverage for float and long-double rounding calls plus direct extern-registry coverage.

- `4149979 feat(runtime): execute tgmath cbrt`
  - Adds `<tgmath.h>` pseudo-function plumbing for the real-only `cbrt` macro.
  - Registers real runtime externs for `cbrt` across float/double/long-double suffix variants.
  - Adds runtime coverage for `cbrt(8)` and direct extern-registry coverage.

- `5296871 feat(runtime): execute tgmath hypot`
  - Adds `<tgmath.h>` pseudo-function plumbing for the binary `hypot` macro.
  - Registers real runtime externs for `hypot` across float/double/long-double suffix variants.
  - Adds runtime coverage for `hypot(3, 4)` and direct extern-registry coverage.

- `0d3afe8 feat(runtime): execute tgmath atan2`
  - Adds `<tgmath.h>` pseudo-function plumbing for the binary `atan2` macro.
  - Extends tgmath pseudo-call arity handling beyond `pow`.
  - Registers real runtime externs for `atan2` and adds runtime/extern-registry coverage.

- `f0215eb feat(runtime): execute tgmath atanh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `atanh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `atanh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `atanh`, complex `atanh`, and the direct real extern registry.

- `2c8756a feat(runtime): execute tgmath acosh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `acosh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `acosh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `acosh`, complex `acosh`, and the direct real extern registry.

- `a36041b feat(runtime): execute tgmath asinh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `asinh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `asinh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `asinh`, complex `asinh`, and the direct real extern registry.

- `897d59f feat(runtime): execute tgmath atan`
  - Adds `<tgmath.h>` pseudo-function plumbing for `atan` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `atan`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `atan`, complex `atan`, and the direct real extern registry.

- `b2e42fe feat(runtime): execute tgmath acos`
  - Adds `<tgmath.h>` pseudo-function plumbing for `acos` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `acos`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `acos`, complex `acos`, and the direct real extern registry.

- `10ea324 feat(runtime): execute tgmath asin`
  - Adds `<tgmath.h>` pseudo-function plumbing for `asin` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `asin`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `asin`, complex `asin`, and the direct real extern registry.

- `aeb337f feat(runtime): execute tgmath tanh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `tanh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `tanh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `tanh`, complex `tanh`, and the direct real extern registry.

- `c0c66b9 feat(runtime): execute tgmath cosh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `cosh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `cosh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `cosh`, complex `cosh`, and the direct real extern registry.

- `b61e315 feat(runtime): execute tgmath sinh`
  - Adds `<tgmath.h>` pseudo-function plumbing for `sinh` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `sinh`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `sinh`, complex `sinh`, and the direct real extern registry.

- `5fb4eff feat(runtime): execute tgmath log`
  - Adds `<tgmath.h>` pseudo-function plumbing for `log` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `log`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `log`, complex `log`, and the direct real extern registry.

- `1a81cb5 feat(runtime): execute tgmath tan`
  - Adds `<tgmath.h>` pseudo-function plumbing for `tan` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `tan`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `tan`, complex `tan`, and the direct real extern registry.

- `ed66108 feat(runtime): execute tgmath cos`
  - Adds `<tgmath.h>` pseudo-function plumbing for `cos` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `cos`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `cos`, complex `cos`, and the direct real extern registry.

- `103dec8 feat(runtime): execute tgmath sqrt`
  - Adds `<tgmath.h>` pseudo-function plumbing for `sqrt` through preprocessor, sema, and codegen.
  - Registers real and complex runtime externs for `sqrt`, including float/double/long-double suffix variants.
  - Adds runtime coverage for real `sqrt`, complex `sqrt`, and the direct real extern registry.

- `038471a test(runtime): cover signed unsigned conversions`
  - Adds runtime coverage derived from `Wsign-conversion.c`.
  - Covers signed/unsigned conditional conversions involving `SCHAR_MIN`, `INT_MIN`, and unsigned branches.
  - Covers signed/unsigned boundary conversions through local assignments and function arguments.

- `7873756 test(runtime): cover struct function pointer calls`
  - Adds runtime coverage for calling a function pointer that returns a struct containing a complex field.
  - Adds runtime coverage for passing a struct containing a complex field by value through a function pointer.
  - Confirms by-value struct arguments are copied before callee mutation.

- `e5784fa test(runtime): cover variadic function calls`
  - Adds runtime coverage for direct variadic function calls with extra arguments.
  - Adds runtime coverage for indirect calls through variadic function pointers.

- `08876db test(runtime): cover function designator expressions`
  - Adds runtime coverage for conditional function designators used to initialize function pointers.
  - Adds runtime coverage for function designator assignment and comma expressions consumed as function pointers.

- `521f2d0 fix(sema): decay function designators in returns`
  - Applies function designator decay to conditional operands.
  - Applies function designator decay to return expressions before assignment conversion.
  - Adds runtime coverage for returning a conditional function designator as a function pointer.

- `c2c1a55 fix(sema): bind parameters in function pointer returns`
  - Finds the function-definition parameter list closest to the declared function name inside nested declarators.
  - Fixes parameter binding for definitions such as functions returning function pointers.
  - Adds runtime coverage for returned function pointers called immediately and function pointer parameters called inside callees.

- `7ffd14e test(runtime): cover function pointer member calls`
  - Adds runtime coverage for indirect calls through function pointers stored in struct fields.
  - Adds runtime coverage for static function-pointer field initializers and function-pointer arrays.

- `f1bd6cc test(runtime): cover static pointer initializers`
  - Adds runtime coverage for static pointer fields initialized from static arrays.
  - Adds runtime coverage for static pointer arrays initialized with array addresses and pointer offsets.

- `9b50fb2 test(runtime): cover addressable pointer updates`
  - Adds runtime coverage for pointer compound assignment through pointer array elements.
  - Adds runtime coverage for pointer `++`/`--` through struct fields and array elements.

- `09e04f9 test(runtime): cover initialized pointer field updates`
  - Adds runtime coverage for pointer fields initialized from arrays.
  - Covers compound pointer updates after struct-field initialization.

- `d0696f2 fix(codegen): lower addressable pointer compound assignment`
  - Adds load/add-or-subtract/store lowering for pointer compound assignment on addressable lvalues.
  - Keeps expression results as the updated pointer value.
  - Adds runtime coverage for `+=` and `-=` on a pointer stored in a struct field.

- `0358faa fix(codegen): return stored bit-field expression values`
  - Reloads bit-fields after assignment and compound assignment so expression values reflect stored bit-field values.
  - Reloads pre-increment/pre-decrement bit-fields after storage so truncated/wrapped values are returned.
  - Adds runtime coverage for unsigned bit-field truncation through assignment, compound assignment, and pre-increment.

- `e185413 fix(codegen): lower bit-field compound updates`
  - Adds bit-field-specific compound-assignment lowering using `BitFieldLoad` and `BitFieldStore`.
  - Adds bit-field-specific `++`/`--` lowering without taking bit-field addresses.
  - Adds runtime coverage for unsigned integer bit-field and `_Bool` bit-field compound/incdec updates.

- `6001d68 test(runtime): cover floating assignment and logical expressions`
  - Adds runtime coverage for consuming scalar floating assignment-expression results.
  - Adds runtime coverage for floating `&&`, `||`, and logical negation through bool conversion.

- `5e282a5 test(runtime): cover floating compound expression values`
  - Extends runtime coverage to consume floating compound-assignment expression results.
  - Covers both local-slot `float += double` and addressable `long double` struct-field compound assignment.

- `04a0380 fix(codegen): lower floating increment and decrement`
  - Reuses the shared increment/decrement operation for local-slot scalar values.
  - Extends local and addressable `++`/`--` lowering to floating value types.
  - Adds runtime coverage for `float` local inc/dec and `long double` field inc/dec.

- `323f57c fix(codegen): lower floating compound assignments`
  - Extends compound-assignment lowering from integer-only arithmetic to scalar floating arithmetic.
  - Adds runtime coverage for `long double` compound assignment and `float += double` RHS conversion.
  - Adds runtime coverage for `long double` by-value arguments and returns.

- `ac015c2 test(runtime): cover integer conversion execution`
  - Adds runtime coverage derived from the Wconversion integer fixtures.
  - Covers signed-to-unsigned conversion, unsigned-char narrowing, conditional integer conversions, and converted function arguments.

- `4df3d7b test(runtime): cover VLA execution fixtures`
  - Adds runtime coverage derived from `vla-2.c` for VLA members inside local structs and unions.
  - Adds runtime coverage derived from `vla-26.c` for VLA parameter indexing with dynamic stride metadata.

- `17c7eb5 test(runtime): cover aggregate assignment expression values`
  - Adds runtime coverage for consuming struct assignment-expression results through member access.
  - Adds runtime coverage for passing struct assignment-expression results to by-value functions.

- `c2b6c96 test(runtime): cover aggregate rvalue arguments`
  - Adds runtime coverage for a conditional struct-return rvalue passed directly as a by-value function argument.
  - Adds runtime coverage proving nested structs containing complex fields are copied for by-value arguments.

- `d7d80ee test(runtime): cover tgmath complex dereference arguments`
  - Adds runtime coverage for complex tgmath `exp(*q)` from a pointer dereference.
  - Adds runtime coverage for complex tgmath `pow(*q, scalar)` from a pointer dereference.

- `99217c1 test(runtime): cover complex dereference arguments`
  - Adds runtime coverage for direct `__builtin_cabs(*q)` from a complex pointer dereference.
  - Adds runtime coverage for passing `*q` as a user-defined complex by-value argument.

- `dbf69fd test(runtime): cover complex dereference binary reads`
  - Adds runtime coverage for complex binary expressions reading through pointer dereferences.
  - Adds runtime coverage for complex comparisons reading through pointer dereferences.

- `ca39dd2 fix(codegen): align complex dereference copies conservatively`
  - Uses expression-aware source alignment when copying complex values.
  - Adds runtime coverage for returning and initializing complex values from pointer dereferences.

- `607e6d1 fix(codegen): align complex pointer dereferences conservatively`
  - Uses byte alignment for complex object accesses through pointer dereference expressions.
  - Adds runtime coverage for pointer assignment and compound assignment to complex struct fields.

- `4019119 test(runtime): cover comma aggregate rvalues`
  - Adds runtime coverage for comma-expression struct-return assignment with complex fields.
  - Adds runtime coverage for comma-expression struct-return initialization with complex fields.

- `0f8a090 fix(codegen): assign conditional aggregate rvalues`
  - Saves rvalue aggregate assignment RHS object addresses before emitting the LHS address.
  - Adds runtime coverage for conditional struct-return initialization and assignment with complex fields.

- `c4e3b3c fix(codegen): assign aggregate rvalues`
  - Reuses rvalue object-address emission when copying aggregate assignment RHS values.
  - Adds runtime coverage for assignment from a returned struct containing a complex field.

- `ef0e77f fix(codegen): initialize aggregate return values`
  - Keeps aggregate expression initializers as whole-object initializers instead of first-field scalar initialization.
  - Copies rvalue aggregate initializer sources by value during object initialization.
  - Adds runtime coverage for struct by-value parameters and struct returns containing complex fields.

- `093d8bc fix(codegen): copy object parameters by value`
  - Creates local objects for `TypeObjectAddr` parameters and copies argument object contents at function entry.
  - Routes complex local-slot compound assignment before pointer compound assignment.
  - Adds runtime coverage that mutating complex parameters does not mutate caller objects, including unaligned struct fields.

- `d1f73de test(runtime): cover complex arrow field assignments`
  - Adds runtime coverage for complex `->` field assignment.
  - Adds runtime coverage for complex `->` field compound assignment.

- `187a497 fix(codegen): handle complex field assignments`
  - Preserves unaligned member access information for complex assignment destinations.
  - Uses member-aware access alignment for complex compound assignment loads/stores.
  - Adds runtime coverage for complex struct field assignment and compound assignment.

- `015d31a test(runtime): cover nested complex struct initializers`
  - Adds runtime coverage for nested struct initializers containing complex fields.
  - Adds runtime coverage for nested designated struct initializers containing complex fields.

- `06a0937 test(runtime): cover designated complex struct initializers`
  - Adds runtime coverage for local designated struct initializers with complex fields.
  - Adds runtime coverage for static designated struct initializers with complex fields.

- `8e3392b fix(runtime): handle unaligned complex fields`
  - Tracks unaligned field destinations through initializer address lowering.
  - Lets complex runtime externs read object-address complex arguments with byte alignment.
  - Adds runtime coverage for local and static struct fields containing complex values.

- `44ecd54 test(runtime): cover complex array initializers`
  - Adds runtime coverage for local complex array initializers.
  - Adds runtime coverage for static complex array initializers.

- `28b6881 test(runtime): cover complex by-value arguments`
  - Adds runtime coverage for passing complex values to user-defined functions.
  - Adds runtime coverage for returning complex by-value parameters.

- `a0701bd fix(codegen): materialize builtin complex values`
  - Treats `__builtin_complex(...)` as a complex rvalue in general expression emission.
  - Adds runtime coverage for direct `__builtin_complex` arguments to complex externs.

- `f79bfeb fix(codegen): lower complex comma expressions`
  - Materializes complex comma-expression rvalues by evaluating and popping the left operand before using the right operand.
  - Adds runtime coverage for complex comma expressions with assignment side effects.

- `a0c2b84 test(runtime): cover complex assignment expression values`
  - Adds runtime coverage for simple complex assignment expressions consumed as values.
  - Adds runtime coverage for complex compound assignment expressions consumed as values.

- `4827c91 fix(codegen): lower static complex conditional constants`
  - Lets sema evaluate static conditional constants whose selected branch is `__builtin_complex`.
  - Writes static complex conditional constant initializers into bytecode static data.

- `534abc4 test(runtime): cover complex conditional execution`
  - Adds runtime coverage for complex conditional returns.
  - Adds runtime coverage for complex conditional local initialization.

- `9b12c65 test(runtime): cover long double complex pow`
  - Adds runtime coverage for complex long double tgmath `pow`.

- `b1f7024 test(runtime): cover long double complex execution`
  - Adds runtime coverage for `__builtin_cabsl` on `__complex__ long double`.
  - Adds runtime coverage for complex long double tgmath `exp`.

- `e030206 fix(sema): evaluate complex constant expressions`
  - Extends constant evaluation to carry complex real and imaginary components.
  - Writes static complex constant-expression initializers into bytecode static data.
  - Adds runtime coverage for automatic and static `3.0 + 4.0 * 1.0iF` initializers.

- `97b2f37 test(runtime): cover scalar complex conversions`
  - Extends scalar RHS complex compound runtime coverage across `+=`, `-=`, `*=`, and `/=`.
  - Adds runtime coverage for scalar return conversion into a complex return type.

- `0f79f2d fix(codegen): lower complex compound scalar operands`
  - Materializes scalar RHS operands as complex temporaries for complex compound assignment.
  - Adds runtime coverage for `z += 2.0` and `z -= 0.0`.

- `010b17a fix(codegen): lower explicit complex casts`
  - Avoids recursive rvalue materialization for explicit complex-to-complex casts.
  - Adds runtime coverage for narrowing `(__complex__ float)z`.

- `8224ea3 fix(codegen): lower explicit casts to complex`
  - Treats explicit scalar-to-complex casts as object-backed complex rvalues.
  - Adds runtime coverage for `(__complex__ double)3.0`.

- `ea00552 fix(codegen): lower complex equality`
  - Lowers complex `==` and `!=` by comparing real and imaginary components.
  - Adds runtime coverage for equal complex locals and an unequal `__builtin_complex` rvalue.

- `123312f test(runtime): cover imaginary complex bool conversion`
  - Adds runtime coverage proving `_Bool` conversion checks a nonzero imaginary component.

- `9ab00f8 fix(codegen): convert complex values to scalars`
  - Lowers complex-to-scalar casts by reading the real component instead of bit-casting object addresses.
  - Handles complex-to-bool by checking both real and imaginary components.
  - Adds runtime coverage for local `double d = 3.0i`.

- `9403ff2 fix(sema): model integer imaginary literals`
  - Recognizes integer constants with `i`/`j` suffixes as pure imaginary complex literals.
  - Adds runtime coverage for `1i` initialization as `0 + 1i`.

- `7f8e6a7 fix(codegen): lower imaginary complex literals`
  - Models floating constants with `i`/`j` suffixes as pure imaginary complex literals in sema.
  - Lowers imaginary literals and complex unary minus through object-backed complex rvalues.
  - Adds runtime coverage derived from GCC `complex-4.c` for `1.0 / x * -1.0i`.

- `f69aba2 feat(runtime): execute complex tgmath sin`
  - Registers `__cvm_tgmath_csin*` runtime externs.
  - Adds runtime coverage for complex double `sin` and complex float `sin`.

- `be841cd test(runtime): cover complex tgmath pow return promotion`
  - Extends complex tgmath `pow` runtime coverage to the original GCC-style shape where a helper returns `complex double` from `pow(complex float, float)`.
  - Verifies the caller consumes the promoted complex return object after the callee frame exits.

- `f638c82 fix(runtime): copy object returns before frame exit`
  - Lowers object-address returns through `OpReturnObject`.
  - Runtime copies the designated return object before freeing the callee frame.
  - Adds coverage for complex float returns promoted to complex double in the caller.

- `cabaaaf feat(runtime): execute complex tgmath pow`
  - Registers `__cvm_tgmath_cpow*` runtime externs.
  - Adds runtime coverage for complex double `pow` and complex float `pow`.

- `433a404 feat(runtime): execute complex tgmath exp`
  - Registers `__cvm_tgmath_cexp*` runtime externs.
  - Allows complex-return calls to be consumed as complex initializer sources.
  - Adds runtime coverage for complex double `exp` and complex float `exp`.

- `9964b6f fix(codegen): lower complex compound division`
  - Extends complex compound assignment lowering to `/=`.
  - Adds runtime coverage for complex compound division.

- `8254e14 fix(codegen): lower complex compound add and subtract`
  - Extends complex compound assignment lowering to `+=` and `-=`.
  - Adds runtime coverage using `__builtin_complex` rvalue operands and `__builtin_cabs`.

- `b9f79bf fix(codegen): lower complex binary arithmetic`
  - Lowers complex `+`, `-`, `*`, and `/` by materializing result objects from real/imag component formulas.
  - Supports complex rvalue sources such as `__builtin_complex(...)` and cast-wrapped complex expressions.
  - Adds runtime coverage for complex add/sub/mul/div and keeps GCC complex compile fixtures green.

- `bb13716 test(runtime): cover complex float copy execution`
  - Adds runtime coverage for double-complex to float-complex copy and `__builtin_cabsf`.

- `c159133 test(runtime): cover complex compound multiply execution`
  - Adds runtime coverage for complex `*=` lowering using `__builtin_cabs` to validate the result magnitude.

- `113cc36 fix(codegen): lower builtin complex local initializers`
  - Lowers local `_Complex` initializers from `__builtin_complex(real, imag)` by writing real and imaginary components directly.

- `f5012ad feat(runtime): support builtin complex absolute value`
  - Registers `__builtin_cabs`, `__builtin_cabsf`, and `__builtin_cabsl` runtime externs.
  - Adds execution coverage for `__builtin_cabs` reading a static complex object.

- `f857cb2 fix(codegen): forward transitive nested captures`
  - Propagates capture requirements through direct nested-function call chains.
  - Covers a three-level nested runtime case where the middle function forwards an outer capture to the inner function.

- `38fe90e test(runtime): cover nested VLA capture execution`
  - Adds runtime coverage for a nested function reading and updating an enclosing VLA.

- `b0af489 fix(codegen): load captured nested variables by address`
  - Loads and stores captured scalar variables through the hidden object-address static-chain parameter.
  - Adds runtime coverage for direct nested function calls that read, update, and return captured state.

- `a3da8ef test(codegen): guard GCC bytecode fixture coverage`
  - Adds a codegen package test that fails when imported GCC accept `.c` fixtures from the tracked roots are missing from `gcc-bytecode-compile.tsv`.

- `c565aca fix(codegen): reject escaping capturing nested functions`
  - Rejects taking a function pointer to a capturing GNU nested function with a clear codegen diagnostic.
  - Keeps direct static-chain nested calls supported while avoiding bytecode that would drop captured state.
  - Superseded by `06a90b9`, which added runtime-managed closure/trampoline pointers for this case.

- `34612f9 test(codegen): cover C99 float preprocess fixture`
  - Adds the final imported C99 accept `.c` fixture that was still outside the bytecode compile manifest.

- `e3838e1 fix(codegen): pass static chain captures to nested functions`
  - Adds minimal static-chain support for GNU nested functions that capture enclosing automatic objects.
  - Adds capture discovery and passes static-chain environment objects through direct nested calls.
  - Covers `pr59011.c`.

- `0bb5caa test(codegen): cover pragma-only GCC fixtures`
  - Covers imported accept fixtures whose parser output is intentionally empty after preprocessing.

- `e95a579 test(codegen): cover GCC vector shuffle fixture`
  - Adds compile validation for vector shuffle coverage already supported by codegen.

- `cd7e7bb fix(sema): analyze nested function bodies in scope`
  - Analyzes GNU nested function bodies while the enclosing function scope is still active.

- `c308b7f fix(codegen): lower conditional aggregate initializers`
  - Lowers conditional aggregate initializer cases needed by imported GCC fixtures.

## Current Design Notes

### Bytecode File Format

Earlier in this work, the direction was set to generate a pure binary single-file bytecode artifact that contains all data required by the runtime:

- instruction stream
- signatures
- globals and extern references
- static data
- object layouts
- string/static data
- entry metadata

The design intent is that the runtime can start directly from this single binary file later.

### Complex Values

Complex values are currently lowered as object-address values (`bytecode.TypeObjectAddr`) instead of scalar VM stack pairs.

The current codegen support includes:

- complex locals and parameters represented by object storage or `ObjectAddr` slots
- real scalar to complex initialization
- `__builtin_complex(real, imag)` local/static initialization
- complex to complex assignment/copy with real/imag component casts
- explicit scalar-to-complex and complex-to-complex casts
- complex tgmath extern dispatch

Runtime execution of complex arithmetic is still incomplete. Current runtime support includes `__builtin_cabs*` externs for object-address complex arguments.
Runtime integration coverage now includes local/static `__builtin_complex` initialization, local/static complex array and struct-field initialization including designated and nested struct fields, complex struct field assignment and compound assignment through `.`, `->`, and pointer dereference, complex pointer dereference reads/copies including direct arguments, tgmath arguments, binary expressions, and comparisons, direct `__builtin_complex` arguments, user-defined complex by-value parameters and returns with mutation isolation, struct by-value parameters and returns/assignments containing complex fields including conditional and comma aggregate rvalues, aggregate rvalues passed directly as by-value arguments, aggregate assignment-expression results consumed through member access and by-value calls, nested struct by-value argument mutation isolation, complex `+`, `-`, `*`, `/`, `==`, `!=`, `*=`, `+=`, `-=`, `/=`, scalar RHS complex compound assignment, complex assignment and comma expressions consumed as values, object-return copy-out, double-to-float complex copy, complex float return promoted to complex double, complex conditional returns and local initialization, `__builtin_cabs`, `__builtin_cabsf`, `__builtin_cabsl`, complex tgmath `sin`, complex tgmath `sinh`, complex tgmath `asinh`, complex tgmath `asin`, complex tgmath `acosh`, complex tgmath `acos`, complex tgmath `atan`, complex tgmath `atanh`, complex tgmath `cos`, complex tgmath `cosh`, complex tgmath `tan`, complex tgmath `tanh`, complex tgmath `exp`, complex long double tgmath `exp`, complex tgmath `log`, complex tgmath `pow`, complex long double tgmath `pow`, and complex tgmath `sqrt`.
GCC-derived complex runtime coverage also includes imaginary floating constants such as `-1.0i` and integer imaginary constants such as `1i`.
Complex-to-scalar runtime coverage includes local initialization from imaginary literals, where the real component is selected.
Complex constant-expression coverage includes automatic and static complex initializers with arithmetic over imaginary literals and static conditional initializers selecting `__builtin_complex`.
VLA runtime coverage includes local VLA dynamic object allocation, VLA fields inside local structs and unions, VLA parameter dynamic strides, dynamic-size `__builtin_memcpy` over local VLAs, and nested-function VLA capture cases.
Integer conversion runtime coverage includes Wconversion-derived signed-to-unsigned conversion, unsigned-char narrowing, conditional conversions, function argument conversions, and Wsign-derived signed/unsigned boundary conversions.
C90-as-C99 runtime coverage includes declaration-after-statement fixture execution with `abort`/`exit` assertions.
Static data runtime coverage includes the `pr27639.c` large static character array write loop with a fixture-specific step budget.
Inline runtime coverage includes repeated inline calls with volatile global side effects from both `pr71969-1.c` and the GNU89 inline variant `pr71969-3.c`.
Scalar floating runtime coverage includes `long double` local arithmetic, `long double` by-value arguments and returns, mixed-width floating compound assignment such as `float += double`, floating assignment and compound-assignment expression results for local slots and addressable fields, floating logical expressions through bool conversion, and floating `++`/`--` for local slots and addressable fields.
Direct builtin runtime coverage includes `__builtin_pow`, `__builtin_huge_val*`, `__builtin_nan`, plain `nan`, `nanf`, and `nanl`, plain math helpers `fabs`, `fabsf`, `fabsl`, `sqrt`, `sqrtf`, `sqrtl`, `sin`, `sinf`, `sinl`, `cos`, `cosf`, `cosl`, `tan`, `tanf`, `tanl`, `sinh`, `sinhf`, `sinhl`, `cosh`, `coshf`, `coshl`, `tanh`, `tanhf`, `tanhl`, `asin`, `asinf`, `asinl`, `acos`, `acosf`, `acosl`, `atan`, `atanf`, `atanl`, `asinh`, `asinhf`, `asinhl`, `acosh`, `acoshf`, `acoshl`, `atanh`, `atanhf`, `atanhl`, `cbrt`, `cbrtf`, `cbrtl`, `erf`, `erff`, `erfl`, `erfc`, `erfcf`, `erfcl`, `tgamma`, `tgammaf`, `tgammal`, `lgamma`, `lgammaf`, `lgammal`, `exp`, `expf`, `expl`, `exp2`, `exp2f`, `exp2l`, `expm1`, `expm1f`, `expm1l`, `log`, `logf`, `logl`, `log10`, `log10f`, `log10l`, `log1p`, `log1pf`, `log1pl`, `log2`, `log2f`, `log2l`, `ceil`, `ceilf`, `ceill`, `floor`, `floorf`, `floorl`, `trunc`, `truncf`, `truncl`, `round`, `roundf`, `roundl`, `nearbyint`, `nearbyintf`, `nearbyintl`, `rint`, `rintf`, `rintl`, `logb`, `logbf`, `logbl`, `ilogb`, `ilogbf`, `ilogbl`, `lrint`, `lrintf`, `lrintl`, `lround`, `lroundf`, `lroundl`, `llrint`, `llrintf`, `llrintl`, `llround`, `llroundf`, `llroundl`, `scalbn`, `scalbnf`, `scalbnl`, `scalbln`, `scalblnf`, `scalblnl`, `ldexp`, `ldexpf`, `ldexpl`, `frexp`, `frexpf`, `frexpl`, `modf`, `modff`, `modfl`, `remquo`, `remquof`, `remquol`, `pow`, `powf`, `powl`, `atan2`, `atan2f`, `atan2l`, `hypot`, `hypotf`, `hypotl`, `fdim`, `fdimf`, `fdiml`, `fmax`, `fmaxf`, `fmaxl`, `fmin`, `fminf`, `fminl`, `fmod`, `fmodf`, `fmodl`, `remainder`, `remainderf`, `remainderl`, `copysign`, `copysignf`, `copysignl`, `fma`, `fmaf`, `fmal`, `nextafter`, `nextafterf`, `nextafterl`, `nexttoward`, `nexttowardf`, and `nexttowardl`, plain complex pow helpers `cpow`, `cpowf`, and `cpowl`, plain complex exp/log/sqrt helpers `cexp`, `cexpf`, `cexpl`, `clog`, `clogf`, `clogl`, `csqrt`, `csqrtf`, and `csqrtl`, plain complex inverse hyperbolic helpers `casinh`, `casinhf`, `casinhl`, `cacosh`, `cacoshf`, `cacoshl`, `catanh`, `catanhf`, and `catanhl`, plain complex inverse trigonometric helpers `casin`, `casinf`, `casinl`, `cacos`, `cacosf`, `cacosl`, `catan`, `catanf`, and `catanl`, plain complex hyperbolic helpers `csinh`, `csinhf`, `csinhl`, `ccosh`, `ccoshf`, `ccoshl`, `ctanh`, `ctanhf`, and `ctanhl`, plain complex trig helpers `csin`, `csinf`, `csinl`, `ccos`, `ccosf`, `ccosl`, `ctan`, `ctanf`, and `ctanl`, plain complex unary helpers `conj`, `conjf`, `conjl`, `cproj`, `cprojf`, and `cprojl`, plain complex absolute-value helpers `cabs`, `cabsf`, and `cabsl`, plain complex projection helpers `creal`, `crealf`, `creall`, `cimag`, `cimagf`, `cimagl`, `carg`, `cargf`, and `cargl`, `abort`, `__builtin_abort`, assert-failure expansion through `<assert.h>`, `exit`, `_Exit`, `__builtin_va_start`, `__builtin_va_end`, allocation helpers `__builtin_malloc`, `malloc`, `__builtin_calloc`, `calloc`, `realloc`, `__builtin_strdup`, `strdup`, `strndup`, and `free`, object-size helpers `__builtin_object_size` and `__builtin_dynamic_object_size`, memory operations `__builtin_memcpy`, `memcpy`, `__builtin_memmove`, `memmove`, `__builtin_mempcpy`, `mempcpy`, `memccpy`, `bcopy`, `__builtin_memset`, `memset`, `__builtin_bzero`, and `bzero`, memory comparison helpers `memcmp` and `bcmp`, checked memory operations `__builtin___memcpy_chk`, `__builtin___memmove_chk`, `__builtin___mempcpy_chk`, and `__builtin___memset_chk`, read-only string helpers `__builtin_strlen`, `strlen`, `strnlen`, `strerror`, `__builtin_strchr`, `__builtin_strstr`, `strncmp`, `strcoll`, `memchr`, `strrchr`, `strpbrk`, `strspn`, `strcspn`, and tokenizer helper `strtok`, string-writing helpers `strxfrm`, `__builtin_strcpy`, `strcpy`, `__builtin_stpcpy`, `stpcpy`, `__builtin_strcat`, `strcat`, `__builtin_strncpy`, `strncpy`, `__builtin_stpncpy`, `stpncpy`, `__builtin_strncat`, and `strncat`, checked string-writing helpers `__builtin___strcpy_chk`, `__builtin___stpcpy_chk`, `__builtin___strcat_chk`, `__builtin___strncpy_chk`, `__builtin___stpncpy_chk`, and `__builtin___strncat_chk`, stdlib helpers `abs`, `labs`, `llabs`, `div`, `ldiv`, `lldiv`, `atoi`, `atol`, `atoll`, `atof`, `strtol`, `strtoul`, `strtoll`, `strtoull`, `strtod`, `strtof`, `strtold`, `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, `wcstombs`, `rand`, `srand`, `getenv`, `system`, and `atexit`, locale helper `setlocale`, time helpers `clock`, `difftime`, and `time`, errno extern variable `errno`, ctype helpers `isdigit`, `isalpha`, `isalnum`, `isspace`, `islower`, `isupper`, `isxdigit`, `isprint`, `isblank`, `iscntrl`, `isgraph`, `ispunct`, `tolower`, and `toupper`, stdio output/status helpers `remove`, `rename`, `fopen`, `freopen`, `tmpfile`, `tmpnam`, `fseek`, `ftell`, `rewind`, `fgetpos`, `fsetpos`, `puts`, `puts_unlocked`, `putchar`, `putchar_unlocked`, `getchar`, `getchar_unlocked`, `fputc`, `fputc_unlocked`, `putc`, `putc_unlocked`, `fgetc`, `fgetc_unlocked`, `getc`, `getc_unlocked`, `ungetc`, `fgets`, `fgets_unlocked`, `fputs`, `fputs_unlocked`, `fflush`, `fflush_unlocked`, `fclose`, `fileno`, `fileno_unlocked`, `setbuf`, `setvbuf`, `flockfile`, `ftrylockfile`, `funlockfile`, `perror`, `ferror`, `ferror_unlocked`, `clearerr`, `clearerr_unlocked`, `feof`, `feof_unlocked`, `fwrite`, `fwrite_unlocked`, pushback-backed `fread`, and `fread_unlocked`, buffer formatting helpers `__builtin_sprintf`, `__builtin_snprintf`, plain `sprintf`, plain `snprintf`, `__builtin_vsprintf`, and `__builtin_vsnprintf`, checked buffer formatting helpers `__builtin___sprintf_chk`, `__builtin___snprintf_chk`, `__builtin___vsprintf_chk`, and `__builtin___vsnprintf_chk`, stdout/FILE formatting helpers `__builtin_printf`, `__builtin_printf_unlocked`, `__builtin_fprintf`, `__builtin_fprintf_unlocked`, `__builtin_vprintf`, `vprintf_unlocked`, `__builtin_vfprintf`, and `vfprintf_unlocked`, checked stdout/FILE formatting helpers `__builtin___printf_chk`, `__builtin___fprintf_chk`, `__builtin___vprintf_chk`, and `__builtin___vfprintf_chk`, and shared formatter coverage for integer/pointer/floating formats such as `%ld`, `%llu`, `%zu`, `%x`, `%X`, `%o`, `%p`, `%f`, `%e`, `%g`, `%a`, and sized `%n`, plus literal/dynamic width, literal/dynamic precision, `-`, `0`, `+`, space, and `#` flags. The v-format externs also consume CVM memory-backed `va_list` layouts for integer, pointer, double, and `%n` payloads; ordinary C source-level `va_arg` lowering is still separate from this extern ABI.
Runtime environment coverage includes configured `argc`/`argv` for `main(int, char **)`, `cvm run` argument forwarding, configured stdin consumption for `getchar`, `fgetc`, `fgets`, and `fread` after pushback, and closed standard stream handle rejection after `fclose`.
Bit-field runtime coverage includes simple assignment, compound assignment, and `++`/`--` for integer and `_Bool` bit-fields, including expression values after bit-field truncation/wrapping.
Pointer runtime coverage includes local and addressable pointer compound assignment, initialized pointer fields updated with `+=` and `-=`, pointer array element compound assignment, pointer `++`/`--` through struct fields and array elements, and static pointer field/array initializers with relocations.
Function pointer runtime coverage includes indirect calls through local arrays, struct fields, static struct-field initializers, function pointer parameters, returned function pointers, struct returns and struct by-value arguments through function pointers, function designator initialization/assignment/comma expressions, and return conversion from function designators to function pointers.
Variadic runtime coverage includes direct variadic calls and indirect calls through variadic function pointers with extra arguments, `<stdarg.h>` `va_start`/`va_end` expansion through no-op builtin externs, v-format output using literal-only calls and CVM memory-backed extern `va_list` payloads, plus first-stage `OpVaStart`/`OpVaArg`/`OpVaEnd` execution for local variadic functions reading extra arguments from a runtime frame cursor.

### GNU Nested Functions

Sema now analyzes GNU nested function bodies in their lexical function scope, and codegen/runtime have static-chain capture support for direct nested function calls and runtime-managed closure pointers that reference enclosing automatic objects.

The current support includes:

- scalar captured variables loaded and stored through hidden object-address parameters
- captured VLA object addresses and dynamic size metadata
- transitive forwarding for direct nested-function call chains
- indirect calls through runtime-managed closure pointers while the creating frame is alive
- VLA and transitive capture forwarding through those closure pointers

Current limits:

- closure pointers follow stack-trampoline lifetime rules; using one after the enclosing frame returns remains invalid
- host ABI/native executable trampolines are not serialized or exposed outside the CVM runtime
- the current support is intentionally scoped to validated bytecode/runtime execution inside CVM

### Tgmath

`<tgmath.h>` is represented through pseudo functions such as:

- `__cvm_tgmath_sin`
- `__cvm_tgmath_sinh`
- `__cvm_tgmath_asinh`
- `__cvm_tgmath_asin`
- `__cvm_tgmath_acosh`
- `__cvm_tgmath_acos`
- `__cvm_tgmath_atan`
- `__cvm_tgmath_atanh`
- `__cvm_tgmath_atan2`
- `__cvm_tgmath_hypot`
- `__cvm_tgmath_cbrt`
- `__cvm_tgmath_ceil`
- `__cvm_tgmath_floor`
- `__cvm_tgmath_trunc`
- `__cvm_tgmath_round`
- `__cvm_tgmath_exp2`
- `__cvm_tgmath_expm1`
- `__cvm_tgmath_fdim`
- `__cvm_tgmath_fmax`
- `__cvm_tgmath_fmin`
- `__cvm_tgmath_fmod`
- `__cvm_tgmath_remainder`
- `__cvm_tgmath_copysign`
- `__cvm_tgmath_fma`
- `__cvm_tgmath_nextafter`
- `__cvm_tgmath_nexttoward`
- `__cvm_tgmath_erf`
- `__cvm_tgmath_erfc`
- `__cvm_tgmath_tgamma`
- `__cvm_tgmath_lgamma`
- `__cvm_tgmath_nearbyint`
- `__cvm_tgmath_rint`
- `__cvm_tgmath_logb`
- `__cvm_tgmath_scalbn`
- `__cvm_tgmath_scalbln`
- `__cvm_tgmath_ilogb`
- `__cvm_tgmath_lrint`
- `__cvm_tgmath_lround`
- `__cvm_tgmath_llrint`
- `__cvm_tgmath_llround`
- `__cvm_tgmath_frexp`
- `__cvm_tgmath_modf`
- `__cvm_tgmath_ldexp`
- `__cvm_tgmath_remquo`
- `__cvm_tgmath_log10`
- `__cvm_tgmath_log1p`
- `__cvm_tgmath_log2`
- `__cvm_tgmath_cos`
- `__cvm_tgmath_cosh`
- `__cvm_tgmath_tan`
- `__cvm_tgmath_tanh`
- `__cvm_tgmath_exp`
- `__cvm_tgmath_log`
- `__cvm_tgmath_pow`
- `__cvm_tgmath_sqrt`
- `__cvm_tgmath_fabs`
- `__cvm_tgmath_creal`
- `__cvm_tgmath_cimag`
- `__cvm_tgmath_conj`
- `__cvm_tgmath_carg`
- `__cvm_tgmath_cproj`

Sema preserves argument types for these pseudo calls, and codegen dispatches to concrete synthetic externs:

- real: `__cvm_tgmath_sinf`, `__cvm_tgmath_sinh`, `__cvm_tgmath_asinh`, `__cvm_tgmath_asin`, `__cvm_tgmath_acosh`, `__cvm_tgmath_acos`, `__cvm_tgmath_atan`, `__cvm_tgmath_atanh`, `__cvm_tgmath_atan2`, `__cvm_tgmath_hypot`, `__cvm_tgmath_cbrt`, `__cvm_tgmath_ceil`, `__cvm_tgmath_floor`, `__cvm_tgmath_trunc`, `__cvm_tgmath_round`, `__cvm_tgmath_exp2`, `__cvm_tgmath_expm1`, `__cvm_tgmath_fdim`, `__cvm_tgmath_fmax`, `__cvm_tgmath_fmin`, `__cvm_tgmath_fmod`, `__cvm_tgmath_remainder`, `__cvm_tgmath_copysign`, `__cvm_tgmath_fma`, `__cvm_tgmath_nextafter`, `__cvm_tgmath_nexttoward`, `__cvm_tgmath_erf`, `__cvm_tgmath_erfc`, `__cvm_tgmath_tgamma`, `__cvm_tgmath_lgamma`, `__cvm_tgmath_nearbyint`, `__cvm_tgmath_rint`, `__cvm_tgmath_logb`, `__cvm_tgmath_scalbn`, `__cvm_tgmath_scalbln`, `__cvm_tgmath_ilogb`, `__cvm_tgmath_lrint`, `__cvm_tgmath_lround`, `__cvm_tgmath_llrint`, `__cvm_tgmath_llround`, `__cvm_tgmath_frexp`, `__cvm_tgmath_modf`, `__cvm_tgmath_ldexp`, `__cvm_tgmath_remquo`, `__cvm_tgmath_fabs`, `__cvm_tgmath_carg`, `__cvm_tgmath_creal`, `__cvm_tgmath_cimag`, `__cvm_tgmath_log10`, `__cvm_tgmath_log1p`, `__cvm_tgmath_log2`, `__cvm_tgmath_cos`, `__cvm_tgmath_cosh`, `__cvm_tgmath_tan`, `__cvm_tgmath_tanh`, `__cvm_tgmath_exp`, `__cvm_tgmath_log`, `__cvm_tgmath_sqrtl`, `__cvm_tgmath_powl`, etc.
- complex: `__builtin_cabs`, `__cvm_tgmath_conj`, `__cvm_tgmath_cproj`, `__cvm_tgmath_csinh`, `__cvm_tgmath_casinh`, `__cvm_tgmath_casin`, `__cvm_tgmath_cacosh`, `__cvm_tgmath_cacos`, `__cvm_tgmath_catan`, `__cvm_tgmath_catanh`, `__cvm_tgmath_ccos`, `__cvm_tgmath_ccosh`, `__cvm_tgmath_ctan`, `__cvm_tgmath_ctanh`, `__cvm_tgmath_cexp`, `__cvm_tgmath_clog`, `__cvm_tgmath_csqrt`, `__cvm_tgmath_cpowf`, etc.

Runtime support exists for real math externs and for the Phase 1 complex math extern surface. Plan 80's scan found no remaining C99 plain `<complex.h>` header or runtime registry gap, and the final maths scan confirmed all 66 plain complex names plus all 61 tgmath pseudo functions are covered by declarations, registry entries, and tests. Phase 1 maths is therefore considered complete.

## Known Limits

- The interpreter/runtime is intentionally not complete yet.
- Bytecode design is intended to be complete enough for the compiler artifact, but execution support is still catching up.
- Broader complex runtime execution outside the Phase 1 math extern surface is not complete.
- Broader long double runtime memory/operations outside the Phase 1 math extern surface are still limited in places.
- `va_list` execution has two bounded paths: source-level `va_arg(ap, type)` uses bytecode `OpVa*` over an interpreter-internal frame cursor, while v-format externs consume the CVM memory-backed `va_list` layout.
- Standard input can be configured through the default extern registry constructors or `cvm run --stdin text`. File streams can be configured hermetically through `ExternRegistry.AddFile`, `fopen`, `tmpfile`, `remove`, `rename`, `fseek`, `ftell`, and stdio read/write helpers; direct host filesystem access and interactive terminal behavior are intentionally not modeled.
- `fclose` now invalidates known host stream handles. Broader FILE lifetime behavior beyond standard host streams remains intentionally narrow.
- The imported `.c` GCC accept fixtures from the three tracked roots are covered by bytecode compile validation. Phase 1 runtime execution has also closed the low-risk non-math `main` candidate set; remaining compile-only `main` fixtures are intentionally not runtime targets for the reasons listed above.
- Static-chain support now includes runtime-managed closure/trampoline pointers for capturing GNU nested functions while their creating frame is alive. Calling one after the enclosing frame has returned remains intentionally invalid, matching stack-trampoline lifetime constraints.
- Builtin headers share a guarded `size_t` typedef, so combinations such as `<stdio.h>` plus `<string.h>` analyze without relaxing ordinary C99 typedef redeclaration diagnostics.

## Suggested Next Work

The compile manifest has caught up with the imported `.c` GCC accept fixtures in the tracked roots, Phase 1 maths is complete, Phase 1 non-math low-risk runtime `main` fixture closure is complete, and Phases 2 through 8 are closed. Suggested next directions:

- Scope any further hosted runtime work as a new phase, starting from the residual limits recorded in the Phase 8 gap map.
- Expand runtime execution coverage only when a fixture has stable deterministic semantics or when the required hosted-runtime semantics are explicitly brought into scope.

Recommended workflow:

1. Add one small runtime fixture group or one focused compiler coverage fixture group.
2. Run a focused test:

   ```bash
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -run 'TestGCCBytecodeCompileSuite/<fixture-name>' -count=1
   ```

3. If it fails, find the exact sema/codegen root cause before changing code.
4. After green, run:

   ```bash
   git diff --check
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
   env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
   ```

5. Commit each coherent increment separately.

## Useful Commands

Show current branch state:

```bash
git status --short --branch
git log --oneline --decorate -12
```

List still-uncovered imported GCC accept `.c` fixtures:

```bash
comm -23 \
  <(find sema/testdata/gcc-c99/accept sema/testdata/gcc-c99-extra/accept sema/testdata/gcc-c90-as-c99/accept -maxdepth 1 -type f -name '*.c' | sort) \
  <(tail -n +2 codegen/testdata/gcc-bytecode-compile.tsv | cut -f1 | sort)
```

Run the full test suite:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```
