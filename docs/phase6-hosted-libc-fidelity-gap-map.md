# Phase 6 Hosted Libc Fidelity Gap Map

Date: 2026-05-19
Branch: `codex/bytecode-runtime-phase-6`

Phase 6 closed a hosted-libc fidelity sweep for deterministic runtime surfaces that are needed by GCC-style C programs but can stay hermetic inside CVM.

## Closed Surfaces

- File positioning:
  - `fgetpos` stores configured-file offsets as `fpos_t`.
  - `fsetpos` restores configured-file offsets, rejects negative positions, clears pushback/EOF, and resets update-mode direction.
  - Standard streams keep deterministic failure return values.

- `freopen`:
  - Configured read files reopen existing streams.
  - `w`/`w+` truncate or create hermetic registry files.
  - `a`/`a+` preserve append-at-end write behavior.
  - Missing read files and invalid modes return `NULL` without replacing the original stream.

- Temporary names:
  - `tmpnam(NULL)` returns a deterministic static C string.
  - `tmpnam(buf)` writes deterministic names and returns `buf`.
  - Caller-buffer calls advance a registry-local sequence.
  - Generated names work with hermetic `fopen` write/close/reopen/read flows.

- Buffer controls:
  - `setvbuf` validates `_IOFBF`, `_IOLBF`, `_IONBF`, rejects invalid modes, accepts NULL buffers, checks provided buffer ranges, and rejects closed streams.
  - `setbuf` has direct and source-level coverage for explicit and NULL buffers.

- C-locale multibyte helpers:
  - `mblen`, `mbtowc`, `wctomb`, `mbstowcs`, and `wcstombs` cover null/reset paths, ASCII/NUL cases, truncation, terminators, and high-bit rejection.
  - A source-level workflow exercises all five helpers together.

## Rechecks

- Header declarations, header smoke tests, runtime registry registration, direct smoke coverage, and source-level runtime tests were rechecked for touched stdio/stdlib surfaces.
- `TestGCCExecutionGapReportIsCurrent` passed.
- Scans for newly unblocked GCC fixtures touching these surfaces found no low-risk imported fixture to add.

## Remaining Bounds

- Locale-specific multibyte encodings are not modeled; behavior is C-locale ASCII only.
- Hosted files remain deterministic registry-backed streams rather than native OS file descriptors.
- `tmpnam` is deterministic and hermetic, not a native temporary-file API.
- Exact native libc buffering effects remain no-op modeled except for validation and return values.
