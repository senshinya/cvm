# Phase 11 Wide String Memory Fidelity Gap Map

Date: 2026-05-20
Branch: `codex/bytecode-runtime-phase-11`

Phase 11 closes deterministic C-locale runtime support for the standard wide memory and wide string helpers exposed through builtin `<wchar.h>`.

## Closed Surfaces

- Builtin `<wchar.h>` declarations for the Phase 11 wide memory and string surface:
  - `wmemchr`, `wmemcmp`, `wmemcpy`, `wmemmove`, and `wmemset`.
  - `wcslen`, `wcscmp`, `wcsncmp`, `wcscoll`, and `wcsxfrm`.
  - `wcschr`, `wcsrchr`, `wcsstr`, `wcspbrk`, `wcsspn`, and `wcscspn`.
  - `wcscpy`, `wcsncpy`, `wcscat`, `wcsncat`, and `wcstok`.

- Runtime extern registration and direct extern coverage for every declared Phase 11 helper:
  - Wide memory helpers cover hit/miss, ordering, copy/move overlap, fill, zero-count, NUL, and high wide values.
  - Wide read/search/span helpers cover equal/less/greater ordering, prefixes, empty strings/sets, NUL terminators, substring search, set search, and high wide values.
  - Wide write helpers cover terminator copying, padding, truncation, bounded append, empty sources, return pointers, and high wide values.
  - `wcscoll` and `wcsxfrm` model deterministic C-locale code-point ordering and transformation.
  - `wcstok` implements C99 three-argument save-pointer tokenization with repeated delimiters, delimiter changes, exhaustion, empty delimiter sets, and save-pointer updates.

- Source-level runtime workflows cover:
  - Wide memory operations.
  - Wide string search and span operations.
  - Wide copy and concatenation operations.
  - Wide collation and transformation operations.
  - Wide tokenization with `wcstok`.

## Recheck Results

- `TestGCCExecutionGapReportIsCurrent` passed.
- Header declarations, registry entries, direct extern tests, and source-level runtime tests were rechecked for the Phase 11 surface.
- Imported GCC accept scans for `wcs`, `wmem`, `wcstok`, `wcscoll`, and `wcsxfrm` found no new low-risk run fixture to add.
- The relevant imported wide-character fixture remains `c99-init-1.c`, which already appears in runtime fixture coverage for wide-string initialization.

## Residual Limits

- `wcscoll` and `wcsxfrm` intentionally model deterministic C-locale ordering, not host locale collation tables.
- `wchar_t` storage is the existing CVM target model: 32-bit wide elements addressed in 4-byte units.
- Wide string and wide memory helpers validate through CVM memory access rules; undefined C cases such as invalid pointers, insufficient destination storage, and overlapping `wmemcpy`/string-copy inputs are not converted into native-libc compatibility behavior.
- Source workflows avoid relying on broader codegen fidelity for local wide aggregate initializers; runtime behavior is covered through wide literals and runtime assignments.
- Stateful encodings, locale-specific multibyte behavior, and wide formatted I/O remain outside this phase.
