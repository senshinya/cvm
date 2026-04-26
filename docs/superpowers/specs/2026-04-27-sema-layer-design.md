# Sema Layer Design

Date: 2026-04-27
Status: Approved

## Context

The cvm pipeline currently runs Lexer → Parser. Parser returns `[]*entity.AstNode` (1+ candidate parse trees) with cheap fork pruning, no semantic analysis. The downstream goal is a stack-based, JVM-like bytecode VM. The intermediate stage between Parser and Codegen is the Sema layer specified here.

This design supersedes the earlier "separated Humanize + Sema" plan (Plan 1). The unified Sema layer (Plan 2, clang-style) absorbs structural rewrite, symbol tables, type checking, implicit-conversion insertion, constant evaluation, and forest disambiguation into one pass.

Scope: full C99 (matching the parser's grammar coverage). No staged subsetting.

## Architecture

```
Lexer → Parser → PreFilter → Sema (concurrent per candidate) → Codegen → VM
                  []*AstNode  pick winner: *Program
```

**Packages:**
- `sema/` (new): `prefilter.go`, `sema.go`, `type.go`, `scope.go`, `ir.go` (typed IR node definitions), `expr.go`, `stmt.go`, `decl.go`, `const.go`, `error.go`.
- `parser/timbersaw.go` is deleted; logic moves to `sema/prefilter.go`.
- `parser.Parse()` no longer calls `chopForest`; it returns the raw candidate forest. Caller (compiler) calls `sema.Analyze(candidates)`.

IR nodes live in `sema/` (not `entity/`) because they reference `Type`, `Symbol`, `Field`, etc. which are sema-owned. `entity/` continues to host parser-stage tokens and parse-tree nodes only. Codegen imports `sema` to consume the IR.

**Data flow:**
1. Parser returns `[]*entity.AstNode` (≥1 by B-stage guarantee).
2. PreFilter performs lightweight typedef-disambiguation scan; returns survivors and any errors from culled candidates (errors enter best-error pool).
3. Each survivor is analyzed in its own goroutine with its own `Sema` instance (independent `TypeTable`, `SymbolTable`).
4. Aggregator picks the winner: exactly one error-free result → return its `*Program`. Multiple error-free → genuine ambiguity error (should not occur for valid C). Zero error-free → return errors from the result with the fewest / deepest errors.

## IR

### Type hierarchy (`sema/type.go`)

```go
type Type interface { isType(); String() string }

type BuiltinKind int
const (
    Void BuiltinKind = iota
    Bool
    Char; SChar; UChar
    Short; UShort
    Int; UInt
    Long; ULong
    LongLong; ULongLong
    Float; Double; LongDouble
    FloatComplex; DoubleComplex; LongDoubleComplex
)

type BuiltinType struct { Kind BuiltinKind }
type PointerType struct { Pointee Type }

type ArraySizeKind int
const (
    ArrayUnsized ArraySizeKind = iota   // int a[]
    ArrayConstantSize                    // int a[5]
    ArrayVLA                             // int a[n] (n runtime)
    ArrayStarSize                        // int f(int a[*])
)
type ArrayType struct {
    Elem     Type
    Size     int64       // ArrayConstantSize
    SizeExpr Expr        // ArrayVLA
    SizeKind ArraySizeKind
}

type FunctionType struct {
    Ret      Type
    Params   []Type
    Variadic bool
    HasProto bool         // false = K&R style declaration `int f();`
}

type StructType struct { Tag *TagID; Fields []*Field; Complete bool }
type UnionType  struct { Tag *TagID; Fields []*Field; Complete bool }
type EnumType   struct { Tag *TagID; Underlying Type; Enumerators []*Enumerator }

type Field struct {
    Name       string
    T          Type
    BitWidth   int
    IsBitField bool
    Offset     int64
}

type QualType struct { Base Type; Const, Volatile, Restrict bool }
type ErrorType struct{}    // singleton
```

**Interning rules:**
- `BuiltinType`: pre-built singletons.
- `PointerType` / `FunctionType` / complete `StructType` / `UnionType` / `EnumType`: interned by structure + tag identity.
- `ArrayType`: `ArrayConstantSize` interned by `(Elem, Size)`; `ArrayVLA` **not** interned (each declaration is a unique pointer; C99 variably modified type semantics); `ArrayUnsized` and `ArrayStarSize` interned by element type.
- `QualType`: interned by `(Base, qualifier bitset)`.

`TagID` is the AST node pointer of the tag's defining (or first-forward) declaration. Tags are nominal: `struct A {int x;}` and `struct B {int x;}` are distinct types.

Each Sema instance owns its `TypeTable`. No global state, no cross-candidate sharing.

### Node hierarchy (`sema/ir.go`)

```go
type Node interface { Pos() entity.SourceRange }

type Program struct {
    Globals []Decl
    Funcs   []*FuncDef
    Types   *TypeTable
    SymTab  *SymbolTable
}

// Statements
type Stmt interface { Node; isStmt() }
type Block        struct { Items []Stmt; Scope *Scope }
type IfStmt       struct { Cond Expr; Then, Else Stmt }
type WhileStmt    struct { Cond Expr; Body Stmt; DoWhile bool }
type ForStmt      struct { Init Stmt; Cond, Post Expr; Body Stmt; Scope *Scope }
type SwitchStmt   struct { Cond Expr; Body Stmt; Cases []*CaseStmt; Default *DefaultStmt }
type ReturnStmt   struct { Value Expr }
type BreakStmt    struct{}
type ContinueStmt struct{}
type GotoStmt     struct { Target *LabeledStmt }
type LabeledStmt  struct { Name string; Body Stmt }
type CaseStmt     struct { Value int64; Body Stmt }
type DefaultStmt  struct { Body Stmt }
type ExprStmt     struct { Expr Expr }
type DeclStmt     struct { Decls []Decl }
type EmptyStmt    struct{}

// Expressions
type ValueCategory int
const ( RValue ValueCategory = iota; LValue )

type Expr interface {
    Node
    isExpr()
    GetType() Type
    GetCategory() ValueCategory
}

type IntLit         struct { Value int64; T Type }
type FloatLit       struct { Value float64; T Type }
type StringLit      struct { Value string; T Type }
type CharLit        struct { Value int32; T Type }
type VarRef         struct { Sym *Symbol; T Type }
type EnumRef        struct { Enumerator *Enumerator; T Type }
type BinOp          struct { Op BinaryOp; L, R Expr; T Type }
type UnOp           struct { Op UnaryOp; X Expr; T Type }
type AssignExpr     struct { L, R Expr; T Type }
type CompoundAssign struct { Op BinaryOp; L, R Expr; T Type }
type CallExpr       struct { Callee Expr; Args []Expr; T Type }
type MemberExpr     struct { Base Expr; Field *Field; Arrow bool; T Type }
type IndexExpr      struct { Base, Index Expr; T Type }
type CondExpr       struct { Cond, Then, Else Expr; T Type }
type SizeofExpr     struct { Operand SizeofOperand; T Type }
type SizeofOperand  struct { Expr Expr; Type Type }
type CommaExpr      struct { L, R Expr; T Type }
type CompoundLit    struct { T Type; Init *InitList }
type InitList       struct { Elems []InitElem; T Type }
type InitElem       struct { Designators []Designator; Value Expr }
type Designator     struct { Kind DesignatorKind; Index int64; Field *Field }
type ImplicitCast   struct { From, To Type; X Expr; Kind CastKind }
type ExplicitCast   struct { To Type; X Expr }
type AddrConst      struct { Sym *Symbol; Offset int64; T Type }

// Declarations
type Decl interface { Node; isDecl() }
type VarDecl     struct { Sym *Symbol; T Type; Init Expr; Storage StorageClass; IsParam bool }
type FuncDecl    struct { Sym *Symbol; T *FunctionType; Storage StorageClass }
type FuncDef     struct {
    Sym    *Symbol
    T      *FunctionType
    Params []*VarDecl
    Body   *Block
    Locals []*VarDecl
    Labels map[string]*LabeledStmt
}
type TypedefDecl struct { Sym *Symbol; T Type }
type TagDecl     struct { Tag *TagID; T Type }
```

**Notes:**
- Bit-fields: `Field.IsBitField + BitWidth` carries the info. The "not addressable" rule is enforced at `&` checking by inspecting `Field.IsBitField`; no separate `BitFieldRefExpr` node.
- VLA: `ArrayType.SizeExpr` runs at runtime. `sizeof(VLA)` is a runtime expression (Sema marks `T = size_t` but codegen emits evaluation, not const push). VLA only allowed in block scope and function prototypes.
- K&R function definitions: Sema converts to prototyped form in IR. The corresponding declaration `int f();` retains `HasProto=false`. Calls to `HasProto=false` functions apply default argument promotion at the call site.
- Designated initializers: `InitElem.Designators` is the chain `[3].field[0] = ...`; Sema constant-evaluates indices and resolves field names.
- Compound literals: `CompoundLit { T, Init }` is an anonymous lvalue; lifetime per C99 (block scope = enclosing block, file scope = entire program). Codegen allocates anonymous slots.
- Switch: `SwitchStmt.Cases` collects all `CaseStmt` from nested blocks at Sema time; `Default` is a direct pointer.
- Goto: resolved in a second pass per function (`FuncDef.Labels` is built first, `GotoStmt.Target` filled from it).
- Enum constants are `EnumRef` (separate from `VarRef`) so codegen knows they fold to immediate values, no slot.

## Symbol Table & Scope

### Namespaces (4 per C99)

```go
type SymbolNamespace int
const (
    NSOrdinary SymbolNamespace = iota   // var / func / typedef / enum constant
    NSTag                                // struct / union / enum tag
    NSLabel                              // goto label (function scope)
    NSMember                             // struct/union member (per-struct)
)
```

`NSLabel` lives on `FuncDef.Labels`. `NSMember` lives on `StructType.Fields`. The scope chain manages only `NSOrdinary` and `NSTag`.

### Scope

```go
type ScopeKind int
const (
    ScopeFile ScopeKind = iota
    ScopeBlock
    ScopeFuncProto
    ScopeFunc
)

type Scope struct {
    Kind     ScopeKind
    Parent   *Scope
    Ordinary map[string]*Symbol
    Tags     map[string]*TagInfo
}
```

### Symbol

```go
type Symbol struct {
    Name     string
    Kind     SymbolKind   // SymVar | SymFunc | SymTypedef | SymEnumerator | SymParam
    T        Type
    Storage  StorageClass // None | Auto | Register | Static | Extern | Typedef
    Linkage  Linkage      // None | Internal | External
    Decl     Decl
    Defs     []Decl       // all declarations of this symbol; new ones appended after compatibility check
    Pos      entity.SourcePos
    SlotID   int          // local-frame slot
    GlobalID int          // global / function index
}

type TagInfo struct {
    Tag      *TagID
    T        Type
    Decl     Decl
    Complete bool
}
```

### Function prototype scope

`int f(int a, int b[a]);` creates a `ScopeFuncProto` so `a` is visible to the rest of the prototype. If the prototype is part of a function definition, parameter symbols migrate from `ScopeFuncProto` into the merged `ScopeFunc` so the body sees them.

### Tag forward declarations

`struct Node;` creates an incomplete `StructType`, `Complete=false`. Later `struct Node { ... };` finds the existing `TagInfo`, fills `Fields` in place, and sets `Complete=true`. The Type pointer must NOT change (consumers may already hold `*StructType` via `struct Node *p;`). `enum` does not support forward declarations in C99; Sema rejects them.

### Lookup

```go
func (s *Scope) Lookup(name string, ns SymbolNamespace) *Symbol         // walks Parent chain
func (s *Scope) LookupCurrent(name string, ns SymbolNamespace) *Symbol  // current scope only
func (s *Scope) Insert(name string, sym *Symbol) error                  // same-NS redefinition check
```

Same-scope same-namespace redefinition: error unless the new declaration is a compatible redeclaration (compatible types + valid linkage rules). Different namespaces never conflict.

### File-scope vs block-scope declaration semantics

- File scope: `int x;` is a tentative definition. Multiple tentatives merge. If no non-tentative is seen, the variable is treated as zero-initialized.
- Block scope: `int x;` is a normal definition. `extern int x;` is a declaration that links to an external symbol.

### Typedef shadowing

`typedef int T; { int T; T x; }` — inner `int T;` shadows the typedef. Sema's lookup order ensures subsequent `T x;` resolves `T` as the inner variable, making it an invalid expression (Sema reports the error). This works because PreFilter does NOT cull the variable-interpretation candidate when the name is in an outer typedef stack.

## Sema Flow

### Two phases per candidate

**Phase 1 (file scope, single-threaded):** walk top-level declarations in source order:

1. Tag declaration: register or merge `TagInfo`; fill `Fields`/`Enumerators` and set `Complete` on definition.
2. Typedef declaration: compute Type from declaration_specifiers + declarator; insert `SymTypedef`.
3. Function declaration (no body): compute `FunctionType`; insert `SymFunc`. Re-declarations merge after compatibility check.
4. Variable declaration (with init): compute Type, constant-evaluate init, insert `SymVar`, emit IR `VarDecl`.
5. Function definition: compute `FunctionType`; insert/merge `SymFunc`; **enqueue body for phase 2.** Outer walk does NOT enter the body.

**Phase 2 (function body, parallelizable across functions):**

1. Build `ScopeFunc` with parameters.
2. Walk body once:
   - Block / ForStmt / SwitchStmt: push new scope.
   - DeclStmt: register symbols in current scope, type-check inits.
   - LabeledStmt: register in `funcCtx.Labels`.
   - GotoStmt: record for back-patch; `Target` left nil for now.
   - Expressions: bottom-up typing + ImplicitCast insertion.
   - return / case / sizeof: corresponding semantic checks.
3. Second pass: walk to back-patch `GotoStmt.Target` from `funcCtx.Labels`. Unresolved labels → error.
4. SwitchStmt: collect all `CaseStmt`/`DefaultStmt` (even from nested blocks) into `Cases` / `Default`; verify case values are constants and unique.

Phase 1 is single-threaded because top-level symbol registration is order-dependent. Phase 2 is parallelizable across functions because all top-level symbols are already registered.

### Expression typing

Each typing function:

```go
func (s *Sema) typeExpr(node *entity.AstNode, scope *Scope) Expr
```

Returns IR `Expr` with type and value category set. The detailed conversion rules per construct are in §6 below.

### Error continuation: ErrorType sentinel + cascade suppression

```go
var ErrorTypeSingleton = &ErrorType{}
func IsError(t Type) bool { return t == ErrorTypeSingleton }
```

Each typing rule's prologue:

```go
func (s *Sema) typeBinOp(...) Expr {
    l := s.typeExpr(left, scope)
    r := s.typeExpr(right, scope)
    if IsError(l.GetType()) || IsError(r.GetType()) {
        return &BinOp{ L: l, R: r, T: ErrorTypeSingleton }
    }
    // typing logic; on error: s.report(...) then return ErrorType node
}
```

The first error at a given position is reported; cascaded uses of `ErrorType` produce no further errors.

### SemaResult

```go
type SemaResult struct {
    Program *Program
    Errors  []*common.CvmError
    Source  *entity.AstNode
}
```

### Forest aggregation

```go
func Analyze(candidates []*entity.AstNode) (*Program, error) {
    survivors, prefilterErrs := preFilter(candidates)
    if len(survivors) == 0 {
        return nil, mergeErrors(prefilterErrs)
    }

    results := make([]*SemaResult, len(survivors))
    var wg sync.WaitGroup
    for i, tree := range survivors {
        wg.Add(1)
        go func(i int, tree *entity.AstNode) {
            defer wg.Done()
            results[i] = NewSema().Analyze(tree)
        }(i, tree)
    }
    wg.Wait()

    var clean []*SemaResult
    for _, r := range results {
        if len(r.Errors) == 0 {
            clean = append(clean, r)
        }
    }
    switch len(clean) {
    case 1:
        return clean[0].Program, nil
    case 0:
        best := pickBestErrorResult(results)
        return nil, mergeErrors(best.Errors)
    default:
        return nil, AmbiguousParse(clean)
    }
}
```

`pickBestErrorResult`: sort by `len(Errors)` ascending, tiebreak by first-error source position descending (deeper position = closer to the actual issue).

## Constant Evaluation

### Use sites

| Position | Required form | Failure |
|---|---|---|
| Array size `int a[N]` | integer constant expression (or VLA) | non-constant → mark ArrayVLA; non-integer / negative / 0 → error |
| Case label `case N:` | integer constant expression | error if non-constant |
| Enum value | integer constant expression | error |
| Bit-field width `int x : N;` | integer constant expression | error; also 0 ≤ N ≤ type_bit_width |
| File-scope / static initializer | constant expression (incl. address constant) | error |
| Designator index `[3] = ...` | integer constant expression | error |
| File-scope compound literal | constant expression | error |

### Evaluator interface (`sema/const.go`)

```go
type ConstKind int
const (
    ConstInt ConstKind = iota
    ConstUint
    ConstFloat
    ConstAddress
    ConstString
)

type ConstValue struct {
    Kind  ConstKind
    Int   int64
    Uint  uint64
    Float float64
    Addr  AddressConst
    T     Type
}

type AddressConst struct {
    Sym    *Symbol
    Offset int64
}

type Evaluator struct { sema *Sema }

func (e *Evaluator) EvalIntegerConstant(expr Expr) (ConstValue, bool)
func (e *Evaluator) EvalConstant(expr Expr)        (ConstValue, bool)
func (e *Evaluator) EvalArithmetic(expr Expr)      (ConstValue, bool)
```

Returns `bool` for "is a valid constant expression of this kind." Caller decides whether failure is an error (case label) or graceful fallback (VLA).

### Rules: strict C99 §6.6

Integer constant expression supports: integer literals, char literals, `sizeof(T)`/`sizeof(expr)` (excluding VLA), enum constants, arithmetic / bitwise / shift / comparison / logical / `?:`, casts from integer/floating to integer.

Disallowed: variable references (except enum constants), function calls, `++`/`--`, assignment, comma, `&`, `*`, member access.

Constant expression = integer constant expression ∪ floating constant expression ∪ address constant expression. Address constants: `&` of static-storage lvalue, string literals, compound literals, plus integer constant offsets, array name (decay), function name (decay), casts (int↔ptr, ptr↔ptr).

### Sema interaction

- Integer-constant contexts: Sema calls `EvalIntegerConstant`, writes the result into the IR field (`ArrayType.Size`, `CaseStmt.Value`, `Designator.Index`, `Enumerator.Value`, `Field.BitWidth`), discards the original Expr.
- Static initializer: Sema computes typed Expr with ImplicitCasts, then calls `EvalConstant`. On success, **replaces** `VarDecl.Init` with a folded literal node (`IntLit` / `FloatLit` / `AddrConst`); the original tree is discarded. Codegen never re-evaluates static initializers.
- Block-scope auto-variable initializer: NOT folded. IR keeps the full Expr; codegen translates it to ordinary store bytecode.

### Long double / Complex

`long double` is implemented as `double` (the C standard explicitly allows this). `_Complex` is represented as a pair of `float64`. `float`/`double` use `float64` internally during evaluation.

## Implicit Conversion Rules (CastKind catalog)

```go
type CastKind int
const (
    LValueToRValue CastKind = iota
    ArrayDecay
    FunctionDecay
    IntegerPromotion
    UsualArithmetic
    IntegralConversion
    FloatingConversion
    IntToFloat
    FloatToInt
    IntToPointer
    PointerToInt
    PointerConversion
    VoidPointerConversion
    NullPointerConstant
    BoolConversion
    DefaultArgPromotion
)
```

Note: assignment-context conversions (assignment, return value, function-call argument, scalar initializer) are decomposed during Sema into the concrete casts they entail (`LValueToRValue`, `IntegralConversion`, `FloatingConversion`, `IntToFloat`, `FloatToInt`, `PointerConversion`, `NullPointerConstant`, `VoidPointerConversion`, `BoolConversion`, etc.). There is no separate `AssignmentConversion` `CastKind` in finalized IR; "assignment-context cast chain" below is shorthand for "apply that decomposition for the target type."

### Trigger sites (per construct)

**Arithmetic binop** `+ - * / % & | ^ << >>`:
1. Both sides → `LValueToRValue`.
2. Each side → `IntegerPromotion` if char/short/_Bool/bit-field.
3. `UsualArithmetic` to balance to common type. (Shift operands are integer-promoted but NOT UAC-balanced; result type = promoted left-operand type.)
4. `BinOp.T` = common type.

**Comparison / relational** `< <= > >= == !=`:
1. `LValueToRValue`.
2. Arithmetic + arithmetic: `IntegerPromotion` + `UsualArithmetic`.
3. Pointer + pointer: must be compatible; `void *` partner → `VoidPointerConversion`; null pointer constant → `NullPointerConstant` cast to other side's type.
4. Pointer + 0: `NullPointerConstant`.
5. Result type: `int`.

**Logical** `&& || !`:
1. Operands → `LValueToRValue`.
2. Each → `BoolConversion`.
3. Result type: `int`.

**Conditional** `?:`:
1. Cond → `LValueToRValue` + `BoolConversion`.
2. Then/else: `UsualArithmetic` for arithmetic; pointer combinations via `PointerConversion` / `VoidPointerConversion` / `NullPointerConstant`.
3. Result type: common type.

**Assignment** `=`:
- Right-hand → `LValueToRValue`, then assignment-context cast chain to target type.

**Compound assignment** `+= -= *= ...`:
- Equivalent to `lhs = (lhs) op (rhs)` but lhs evaluated only once. IR retains `CompoundAssign`; cast chain encodes the equivalent typing internally.

**Unary `+ - ~`**: `LValueToRValue` + `IntegerPromotion`. (`~` integer-only; `+ -` accept floating too.)

**Unary `*`**: `LValueToRValue` + `ArrayDecay` (if operand is array). Result is lvalue of pointee type.

**Unary `&`**: operand must be lvalue (and non-register, non-bit-field). Does NOT apply `LValueToRValue` / `ArrayDecay` / `FunctionDecay`. Result type `*T` (rvalue).

**`++` `--`**: operand must be modifiable lvalue. IR `UnOp(IncPre/IncPost/DecPre/DecPost)` with implied cast chain.

**`sizeof`**: operand does NOT undergo `LValueToRValue` / `ArrayDecay` / `FunctionDecay`. `sizeof(VLA)` is a runtime expression (Sema marks `T = size_t`, codegen emits eval bytecode); other forms are constant.

**Cast `(T)expr`**: `LValueToRValue` then wrap `ExplicitCast` to T. Sema verifies T is reachable from source type (any arithmetic↔arithmetic, integer↔pointer, pointer↔pointer including void*); struct/union/array cannot be directly cast.

**Index `a[i]`**: equivalent to `*(a + i)` semantically; IR keeps `IndexExpr` with `ArrayDecay` on `a`.

**Member `a.f` / `a->f`**: `.` requires struct/union (lvalue or rvalue). `->` requires pointer to struct/union; `LValueToRValue` first, then use pointee. `.` member inherits operand's category; `->` member is always lvalue. Bit-field members are lvalue but not `&`-addressable.

**Call `f(args...)`**:
1. Callee → `LValueToRValue` + `FunctionDecay` → must be function pointer.
2. `HasProto=true`: each arg → `LValueToRValue` + the assignment-context cast chain to corresponding param type; variadic tail → `DefaultArgPromotion`.
3. `HasProto=false`: each arg → `LValueToRValue` + `DefaultArgPromotion`. If callee is defined in the same TU, Sema additionally checks against the definition's prototype.

**Return `return e;`**: e → `LValueToRValue` + the assignment-context cast chain to function return type. Bare `return;` only allowed in void functions.

**If/While/For/DoWhile cond**: cond → `LValueToRValue` + `BoolConversion`; type must be scalar. Empty `for(;;)` cond is treated as `1`.

**Switch cond**: cond → `LValueToRValue` + `IntegerPromotion`; type must be integer. Case values are compared in the cond's promoted type.

**Initializer**:
- Scalar: right side → `LValueToRValue` + the assignment-context cast chain to target type.
- Array: each `InitElem.Value` → init rule for element type. String-literal initializing `char[]` is a special path (no decay; copies chars).
- Struct/union: per field, in order or by designator; each → init rule for field type.
- Nested initializer: same recursion.

**Designator index**: integer constant; evaluated and stored in `Designator.Index`. No cast chain needed.

### Qualifier handling

- `LValueToRValue` discards all qualifiers from the result type.
- Pointer assignment / comparison: destination's pointee qualifiers must be a superset of source's pointee qualifiers (qualifiers can be added implicitly, not removed). Violation → `PointerConversion` rejected → "discards qualifier" error.
- `const int *p = ...; int *q = p;` requires explicit cast.

### Null pointer constant identification

An integer constant expression with value 0, or such an expression cast to `void *`, is a null pointer constant. Sema inserts `NullPointerConstant` cast in pointer-target contexts (assignment / comparison / call). Other integer 0 values in pointer contexts require explicit cast.

### Bit-field special rules

- Integer promotion for bit-field: if width + signedness fits in `int` → promote to `int`; else `unsigned int`.
- Bit-field is lvalue but NOT addressable: Sema verifies `Field.IsBitField == false` at `&` checking.
- Bit-field assignment: right side casts to the bit-field's declared type (not promoted type).

## PreFilter

`parser/timbersaw.go` is deleted; logic moves to `sema/prefilter.go`. `parser.Parse()` no longer calls `chopForest`. New entry:

```go
func preFilter(candidates []*entity.AstNode) (survivors []*entity.AstNode, errs []*common.CvmError)
```

Algorithm: walk each candidate AST with a lightweight typedef-vs-variable scope tracker (mirroring TimberSaw's structure). For each `PrimaryExpression { IDENTIFIER }`, the identifier must NOT be a typedef. For each `TypedefName { IDENTIFIER }`, the identifier MUST be a typedef. Mismatches drop the candidate.

**Scope changes vs current TimberSaw:**
- Keep: typedef-vs-variable disambiguation (the parser-residual ambiguity).
- Drop: the "undeclared identifier" check. Move that to Sema, where richer context (variable vs function vs typedef, link semantics) gives a more precise error.

PreFilter errors enter the best-error pool when no survivors remain.

## Testing

### Layout

```
sema/
  type_test.go       TypeTable interning
  scope_test.go      scope chain, namespace isolation, redeclaration merging
  const_test.go      constant evaluation (one case per C99 §6.6 rule)
  expr_test.go       per typeXxx function
  stmt_test.go       if/while/for/switch/return/goto typing
  decl_test.go       tag forward, typedef, function-decl merging
  prefilter_test.go
  sema_test.go       end-to-end source → IR
  testdata/
    pass/   *.c + *.ir.golden     (snapshot of pretty-printed IR)
    fail/   *.c + *.err.golden    (line:col + message substring)
```

### Coverage targets

- TypeTable: pointer equality for interned, inequality for distinct (incl. tag identity, qualifier sets, VLA non-interning).
- Scope: namespace separation, shadow chains, redeclaration compatibility.
- Constant evaluation: every allowed and disallowed form per §6.6.
- Expression typing: each `ImplicitCast.Kind` has at least one positive test asserting the cast appears in IR at the expected node.
- Stmt/decl: goto resolution, switch case collection across nested blocks, VLA scope rules, bit-field constraints, struct forward + completion preserves Type pointer, K&R definitions normalized to prototype.
- PreFilter: typedef-shadow disambiguation; undeclared identifier NOT pre-filtered (downstream to Sema).
- End-to-end: golden snapshots for pass cases; line/column + message substring for fail cases.
- Concurrency: `TestConcurrentSemaIsolation` — many candidates, 100 iterations, `-race` clean, distinct TypeTable identity per result.
- Migrate `TestSimpleFib`, `TestSqrt` and other end-to-end tests from `parser/` to `sema/`.

### Out of scope

- Multi-translation-unit linking
- Preprocessor
- C11 features (`_Generic`, `_Atomic`, `_Alignas`, `_Thread_local`, `static_assert`)

## Open Items

None. All design decisions resolved during brainstorming.
