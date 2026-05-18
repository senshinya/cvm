# Runtime Design

Date: 2026-05-16
Status: Pending written-spec review

## Context

The current cvm pipeline can lower C source into a complete bytecode module and
write that module as a pure binary `.cvmbc` file:

```text
preprocessor -> parser -> sema.Analyze -> codegen.Generate -> bytecode.EncodeModule
```

The bytecode module already carries the data a runtime needs to start a
program: module version, target layout metadata, entry point metadata, globals,
static initialization bytes, relocations, string constants, object layouts,
function signatures, functions, and instructions.

There is no runtime or interpreter package yet. This design adds the first
runtime stage. The runtime may be implemented in phases, but its package
boundaries and core data model should be final enough that later opcode,
extern, and debugging work extends the same structure instead of replacing it.

## Goals

- Load a `.cvmbc` file directly and execute its module entry function.
- Return the integer result of `main` as the process exit code.
- Map `exit(n)` to the same exit status behavior.
- Provide a small but real stack-machine interpreter core.
- Build static data, string data, function pointers, and relocations before
  execution.
- Provide a minimal extern registry with `exit`, `abort`, `puts`, and `fputs`.
- Report unsupported runtime features with clear load or trap errors.
- Keep runtime independent from `sema` and `codegen`; it consumes only
  `bytecode.Module`.
- Layer tests so runtime execution coverage can grow without weakening existing
  compile-only bytecode tests.

## Non-Goals

- Full C standard library support.
- Native host C ABI or dynamic library binding.
- Full floating-point, varargs, VLA, bit-field, and stdio behavior in phase 1.
- Revalidating C language semantics at runtime.
- Optimizing bytecode execution.
- Replacing the bytecode validator.

## Recommended Approach

Use a small but complete interpreter kernel. Phase 1 should create the actual
runtime package, loader, memory model, frame model, extern registry, CLI entry,
and execution loop. It should then implement enough opcode handlers to execute
integer, pointer, static-data, function-call, and control-flow programs.

Unsupported opcode families should trap cleanly. This keeps the runtime usable
early while preserving a stable architecture for later phases.

## Package Architecture

Add a new package:

```text
runtime/
```

The dependency direction is:

```text
main/compiler -> runtime -> bytecode
runtime       -> no sema or codegen dependency
```

The runtime package should expose a small public API:

```go
func Load(r io.Reader, opts LoadOptions) (*Program, error)
func Run(ctx context.Context, p *Program, opts RunOptions) (ExitStatus, error)
```

`Load` decodes and prepares a program. `Run` executes an already loaded program.
Convenience helpers may wrap these two steps, but tests should be able to call
them separately.

### Program

`Program` is the executable form of a decoded module. It owns:

- the validated `*bytecode.Module`;
- global memory blocks;
- string memory blocks;
- function pointer tokens;
- global-address and function-address tables;
- extern bindings;
- entry function metadata.

`Program` represents "loaded and linked, not yet running." It should not own
the operand stack or active call frames.

### VM

`VM` owns execution state:

- operand stack;
- call frame stack;
- current function and program counter;
- memory accessor;
- extern registry;
- run options such as trace hooks or step limits.

`VM` starts at the module entry function. It returns `ExitStatus{Code: n}` when
the entry function returns an integer or when `exit(n)` is called.

### Memory

Use one VM address abstraction for all bytecode pointers. Pointer values are
stored as 64-bit values because the current bytecode target uses 8-byte
pointers. The memory implementation maps those values to blocks and offsets.

Memory block classes:

- global variable blocks;
- readonly string blocks;
- local object blocks owned by call frames;
- dynamic object blocks owned by call frames;
- optional host-handle blocks for later extern variables or stdio objects.

The memory layer must check bounds, alignment, readonly writes, and use-after-
free for dynamic objects. These checks should produce traps with enough context
to identify the failing function, instruction, address, and access size.

### ExternRegistry

Externs are resolved by C ABI import name. Phase 1 includes:

- `exit(int)`: terminate execution with that exit code;
- `abort(void)`: terminate with an abort trap;
- `puts(const char *)`: read a NUL-terminated VM string and write it to the
  configured output writer with a trailing newline;
- `fputs(const char *, FILE *)`: accept only the built-in `stdout` and `stderr`
  host handles described below.

Phase 1 supports only two extern variables: `stdout` and `stderr`. They are not
general C objects; the loader binds them to opaque host-handle blocks that only
the built-in stdio externs understand. All other extern variables remain
unsupported until the hosted runtime phase.

Unresolved externs should be rejected during load when possible. If a binding
depends on a runtime option that is unavailable, calls must fail with a clear
trap rather than silently doing nothing.

## Loading And Startup

Runtime startup follows a fixed sequence:

1. Decode `.cvmbc` with `bytecode.DecodeModule`.
2. Validate that `Module.Entry` names a defined function. A module with
   `entry=none` is not runnable by `cvm run`.
3. Allocate global variable blocks for all `GlobalVar` entries.
4. Copy `Init.Bytes` and apply zero fill according to `Global.Size`.
5. Allocate readonly blocks for `StringConst.Bytes`.
6. Create function pointer tokens for defined and extern functions.
7. Apply `Init.Relocations` by writing pointer-sized addresses using
   `Target.Endian` and `Target.PointerSize`.
8. Resolve extern function globals through `ExternRegistry`.
9. Resolve the built-in `stdout` and `stderr` extern variables to host-handle
   blocks when they are present. Reject other extern variables.
10. Check the entry signature. Phase 1 officially supports no-argument `main`
   returning an integer-compatible type.
11. Create the initial frame and enter the interpreter loop.

The runtime must not infer missing static data from source-level C constructs.
If a required object, relocation, signature, or entry is absent from the
bytecode module, loading should fail.

## Execution Semantics

The interpreter is a typed stack machine following the existing bytecode
instruction model:

- instructions consume and produce operand-stack values;
- local slots are typed scalar storage;
- addressable objects are accessed through VM pointers;
- branches target bytecode labels;
- calls use function signatures and argument counts;
- aggregate values are moved through memory, not as generic stack values.

Phase 1 opcode coverage should include:

- constants and null pointer values;
- stack operations: `Dup`, `Pop`, `Swap`;
- local slot load/store;
- global, string, function, and local-object addresses;
- typed load/store;
- memory copy and memory set;
- byte offsets and pointer arithmetic;
- pointer difference;
- integer binary operations, unary negation, casts, and comparisons;
- labels, unconditional jumps, conditional jumps, and switch;
- direct calls, indirect calls, returns, and void returns;
- minimal extern calls.

Phase 1 may trap on:

- floating-point arithmetic not yet implemented;
- bit-field load/store;
- dynamic object allocation/free when the needed semantics are incomplete;
- varargs opcodes;
- extern variables other than `stdout` and `stderr`;
- host-specific stdio objects.

These traps are acceptable only when their messages identify the unsupported
opcode or runtime feature.

## Exit Semantics

`main` return value defines the process exit code. `exit(n)` uses the same
mechanism and bypasses remaining frames.

The CLI should use `ExitStatus.Code` as the OS process exit code. `abort()`
should not be treated as a normal exit; it should return a trap that the CLI
maps to a nonzero exit code after printing a diagnostic.

## Error Model

Use distinct error categories:

### LoadError

Raised before execution starts:

- invalid file or module;
- missing entry;
- unsupported entry signature;
- invalid relocation;
- unresolved extern;
- unsupported target pointer size or endian mode.

### TrapError

Raised during execution:

- illegal memory access;
- alignment violation;
- readonly write;
- division by zero;
- stack underflow or type mismatch;
- invalid label target;
- unsupported opcode;
- abort trap.

Trap errors should include function name, function ID, PC, opcode, and a compact
call stack where available.

### ExitStatus

Normal program completion:

```go
type ExitStatus struct {
    Code int
}
```

Normal exits are not errors.

## CLI

Add a bytecode execution command:

```text
cvm run file.cvmbc
```

This is preferred over adding another top-level flag because it creates a clear
space for future subcommands such as compile, dump, and trace. Existing flags
such as `--dump-ir`, `--dump-bytecode`, and `--emit-bytecode` can remain
available for the current compiler mode.

The command should:

- open the `.cvmbc` file;
- load the program;
- run it;
- write extern output to stdout by default;
- print load and trap diagnostics to stderr;
- exit with the status returned by `main` or `exit`.

## Testing Strategy

### Runtime Unit Tests

Test the runtime package directly:

- loader rejects missing entry, bad entry signatures, invalid relocations, and
  unresolved externs;
- global bytes, string bytes, and relocations produce the expected memory;
- memory bounds, alignment, readonly, and freed-block checks trap;
- hand-written bytecode modules return expected exit codes.

### Opcode Golden Tests

Create compact module fixtures by opcode family. Each test should assert one of:

- returned exit code;
- final memory value;
- specific trap category and message.

Unsupported opcode tests should be explicit so missing coverage is visible and
intentional.

### Compiler-To-Runtime Tests

Compile C source to `.cvmbc`, load it through runtime, and execute it. Start
with small self-owned fixtures:

- `return 0`;
- integer arithmetic;
- if/else;
- while and for loops;
- function calls and recursion where supported;
- global variables and static initialization;
- string constants and `puts`.

### GCC Execution Fixtures

Do not reuse the existing GCC compile fixture set wholesale. Add a separate
execution fixture layer under `runtime/testdata/gcc-exec/` or a similarly
runtime-owned directory.

Each fixture should declare:

- expected exit code;
- required externs;
- unsupported-feature skip reason when skipped;
- whether the test depends on hosted C library behavior.

Only include GCC fixtures that are deterministic under the cvm target and whose
requirements match implemented runtime features.

## Phasing

### Phase 1: Runnable Core

- `.cvmbc` loader and CLI `cvm run`;
- module entry startup;
- static globals, strings, function tokens, and relocations;
- VM operand stack and call frames;
- integer, pointer, control-flow, local, memory, and call opcodes needed by
  current codegen smoke tests;
- minimal externs: `exit`, `abort`, `puts`, and limited `fputs`;
- hand-written module tests and small compiler-to-runtime tests.

### Phase 2: Complete C Object Semantics

- aggregate copy and return coverage;
- bit-field load/store;
- dynamic objects and VLA lifetime;
- stricter volatile memory behavior;
- broader codegen fixture execution.

### Phase 3: Hosted Runtime Surface

- floating-point execution;
- varargs;
- extern variables;
- general stdio host handles, including `stdin`;
- additional libc shim functions needed by accepted fixtures.

### Phase 4: Execution Coverage And Tooling

- layered GCC execution fixtures;
- trace and single-step hooks;
- call-stack and memory diagnostics;
- optional bytecode execution golden printer for debugging.

## Acceptance Criteria

Phase 1 is complete when:

- `cvm --emit-bytecode out.cvmbc input.c` followed by `cvm run out.cvmbc`
  executes no-argument integer-returning `main`;
- `return n;` from `main` produces process exit code `n`;
- `exit(n)` produces process exit code `n`;
- static globals, string constants, and relocations are observable in execution;
- unsupported externs and unsupported opcodes fail with clear diagnostics;
- runtime unit tests and compiler-to-runtime integration tests pass;
- existing `go test ./... -count=1` remains green.

## Open Decisions Fixed By This Spec

- Phase 1 uses the small complete interpreter-kernel approach.
- Phase 1 supports minimal built-in externs rather than full native ABI binding.
- Entry return values map to OS process exit codes.
- Runtime consumes the pure binary `.cvmbc` file directly.
- Existing bytecode compile tests remain compile-only; runtime execution gets a
  separate fixture layer.
