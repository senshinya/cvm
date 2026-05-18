# Phase 1 Non-Math Runtime Fixture Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the Phase 1 non-math runtime fixture milestone by executing every remaining low-risk GCC accept `main` candidate and documenting all non-runtime candidates.

**Architecture:** Use the existing runtime GCC fixture helpers instead of adding a new manifest category. Keep the `{ dg-do run }` manifest limited to true GCC runtime directives, and add direct runtime tests only for compile-only accept fixtures that have stable exit-0 semantics. Record rejected candidates with concrete reasons in the roadmap and handoff docs.

**Tech Stack:** Go runtime tests, GCC accept fixture sources under `sema/testdata`, builtin runtime fixture helpers in `runtime/gcc_exec_test.go`, docs in `docs/bytecode-runtime-handoff.md` and the rolling roadmap.

---

### Task 1: Baseline Candidate Scan

**Files:**
- Inspect: `runtime/gcc_exec_test.go`
- Inspect: `runtime/testdata/gcc-exec/manifest.tsv`
- Inspect: `sema/testdata/gcc-c99/accept`
- Inspect: `sema/testdata/gcc-c99-extra/accept`
- Inspect: `sema/testdata/gcc-c90-as-c99/accept`

- [x] **Step 1: List accept fixtures with `main` not already covered by manifest or direct runtime tests**

Run:

```bash
node - <<'NODE'
const fs=require('fs'); const path=require('path');
const roots=['sema/testdata/gcc-c99/accept','sema/testdata/gcc-c99-extra/accept','sema/testdata/gcc-c90-as-c99/accept'];
const manifest=new Set(fs.readFileSync('runtime/testdata/gcc-exec/manifest.tsv','utf8').split(/\n/).slice(1).map(l=>l.split('\t')[0]).filter(Boolean));
const tests=fs.readFileSync('runtime/gcc_exec_test.go','utf8');
for (const root of roots) for (const f of fs.readdirSync(root).filter(f=>f.endsWith('.c')).sort()) {
  const p=path.join(root,f); const s=fs.readFileSync(p,'utf8');
  if (/\bmain\s*\(/.test(s) && !(manifest.has(p) || tests.includes(p) || tests.includes(f))) console.log(p);
}
NODE
```

Expected candidates:

```text
sema/testdata/gcc-c99-extra/accept/Wstrict-aliasing-bogus-vla-1.c
sema/testdata/gcc-c99-extra/accept/inline-10.c
sema/testdata/gcc-c99-extra/accept/inline-8.c
sema/testdata/gcc-c99-extra/accept/overflow-2.c
sema/testdata/gcc-c99-extra/accept/pr70418.c
sema/testdata/gcc-c99-extra/accept/transparent-union-1.c
```

### Task 2: Add Direct Runtime Coverage For The Remaining Exit-0 Candidate

**Files:**
- Modify: `runtime/gcc_exec_test.go`
- Modify: `runtime/program.go`
- Modify: `runtime/vm.go`
- Modify: `runtime/vm_test.go`

- [x] **Step 1: Add a direct fixture test for `Wstrict-aliasing-bogus-vla-1.c`**

Add:

```go
func TestGCCStrictAliasingVLADereferenceExecutesThroughRuntime(t *testing.T) {
	st := runGCCAcceptFixture(t, "sema/testdata/gcc-c99-extra/accept/Wstrict-aliasing-bogus-vla-1.c")
	if st.Code != 0 {
		t.Fatalf("exit code = %d, want 0", st.Code)
	}
}
```

- [x] **Step 2: Run focused runtime coverage**

Run:

```bash
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./runtime -run 'TestRunPassesDefaultArgcToMain|TestGCCStrictAliasingVLADereferenceExecutesThroughRuntime|TestGCCVLAMemcpyDynamicSizeExecutesThroughRuntime' -count=1 -v
```

Expected: all three tests pass.

- [x] **Step 3: Resolve the `main(int, char **)` loader gap**

The direct fixture exposed that Phase 1 runtime loading only accepted no-argument `main`. `runtime.Load` now accepts integer-returning `main(int, char **)`, initializes default entry arguments as `argc = 1` and `argv = {"cvm", NULL}`, and `runtime.Run` passes those entry arguments into the first frame.

- [x] **Step 4: Run full verification and commit**

Run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Commit:

```bash
git add runtime/gcc_exec_test.go runtime/program.go runtime/vm.go runtime/vm_test.go docs/superpowers/plans/2026-05-19-phase1-non-math-runtime-fixture-closure.md
git commit -m "test(runtime): execute GCC strict aliasing VLA fixture"
git push origin codex/bytecode-runtime-phase-1
```

### Task 3: Document Milestone Closure

**Files:**
- Modify: `docs/bytecode-runtime-handoff.md`
- Modify: `docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md`
- Modify: `docs/superpowers/plans/2026-05-19-phase1-non-math-runtime-fixture-closure.md`

- [x] **Step 1: Record remaining non-runtime candidates**

Record these classifications:

```text
inline-8.c: inline declaration of main only; compile validation target, no runtime entry body.
inline-10.c: inline main returns 1 and is not an exit-0 runtime candidate.
overflow-2.c: compile-only overflow diagnostic fixture; system run exits 1.
pr70418.c: GNU nested function with VLA-in-struct parameter extension; no stable Phase 1 runtime assertion.
transparent-union-1.c: compile-only warning fixture with unresolved extern calls; not runnable without artificial extern definitions.
```

- [x] **Step 2: Run full verification and commit docs**

Run:

```bash
git diff --check
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./codegen -count=1
env GOCACHE=/private/tmp/cvm-go-build-cache go test ./... -count=1
```

Commit:

```bash
git add docs/bytecode-runtime-handoff.md docs/superpowers/plans/2026-05-18-runtime-phase1-followup-roadmap.md docs/superpowers/plans/2026-05-19-phase1-non-math-runtime-fixture-closure.md
git commit -m "docs: record phase 1 non-math runtime closure"
git push origin codex/bytecode-runtime-phase-1
```
