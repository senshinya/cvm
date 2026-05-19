# Phase 5 Nested Closure Lifetime Design

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-5`

## Goal

Phase 5 closes the bounded runtime safety gap for GNU nested-function closure pointers after their creating frame has returned. The runtime should keep valid in-frame closure calls working, but calling an escaped capturing nested-function pointer after its enclosing frame exits should produce an explicit lifecycle trap instead of falling through to a generic invalid indirect-call target.

## Scope

Included:

- Track closure pointer lifecycle in the VM after the owning frame exits.
- Preserve existing stack-trampoline semantics: closure pointers remain valid only while the creating frame is alive.
- Produce an explicit trap reason for expired closure pointer calls.
- Add focused source-level runtime coverage for an escaped nested-function pointer.
- Re-run existing nested closure coverage to ensure valid in-frame calls, VLA captures, and transitive captures still work.
- Update handoff and gap-map docs so Phase 5 is closed.

Excluded:

- Making escaped GNU nested-function pointers valid after the enclosing frame returns.
- Serializing closures into native executable trampolines.
- Changing codegen capture propagation or source-level sema rules beyond tests needed for this lifecycle behavior.

## Architecture

`runtime.VM` already records live closure pointers in `vm.closures` and deletes them when the owning frame exits. Phase 5 adds a small expired-closure registry keyed by the same runtime pointer address. When a frame exits, each owned closure is moved from live to expired before its backing local block is freed. `OpCallIndirect` checks expired closures before attempting plain function-address lookup, so stale closure pointers produce a precise trap.

This keeps the data flow local to `runtime/vm.go`: codegen still emits `OpMakeClosure`, valid calls still dispatch through `vm.closures`, and plain function pointers still use the existing global-address lookup.

## Testing

Each milestone starts with a current-state calibration. Implementation increments add focused tests and then run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Each coherent increment is committed and pushed before moving to the next milestone.
