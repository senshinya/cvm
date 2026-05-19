# Phase 13 Formatted I/O Spec Fidelity Gap Map

Date: 2026-05-20
Branch: `codex/bytecode-runtime-phase-13`

Phase 13 closes the deterministic C-locale formatted I/O fidelity slice that builds on Phase 12 wide formatted stdio.

## Closed Surfaces

- `printf` integer length semantics:
  - Signed `%d`/`%i` now honor `hh`, `h`, default `int`, `l`, `ll`, `j`, `z`, and `t`.
  - Unsigned `%u`, `%x`, `%X`, and `%o` now honor `hh`, `h`, default `unsigned int`, `l`, `ll`, `j`, `z`, and `t`.
  - `%n` count stores are covered for narrow and wide count types.

- `printf` text and pointer semantics:
  - `%lc` supports deterministic ASCII wide-character output and rejects non-ASCII wide characters.
  - `%ls` reads wide strings through the same deterministic ASCII bridge used by wide format strings.
  - `%p` deterministic `0x` formatting is covered for null, non-null, width, and left alignment.

- `printf` width, precision, alternate form, and floating edges:
  - Dynamic width and precision cover negative width, negative precision, and combined `*.*`.
  - Zero integer precision covers `%.0d`, `%.0u`, `%.0x`, `%.0o`, and `%#.0o`.
  - Alternate float forms preserve required decimal points for `%#f`, `%#g`, and `%#a`.
  - Float width and precision cover `%f`, `%e`, `%E`, `%g`, `%G`, `%a`, and `%A`.
  - Special values use deterministic C-style `inf`/`INF` and `nan`/`NAN` spelling, with signs handled for infinities and signed zero kept on the finite path.

- `scanf` length and text semantics:
  - Signed integer scans cover `hh`, `h`, default, `l`, `ll`, `j`, `z`, and `t`.
  - Unsigned decimal, hexadecimal, and octal scans cover narrow, default, and wide integer destinations.
  - `%n` count stores cover all supported integer length modifiers and do not increase the assignment count.
  - `%c`, `%lc`, `%s`, `%ls`, `%[`, and `%l[` cover width-limited narrow and wide storage plus NUL termination where required.
  - Literal percent, whitespace skipping, literal mismatch, and unread preservation are covered.

- Stream formatted I/O:
  - Source workflows cover `fscanf` followed by `fgetc` and `fwscanf` followed by `fgetwc`, preserving unread input after formatted scans.
  - Existing stream-backed `scanf`, `fscanf`, `wscanf`, and `fwscanf` direct tests remain in force.

## Recheck Results

- `TestGCCExecutionGapReportIsCurrent` passed.
- Header declarations and extern registry entries were rechecked for the touched narrow and wide formatted I/O families.
- Imported GCC accept/reject scans found formatted I/O diagnostic fixtures but no newly unblocked low-risk runtime fixture to add.
- Standard gates passed for every coherent increment:
  - `git diff --check`
  - `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1`
  - `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1`

## Residual Limits

- Formatted I/O remains a deterministic C-locale subset. Native locale, multibyte state machines, and platform libc spelling quirks remain out of scope.
- Wide formatted text uses ASCII bridging; non-ASCII wide formatted output is rejected rather than transcoded.
- `long double` storage continues to use the runtime's current binary64-backed model.
- Native filesystem side effects, native file descriptors, and host libc buffering behavior remain outside the hermetic stream model.
- Undefined C behavior, invalid pointers, insufficient destination storage, and incompatible argument types are governed by CVM memory/runtime checks rather than native-libc compatibility.
