# GCC Runtime Execution Fixtures Design

## Purpose

The bytecode runtime now supports loading and running a single binary module.
The next test layer should prove that GCC-derived C fixtures can execute through
the full pipeline:

`preprocessor -> parser -> sema -> codegen -> bytecode binary -> runtime`.

This is not a wholesale import of GCC `gcc.dg`. Most existing GCC fixtures are
compile-time or diagnostic tests, not deterministic runtime tests. The execution
suite should include only cases with clear runtime behavior and stable expected
exit codes.

## Scope

Add a curated runtime execution gate under `runtime/testdata/gcc-exec`.

The initial suite is selected from fixtures that already pass the codegen GCC
bytecode compile gate:

- source path appears in `codegen/testdata/gcc-bytecode-compile.tsv`;
- the source is accepted by the current frontend and codegen;
- the generated binary bytecode can be loaded by `runtime.Load`;
- `runtime.Run` returns a deterministic exit code;
- the case does not require unsupported hosted-library behavior, GCC target
  harness behavior, bit-field runtime support, varargs runtime support, floating
  runtime support, or GCC diagnostics.

## Non-Goals

This task will not:

- read directly from a local GCC checkout such as `/tmp/cvm-gcc-c99`;
- parse full DejaGNU semantics;
- execute every `codegen` GCC fixture;
- add runtime support solely to make one unsuitable GCC fixture pass;
- include skip-heavy manifests that weaken the signal of the runtime gate.

Unsupported or unsuitable cases remain outside the runtime execution manifest
until the runtime implements the required behavior.

## Manifest

Create `runtime/testdata/gcc-exec/manifest.tsv`.

Columns:

1. `path`: repository-relative source path, for example
   `sema/testdata/gcc-c99/accept/example.c`.
2. `exit`: expected process exit code from `main`.
3. `category`: short coverage label, such as `arithmetic`, `global`, `loop`,
   `array`, `call`, or `pointer`.
4. `reason`: one-line explanation of why the fixture is deterministic and what
   it covers.

The manifest should contain no skipped rows in the initial version. A future
skip column may be added only if runtime execution fixtures need to track known
execution gaps without enabling them.

## Runner

Extend `runtime/gcc_exec_test.go` with a table-driven runner:

1. Read `runtime/testdata/gcc-exec/manifest.tsv`.
2. Validate each row has exactly four fields.
3. Reject paths outside the existing GCC fixture roots:
   - `sema/testdata/gcc-c99/accept`
   - `sema/testdata/gcc-c99-extra/accept`
   - `sema/testdata/gcc-c90-as-c99/accept`
4. Read the fixture source.
5. Strip DejaGNU directive-only lines using the same line-preserving behavior as
   the existing GCC fixture runners.
6. Run the full compile-to-runtime path:
   `preprocessor.PreprocessSource`, `parser.Parse`, `sema.AnalyzeWithOptions`
   using GCC-compatible options, `codegen.Generate`, `bytecode.EncodeModule`,
   `runtime.Load`, `runtime.Run`.
7. Assert the returned exit code equals the manifest value.

The runner should fail fast on malformed manifest rows, missing files, duplicate
fixture paths, unsupported fixture roots, and runtime errors.

## First Import Batch

An empirical probe of the current repository GCC accept fixtures shows that the
Phase 1 runtime can execute four deterministic GCC-derived fixtures end-to-end.
The first batch should include all four current runnable cases rather than
padding the manifest with compile-only or unsuitable cases. As runtime support
grows, expand this gate toward a small stable set of roughly 10-25 fixtures.

Prefer future fixtures covering:

- integer arithmetic and conversions;
- local variables;
- global variables;
- branches and loops;
- arrays and local object addressing;
- simple direct function calls;
- simple pointer arithmetic.

Avoid fixtures requiring:

- standard library calls other than already-supported built-ins;
- target-specific assumptions not modeled by `bytecode.DefaultTarget`;
- undefined behavior as the expected result;
- diagnostics or warnings;
- bit-fields, varargs, floating-point execution, inline assembly, atomics, or
  GCC builtins beyond current frontend/runtime support.

## Error Handling

Manifest and fixture setup errors should use `t.Fatalf` with the fixture path and
failing stage. Runtime errors should report both the fixture path and the runtime
trap/error text.

If a fixture compiles but traps at runtime, do not add it to the manifest unless
the trap is itself the intended behavior and the manifest format has been
explicitly extended to support trap expectations.

## Verification

Required checks:

- `go test ./runtime -run TestGCCExecution -count=1`
- `go test ./runtime -count=1`
- `go test ./codegen -run TestGCCBytecodeCompileSuite -count=1`
- `go test ./... -count=1`

The suite is complete when the runtime GCC execution manifest exists, the runner
executes all listed fixtures through binary bytecode loading, and all required
checks pass.
