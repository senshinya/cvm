# Phase 2 Runtime Environment Design

## Goal

Close the next Phase 2 runtime milestone by replacing Phase 1's fixed process environment with a small hosted model for program arguments, stdin-backed input, and stream lifetime.

## Scope

This milestone covers three concrete runtime gaps from the current handoff:

- configurable `argc`/`argv` for integer-returning `main(int, char **)`
- stdin-backed `getchar`, `fgetc`, `fgets`, and `fread` while preserving existing `ungetc` pushback behavior
- `fclose` marking host stream handles as closed so later operations reject the handle

It does not implement filesystem-backed `FILE *`, seekable stream offsets, memory-backed C ABI `va_list`, or GCC nested-function trampolines. Those are larger follow-up milestones after this environment layer is stable.

## Design

`runtime.LoadOptions` gains an `Args []string` field. If `Args` is nil, runtime keeps the current deterministic default of `argv = {"cvm", NULL}`. When callers provide args, `Load` allocates one C string per argument plus a null-terminated pointer vector and passes `argc`/`argv` into the entry frame.

`runtime.ExternRegistry` gains a stdin reader while keeping existing `NewExternRegistry(stdout, stderr)` compatibility. New helper constructors accept stdin explicitly, and `DefaultExternRegistry` continues to default stdin to empty input. Character and block input first consume the existing `ungetc` pushback stack, then read from the configured stdin reader, setting EOF only when no byte is available.

Host stream lifetime is tracked in the registry. `fclose(stream)` validates the stream and then marks it closed. Later stdio operations that use the same handle return the existing unknown-handle style error, which is enough for Phase 2 to prevent accidental reuse without modeling all C undefined behavior.

## Testing

Focused tests cover:

- direct VM/program behavior for custom `argc`/`argv`
- CLI `cvm run file.cvmbc arg...` forwarding arguments
- direct extern stdin reads for character, line, and block input
- GCC runtime snippets that read from configured stdin
- `fclose` followed by a later operation rejecting the closed handle

Every implementation increment must pass:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

## Completion Criteria

This Phase 2 environment milestone is complete when:

- runtime callers can provide program arguments and observe them through `main(int, char **)`
- the CLI `run` subcommand forwards arguments after the bytecode path
- stdin-backed stdio reads consume real configured input after pushback is exhausted
- `fclose` prevents later use of the closed host stream handle
- handoff docs no longer list fixed `argv`, no-input stdin, or no-op `fclose` as current limits
