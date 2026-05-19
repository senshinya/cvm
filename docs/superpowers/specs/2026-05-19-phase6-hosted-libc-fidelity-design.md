# Phase 6 Hosted Libc Fidelity Design

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-6`

## Goal

Phase 6 closes a broad but bounded hosted C library fidelity sweep. It focuses on C-standard runtime surfaces that are already declared or registered but still behave as conservative stubs or have narrow edge coverage: file positioning, `freopen`, deterministic temporary names, buffering controls, C-locale multibyte/wide-character helpers, and final fixture/documentation closure.

## Scope

Included:

- Real `fgetpos`/`fsetpos` behavior for hermetic configured streams.
- More useful `freopen` behavior over configured files and existing `FILE *` handles.
- Deterministic hermetic `tmpnam` names for caller-provided buffers and `NULL`.
- Buffer-control validation for `setbuf` and `setvbuf`.
- C-locale single-byte edge behavior for `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, and `wcstombs`.
- Focused source-level runtime tests for each surface.
- GCC runtime fixture rechecks and closure docs.

Excluded:

- Ambient host filesystem access.
- Native OS file descriptors and descriptor syscalls.
- Locale-specific encodings beyond the existing deterministic C-locale model.
- Long-double precision model changes.

## Architecture

Phase 6 keeps the runtime hermetic. `ExternRegistry` remains the owner of configured files and host stream handles. File positioning improvements reuse `hostFile.pos`; `fpos_t` stores the same signed 64-bit offset already used by `ftell`. `freopen` replaces the target handle's in-registry stream state without touching host files. `tmpnam` returns deterministic virtual names that callers may pass to existing hermetic file APIs, but it does not create files.

The multibyte/wide-character helpers remain C-locale single-byte helpers. They should validate buffers, handle null-pointer state queries consistently, and keep conversion counts deterministic.

## Testing

Every milestone begins by checking current implementation and test coverage for that exact surface. Each coherent increment adds direct extern coverage and source-level runtime coverage when source-level behavior matters. The standard verification chain for each coherent increment is:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Each coherent increment is committed and pushed before moving to the next milestone.
