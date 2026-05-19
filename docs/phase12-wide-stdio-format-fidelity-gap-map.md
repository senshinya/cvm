# Phase 12 Wide Stdio Format Fidelity Gap Map

Date: 2026-05-20
Branch: `codex/bytecode-runtime-phase-12`

Phase 12 closes deterministic C-locale runtime support for the wide stdio and wide formatted I/O subset exposed through builtin `<wchar.h>`.

## Closed Surfaces

- Builtin `<wchar.h>` declarations and registry entries for:
  - `fwide`.
  - `fputwc`, `putwc`, and `putwchar`.
  - `fgetwc`, `getwc`, `getwchar`, and `ungetwc`.
  - `fputws` and `fgetws`.
  - `wprintf`, `fwprintf`, `swprintf`, `vwprintf`, `vfwprintf`, and `vswprintf`.
  - `wscanf`, `fwscanf`, and `swscanf`.

- Runtime stream orientation support:
  - `fwide` queries and locks byte/wide orientation.
  - Wide input and output helpers orient streams consistently and reject byte-oriented conflicts.
  - EOF, error, pushback, closed-stream, and update-mode behavior are covered for the implemented wide paths.

- Wide character and string stdio helpers:
  - Wide output covers ASCII writes, NUL writes, high wide-character rejection, stream orientation, aliases, and closed streams.
  - Wide input covers ASCII reads, NUL, EOF, invalid high input, aliases, pushback, and EOF clearing after `ungetwc`.
  - `fputws` and `fgetws` cover strings, empty strings, line/width behavior, EOF, invalid input, and closed streams.

- Wide formatted output:
  - Wide format strings are converted through deterministic C-locale ASCII bridging.
  - `swprintf` and `vswprintf` write wide output buffers and model truncation as failure.
  - `wprintf`, `fwprintf`, `vwprintf`, and `vfwprintf` write through stdout or host streams with wide orientation.
  - Direct tests cover integers, strings, chars, percent literals, truncation, unsupported high format characters, streams, and memory-backed `va_list` values.

- Wide formatted input:
  - Wide format and wide input buffers are converted through deterministic C-locale ASCII bridging.
  - `swscanf`, `wscanf`, and `fwscanf` reuse the existing scanner and support integer, `%ls`, `%lc`, `%n`, width, assignment suppression, mismatch, EOF-like empty input, unread preservation, and hermetic file state.

- Source-level runtime workflows cover:
  - Wide stdio character/string input and output through `fwide`, `fgetwc`, `ungetwc`, `fgetws`, `fputwc`, and `fputws`.
  - Wide formatted memory/stdout/stderr/stdin paths through `swprintf`, `swscanf`, `fwprintf`, `wprintf`, and `wscanf`.
  - File-backed wide formatted scanning through `fwscanf`.

- Runtime/codegen alignment fixes needed by the source workflows:
  - String constants are loaded with at least 4-byte alignment so wide literals can be read safely.
  - Aggregate initializer field stores respect the effective alignment guaranteed by the containing object.

## Recheck Results

- `TestGCCExecutionGapReportIsCurrent` passed.
- Header declarations, registry entries, direct extern tests, and source-level runtime tests were rechecked for the Phase 12 surface.
- Imported `gcc-c99-extra/accept` scans for `wprintf`, `wscanf`, `fwide`, `fgetwc`, `fputwc`, `fgetws`, and `fputws` found no new low-risk accept fixture to add.
- The source workflow recheck added `fwprintf` stderr coverage to the wide format workflow.

## Residual Limits

- Wide stdio and wide formatted I/O intentionally use deterministic C-locale ASCII bridging; locale-specific wide I/O and stateful encodings are not modeled.
- High wide characters in wide format/input/output paths are rejected or reported through the existing stream-error conventions instead of being transcoded.
- Wide formatted output reuses the existing formatter subset; unsupported native-libc format behavior remains outside this phase.
- Source-level `v*wide printf` workflows are not added; direct tests cover memory-backed `va_list` handling deterministically.
- File behavior remains the runtime's hermetic host-file model, not native filesystem side effects.
- Undefined C cases such as invalid pointers, insufficient destination storage, and incompatible stream orientation remain governed by CVM memory and stream error behavior rather than native-libc compatibility.
