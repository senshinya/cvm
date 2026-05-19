# Phase 8 Stdlib Time Locale Fidelity Gap Map

Phase 8 is closed on `codex/bytecode-runtime-phase-8`.

## Closed Surface

The runtime now has direct extern and source-level execution coverage for:

- `strtol`, `strtoul`, `strtoll`, and `strtoull` base detection, signedness, boundary values, no-conversion, and `endptr` behavior.
- `atoi`, `atol`, `atoll`, `atof`, `strtod`, `strtof`, and `strtold` decimal, exponent, hex-float, no-conversion, f32, and current flong behavior.
- `rand` and `srand` deterministic default seed, reseed repeatability, and value range.
- `getenv` missing variables, configured values, stable pointers, and per-memory static string materialization.
- Hermetic `system` behavior for NULL and non-NULL commands.
- `atexit`, `exit`, and `_Exit` callback ordering and cleanup-control behavior.
- `setlocale` deterministic C-locale query, explicit `C`, empty locale, and unsupported-locale behavior.
- `clock`, `time`, and `difftime` deterministic arithmetic and storage behavior.
- Ctype classification and case conversion EOF, ASCII, punctuation, and high-int unsigned-byte behavior.

## Recheck Results

- Header declarations, sema builtin typing, extern registry entries, direct extern tests, and source-level runtime tests were rechecked for the Phase 8 symbols.
- `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1` passed.
- `runtime/testdata/gcc-exec/gap-report.md` remains closed with 18 runnable manifest candidates and no failures.
- `codegen/testdata/gcc-bytecode-compile.tsv` remains at 232 lines including the header.
- Imported GCC accept scans for conversion, random/env/process, locale, time, and ctype terms did not identify a new stable low-risk runtime fixture.

## Residual Limits

- Floating conversion uses deterministic C-locale parsing through Go `strconv.ParseFloat`; errno, range diagnostics, and exact native libc overflow behavior are not modeled.
- `time` and `clock` return deterministic zero values rather than host wall-clock or CPU time.
- `system` does not execute host commands; non-NULL commands fail deterministically.
- Locale behavior remains limited to deterministic C-locale semantics.
- Ctype behavior is byte-oriented and deterministic; locale-specific classification and wide-character classification remain out of scope.
