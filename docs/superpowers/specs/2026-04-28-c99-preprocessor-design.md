# C99 Preprocessor Design

Date: 2026-04-28
Status: Draft for review

## Context

The current cvm frontend runs:

```text
lexer -> parser -> sema
```

The C99 semantic gate now passes for all `gcc.dg/c99*.c` cases that do not
depend on preprocessing. The remaining primary C99 skips are all classified as
preprocessor or system macro/header dependencies. The next frontend step is a
C99 preprocessor that can feed the existing parser and sema pipeline while
preserving useful source locations.

Scope for this design:

- Implement C99-standard preprocessing behavior needed by the frontend.
- Use the remaining GCC C99 preprocessor-dependent cases as the first compliance
  gate.
- Support local quoted includes and built-in C99 standard header shims.
- Preserve original file, line, column, include, and macro expansion provenance
  for diagnostics.

Out of scope for the first implementation:

- Real macOS/GCC system header search.
- GCC-specific `cpp` compatibility as a goal in itself.
- Full `#pragma` semantics, target runtime behavior, and backend/codegen work.
- Broad GNU extension support unless a test requires accepting or rejecting it
  at the preprocessing boundary.

## Architecture

The frontend pipeline becomes:

```text
source files
  -> preprocessor
  -> []entity.Token
  -> parser
  -> sema
```

The new `preprocessor` package owns preprocessing-token scanning, directive
handling, macro expansion, include resolution, and conversion to parser tokens.
The existing `lexer` package continues to define ordinary C tokenization rules
where they are reusable, but preprocessing has its own scanner because C99
preprocessing tokens include forms that are not parser tokens, such as header
names, skipped conditional text, and macro replacement lists before final
conversion.

The preprocessor should produce `[]entity.Token` so `parser` and `sema` do not
need a broad API rewrite. `entity.SourcePos` and `entity.SourceRange` will grow
enough source identity/provenance information for diagnostics to map errors back
to original files and macro expansion sites.

The initial public API should be small:

```go
type Options struct {
	IncludePaths []string
	Std          Standard
}

type Result struct {
	Tokens  []entity.Token
	Sources *SourceManager
}

func PreprocessFile(path string, opts Options) (*Result, error)
func PreprocessSource(name, source string, opts Options) (*Result, error)
```

`Compiler` stores the returned `SourceManager` and passes `Result.Tokens` to the
parser. Tests that need direct token output can use `PreprocessSource`.

## Components

### `preprocessor.Scanner`

Scans preprocessing tokens from source files.

Responsibilities:

- Apply line splicing for backslash-newline before token recognition.
- Replace comments with whitespace while preserving line structure.
- Recognize preprocessing identifiers, pp-numbers, string literals, character
  literals, punctuators, header-name tokens in include directives, and newline
  boundaries.
- Preserve original source ranges for each preprocessing token.
- Report lexical preprocessing errors, such as unterminated comments or
  literals.

The scanner should keep comments and line splicing in this layer, not in parser
tests or fixture stripping, because C99 translation phases make them
preprocessing behavior.

### `preprocessor.DirectiveParser`

Parses one logical preprocessing line at a time.

Supported directives:

- `#define`
- `#undef`
- `#include`
- `#if`
- `#ifdef`
- `#ifndef`
- `#elif`
- `#else`
- `#endif`
- `#line`
- `#error`
- empty directive

Unknown directives are diagnostics unless they are explicitly accepted as
implementation-defined no-ops in a later plan. `#pragma` is not semantically
implemented in the first pass; tests that require target pragma behavior remain
outside this scope.

### `preprocessor.MacroTable`

Stores macro definitions.

Supported macro forms:

- Object-like macros.
- Function-like macros.
- C99 variadic function-like macros with `__VA_ARGS__`.
- Built-in macros:
  - `__STDC__`
  - `__STDC_VERSION__`
  - `__FILE__`
  - `__LINE__`

`__DATE__` and `__TIME__` may be added later if tests require them. To keep test
output deterministic, they should not be introduced casually.

### `preprocessor.Expander`

Expands ordinary token streams after directive handling.

Responsibilities:

- Macro lookup and replacement.
- Function-like macro argument collection.
- Argument prescan rules.
- Recursive expansion suppression.
- Stringification with `#`.
- Token pasting with `##`.
- Variadic macro substitution.
- `defined` evaluation inside `#if` expressions.
- Conversion of expanded preprocessing tokens into `entity.Token`.

The expander must avoid text-level concatenation except where the C99 rules
explicitly require retokenization after token pasting. Retokenization should run
through the preprocessing-token scanner for the pasted spelling and then convert
the result to a parser token.

### `preprocessor.IncludeResolver`

Resolves include directives without reading real system headers by default.

Resolution rules:

- `"file.h"` searches relative to the including file first, then configured
  include paths.
- `<header.h>` searches the built-in C99 header shim table first.
- Real system include directories are not searched in the first implementation.
- Include recursion has a hard depth limit and a clear diagnostic.

Missing local quoted headers are errors. Missing angle-bracket headers are also
errors unless a future compiler option explicitly allows ignoring unknown
standard headers.

### `preprocessor.Headers`

Provides built-in C99 standard header shims. The first set should focus on the
headers needed by the GCC C99 skipped cases and current frontend tests:

- `stdbool.h`
- `stdint.h`
- `stddef.h`
- `limits.h`
- `float.h`
- `iso646.h`

Additional lightweight shims may be added when test evidence requires them,
especially for `math.h` and `tgmath.h`. Shims should define the smallest useful
set of macros and typedef-friendly text needed by cvm's frontend. They should
not try to mirror platform libc headers.

### `preprocessor.SourceManager`

Owns source identity and diagnostic provenance.

Responsibilities:

- Store each loaded source file, built-in header source, and line table.
- Resolve `SourcePos` to file name, line, column, and source line text.
- Track include stack entries.
- Track macro definition sites and macro invocation sites.
- Provide rendering data for `Compiler.handleError`.

Existing `Compiler` state should move away from a single `FileName`, `Source`,
and `Lines` assumption. It can keep a simple `RunSource` path for tests, but
diagnostics should ask `SourceManager` for display information.

## Source Location Model

`entity.SourcePos` currently stores only line and column. The preprocessor needs
file identity and expansion provenance. The implementation should extend the
model in this direction:

```go
type SourcePos struct {
	FileID      int
	Line        int
	Column      int
	ExpansionID int
}
```

`FileID == 0` remains usable for legacy tests that create tokens without a
source manager. `ExpansionID == 0` means the token is not the result of macro
expansion.

`SourceManager` owns the tables behind those IDs:

- Physical file ID.
- Physical line and column.
- Optional macro expansion ID or expansion stack reference.
- Optional spelling location and expansion location distinction.

Diagnostic policy:

- Normal parser/sema errors point to the spelling location in user source.
- Errors caused by macro expansion point to the expansion use site and add notes
  for the macro definition or argument source when useful.
- Include-file errors display the included file location and can add a note for
  the include directive.

The first implementation does not need byte-perfect GCC/Clang diagnostic text,
but it must avoid losing the original file and line after preprocessing.

## Data Flow

Top-level preprocessing flow:

```text
PreprocessFile(path)
  -> SourceManager.Load(path)
  -> Scanner.ScanPreprocessingTokens(file)
  -> DirectiveParser processes logical lines
  -> inactive conditional groups are skipped
  -> IncludeResolver recursively injects included token streams
  -> Expander expands ordinary token streams
  -> ConvertToCompilerTokens()
  -> parser.NewParser(tokens).Parse()
```

`Compiler.RunFile` should call the preprocessor first. `Compiler.RunSource`
should create an in-memory source file, preprocess it, and then call parser and
sema. Tests that intentionally bypass preprocessing can continue using the
existing lexer helpers until they are migrated.

## Conditional Inclusion

The conditional inclusion engine should maintain a stack of groups. Each group
tracks:

- Parent active state.
- Whether any earlier branch in the group has been selected.
- Whether the current branch is active.
- Whether `#else` has already appeared.

Inactive branches are still scanned enough to find nested conditional
directives, but their ordinary tokens are not macro-expanded or emitted.

`#if` expressions use C99 preprocessing integer expression rules:

- Expand macros first, except for `defined`.
- Replace remaining identifiers with `0`.
- Evaluate integer constants and operators used by preprocessing expressions.
- Reject unsupported or malformed expressions with a preprocessing diagnostic.

## Standard Header Shim Policy

Built-in headers are source text owned by cvm. They should be simple and
portable:

```c
#define true 1
#define false 0
#define bool _Bool
```

Typedef-style definitions that the parser and sema need should be emitted as C
source, not injected directly into sema. For example, `stddef.h` can use an
explicit predefined macro:

```c
#define __SIZE_TYPE__ unsigned long
typedef __SIZE_TYPE__ size_t;
```

If a shim requires frontend support for compiler predefined type macros such as
`__WCHAR_TYPE__`, that support should be modeled explicitly as predefined
macros.

The shim layer should be test-driven against actual skipped cases. It should not
grow into a fake libc.

## Error Handling

Preprocessor errors return `*common.CvmError` with source positions managed by
`SourceManager`.

Expected error categories:

- Bad directive syntax.
- Unterminated conditional inclusion.
- Unexpected `#elif`, `#else`, or `#endif`.
- Macro argument count mismatch.
- Invalid use of `#` or `##`.
- Token paste that does not form a valid preprocessing token.
- Include file not found.
- Include recursion depth exceeded.
- `#error`.
- Malformed `#if` expression.

Parser and sema errors continue to use `common.CvmError`; `Compiler.handleError`
is responsible for rendering those positions through the source manager when
available.

## Testing Strategy

### Unit Tests

Create focused `preprocessor` tests for:

- Comment replacement and line splicing.
- Object-like macros.
- Function-like macros.
- Recursive macro suppression.
- Argument prescan behavior.
- Stringification.
- Token pasting.
- Variadic macros and `__VA_ARGS__`.
- `defined` and `#if` expressions.
- `#ifdef`, `#ifndef`, `#elif`, `#else`, `#endif`.
- Local quoted include.
- Built-in standard header include.
- `#line`.
- `#error`.
- `__FILE__` and `__LINE__`.
- Source mapping for direct tokens, included tokens, and macro-expanded tokens.

Test comments in new code should be Chinese, matching the project preference
for this work.

### Integration Tests

Add tests that run:

```text
preprocessor -> parser -> sema
```

Representative cases:

- `#define T int` followed by declarations using `T`.
- Function-like macro producing an expression.
- Include of a local header that defines a typedef or macro.
- Include of `<stdbool.h>`, `<stdint.h>`, and `<stddef.h>` shims.
- Conditional compilation that selects only the valid C branch.

### GCC C99 Gate

Reclassify `sema/testdata/gcc-c99/manifest.tsv` after the preprocessor is wired
into the test runner:

- Move GCC C99 preprocessor-dependent accept cases into `accept/` when they pass.
- Move GCC C99 preprocessor-dependent reject cases into `reject/` when cvm
  rejects them for the right frontend reason.
- Keep only cases requiring out-of-scope behavior skipped, with specific
  reasons such as `requires real system header`, `requires GCC pragma semantics`,
  or `requires GNU extension`.

The generic reason `requires preprocessor or system macro handling` should
disappear from the primary `gcc-c99` manifest after this work unless a remaining
case is deliberately deferred with a more exact note in the implementation plan.

## Acceptance Criteria

- `GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1 -v` passes.
- `GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99|TestGCCC99Extra|TestGCCC90AsC99' -count=1 -v` passes.
- `GOCACHE=/tmp/cvm-go-cache go test ./... -count=1` passes.
- `Compiler.RunFile` uses preprocessing before parsing.
- Existing lexer/parser/sema tests remain green.
- The primary GCC C99 manifest has fewer skipped cases, and remaining skipped
  rows have specific non-generic reasons.
- Standard header shims are minimal, documented, and covered by integration
  tests.
- Source diagnostics preserve original file/line/column for ordinary files,
  includes, and macro expansions.

## Implementation Notes

- Keep `preprocessor` independent from `sema`; it may import `entity` and
  `common`, but sema must not become part of macro expansion.
- Keep macro expansion token-based. Avoid whole-source string rewriting as the
  primary representation.
- Add small files by responsibility instead of a single large preprocessor file.
- Prefer deterministic predefined macro behavior in tests.
- Do not search host system headers in the first implementation.
- Use Chinese comments for newly added non-obvious implementation logic.
