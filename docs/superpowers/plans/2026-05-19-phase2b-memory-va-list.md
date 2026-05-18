# Phase 2B Memory `va_list` Runtime Plan

## Common Pre-Plan Adjustment

After closing the Phase 2 runtime environment milestone, the remaining Phase 2B candidates are memory-backed `va_list`, escaping GNU nested-function trampolines, and filesystem-backed streams. The next safest increment is memory-backed `va_list` consumption inside v-format externs: the extern signatures already accept `void *`/`va_list` pointers, and this can improve runtime behavior without changing bytecode function pointer or trampoline representation.

This plan deliberately does not claim full C ABI `va_arg` support. `<stdarg.h>` still expands `va_arg(ap, type)` to a placeholder, so generating `OpVaArg` from ordinary C source remains a later sema/preprocessor/codegen milestone. The current milestone is a runtime ABI for extern consumers of a `va_list` pointer.

## Runtime Encoding

Use a small CVM-owned memory layout:

```text
uint64 count
repeat count times:
  uint64 type_tag
  uint64 payload
```

The initial type tags cover integer, pointer, and double payloads, matching the formatter's current argument surface and C default argument promotions.

## Task 1: v-format Memory `va_list`

**Files:**
- Modify: `runtime/extern.go`
- Modify: `runtime/extern_test.go`
- Modify: `runtime/gcc_exec_test.go`
- Modify: `docs/superpowers/plans/2026-05-19-phase2b-memory-va-list.md`

- [x] **Step 1: Add failing direct extern coverage**

Add direct tests where `vsprintf`/`vprintf` receive a memory-backed argument list and consume `%d`, `%s`, and related simple formats.

- [x] **Step 2: Add failing GCC runtime coverage**

Add a small GCC runtime source that constructs the CVM `va_list` memory layout as an `unsigned long[]`, calls `vsprintf`, and checks formatted output.

- [x] **Step 3: Implement memory `va_list` decoding**

Add a helper that decodes the CVM memory layout into formatter `Value` arguments. Keep null `va_list` behavior unchanged for literal-only formats.

- [x] **Step 4: Focused tests**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestVFormatExternsReadMemoryVaList|TestGCCVFormatMemoryVaListExecutesThroughRuntime' -count=1 -v
```

- [x] **Step 5: Full verification, commit, push**

Run Common Verification, then commit:

```bash
git add runtime/extern.go runtime/extern_test.go runtime/gcc_exec_test.go docs/superpowers/plans/2026-05-19-phase2b-memory-va-list.md
git commit -m "feat(runtime): consume memory backed va lists"
git push origin codex/bytecode-runtime-phase-1
```
