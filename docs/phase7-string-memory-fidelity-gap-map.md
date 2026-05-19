# Phase 7 String And Memory Fidelity Gap Map

Phase 7 is closed on `codex/bytecode-runtime-phase-7`.

## Closed Surface

The runtime now has direct extern and source-level execution coverage for:

- `memchr`, `memcmp`, `strncmp`, and `strnlen` zero-length, bounded-read, and unsigned-byte behavior.
- `strchr`, `strrchr`, `strstr`, `strpbrk`, `strspn`, and `strcspn` boundary behavior.
- `strtok` delimiter changes, exhaustion, all-delimiter inputs, and restart with a new string.
- `strxfrm`, `strcoll`, and deterministic per-memory `strerror` storage behavior.
- `memccpy`, `bcopy`, `memset`, and `bzero` byte masking, overlap, zero-length, and exact-bound behavior.
- `strcpy`, `stpcpy`, `strncpy`, `stpncpy`, `strcat`, `strncat`, and checked string/memory builtin success/failure boundaries.

## Recheck Results

- Header, sema builtin, registry, direct extern, and source-level runtime coverage were rechecked for the Phase 7 symbols.
- `env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1` passed.
- `runtime/testdata/gcc-exec/gap-report.md` remains closed with 18 runnable manifest candidates and no failures.
- `codegen/testdata/gcc-bytecode-compile.tsv` remains at 232 lines including the header.
- Imported GCC accept scans found checked string/memory warning fixtures already represented in the compile manifest and no new low-risk runtime manifest candidate.

## Residual Limits

- `strerror` intentionally returns a deterministic per-memory static `"error"` string rather than host-specific errno text.
- Checked builtin object-size failures are modeled as runtime traps in direct extern coverage; warning-only GCC accept fixtures remain compile-time coverage, not runtime manifest targets.
- Broader libc behavior outside the deterministic hosted subset remains scoped to later phases.
