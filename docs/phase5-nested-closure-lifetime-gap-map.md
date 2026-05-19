# Phase 5 Nested Closure Lifetime Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-5`
Base: closed Phase 4 tip `76dd4f8`

## Calibration Summary

Phase 4 closed the hosted stdio/input surface and left several possible Phase 5 candidates. Native descriptors, locale/multibyte input, and a different long-double model are broad runtime model changes. The bounded, high-value runtime safety gap was GNU nested-function closure lifecycle diagnostics: escaped closure pointers were already invalid after their creating frame returned, but calling one reported only a generic invalid indirect-call target.

## Closed During Phase 5

- The VM now records expired closure pointer addresses when their owning frame exits.
- `OpCallIndirect` checks expired closure pointers before falling back to normal function-address lookup.
- Escaped capturing GNU nested-function pointers now trap with an explicit `expired closure pointer` diagnostic.
- Existing live closure behavior remains covered and passing for direct nested calls, function-pointer calls while the creator frame is alive, closure pointers passed to another callee, VLA captures, and transitive captures.
- GCC runtime fixture recheck remains closed: 18 directive-based runnable manifest candidates pass, and no low-risk nested-function pointer fixture with stable behavior was found.

## Current High-Value Gaps

No current Phase 5 implementation gap remains.

Further runtime work should be scoped as a new phase or review follow-up. Candidate Phase 6 surfaces, if needed, are:

- Native-like descriptor-backed file behavior rather than only hermetic configured streams.
- Locale-specific and multibyte/wide-character formatted input.
- A different long-double storage/arithmetic model if future fixtures require precision beyond the current binary64-backed representation.
- Optional public diagnostics or documentation for detecting escaped nested-function pointer use before runtime execution.
