# C99 Preprocessor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a C99 token-level preprocessor that feeds cvm's existing parser and sema pipeline while preserving source locations for files, includes, macro expansions, and `#line`.

**Architecture:** Add a new `preprocessor` package with preprocessing tokens, a source manager, scanner, token-source context stack, directive parser, macro expander, include resolver, target-driven header shims, and parser-token conversion. Keep parser and sema mostly unchanged by returning `[]entity.Token`; only extend `entity.SourcePos` and compiler error rendering to understand source-manager-backed locations.

**Tech Stack:** Go 1.23 stdlib, existing `common`, `entity`, `lexer`, `parser`, and `sema` packages, local GCC fixture directories under `sema/testdata`.

---

## File Structure

Create:

- `preprocessor/doc.go` - package overview.
- `preprocessor/token.go` - preprocessing token type, token flags, token-source interfaces, and parser-token conversion helpers.
- `preprocessor/source.go` - `SourceManager`, file table, line table, physical/presumed/spelling/expansion location lookup, and diagnostic line rendering.
- `preprocessor/options.go` - `Options`, `Standard`, `TargetInfo`, `FileSystem`, defaults, and command-line-style define/undefine application.
- `preprocessor/scanner.go` - preprocessing scanner for trigraphs, line splicing, comments, pp tokens, newlines, and flags.
- `preprocessor/error.go` - preprocessor diagnostics returning `*common.CvmError`.
- `preprocessor/expr.go` - `#if` expression evaluator with preprocessing integer semantics.
- `preprocessor/directive.go` - directive parser and conditional-inclusion state machine.
- `preprocessor/macro.go` - macro definition model, macro table, builtin macros, disable/enable state.
- `preprocessor/expand.go` - token-source context stack, macro expansion, function-like macro arguments, blue paint, `#`, `##`, variadics, `_Pragma`.
- `preprocessor/include.go` - include resolver, include stack, include depth guard, and virtual filesystem support.
- `preprocessor/headers.go` - built-in C99 standard header shim source.
- `preprocessor/preprocessor.go` - public `PreprocessFile` / `PreprocessSource` API and top-level orchestration.
- `preprocessor/*_test.go` - focused unit tests for each component.

Modify:

- `entity/token.go` - extend `SourcePos` with `LocationID` while keeping legacy line/column constructors working.
- `compiler.go` - call `preprocessor` before `parser`, store the returned `SourceManager`, and render diagnostics through it.
- `sema/gcc_c99_test.go` - route GCC C99 fixtures through the preprocessor once integration is ready.
- `sema/gcc_c99_extra_test.go` and `sema/gcc_c90_as_c99_test.go` - keep current imported cases green; only expand scope when evidence supports it.
- `sema/testdata/gcc-c99/manifest.tsv` and related fixture directories - reclassify preprocessor-dependent C99 cases after the preprocessor is wired.
- `go.mod` only if an implementation step proves a small dependency is necessary; prefer stdlib.

Do not modify:

- `parser/glr_table.go` unless a preprocessed token exposes a real grammar gap.
- `sema` semantics as part of preprocessor tasks unless a GCC case clearly moves from "preprocessor missing" to a real sema mismatch; record that as a separate follow-up.

## Task 1: Source Locations And Public API Skeleton

**Files:**
- Modify: `entity/token.go`
- Create: `preprocessor/doc.go`
- Create: `preprocessor/options.go`
- Create: `preprocessor/source.go`
- Create: `preprocessor/error.go`
- Create: `preprocessor/preprocessor.go`
- Test: `preprocessor/source_test.go`
- Test: `preprocessor/preprocessor_test.go`

- [ ] **Step 1: Write failing source-manager tests**

Create `preprocessor/source_test.go`:

```go
package preprocessor

import "testing"

func TestSourceManagerPhysicalAndPresumedLocations(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "int a;\n#line 42 \"virt.c\"\nint b;\n")
	locA := sm.Location(fileID, 0)
	gotA := sm.DisplayLocation(locA)
	if gotA.File != "main.c" || gotA.Line != 1 || gotA.Column != 1 {
		t.Fatalf("physical location = %#v, want main.c:1:1", gotA)
	}

	lineDirectiveOffset := len("int a;\n")
	sm.SetPresumedLine(fileID, lineDirectiveOffset+len("#line 42 \"virt.c\"\n"), "virt.c", 42)
	locB := sm.Location(fileID, len("int a;\n#line 42 \"virt.c\"\n"))
	gotB := sm.DisplayLocation(locB)
	if gotB.File != "virt.c" || gotB.Line != 42 || gotB.Column != 1 {
		t.Fatalf("presumed location = %#v, want virt.c:42:1", gotB)
	}
}

func TestSourceManagerExpansionLocations(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "#define A 1\nA\n")
	defLoc := sm.Location(fileID, len("#define "))
	useLoc := sm.Location(fileID, len("#define A 1\n"))
	spellingLoc := sm.Location(fileID, len("#define A "))
	expLoc := sm.AddExpansion(spellingLoc, useLoc, defLoc)

	got := sm.ExpansionTrace(expLoc)
	if got.Spelling.LocationID != spellingLoc.LocationID {
		t.Fatalf("spelling loc = %#v, want %#v", got.Spelling, spellingLoc)
	}
	if got.Expansion.LocationID != useLoc.LocationID {
		t.Fatalf("expansion loc = %#v, want %#v", got.Expansion, useLoc)
	}
	if got.Definition.LocationID != defLoc.LocationID {
		t.Fatalf("definition loc = %#v, want %#v", got.Definition, defLoc)
	}
}
```

- [ ] **Step 2: Write failing API skeleton tests**

Create `preprocessor/preprocessor_test.go`:

```go
package preprocessor

import "testing"

func TestPreprocessSourceReturnsEOFTokenAndSourceManager(t *testing.T) {
	res, err := PreprocessSource("main.c", "int x;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if res.Sources == nil {
		t.Fatalf("Sources is nil")
	}
	if len(res.Tokens) == 0 {
		t.Fatalf("Tokens is empty")
	}
	if got := res.Tokens[len(res.Tokens)-1].Lexeme; got != "" {
		t.Fatalf("EOF lexeme = %q, want empty", got)
	}
}
```

- [ ] **Step 3: Run tests and verify they fail**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestSourceManager|TestPreprocessSource' -count=1 -v
```

Expected: FAIL because package `preprocessor` and its API do not exist.

- [ ] **Step 4: Extend `entity.SourcePos` compatibly**

Modify `entity/token.go` so `SourcePos` keeps line/column and gains a location handle:

```go
type SourcePos struct {
	LocationID int
	Line       int
	Column     int
}
```

Keep `NewToken` unchanged except that it leaves `LocationID` as zero:

```go
SourceStart: SourcePos{Line: line, Column: sc},
SourceEnd:   SourcePos{Line: line, Column: ec},
```

- [ ] **Step 5: Add API and source-manager implementation**

Create `preprocessor/options.go`:

```go
package preprocessor

type Standard int

const (
	StandardC99 Standard = iota
)

type TargetInfo struct {
	SizeType    string
	PtrdiffType string
	WCharType   string
	CharSigned  bool
	Hosted      bool
}

func DefaultTarget() TargetInfo {
	return TargetInfo{
		SizeType:    "unsigned long",
		PtrdiffType: "long",
		WCharType:   "int",
		CharSigned:  true,
		Hosted:      true,
	}
}

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Options struct {
	IncludePaths []string
	Defines      []string
	Undefines    []string
	Std          Standard
	Target       TargetInfo
	FileSystem   FileSystem
}

func normalizeOptions(opts Options) Options {
	if opts.Target.SizeType == "" {
		opts.Target = DefaultTarget()
	}
	if opts.Std != StandardC99 {
		opts.Std = StandardC99
	}
	return opts
}
```

Create `preprocessor/source.go` with these exported types and methods:

```go
package preprocessor

import (
	"sort"
	"strings"

	"shinya.click/cvm/entity"
)

type DisplayLocation struct {
	File   string
	Line   int
	Column int
	Text   string
}

type ExpansionTrace struct {
	Spelling   entity.SourcePos
	Expansion  entity.SourcePos
	Definition entity.SourcePos
}

type SourceManager struct {
	files      []sourceFile
	locations []sourceLocation
	expansions []ExpansionTrace
}

type sourceFile struct {
	name      string
	content   string
	lineStart []int
	presumed  []presumedLine
}

type sourceLocation struct {
	fileID int
	offset int
}

type presumedLine struct {
	offset int
	file   string
	line   int
}

func NewSourceManager() *SourceManager {
	return &SourceManager{
		files: []sourceFile{{name: "<invalid>", lineStart: []int{0}}},
		locations: []sourceLocation{{}},
		expansions: []ExpansionTrace{{}},
	}
}

func (sm *SourceManager) AddFile(name, content string) int {
	starts := []int{0}
	for i, b := range content {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	sm.files = append(sm.files, sourceFile{name: name, content: content, lineStart: starts})
	return len(sm.files) - 1
}

func (sm *SourceManager) Location(fileID, offset int) entity.SourcePos {
	sm.locations = append(sm.locations, sourceLocation{fileID: fileID, offset: offset})
	return entity.SourcePos{LocationID: len(sm.locations) - 1}
}

func (sm *SourceManager) SetPresumedLine(fileID, offset int, file string, line int) {
	f := &sm.files[fileID]
	f.presumed = append(f.presumed, presumedLine{offset: offset, file: file, line: line})
	sort.Slice(f.presumed, func(i, j int) bool { return f.presumed[i].offset < f.presumed[j].offset })
}

func (sm *SourceManager) AddExpansion(spelling, expansion, definition entity.SourcePos) entity.SourcePos {
	sm.expansions = append(sm.expansions, ExpansionTrace{Spelling: spelling, Expansion: expansion, Definition: definition})
	idx := len(sm.expansions) - 1
	return entity.SourcePos{LocationID: -idx}
}

func (sm *SourceManager) ExpansionTrace(pos entity.SourcePos) ExpansionTrace {
	if pos.LocationID >= 0 {
		return ExpansionTrace{}
	}
	idx := -pos.LocationID
	if idx <= 0 || idx >= len(sm.expansions) {
		return ExpansionTrace{}
	}
	return sm.expansions[idx]
}

func (sm *SourceManager) DisplayLocation(pos entity.SourcePos) DisplayLocation {
	if pos.LocationID == 0 {
		return DisplayLocation{Line: pos.Line, Column: pos.Column}
	}
	if pos.LocationID < 0 {
		trace := sm.ExpansionTrace(pos)
		if trace.Expansion.LocationID != 0 {
			return sm.DisplayLocation(trace.Expansion)
		}
		return DisplayLocation{}
	}
	if pos.LocationID >= len(sm.locations) {
		return DisplayLocation{}
	}
	loc := sm.locations[pos.LocationID]
	if loc.fileID <= 0 || loc.fileID >= len(sm.files) {
		return DisplayLocation{}
	}
	f := sm.files[loc.fileID]
	line := sort.Search(len(f.lineStart), func(i int) bool { return f.lineStart[i] > loc.offset })
	if line == 0 {
		line = 1
	}
	lineStart := f.lineStart[line-1]
	display := DisplayLocation{File: f.name, Line: line, Column: loc.offset - lineStart + 1}
	if len(f.presumed) > 0 {
		idx := sort.Search(len(f.presumed), func(i int) bool { return f.presumed[i].offset > loc.offset }) - 1
		if idx >= 0 {
			p := f.presumed[idx]
			baseLine := sort.Search(len(f.lineStart), func(i int) bool { return f.lineStart[i] > p.offset })
			if baseLine == 0 {
				baseLine = 1
			}
			display.File = p.file
			display.Line = p.line + (line - baseLine)
		}
	}
	lineEnd := strings.IndexByte(f.content[lineStart:], '\n')
	if lineEnd < 0 {
		display.Text = f.content[lineStart:]
	} else {
		display.Text = f.content[lineStart : lineStart+lineEnd]
	}
	return display
}
```

Create `preprocessor/preprocessor.go` with a temporary delegation to the existing lexer so Task 1 passes while later tasks replace the internals:

```go
package preprocessor

import (
	"os"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/lexer"
)

type Result struct {
	Tokens  []entity.Token
	Sources *SourceManager
}

func PreprocessSource(name, source string, opts Options) (*Result, error) {
	opts = normalizeOptions(opts)
	_ = opts
	sm := NewSourceManager()
	sm.AddFile(name, source)
	tokens, err := lexer.NewLexer(source).ScanTokens()
	if err != nil {
		return nil, err
	}
	return &Result{Tokens: tokens, Sources: sm}, nil
}

func PreprocessFile(path string, opts Options) (*Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return PreprocessSource(path, string(content), opts)
}
```

Create `preprocessor/error.go`:

```go
package preprocessor

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func ppError(pos entity.SourcePos, format string, args ...any) *common.CvmError {
	return common.NewCvmError(common.NewErrorMessage(pos, fmt.Sprintf(format, args...)))
}
```

Create `preprocessor/doc.go`:

```go
// Package preprocessor implements the C99 preprocessing stage for cvm.
package preprocessor
```

- [ ] **Step 6: Run tests and full suite**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestSourceManager|TestPreprocessSource' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add entity/token.go preprocessor
git commit -m "feat(preprocessor): add source manager and public API"
```

## Task 2: Preprocessing Token Scanner

**Files:**
- Create: `preprocessor/token.go`
- Create: `preprocessor/scanner.go`
- Test: `preprocessor/scanner_test.go`

- [ ] **Step 1: Write failing scanner tests**

Create `preprocessor/scanner_test.go`:

```go
package preprocessor

import "testing"

func TestScannerCommentsSplicesAndFlags(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "int/**/x \\\n= 1;\n#define A 2\n")
	toks, err := scanFile(sm, fileID, "int/**/x \\\n= 1;\n#define A 2\n", Options{})
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}
	want := []string{"int", "x", "=", "1", ";", "\n", "#", "define", "A", "2", "\n"}
	if got := ppLexemes(toks); !sameStrings(got, want) {
		t.Fatalf("lexemes = %#v, want %#v", got, want)
	}
	if !toks[0].StartOfLine {
		t.Fatalf("first token should start a line")
	}
	if !toks[1].LeadingSpace {
		t.Fatalf("comment replacement should leave leading space before x")
	}
	if !toks[2].LeadingSpace {
		t.Fatalf("line splice should preserve leading space before =")
	}
}

func TestScannerTrigraphWhenC99(t *testing.T) {
	sm := NewSourceManager()
	fileID := sm.AddFile("main.c", "??=define A 1\nA\n")
	toks, err := scanFile(sm, fileID, "??=define A 1\nA\n", Options{Std: StandardC99})
	if err != nil {
		t.Fatalf("scanFile failed: %v", err)
	}
	want := []string{"#", "define", "A", "1", "\n", "A", "\n"}
	if got := ppLexemes(toks); !sameStrings(got, want) {
		t.Fatalf("lexemes = %#v, want %#v", got, want)
	}
	if !toks[0].NeedsCleaning {
		t.Fatalf("trigraph-produced token should need cleaning")
	}
}
```

Include test helpers in the same file:

```go
func ppLexemes(toks []PPToken) []string {
	out := make([]string, 0, len(toks))
	for _, tok := range toks {
		out = append(out, tok.Lexeme)
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run scanner tests and verify failure**

Run:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run TestScanner -count=1 -v
```

Expected: FAIL because `PPToken` and `scanFile` do not exist.

- [ ] **Step 3: Add preprocessing token model**

Create `preprocessor/token.go`:

```go
package preprocessor

import "shinya.click/cvm/entity"

type PPTokenKind int

const (
	PPIdentifier PPTokenKind = iota
	PPNumber
	PPString
	PPCharacter
	PPPunctuator
	PPHeaderName
	PPNewline
	PPEOF
)

type PPToken struct {
	Kind          PPTokenKind
	Lexeme        string
	Location      entity.SourcePos
	StartOfLine   bool
	LeadingSpace  bool
	DisableExpand bool
	NeedsCleaning bool
}

type TokenSource interface {
	Lex() (PPToken, error)
}
```

- [ ] **Step 4: Implement scanner**

Create `preprocessor/scanner.go` with a byte scanner that:

- performs trigraph replacement before line splicing;
- replaces block and line comments with a single leading-space marker while preserving newlines;
- emits `PPNewline` tokens;
- emits `StartOfLine` on the first non-newline token after a newline;
- emits `LeadingSpace` when whitespace/comment/splice whitespace preceded the token;
- recognizes identifiers, pp-numbers, string literals, character literals, `#`, `##`, `...`, and ordinary C punctuators;
- sets `NeedsCleaning` on tokens whose spelling came from trigraph or escaped newline cleanup.

Use this exact function signature:

```go
func scanFile(sm *SourceManager, fileID int, source string, opts Options) ([]PPToken, error)
```

Add a helper for trigraphs:

```go
func translateTrigraph(a, b, c byte) (byte, bool) {
	if a != '?' || b != '?' {
		return 0, false
	}
	switch c {
	case '=':
		return '#', true
	case '/':
		return '\\', true
	case '\'':
		return '^', true
	case '(':
		return '[', true
	case ')':
		return ']', true
	case '!':
		return '|', true
	case '<':
		return '{', true
	case '>':
		return '}', true
	case '-':
		return '~', true
	default:
		return 0, false
	}
}
```

Add Chinese comments before the comment-replacement and line-splicing blocks, because those are non-obvious C99 translation-phase behavior.

- [ ] **Step 5: Run scanner tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run TestScanner -count=1 -v
```

Expected: PASS.

- [ ] **Step 6: Run package tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add preprocessor/token.go preprocessor/scanner.go preprocessor/scanner_test.go
git commit -m "feat(preprocessor): scan C99 preprocessing tokens"
```

## Task 3: Parser Token Conversion And Compiler Location Compatibility

**Files:**
- Modify: `preprocessor/token.go`
- Modify: `preprocessor/preprocessor.go`
- Modify: `compiler.go`
- Test: `preprocessor/convert_test.go`
- Test: `compiler_test.go`

- [ ] **Step 1: Write failing conversion tests**

Create `preprocessor/convert_test.go`:

```go
package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestConvertExpandedTokensToParserTokens(t *testing.T) {
	res, err := PreprocessSource("main.c", "int x = 1;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	want := []entity.TokenType{entity.INT, entity.IDENTIFIER, entity.EQUAL, entity.INTEGER_CONSTANT, entity.SEMICOLON, entity.EOF}
	if len(res.Tokens) != len(want) {
		t.Fatalf("token count = %d, want %d: %#v", len(res.Tokens), len(want), res.Tokens)
	}
	for i, typ := range want {
		if res.Tokens[i].Typ != typ {
			t.Fatalf("token[%d] type = %s, want %s", i, res.Tokens[i].Typ, typ)
		}
		if res.Tokens[i].SourceStart.LocationID == 0 && typ != entity.EOF {
			t.Fatalf("token[%d] missing source location: %#v", i, res.Tokens[i])
		}
	}
}
```

- [ ] **Step 2: Run conversion test and verify failure**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run TestConvertExpandedTokensToParserTokens -count=1 -v
```

Expected: FAIL because `PreprocessSource` still delegates to `lexer.NewLexer` and returns legacy locations.

- [ ] **Step 3: Implement parser-token conversion**

In `preprocessor/token.go`, add:

```go
func convertToParserTokens(tokens []PPToken, sm *SourceManager) ([]entity.Token, error)
```

Rules:

- Drop `PPNewline`.
- Map C keywords through the same spelling table as `lexer/identifier.go`.
- Map punctuators using the same spellings as `lexer/punctuator.go`.
- Parse integer and floating constants by scanning the token lexeme with `lexer.NewLexer(tok.Lexeme).ScanTokens()` and copying the literal.
- Parse string and character constants the same way.
- Reject remaining `#`, `##`, and header-name tokens in ordinary output.
- Append EOF with a source location at the last emitted token or file end.

Use local maps in `preprocessor` instead of importing unexported lexer maps:

```go
var keywordTokens = map[string]entity.TokenType{
	"auto": entity.AUTO,
	"break": entity.BREAK,
	"case": entity.CASE,
	"char": entity.CHAR,
	"const": entity.CONST,
	"continue": entity.CONTINUE,
	"default": entity.DEFAULT,
	"do": entity.DO,
	"double": entity.DOUBLE,
	"else": entity.ELSE,
	"enum": entity.ENUM,
	"extern": entity.EXTERN,
	"float": entity.FLOAT,
	"for": entity.FOR,
	"goto": entity.GOTO,
	"if": entity.IF,
	"inline": entity.INLINE,
	"int": entity.INT,
	"long": entity.LONG,
	"register": entity.REGISTER,
	"restrict": entity.RESTRICT,
	"return": entity.RETURN,
	"short": entity.SHORT,
	"signed": entity.SIGNED,
	"sizeof": entity.SIZEOF,
	"static": entity.STATIC,
	"struct": entity.STRUCT,
	"switch": entity.SWITCH,
	"typedef": entity.TYPEDEF,
	"union": entity.UNION,
	"unsigned": entity.UNSIGNED,
	"void": entity.VOID,
	"volatile": entity.VOLATILE,
	"while": entity.WHILE,
	"_Bool": entity.BOOL,
	"_Complex": entity.COMPLEX,
	"_Imaginary": entity.IMAGINARY,
	"_Static_assert": entity.STATIC_ASSERT,
}
```

Add a parallel `punctuatorTokens` map covering every punctuator in `entity/token.go`.

- [ ] **Step 4: Route `PreprocessSource` through scanner and conversion**

Replace the temporary `lexer.NewLexer` delegation in `preprocessor/preprocessor.go`:

```go
func PreprocessSource(name, source string, opts Options) (*Result, error) {
	opts = normalizeOptions(opts)
	sm := NewSourceManager()
	fileID := sm.AddFile(name, source)
	ppTokens, err := scanFile(sm, fileID, source, opts)
	if err != nil {
		return nil, err
	}
	tokens, err := convertToParserTokens(ppTokens, sm)
	if err != nil {
		return nil, err
	}
	return &Result{Tokens: tokens, Sources: sm}, nil
}
```

- [ ] **Step 5: Update compiler to store source manager**

Modify `compiler.go`:

```go
type Compiler struct {
	FileName string
	Source   string
	Lines    []string
	Sources  *preprocessor.SourceManager
}
```

In `RunSource`, call `preprocessor.PreprocessSource(c.FileName, source, preprocessor.Options{})`.

In `handleError`, when `c.Sources != nil`, render each message with:

```go
display := c.Sources.DisplayLocation(message.SourcePos)
file := display.File
if file == "" {
	file = c.FileName
}
line := display.Line
column := display.Column
text := display.Text
```

Keep the old `Lines` fallback when `display.Text == ""` and `message.SourcePos.Line > 0`.

- [ ] **Step 6: Run tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run TestConvertExpandedTokensToParserTokens -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add compiler.go preprocessor/token.go preprocessor/preprocessor.go preprocessor/convert_test.go
git commit -m "feat(preprocessor): convert preprocessing tokens for parser"
```

## Task 4: Directives And Conditional Inclusion

**Files:**
- Create: `preprocessor/directive.go`
- Create: `preprocessor/expr.go`
- Modify: `preprocessor/preprocessor.go`
- Test: `preprocessor/directive_test.go`
- Test: `preprocessor/expr_test.go`

- [ ] **Step 1: Write failing directive and conditional tests**

Create `preprocessor/directive_test.go`:

```go
package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestConditionalInclusionSelectsActiveBranch(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#if 0
int bad = ;
#elif 1
int good;
#else
int also_bad = ;
#endif
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "good") {
		t.Fatalf("expected active branch token good in %#v", res.Tokens)
	}
	if hasIdentifier(res.Tokens, "bad") || hasIdentifier(res.Tokens, "also_bad") {
		t.Fatalf("inactive branch tokens leaked: %#v", res.Tokens)
	}
}

func TestLineDirectiveChangesDisplayedLocation(t *testing.T) {
	res, err := PreprocessSource("main.c", "#line 77 \"virt.c\"\nint x;\n", Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	for _, tok := range res.Tokens {
		if tok.Typ == entity.INT {
			got := res.Sources.DisplayLocation(tok.SourceStart)
			if got.File != "virt.c" || got.Line != 77 {
				t.Fatalf("display location = %#v, want virt.c:77", got)
			}
			return
		}
	}
	t.Fatalf("INT token not found")
}

func hasIdentifier(tokens []entity.Token, name string) bool {
	for _, tok := range tokens {
		if tok.Typ == entity.IDENTIFIER && tok.Lexeme == name {
			return true
		}
	}
	return false
}
```

Create `preprocessor/expr_test.go`:

```go
package preprocessor

import "testing"

func TestIfExpressionDefinedAndIdentifiers(t *testing.T) {
	pp := newPreprocessor("main.c", "", Options{})
	pp.macros.DefineObject("A", []PPToken{{Kind: PPNumber, Lexeme: "4"}})
	got, err := pp.evalIfExpression([]PPToken{
		{Kind: PPIdentifier, Lexeme: "defined"},
		{Kind: PPPunctuator, Lexeme: "("},
		{Kind: PPIdentifier, Lexeme: "A"},
		{Kind: PPPunctuator, Lexeme: ")"},
		{Kind: PPPunctuator, Lexeme: "&&"},
		{Kind: PPIdentifier, Lexeme: "MISSING"},
		{Kind: PPPunctuator, Lexeme: "=="},
		{Kind: PPNumber, Lexeme: "0"},
	})
	if err != nil {
		t.Fatalf("evalIfExpression failed: %v", err)
	}
	if got == 0 {
		t.Fatalf("expression evaluated false, want true")
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestConditional|TestLineDirective|TestIfExpression' -count=1 -v
```

Expected: FAIL because directive parsing and `newPreprocessor` do not exist.

- [ ] **Step 3: Add preprocessor state and directive pipeline**

Create `preprocessor/directive.go` with:

```go
type preprocessor struct {
	opts    Options
	sm      *SourceManager
	macros  *MacroTable
	output  []PPToken
	conds   []conditionalGroup
}

type conditionalGroup struct {
	parentActive bool
	branchTaken  bool
	active       bool
	seenElse     bool
}

func newPreprocessor(name, source string, opts Options) *preprocessor
func (pp *preprocessor) process(tokens []PPToken) ([]PPToken, error)
func (pp *preprocessor) handleDirective(line []PPToken) error
func (pp *preprocessor) isActive() bool
```

Implement line grouping so `#` is treated as a directive only when its token has `StartOfLine`.

Rules:

- Ordinary tokens emit only when all conditional groups are active.
- Inactive branches still parse nested conditional directives.
- `#if`, `#ifdef`, `#ifndef`, `#elif`, `#else`, and `#endif` update `pp.conds`.
- End of file with non-empty `pp.conds` reports an unterminated conditional error.
- `#line N "file"` calls `SourceManager.SetPresumedLine`.
- `#error ...` reports an error with the directive location.

- [ ] **Step 4: Implement `#if` expression evaluator**

Create `preprocessor/expr.go`:

```go
func (pp *preprocessor) evalIfExpression(tokens []PPToken) (int64, error)
```

Implement recursive-descent precedence for:

- unary `+`, `-`, `!`, `~`
- `*`, `/`, `%`
- `+`, `-`
- `<<`, `>>`
- `<`, `<=`, `>`, `>=`
- `==`, `!=`
- `&`, `^`, `|`
- `&&`, `||`
- parentheses
- integer constants parsed with base 8/10/16
- `defined NAME` and `defined(NAME)`
- remaining identifiers become `0`

Use `int64` for signed values in this task. Add a file comment explaining that the implementation models C99 preprocessing integer evaluation and can be widened to explicit `uintmax_t` when GCC cases demand it.

- [ ] **Step 5: Route `PreprocessSource` through directive processing**

In `preprocessor/preprocessor.go`, after `scanFile` and before `convertToParserTokens`, call:

```go
pp := newPreprocessor(name, source, opts)
pp.sm = sm
processed, err := pp.process(ppTokens)
if err != nil {
	return nil, err
}
tokens, err := convertToParserTokens(processed, sm)
```

- [ ] **Step 6: Run tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestConditional|TestLineDirective|TestIfExpression' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add preprocessor/directive.go preprocessor/expr.go preprocessor/directive_test.go preprocessor/expr_test.go preprocessor/preprocessor.go
git commit -m "feat(preprocessor): handle directives and conditionals"
```

## Task 5: Macro Table And Expansion Context Stack

**Files:**
- Create: `preprocessor/macro.go`
- Create: `preprocessor/expand.go`
- Modify: `preprocessor/directive.go`
- Modify: `preprocessor/preprocessor.go`
- Test: `preprocessor/macro_test.go`
- Test: `preprocessor/expand_test.go`

- [ ] **Step 1: Write failing macro tests**

Create `preprocessor/macro_test.go`:

```go
package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestObjectAndFunctionLikeMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define T int
#define ADD(a, b) ((a) + (b))
T x = ADD(1, 2);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	want := []entity.TokenType{entity.INT, entity.IDENTIFIER, entity.EQUAL, entity.LEFT_PARENTHESES, entity.LEFT_PARENTHESES, entity.INTEGER_CONSTANT}
	for i, typ := range want {
		if res.Tokens[i].Typ != typ {
			t.Fatalf("token[%d] = %s, want %s; tokens=%#v", i, res.Tokens[i].Typ, typ, res.Tokens)
		}
	}
}

func TestStringifyPasteAndVariadicMacros(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define STR(x) #x
#define CAT(a, b) a ## b
#define FIRST(fmt, ...) fmt
char *s = STR(hello world);
int CAT(ab, c) = 3;
char *f = FIRST("x", 1, 2);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "abc") {
		t.Fatalf("token paste did not produce abc: %#v", res.Tokens)
	}
	if !hasString(res.Tokens, "\"hello world\"") {
		t.Fatalf("stringification did not produce expected string token: %#v", res.Tokens)
	}
}

func hasString(tokens []entity.Token, lexeme string) bool {
	for _, tok := range tokens {
		if tok.Typ == entity.STRING && tok.Lexeme == lexeme {
			return true
		}
	}
	return false
}
```

Create `preprocessor/expand_test.go`:

```go
package preprocessor

import "testing"

func TestBluePaintRecursiveSuppression(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define A A
int A;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "A") {
		t.Fatalf("self-referential macro should leave disabled A token: %#v", res.Tokens)
	}
}

func TestMacroArgumentPrescan(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#define F(x) x + x
#define ONE 1
int y = F(ONE);
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	count := 0
	for _, tok := range res.Tokens {
		if tok.Lexeme == "1" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expanded 1 count = %d, want 2; tokens=%#v", count, res.Tokens)
	}
}
```

- [ ] **Step 2: Run macro tests and verify failure**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestObject|TestStringify|TestBluePaint|TestMacroArgument' -count=1 -v
```

Expected: FAIL because macro definition and expansion are not implemented.

- [ ] **Step 3: Implement macro definitions**

Create `preprocessor/macro.go` with:

```go
type MacroKind int

const (
	MacroObject MacroKind = iota
	MacroFunction
)

type Macro struct {
	Name       string
	Kind       MacroKind
	Params     []string
	Variadic   bool
	Replacement []PPToken
	Definition entity.SourcePos
	Disabled   bool
}

type MacroTable struct {
	entries map[string]*Macro
}

func NewMacroTable(target TargetInfo) *MacroTable
func (m *MacroTable) DefineObject(name string, replacement []PPToken)
func (m *MacroTable) DefineFunction(name string, params []string, variadic bool, replacement []PPToken, pos entity.SourcePos)
func (m *MacroTable) Undef(name string)
func (m *MacroTable) Lookup(name string) (*Macro, bool)
```

`NewMacroTable` must define `__STDC__`, `__STDC_VERSION__`, `__STDC_HOSTED__`, and target type macros such as `__SIZE_TYPE__`, `__PTRDIFF_TYPE__`, and `__WCHAR_TYPE__`.

Update `#define` and `#undef` handling in `directive.go`:

- object-like macro if no `(` immediately follows the name;
- function-like macro only when `(` immediately follows the name without `LeadingSpace`;
- variadic if the parameter list contains `...`;
- replacement list preserves token flags and locations;
- redefinition with an identical replacement is accepted; conflicting replacement is a diagnostic.

- [ ] **Step 4: Implement expansion context stack**

Create `preprocessor/expand.go` with:

```go
type macroTokenSource struct {
	tokens []PPToken
	index  int
	macro  *Macro
}

func (pp *preprocessor) expand(tokens []PPToken) ([]PPToken, error)
func (pp *preprocessor) expandOne(tok PPToken, rest []PPToken) ([]PPToken, int, error)
func (pp *preprocessor) collectMacroArgs(rest []PPToken, openParenIndex int) ([][]PPToken, int, error)
func (pp *preprocessor) substitute(m *Macro, args [][]PPToken, use entity.SourcePos) ([]PPToken, error)
```

Required rules:

- If `tok.DisableExpand` is true, emit it unchanged.
- If an identifier names a disabled macro, set `DisableExpand` and emit it unchanged.
- Disable a macro while scanning its replacement list; re-enable only after that replacement context is exhausted.
- Prescan macro arguments unless the parameter is adjacent to `#` or `##` in the replacement list.
- Implement `#` by producing a `PPString` token with escaped argument spelling.
- Implement `##` by concatenating adjacent token lexemes, retokenizing the result with `scanFile`, and requiring exactly one non-newline token.
- Implement `__VA_ARGS__` substitution for variadic macros.
- Implement `_Pragma("...")` recognition after macro expansion; first behavior is to consume it and emit no ordinary token.

Add Chinese comments in the disable/reenable and `DisableExpand` blocks explaining the blue-paint rule.

- [ ] **Step 5: Call expansion from the pipeline**

In `preprocessor/preprocessor.go`, after directive processing:

```go
expanded, err := pp.expand(processed)
if err != nil {
	return nil, err
}
tokens, err := convertToParserTokens(expanded, sm)
```

Do not expand directive lines as ordinary output.

- [ ] **Step 6: Run macro tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestObject|TestStringify|TestBluePaint|TestMacroArgument' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add preprocessor/macro.go preprocessor/expand.go preprocessor/macro_test.go preprocessor/expand_test.go preprocessor/directive.go preprocessor/preprocessor.go
git commit -m "feat(preprocessor): expand C99 macros"
```

## Task 6: Include Resolver, Virtual Filesystem, TargetInfo, And Standard Header Shims

**Files:**
- Create: `preprocessor/include.go`
- Create: `preprocessor/headers.go`
- Modify: `preprocessor/options.go`
- Modify: `preprocessor/directive.go`
- Test: `preprocessor/include_test.go`
- Test: `preprocessor/headers_test.go`

- [ ] **Step 1: Write failing include and header tests**

Create `preprocessor/include_test.go`:

```go
package preprocessor

import (
	"fmt"
	"testing"
)

type mapFS map[string]string

func (m mapFS) ReadFile(path string) ([]byte, error) {
	if s, ok := m[path]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("missing %s", path)
}

func TestQuotedIncludeAndMacroExpandedInclude(t *testing.T) {
	fs := mapFS{
		"/work/main.c":  "#define H \"defs.h\"\n#include H\nT x;\n",
		"/work/defs.h": "typedef int T;\n",
	}
	res, err := PreprocessFile("/work/main.c", Options{FileSystem: fs})
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}
	if !hasIdentifier(res.Tokens, "T") {
		t.Fatalf("included typedef token missing: %#v", res.Tokens)
	}
}
```

Create `preprocessor/headers_test.go`:

```go
package preprocessor

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestBuiltinStandardHeaders(t *testing.T) {
	res, err := PreprocessSource("main.c", `
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
bool b = true;
size_t n;
intptr_t p;
`, Options{})
	if err != nil {
		t.Fatalf("PreprocessSource failed: %v", err)
	}
	if !hasToken(res.Tokens, entity.BOOL) {
		t.Fatalf("stdbool bool did not become _Bool: %#v", res.Tokens)
	}
	if !hasIdentifier(res.Tokens, "size_t") || !hasIdentifier(res.Tokens, "intptr_t") {
		t.Fatalf("standard typedef names missing: %#v", res.Tokens)
	}
}

func hasToken(tokens []entity.Token, typ entity.TokenType) bool {
	for _, tok := range tokens {
		if tok.Typ == typ {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run include/header tests and verify failure**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestQuotedInclude|TestBuiltinStandardHeaders' -count=1 -v
```

Expected: FAIL because include resolution and headers are not implemented.

- [ ] **Step 3: Implement filesystem and include resolver**

In `preprocessor/options.go`, add:

```go
type osFileSystem struct{}

func (osFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
```

Import `os`.

Create `preprocessor/include.go`:

```go
const maxIncludeDepth = 64

type IncludeResolver struct {
	opts Options
}

func newIncludeResolver(opts Options) IncludeResolver
func (r IncludeResolver) resolveQuoted(currentFile, name string) (resolvedPath string, content string, err error)
func (r IncludeResolver) resolveAngled(name string) (resolvedPath string, content string, err error)
```

Rules:

- Quoted include first checks the directory of `currentFile`, then `Options.IncludePaths`.
- Angled include checks `builtinHeader(name)` only.
- Missing include reports `include file not found: %s`.
- Track include depth in `preprocessor` and reject depth over `maxIncludeDepth`.

- [ ] **Step 4: Implement built-in headers**

Create `preprocessor/headers.go`:

```go
package preprocessor

import "fmt"

func builtinHeader(name string, target TargetInfo) (string, bool) {
	switch name {
	case "stdbool.h":
		return "#define bool _Bool\n#define true 1\n#define false 0\n#define __bool_true_false_are_defined 1\n", true
	case "stddef.h":
		return fmt.Sprintf("#define __SIZE_TYPE__ %s\n#define __PTRDIFF_TYPE__ %s\ntypedef __SIZE_TYPE__ size_t;\ntypedef __PTRDIFF_TYPE__ ptrdiff_t;\n#define NULL ((void *)0)\n", target.SizeType, target.PtrdiffType), true
	case "stdint.h":
		return fmt.Sprintf("typedef signed char int8_t;\ntypedef short int16_t;\ntypedef int int32_t;\ntypedef long int64_t;\ntypedef unsigned char uint8_t;\ntypedef unsigned short uint16_t;\ntypedef unsigned int uint32_t;\ntypedef unsigned long uint64_t;\ntypedef %s intptr_t;\ntypedef %s uintptr_t;\n#define INT8_MAX 127\n#define INT16_MAX 32767\n#define INT32_MAX 2147483647\n#define INT64_MAX 9223372036854775807L\n", target.PtrdiffType, target.SizeType), true
	case "iso646.h":
		return "#define and &&\n#define and_eq &=\n#define bitand &\n#define bitor |\n#define compl ~\n#define not !\n#define not_eq !=\n#define or ||\n#define or_eq |=\n#define xor ^\n#define xor_eq ^=\n", true
	case "limits.h":
		return "#define CHAR_BIT 8\n#define SCHAR_MAX 127\n#define UCHAR_MAX 255\n#define SHRT_MAX 32767\n#define USHRT_MAX 65535\n#define INT_MAX 2147483647\n#define UINT_MAX 4294967295U\n#define LONG_MAX 9223372036854775807L\n", true
	case "float.h":
		return "#define FLT_RADIX 2\n#define FLT_MANT_DIG 24\n#define DBL_MANT_DIG 53\n#define LDBL_MANT_DIG 64\n", true
	default:
		return "", false
	}
}
```

- [ ] **Step 5: Wire `#include`**

In `directive.go`, implement include handling:

- For `<name>` tokens, call `resolveAngled`.
- For `"name"` string tokens, unquote and call `resolveQuoted`.
- For any other operand, macro-expand operand first, then require a header-name or string token.
- Scan, process, and expand the included source using the same macro table and source manager.
- Built-in headers get virtual file names like `<stdint.h>`.
- Included tokens are injected into `pp.output` at the directive location.

Add Chinese comments before macro-expanded include operand handling.

- [ ] **Step 6: Run tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -run 'TestQuotedInclude|TestBuiltinStandardHeaders' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add preprocessor/include.go preprocessor/headers.go preprocessor/include_test.go preprocessor/headers_test.go preprocessor/options.go preprocessor/directive.go
git commit -m "feat(preprocessor): resolve includes and C99 header shims"
```

## Task 7: Compiler And Sema GCC Fixture Integration

**Files:**
- Modify: `compiler.go`
- Modify: `sema/gcc_c99_test.go`
- Modify: `sema/gcc_c99_extra_test.go`
- Modify: `sema/gcc_c90_as_c99_test.go`
- Test: `compiler_test.go`
- Test: `sema/gcc_c99_test.go`

- [ ] **Step 1: Write failing frontend integration tests**

Add to `compiler_test.go`:

```go
func TestCompilerRunsPreprocessor(t *testing.T) {
	(&Compiler{}).RunSource(`#define T int
T main(void) {
	return 0;
}`)
}
```

Add a new helper in `sema/gcc_c99_test.go`:

```go
func preprocessParseAnalyze(t *testing.T, path string, src string, opts SemaOptions) error {
	t.Helper()
	pp, err := preprocessor.PreprocessSource(path, src, preprocessor.Options{})
	if err != nil {
		return err
	}
	candidates, err := parser.NewParser(pp.Tokens).Parse()
	if err != nil {
		return err
	}
	_, err = AnalyzeWithOptions(candidates, opts)
	return err
}
```

Import `shinya.click/cvm/preprocessor`.

- [ ] **Step 2: Run targeted tests and verify failure or old path**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./ -run TestCompilerRunsPreprocessor -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run TestGCCC99AcceptSuite -count=1 -v
```

Expected before integration: compiler test fails if `Compiler.RunSource` still bypasses preprocessing, or GCC runner does not exercise preprocessing-dependent cases.

- [ ] **Step 3: Update compiler pipeline**

Ensure `compiler.go` imports `preprocessor` and uses:

```go
pp, err := preprocessor.PreprocessSource(c.FileName, source, preprocessor.Options{})
if err != nil {
	c.handleError(err)
	return
}
c.Sources = pp.Sources
candidates, err := parser.NewParser(pp.Tokens).Parse()
```

Remove the direct `lexer` import from `compiler.go`.

- [ ] **Step 4: Update GCC fixture runner**

In `sema/gcc_c99_test.go`, replace direct `lexer.NewLexer(src).ScanTokens()` in `runGCCC99Suite` with `preprocessor.PreprocessSource(path, src, preprocessor.Options{})`.

Do not use `stripCComments` before preprocessing. Keep removal of DejaGNU directive comments only if those comments are not valid source under preprocessing. Change `stripGCCDirectives` to remove only DejaGNU comment lines and leave ordinary C comments intact.

Keep `gccPedanticErrors(originalSrc)` based on the original source.

- [ ] **Step 5: Keep extra suites stable**

In `sema/gcc_c99_extra_test.go` and `sema/gcc_c90_as_c99_test.go`, keep the current imported case set unchanged for this task. If switching their shared runner to preprocessing changes behavior, fix only cases where the new behavior is strictly more C99-correct and note it in the commit message body.

- [ ] **Step 6: Run integration tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./ -run TestCompilerRunsPreprocessor -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99AcceptSuite|TestGCCC99RejectSuite' -count=1 -v
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS for the existing imported suites.

- [ ] **Step 7: Commit**

```bash
git add compiler.go compiler_test.go sema/gcc_c99_test.go sema/gcc_c99_extra_test.go sema/gcc_c90_as_c99_test.go
git commit -m "feat(frontend): run GCC fixtures through preprocessor"
```

## Task 8: Reclassify Primary GCC C99 Preprocessor Cases

**Files:**
- Modify: `sema/testdata/gcc-c99/manifest.tsv`
- Modify: `sema/testdata/gcc-c99/README.md`
- Add or move: `sema/testdata/gcc-c99/accept/*.c`
- Add or move: `sema/testdata/gcc-c99/reject/*.c`
- Test: `sema/gcc_c99_test.go`

- [ ] **Step 1: List current preprocessor skips**

Run:

```bash
awk -F '\t' 'NR>1 && $2=="skipped" {print $1 "\t" $3 "\t" $4}' sema/testdata/gcc-c99/manifest.tsv
```

Expected: the current list includes only rows classified as preprocessor/system macro/header dependent.

- [ ] **Step 2: Refresh candidates from local GCC checkout**

If `/tmp/cvm-gcc-c99/gcc/testsuite/gcc.dg` exists, use it as the source of truth. For each skipped `gcc/testsuite/gcc.dg/<name>.c`, copy the original into a temp directory and run it through the new preprocessor runner manually:

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run TestGCCC99 -count=1 -v
```

Expected: cases that are now supported should pass when moved into `accept/` or `reject/`; unsupported cases should fail for a specific remaining reason.

- [ ] **Step 3: Reclassify supported accept cases**

For each skipped accept case that now passes:

- copy or move the stripped fixture into `sema/testdata/gcc-c99/accept/<name>.c`;
- set manifest status to `imported-accept`;
- set reason to `preprocessed by cvm C99 preprocessor`;
- preserve source comments unless they are DejaGNU-only lines removed by the fixture policy.

Start with cases involving:

- `#include <stdbool.h>`
- `#include <stdint.h>`
- `#include <stddef.h>`
- simple object/function macros
- conditional inclusion
- `#line`

- [ ] **Step 4: Reclassify supported reject cases**

For each skipped reject case that the preprocessor now reaches and sema/parser rejects:

- copy or move it into `sema/testdata/gcc-c99/reject/<name>.c`;
- set manifest status to `imported-reject`;
- set reason to `preprocessed by cvm C99 preprocessor`;
- verify the rejection is not a preprocessor crash or unrelated scanner bug.

- [ ] **Step 5: Replace generic skip reasons**

For remaining skipped rows, replace `requires preprocessor or system macro handling` with one of these exact reasons:

- `requires real system header`
- `requires GCC pragma semantics`
- `requires GNU preprocessor extension`
- `requires target-specific runtime behavior`
- `requires frontend semantic follow-up after preprocessing`

Do not keep the generic reason in the primary `gcc-c99` manifest.

- [ ] **Step 6: Strengthen the manifest gate**

Update `TestGCCC99OnlyPreprocessorSkipsRemain` in `sema/gcc_c99_test.go` to reject the old generic reason and allow only the exact remaining reasons from Step 5.

Expected test body condition:

```go
allowed := map[string]bool{
	"requires real system header": true,
	"requires GCC pragma semantics": true,
	"requires GNU preprocessor extension": true,
	"requires target-specific runtime behavior": true,
	"requires frontend semantic follow-up after preprocessing": true,
}
```

- [ ] **Step 7: Run GCC C99 tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99' -count=1 -v
```

Expected: PASS.

- [ ] **Step 8: Run full suite**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add sema/gcc_c99_test.go sema/testdata/gcc-c99
git commit -m "test(sema): reclassify preprocessed GCC C99 cases"
```

## Task 9: Broaden Extra/C90 Coverage Opportunistically

**Files:**
- Modify: `sema/testdata/gcc-c99-extra/manifest.tsv`
- Modify: `sema/testdata/gcc-c99-extra/README.md`
- Modify: `sema/testdata/gcc-c90-as-c99/manifest.tsv`
- Modify: `sema/testdata/gcc-c90-as-c99/README.md`
- Add or move: `sema/testdata/gcc-c99-extra/accept/*.c`
- Add or move: `sema/testdata/gcc-c99-extra/reject/*.c`
- Add or move: `sema/testdata/gcc-c90-as-c99/accept/*.c`
- Add or move: `sema/testdata/gcc-c90-as-c99/reject/*.c`

- [ ] **Step 1: Identify skipped rows that mention preprocessor/system macros**

Run:

```bash
awk -F '\t' 'NR>1 && $2=="skipped" && $4=="requires preprocessor or system macro handling" {print $1 "\t" $3}' sema/testdata/gcc-c99-extra/manifest.tsv
awk -F '\t' 'NR>1 && $2=="skipped" && $4=="requires preprocessor or system macro handling" {print $1 "\t" $3}' sema/testdata/gcc-c90-as-c99/manifest.tsv
```

Expected: a list of extra and C90-as-C99 candidates that may now be importable.

- [ ] **Step 2: Import only cases covered by implemented behavior**

For each candidate:

- import cases using C99 standard headers, simple macros, include guards, and conditionals;
- skip cases requiring true system headers, GCC attributes hidden in headers, target runtime, or unsupported GNU preprocessor features;
- update manifest reasons specifically, following the reason vocabulary already used in each suite.

- [ ] **Step 3: Run suite-specific gates**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99Extra|TestGCCC90AsC99' -count=1 -v
```

Expected: PASS.

- [ ] **Step 4: Run full suite**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/testdata/gcc-c99-extra sema/testdata/gcc-c90-as-c99
git commit -m "test(sema): broaden GCC suites with preprocessed cases"
```

## Task 10: Final Verification And Documentation Cleanup

**Files:**
- Modify: `docs/superpowers/specs/2026-04-28-c99-preprocessor-design.md` only if implementation exposes a design correction.
- Modify: `README.md` only if the command-line frontend behavior needs user-facing notes.
- Modify: `preprocessor/doc.go` if package comments are out of date.

- [ ] **Step 1: Run preprocessor package tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./preprocessor/ -count=1 -v
```

Expected: PASS.

- [ ] **Step 2: Run GCC frontend gates**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./sema/ -run 'TestGCCC99|TestGCCC99Extra|TestGCCC90AsC99' -count=1 -v
```

Expected: PASS.

- [ ] **Step 3: Run all tests**

```bash
GOCACHE=/tmp/cvm-go-cache go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 4: Check for generic skip reason regression**

```bash
awk -F '\t' 'NR>1 && $2=="skipped" && $4=="requires preprocessor or system macro handling" {print}' sema/testdata/gcc-c99/manifest.tsv
```

Expected: no output.

- [ ] **Step 5: Check worktree**

```bash
git status --short --branch
```

Expected: clean branch except for intentional documentation updates.

- [ ] **Step 6: Commit final docs if needed**

If Step 5 shows documentation-only updates:

```bash
git add README.md docs/superpowers/specs/2026-04-28-c99-preprocessor-design.md preprocessor/doc.go
git commit -m "docs: document C99 preprocessor behavior"
```

If Step 5 is clean, do not create an empty commit.

## Self-Review Checklist

- Spec coverage:
  - Token-source context stack: Task 5.
  - Source manager and spelling/expansion/presumed locations: Task 1 and Task 3.
  - Token flags and preprocessing scanner: Task 2.
  - Directives and conditionals: Task 4.
  - Macro expansion, blue paint, stringification, token paste, variadics, `_Pragma`: Task 5.
  - Macro-expanded include and header shims: Task 6.
  - TargetInfo: Task 6.
  - Compiler and GCC gates: Task 7 through Task 10.
- Placeholder scan completed: the plan contains no unfinished placeholder steps.
- Type consistency:
  - Public API remains `PreprocessFile`, `PreprocessSource`, `Options`, `Result`, `SourceManager`.
  - Location compatibility keeps `entity.SourcePos.Line` and `Column` for legacy callers while adding `LocationID`.
  - Tests use helper functions defined in their own files or earlier tasks.
