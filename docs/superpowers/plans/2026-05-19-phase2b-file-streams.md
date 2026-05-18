# Phase 2B File Stream Runtime Plan

## Common Pre-Plan Adjustment

After the memory `va_list` v-format milestone, the next bounded Phase 2B target is filesystem-backed stream behavior. Direct host filesystem access would make tests environment-dependent, so the first increment uses hermetic files configured on `ExternRegistry`.

This milestone starts with read-only configured files: `fopen(path, "r")` returns a host FILE handle when the path exists in the registry, and existing input helpers (`fgetc`, `fgets`, `fread`) consume bytes from that handle. Write modes, host disk effects, seek offsets, and mutation are left to later increments.

## Task 1: Read-Only Configured `fopen`

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase2b-file-streams.md`

- [x] **Step 1: Add failing direct extern coverage**

Add a direct test that configures file bytes on the registry, opens the path with `fopen(path, "r")`, and reads bytes with `fgetc`/`fread`.

- [x] **Step 2: Add failing GCC runtime coverage**

Add a GCC runtime test using a configured extern registry and `fopen("data.txt", "r")`.

- [x] **Step 3: Implement hermetic read-only file handles**

Add registry state for configured file contents and per-handle read offsets. Keep missing files returning null, and keep non-read modes unsupported for now.

- [x] **Step 4: Focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestFopenReadsConfiguredFile|TestGCCFopenConfiguredFileExecutesThroughRuntime' -count=1 -v
```

- [x] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase2b-file-streams.md
git commit -m "feat(runtime): open configured read only files"
git push origin codex/bytecode-runtime-phase-1
```
