# GCC Execution Fixtures

This directory contains the GCC-derived execution fixtures that currently form
the Phase 1 cvm bytecode runtime gate. The manifest is intentionally curated:
it contains only deterministic GCC accept fixtures that compile to bytecode,
load through the binary runtime path, and have a stable runtime exit code.

The current gate has six fixtures. Keep it small until runtime semantics and
fixture suitability are clear enough to expand without mixing compile-only,
diagnostic-only, or hosted-library behavior into the execution suite.

Manifest rules:

- The header is exactly `path	exit	category	reason`.
- Each row is enabled; unsupported or unsuitable cases stay out of the manifest
  instead of being listed as skipped cases.
- `path` must be a `.c` file under one of the supported GCC accept fixture
  roots:
  - `sema/testdata/gcc-c99/accept/`
  - `sema/testdata/gcc-c99-extra/accept/`
  - `sema/testdata/gcc-c90-as-c99/accept/`
- `exit` is the expected deterministic runtime exit code.
- `category` and `reason` must explain why the fixture belongs in the runtime
  gate.
- The source must contain a `{ dg-do run }` directive.

Do not copy the compile-only GCC fixture set wholesale. Do not add fixtures
whose expected behavior is still diagnostic-driven or unclear for execution
coverage. Compile-only diagnostic cases, including `signbit-sa.c`,
`inline-10.c`, and `overflow-2.c`, are excluded from this runtime gate.

`gap-report.md` tracks every current GCC accept fixture whose source contains
`{ dg-do run`. Refresh it with:

```bash
CVM_UPDATE_GCC_EXEC_GAP_REPORT=1 go test ./runtime -run TestGCCExecutionGapReportIsCurrent -count=1
```
