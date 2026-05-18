# Phase 2B Nested Trampolines Design

## Goal

Support escaping GNU nested functions with captured automatic state while the enclosing stack frame is still alive.

## Design

Add a bytecode-level closure/trampoline pointer instead of treating capturing nested functions as plain static function addresses. A new `MakeClosure` instruction consumes hidden static-chain capture arguments from the stack, records the target function global plus those captured values in the runtime, and pushes a pointer value. Existing C function pointer storage and calls can keep using `ptr`.

`CallIndirect` keeps its existing behavior for plain function addresses. When the callee pointer is a runtime closure pointer, it pops only the source-level function pointer arguments using the visible signature, appends the closure's captured hidden arguments, and invokes the nested function with its full lowered signature.

Closure backing storage is allocated as a local runtime block in the frame where the closure is created. The VM removes the closure record and frees the block when that frame exits, matching the stack lifetime of GCC nested-function trampolines. Calling such a pointer after the enclosing frame returns remains invalid.

## Scope

In scope:

- captured nested function address expressions
- local function pointer variables initialized from capturing nested functions
- indirect calls through those function pointers while the parent frame is alive
- passing such pointers to deeper calls

Out of scope:

- making nested-function pointers valid after their enclosing function returns
- serializing closure state into host ABI trampolines
- changing ordinary non-capturing function pointers

## Verification

Use TDD with focused codegen and runtime GCC execution tests first, then run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```
