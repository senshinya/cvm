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

## Task 3: Configured File Write Mode

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase2b-file-streams.md`

- [x] **Step 1: Add failing direct write-mode coverage**

Cover `fopen(path, "w")`, byte writes, close/reopen, and reads from the registry-backed file content.

- [x] **Step 2: Add failing GCC runtime coverage**

Add a GCC runtime test that writes a small file with stdio calls, closes it, reopens it for reading, and validates bytes.

- [x] **Step 3: Implement hermetic write handles**

Add per-handle write support for configured file streams. Keep writes in the registry file map and avoid host filesystem effects.

- [x] **Step 4: Focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestFopenWritesConfiguredFile|TestGCCFopenConfiguredFileWriteModeExecutesThroughRuntime' -count=1 -v
```

- [x] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase2b-file-streams.md
git commit -m "feat(runtime): write configured files"
git push origin codex/bytecode-runtime-phase-1
```

## Task 2: Configured File Positioning

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase2b-file-streams.md`

- [x] **Step 1: Add failing direct file positioning coverage**

Cover `ftell`, `fseek`, and `rewind` on configured file handles while preserving standard-stream stub behavior.

- [x] **Step 2: Add failing GCC runtime coverage**

Add a GCC runtime test that opens a configured file, seeks from set/current/end, and verifies subsequent reads.

- [x] **Step 3: Implement configured file offsets**

Implement `SEEK_SET`, `SEEK_CUR`, and `SEEK_END` for configured file handles only. Clear EOF/pushback when seeking or rewinding.

- [x] **Step 4: Focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestConfiguredFilePositioning|TestGCCFopenConfiguredFilePositioningExecutesThroughRuntime|TestStdioPositionStubsExecuteThroughRuntime' -count=1 -v
```

- [x] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase2b-file-streams.md
git commit -m "feat(runtime): seek configured read only files"
git push origin codex/bytecode-runtime-phase-1
```
