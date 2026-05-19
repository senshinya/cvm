# Phase 9 Floating Conversion Errno Fidelity Gap Map

Phase 9 is closed on `codex/bytecode-runtime-phase-9`.

## Closed Surface

The runtime now has direct extern and source-level execution coverage for:

- Per-memory `errno` storage stability, initial zero state, and source-level `<errno.h>` read/write behavior.
- Floating conversion success and no-conversion paths preserving an existing nonzero `errno`.
- `strtod` decimal and hex overflow returning signed infinities, preserving the consumed-token `endptr`, and setting `errno` to `ERANGE`.
- `strtod` decimal and hex underflow-to-zero returning signed zero, preserving the consumed-token `endptr`, and setting `errno` to `ERANGE`.
- `strtod` representable subnormal values preserving the parsed value without setting `errno`.
- `strtod` infinity spellings, NaN spellings, and parenthesized NaN payload consumption.
- `strtof` float32 overflow and underflow behavior after target-type narrowing, including signed infinities, signed zero, subnormal preservation, `endptr`, and `ERANGE`.
- `strtold` overflow and underflow behavior under the current binary64-backed `TypeFLong` representation.
- `atof` wrapper range behavior, including overflow, underflow, no-conversion errno preservation, and shared parser range handling.
- `nan`, `nanf`, and `nanl` payload input coverage through direct extern and source-level runtime tests.

## Recheck Results

- Header declarations and macros were rechecked for `<errno.h>`, `<stdlib.h>`, and `<math.h>`.
- `DefaultExternRegistry` registrations were rechecked for `atof`, `strtod`, `strtof`, `strtold`, `nan`, `nanf`, `nanl`, and `errno`.
- Direct extern and source-level runtime sweeps cover the Phase 9 conversion, range, errno, NaN, and infinity surfaces.
- `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1` passed.
- Imported GCC accept scans for float conversion, `errno`, `ERANGE`, NaN, infinity, and range candidates found no new low-risk fixture beyond existing compile-manifest entries.

## Residual Limits

- `long double` remains represented by the current binary64-backed runtime model; `strtold` does not attempt platform-specific extended precision.
- Floating conversion remains deterministic C-locale parsing, not host-locale parsing.
- NaN payload text is consumed deterministically, but payload bits are not preserved.
- Exact native libc errno corner cases beyond overflow and underflow-to-zero remain out of scope until a fixture or workflow requires them.
