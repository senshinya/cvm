# Bytecode Runtime Phase 1 Handoff

Date: 2026-05-17

This document records the current state of the bytecode/runtime work so the branch can be resumed from another device.

## Repository State

- Workspace: `/Users/shinya/Downloads/cvm`
- Branch: `codex/bytecode-runtime-phase-1`
- Latest implementation/coverage commit before this handoff document: `9403ff2 fix(sema): model integer imaginary literals`
- Remote: `origin git@github.com:senshinya/cvm.git`
- Upstream: `origin/codex/bytecode-runtime-phase-1`
- Working tree at handoff time: clean
- Base remote branch used for comparison: `origin/main`

To update this work on another device:

```bash
git fetch origin
git switch -c codex/bytecode-runtime-phase-1 origin/codex/bytecode-runtime-phase-1
```

If the local branch already exists, use `git switch codex/bytecode-runtime-phase-1` instead.

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

### Bytecode GCC Compile Coverage

`codegen/testdata/gcc-bytecode-compile.tsv` currently has 232 lines including the header, so 231 fixture entries.

The bytecode compile suite now covers every imported `.c` GCC accept fixture from:

- `sema/testdata/gcc-c99/accept`
- `sema/testdata/gcc-c99-extra/accept`
- `sema/testdata/gcc-c90-as-c99/accept`

The coverage check currently returns no uncovered `.c` fixtures for those roots.

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
- complex tgmath extern dispatch

Runtime execution of complex arithmetic is still incomplete. Current runtime support includes `__builtin_cabs*` externs for object-address complex arguments.
Runtime integration coverage now includes local/static `__builtin_complex` initialization, complex `+`, `-`, `*`, `/`, `*=`, `+=`, `-=`, `/=`, object-return copy-out, double-to-float complex copy, complex float return promoted to complex double, `__builtin_cabs`, `__builtin_cabsf`, complex tgmath `sin`, complex tgmath `exp`, and complex tgmath `pow`.
GCC-derived complex runtime coverage also includes imaginary floating constants such as `-1.0i` and integer imaginary constants such as `1i`.

### GNU Nested Functions

Sema now analyzes GNU nested function bodies in their lexical function scope, and codegen has initial static-chain capture support for direct nested function calls that reference enclosing automatic objects.

The current direct-call support includes:

- scalar captured variables loaded and stored through hidden object-address parameters
- captured VLA object addresses and dynamic size metadata
- transitive forwarding for direct nested-function call chains

Current limits:

- escaping nested function addresses are rejected during codegen; GCC-style trampolines are not implemented
- indirect calls through nested function pointers with captured state are not implemented
- the current support is intentionally scoped to compile validated bytecode for known imported GCC accept fixtures

### Tgmath

`<tgmath.h>` is represented through pseudo functions such as:

- `__cvm_tgmath_sin`
- `__cvm_tgmath_exp`
- `__cvm_tgmath_pow`

Sema preserves argument types for these pseudo calls, and codegen dispatches to concrete synthetic externs:

- real: `__cvm_tgmath_sinf`, `__cvm_tgmath_exp`, `__cvm_tgmath_powl`, etc.
- complex: `__cvm_tgmath_cexp`, `__cvm_tgmath_cpowf`, etc.

Runtime support exists for real math externs and for the currently covered complex `csin*`/`cexp*`/`cpow*` externs. Broader complex tgmath coverage remains a later phase.

## Known Limits

- The interpreter/runtime is intentionally not complete yet.
- Bytecode design is intended to be complete enough for the compiler artifact, but execution support is still catching up.
- Complex runtime execution is not complete.
- Long double runtime memory/operations are still limited in places.
- The imported `.c` GCC accept fixtures from the three tracked roots are covered by bytecode compile validation, but runtime execution coverage is still much smaller.
- Static-chain support is deliberately narrow and rejects escaping capturing nested functions instead of modeling trampolines.

## Suggested Next Work

The compile manifest has caught up with the imported `.c` GCC accept fixtures in the tracked roots. Suggested next directions:

- Expand runtime execution coverage fixture-by-fixture, starting with cases that already compile cleanly and exercise existing runtime opcodes.
- Continue complex and long double runtime work where execution still lags behind bytecode generation.
- Revisit static-chain support if the bytecode format grows an explicit closure/trampoline representation.

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
