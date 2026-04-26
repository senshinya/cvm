# Sema Plan A — Foundations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `sema` package skeleton with the C99 type system (interned), symbol-table + scope plumbing, and migrate `parser/timbersaw.go` to `sema/prefilter.go`. The existing parser pipeline keeps working with no behavior change.

**Architecture:** New `sema/` package owns Type, Scope, Symbol, and PreFilter. `parser.Parse()` stops calling `chopForest`; the `compiler.go` glue temporarily calls `sema.PreFilter` after parse so existing tests still execute the prefilter. No Sema walker yet — that's Plan B.

**Tech Stack:** Go 1.x stdlib + existing `entity`, `parser`, `common` packages. Tests via `go test`.

**Spec reference:** `docs/superpowers/specs/2026-04-27-sema-layer-design.md` § Architecture, § IR (Type hierarchy), § Symbol Table & Scope, § PreFilter.

---

## File Structure

**Create:**
- `sema/doc.go` — package documentation
- `sema/error.go` — Sema-specific error types
- `sema/type.go` — Type interface + concrete type structs + ErrorType singleton
- `sema/type_table.go` — TypeTable struct + interning getters
- `sema/scope.go` — Symbol, TagInfo, Scope, namespace enum, lookup/insert
- `sema/prefilter.go` — TimberSaw logic migrated from parser
- `sema/type_test.go`, `sema/scope_test.go`, `sema/prefilter_test.go`

**Modify:**
- `parser/parser.go` — remove `chopForest` call from `Parse()`; remove `recordError` callsite for chop errors
- `parser/timbersaw.go` — DELETED after migration
- `compiler.go` — call `sema.PreFilter` on `parser.Parse()` output (temporary glue until Plan B/C)

**Notes:**
- `sema` imports `parser` to access auto-generated `TokenType` constants (`PrimaryExpression`, `TypedefName`, `DirectDeclarator`, etc.). This is a known awkwardness; cleanup is out of Plan A scope.
- Each test file is unit-level; no testdata directory in Plan A.

---

## Task 1: Bootstrap sema package

**Files:**
- Create: `sema/doc.go`
- Create: `sema/error.go`

- [ ] **Step 1: Write package doc**

`sema/doc.go`:
```go
// Package sema implements C99 semantic analysis for the cvm compiler.
//
// Sema consumes the candidate parse forest from package parser, runs
// PreFilter to drop obviously-wrong typedef interpretations, then runs
// per-candidate Sema walkers (Plan B/C) that produce a typed Program IR
// for codegen.
//
// See docs/superpowers/specs/2026-04-27-sema-layer-design.md.
package sema
```

- [ ] **Step 2: Add error stubs**

`sema/error.go`:
```go
package sema

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func UndeclaredIdentifier(pos entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("use of undeclared identifier '%s'", name)),
	)
}

func RedefinitionSymbol(pos, previous entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("redefinition of '%s' as different kind of symbol", name)),
		common.NewNoteMessage(previous, "previous definition is here"),
	)
}

func InvalidSymbolKind(pos, previous entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("invalid symbol kind '%s'", name)),
		common.NewNoteMessage(previous, "symbol definition is here"),
	)
}
```

- [ ] **Step 3: Build the package**

Run: `go build ./sema/...`
Expected: success.

- [ ] **Step 4: Commit**

```bash
git add sema/doc.go sema/error.go
git commit -m "feat: bootstrap sema package with error stubs"
```

---

## Task 2: Type interface + BuiltinType

**Files:**
- Create: `sema/type.go`
- Create: `sema/type_test.go`

- [ ] **Step 1: Write failing test**

`sema/type_test.go`:
```go
package sema

import "testing"

func TestBuiltinTypeKindString(t *testing.T) {
	bt := &BuiltinType{Kind: Int}
	if got := bt.String(); got != "int" {
		t.Fatalf("BuiltinType{Int}.String() = %q, want %q", got, "int")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestBuiltinTypeKindString`
Expected: FAIL (BuiltinType undefined).

- [ ] **Step 3: Implement Type interface + BuiltinType**

`sema/type.go`:
```go
package sema

type Type interface {
	isType()
	String() string
}

type BuiltinKind int

const (
	Void BuiltinKind = iota
	Bool
	Char
	SChar
	UChar
	Short
	UShort
	Int
	UInt
	Long
	ULong
	LongLong
	ULongLong
	Float
	Double
	LongDouble
	FloatComplex
	DoubleComplex
	LongDoubleComplex
)

var builtinNames = [...]string{
	Void:              "void",
	Bool:              "_Bool",
	Char:              "char",
	SChar:             "signed char",
	UChar:             "unsigned char",
	Short:             "short",
	UShort:            "unsigned short",
	Int:               "int",
	UInt:              "unsigned int",
	Long:              "long",
	ULong:             "unsigned long",
	LongLong:          "long long",
	ULongLong:         "unsigned long long",
	Float:             "float",
	Double:            "double",
	LongDouble:        "long double",
	FloatComplex:      "float _Complex",
	DoubleComplex:     "double _Complex",
	LongDoubleComplex: "long double _Complex",
}

type BuiltinType struct {
	Kind BuiltinKind
}

func (*BuiltinType) isType() {}
func (b *BuiltinType) String() string {
	return builtinNames[b.Kind]
}
```

- [ ] **Step 4: Run test to verify pass**

Run: `go test ./sema/ -run TestBuiltinTypeKindString`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_test.go
git commit -m "feat(sema): Type interface and BuiltinType"
```

---

## Task 3: TypeTable with builtin singletons

**Files:**
- Create: `sema/type_table.go`
- Modify: `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestTypeTableBuiltinSingleton(t *testing.T) {
	tt := NewTypeTable()
	a := tt.Builtin(Int)
	b := tt.Builtin(Int)
	if a != b {
		t.Fatalf("Builtin(Int) returned distinct pointers; expected interning")
	}
	c := tt.Builtin(UInt)
	if a == c {
		t.Fatalf("Int and UInt returned same pointer")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestTypeTableBuiltinSingleton`
Expected: FAIL (NewTypeTable undefined).

- [ ] **Step 3: Implement TypeTable**

`sema/type_table.go`:
```go
package sema

type TypeTable struct {
	builtins [len(builtinNames)]*BuiltinType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{}
	for k := Void; int(k) < len(builtinNames); k++ {
		tt.builtins[k] = &BuiltinType{Kind: k}
	}
	return tt
}

func (tt *TypeTable) Builtin(k BuiltinKind) *BuiltinType {
	return tt.builtins[k]
}
```

- [ ] **Step 4: Run test to verify pass**

Run: `go test ./sema/ -run TestTypeTableBuiltinSingleton`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type_table.go sema/type_test.go
git commit -m "feat(sema): TypeTable with interned builtin singletons"
```

---

## Task 4: PointerType with interning

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestPointerTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	p1 := tt.Pointer(intT)
	p2 := tt.Pointer(intT)
	if p1 != p2 {
		t.Fatalf("Pointer(int) interning failed: %p vs %p", p1, p2)
	}
	pp := tt.Pointer(p1)
	if pp == p1 {
		t.Fatalf("Pointer(int*) collided with Pointer(int)")
	}
	if got := pp.String(); got != "int**" {
		t.Fatalf("Pointer(int*).String() = %q, want %q", got, "int**")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestPointerTypeInterning`
Expected: FAIL.

- [ ] **Step 3: Implement PointerType + interning**

Append to `sema/type.go`:
```go
type PointerType struct {
	Pointee Type
}

func (*PointerType) isType() {}
func (p *PointerType) String() string {
	return p.Pointee.String() + "*"
}
```

Append to `sema/type_table.go`:
```go
type pointerKey struct{ pointee Type }

type TypeTable struct {
	builtins [len(builtinNames)]*BuiltinType
	pointers map[pointerKey]*PointerType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		pointers: map[pointerKey]*PointerType{},
	}
	for k := Void; int(k) < len(builtinNames); k++ {
		tt.builtins[k] = &BuiltinType{Kind: k}
	}
	return tt
}

func (tt *TypeTable) Pointer(pointee Type) *PointerType {
	key := pointerKey{pointee}
	if p, ok := tt.pointers[key]; ok {
		return p
	}
	p := &PointerType{Pointee: pointee}
	tt.pointers[key] = p
	return p
}
```

(Replace the existing `TypeTable` struct + `NewTypeTable` body — both versions cannot coexist.)

- [ ] **Step 4: Run all sema tests**

Run: `go test ./sema/`
Expected: PASS (all three tests).

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_test.go
git commit -m "feat(sema): PointerType with TypeTable interning"
```

---

## Task 5: ArrayType with conditional interning (constant-size + non-VLA)

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing tests**

Append to `sema/type_test.go`:
```go
func TestArrayTypeConstantInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayConstant(intT, 5)
	a2 := tt.ArrayConstant(intT, 5)
	if a1 != a2 {
		t.Fatalf("ArrayConstant(int, 5) interning failed")
	}
	a3 := tt.ArrayConstant(intT, 6)
	if a1 == a3 {
		t.Fatalf("Different sizes collided")
	}
}

func TestArrayTypeUnsizedInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayUnsized(intT)
	a2 := tt.ArrayUnsized(intT)
	if a1 != a2 {
		t.Fatalf("ArrayUnsized(int) interning failed")
	}
}

func TestArrayTypeVLANotInterned(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	a1 := tt.ArrayVLA(intT, nil) // SizeExpr Expr is nil here; only validating non-interning
	a2 := tt.ArrayVLA(intT, nil)
	if a1 == a2 {
		t.Fatalf("VLA arrays must NOT be interned")
	}
}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `go test ./sema/ -run TestArrayType`
Expected: FAIL (Array* methods undefined).

- [ ] **Step 3: Implement ArrayType + getters**

Append to `sema/type.go`:
```go
type ArraySizeKind int

const (
	ArrayUnsized ArraySizeKind = iota
	ArrayConstantSize
	ArrayVLA
	ArrayStarSize
)

// SizeExpr is intentionally `any` to avoid an Expr import cycle in Plan A;
// Plan B re-types this to sema.Expr once the IR exists.
type ArrayType struct {
	Elem     Type
	Size     int64
	SizeExpr any
	SizeKind ArraySizeKind
}

func (*ArrayType) isType() {}
func (a *ArrayType) String() string {
	switch a.SizeKind {
	case ArrayConstantSize:
		return a.Elem.String() + "[" + itoa(a.Size) + "]"
	case ArrayUnsized:
		return a.Elem.String() + "[]"
	case ArrayVLA:
		return a.Elem.String() + "[<vla>]"
	case ArrayStarSize:
		return a.Elem.String() + "[*]"
	}
	return a.Elem.String() + "[?]"
}

func itoa(n int64) string {
	// stdlib strconv import would otherwise pull a heavy dep into this file
	// for one call. Tiny inline impl is cleaner here.
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
```

Append to `sema/type_table.go`:
```go
type arrayConstantKey struct {
	elem Type
	size int64
}

func (tt *TypeTable) ArrayConstant(elem Type, size int64) *ArrayType {
	key := arrayConstantKey{elem, size}
	if a, ok := tt.arraysConstant[key]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, Size: size, SizeKind: ArrayConstantSize}
	tt.arraysConstant[key] = a
	return a
}

func (tt *TypeTable) ArrayUnsized(elem Type) *ArrayType {
	if a, ok := tt.arraysUnsized[elem]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, SizeKind: ArrayUnsized}
	tt.arraysUnsized[elem] = a
	return a
}

func (tt *TypeTable) ArrayStar(elem Type) *ArrayType {
	if a, ok := tt.arraysStar[elem]; ok {
		return a
	}
	a := &ArrayType{Elem: elem, SizeKind: ArrayStarSize}
	tt.arraysStar[elem] = a
	return a
}

func (tt *TypeTable) ArrayVLA(elem Type, sizeExpr any) *ArrayType {
	// VLA types are NOT interned per C99 variably-modified-type semantics.
	return &ArrayType{Elem: elem, SizeExpr: sizeExpr, SizeKind: ArrayVLA}
}
```

Update `TypeTable` struct + `NewTypeTable` to include the new maps:
```go
type TypeTable struct {
	builtins       [len(builtinNames)]*BuiltinType
	pointers       map[pointerKey]*PointerType
	arraysConstant map[arrayConstantKey]*ArrayType
	arraysUnsized  map[Type]*ArrayType
	arraysStar     map[Type]*ArrayType
}

func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		pointers:       map[pointerKey]*PointerType{},
		arraysConstant: map[arrayConstantKey]*ArrayType{},
		arraysUnsized:  map[Type]*ArrayType{},
		arraysStar:     map[Type]*ArrayType{},
	}
	for k := Void; int(k) < len(builtinNames); k++ {
		tt.builtins[k] = &BuiltinType{Kind: k}
	}
	return tt
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_test.go
git commit -m "feat(sema): ArrayType with conditional interning (constant + unsized + star; VLA not interned)"
```

---

## Task 6: FunctionType with interning

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestFunctionTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	doubleT := tt.Builtin(Double)
	f1 := tt.Function(intT, []Type{intT, doubleT}, false, true)
	f2 := tt.Function(intT, []Type{intT, doubleT}, false, true)
	if f1 != f2 {
		t.Fatalf("identical function types not interned")
	}
	f3 := tt.Function(intT, []Type{intT, doubleT}, true, true) // variadic
	if f1 == f3 {
		t.Fatalf("variadic flag did not differentiate")
	}
	f4 := tt.Function(intT, nil, false, false) // K&R `int f();`
	f5 := tt.Function(intT, nil, false, true)  // `int f(void);`
	if f4 == f5 {
		t.Fatalf("HasProto flag did not differentiate")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestFunctionTypeInterning`
Expected: FAIL.

- [ ] **Step 3: Implement FunctionType + interning**

Append to `sema/type.go`:
```go
type FunctionType struct {
	Ret      Type
	Params   []Type
	Variadic bool
	HasProto bool
}

func (*FunctionType) isType() {}
func (f *FunctionType) String() string {
	s := f.Ret.String() + " ("
	for i, p := range f.Params {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	if f.Variadic {
		if len(f.Params) > 0 {
			s += ", "
		}
		s += "..."
	}
	if !f.HasProto {
		s += "<noproto>"
	}
	s += ")"
	return s
}
```

Append to `sema/type_table.go`:
```go
type functionKey struct {
	ret      Type
	params   string // joined pointer addresses for hash; len(params) prefix
	variadic bool
	hasProto bool
}

func (tt *TypeTable) Function(ret Type, params []Type, variadic, hasProto bool) *FunctionType {
	key := functionKey{
		ret:      ret,
		params:   paramsKey(params),
		variadic: variadic,
		hasProto: hasProto,
	}
	if f, ok := tt.functions[key]; ok {
		return f
	}
	f := &FunctionType{
		Ret:      ret,
		Params:   append([]Type(nil), params...),
		Variadic: variadic,
		HasProto: hasProto,
	}
	tt.functions[key] = f
	return f
}

func paramsKey(params []Type) string {
	if len(params) == 0 {
		return ""
	}
	// Use Type pointer identity (Types are interned) for the key.
	var b []byte
	for _, p := range params {
		// fmt.Sprintf("%p", ...) would be cleaner but pulls in fmt; build it inline.
		ptr := uintptrOf(p)
		for i := 0; i < 8; i++ {
			b = append(b, byte(ptr>>(i*8)))
		}
		b = append(b, '|')
	}
	return string(b)
}

// uintptrOf erases the interface to a uintptr for keying. Defined in a
// separate file (type_uintptr.go) to localize the unsafe usage.
```

Add `sema/type_uintptr.go`:
```go
package sema

import "unsafe"

func uintptrOf(t Type) uintptr {
	type iface struct {
		_   *struct{}
		ptr unsafe.Pointer
	}
	i := *(*iface)(unsafe.Pointer(&t))
	return uintptr(i.ptr)
}
```

Update `TypeTable` to add `functions` map:
```go
type TypeTable struct {
	builtins       [len(builtinNames)]*BuiltinType
	pointers       map[pointerKey]*PointerType
	arraysConstant map[arrayConstantKey]*ArrayType
	arraysUnsized  map[Type]*ArrayType
	arraysStar     map[Type]*ArrayType
	functions      map[functionKey]*FunctionType
}
```

And init in `NewTypeTable`:
```go
functions: map[functionKey]*FunctionType{},
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_uintptr.go sema/type_test.go
git commit -m "feat(sema): FunctionType with interning by ret/params/variadic/hasProto"
```

---

## Task 7: QualType with interning

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestQualTypeInterning(t *testing.T) {
	tt := NewTypeTable()
	intT := tt.Builtin(Int)
	c1 := tt.Qualified(intT, true, false, false)
	c2 := tt.Qualified(intT, true, false, false)
	if c1 != c2 {
		t.Fatalf("const int interning failed")
	}
	cv := tt.Qualified(intT, true, true, false)
	if c1 == cv {
		t.Fatalf("different qualifier sets collided")
	}
	if got := c1.String(); got != "const int" {
		t.Fatalf("String() = %q, want %q", got, "const int")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestQualTypeInterning`
Expected: FAIL.

- [ ] **Step 3: Implement QualType + interning**

Append to `sema/type.go`:
```go
type QualType struct {
	Base                       Type
	Const, Volatile, Restrict bool
}

func (*QualType) isType() {}
func (q *QualType) String() string {
	s := ""
	if q.Const {
		s += "const "
	}
	if q.Volatile {
		s += "volatile "
	}
	if q.Restrict {
		s += "restrict "
	}
	return s + q.Base.String()
}
```

Append to `sema/type_table.go`:
```go
type qualKey struct {
	base Type
	bits uint8 // bit 0 = const, bit 1 = volatile, bit 2 = restrict
}

func (tt *TypeTable) Qualified(base Type, isConst, isVolatile, isRestrict bool) *QualType {
	var bits uint8
	if isConst {
		bits |= 1
	}
	if isVolatile {
		bits |= 2
	}
	if isRestrict {
		bits |= 4
	}
	key := qualKey{base, bits}
	if q, ok := tt.qualified[key]; ok {
		return q
	}
	q := &QualType{Base: base, Const: isConst, Volatile: isVolatile, Restrict: isRestrict}
	tt.qualified[key] = q
	return q
}
```

Add `qualified` map to `TypeTable` and init in `NewTypeTable`.

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_test.go
git commit -m "feat(sema): QualType with interning by qualifier bitset"
```

---

## Task 8: TagID + Field + StructType (incomplete + complete in-place)

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestStructTypeForwardCompletion(t *testing.T) {
	tt := NewTypeTable()
	tag := tt.NewTagID()
	st := tt.Struct(tag) // incomplete
	if st.Complete {
		t.Fatalf("forward struct should be incomplete")
	}
	// Reference it via pointer before completion.
	pst := tt.Pointer(st)
	// Complete in place.
	intT := tt.Builtin(Int)
	tt.CompleteStruct(st, []*Field{{Name: "x", T: intT}})
	if !st.Complete {
		t.Fatalf("struct still incomplete after CompleteStruct")
	}
	if pst.Pointee != st {
		t.Fatalf("pointer's pointee no longer points to completed struct (lost identity)")
	}
	if len(st.Fields) != 1 || st.Fields[0].Name != "x" {
		t.Fatalf("fields not populated: %+v", st.Fields)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestStructTypeForwardCompletion`
Expected: FAIL.

- [ ] **Step 3: Implement TagID + Field + StructType**

Append to `sema/type.go`:
```go
// TagID identifies a tag declaration. Tags are nominal: two struct
// declarations with identical layout are still distinct types.
type TagID struct {
	id int // assigned monotonically by TypeTable
}

type Field struct {
	Name       string
	T          Type
	BitWidth   int
	IsBitField bool
	Offset     int64
}

type StructType struct {
	Tag      *TagID
	Fields   []*Field
	Complete bool
}

func (*StructType) isType() {}
func (s *StructType) String() string {
	if !s.Complete {
		return "struct<incomplete>"
	}
	return "struct{...}"
}
```

Append to `sema/type_table.go`:
```go
func (tt *TypeTable) NewTagID() *TagID {
	tt.nextTagID++
	return &TagID{id: tt.nextTagID}
}

func (tt *TypeTable) Struct(tag *TagID) *StructType {
	// One *StructType per *TagID; reuse if already created.
	if s, ok := tt.structs[tag]; ok {
		return s
	}
	s := &StructType{Tag: tag, Complete: false}
	tt.structs[tag] = s
	return s
}

func (tt *TypeTable) CompleteStruct(s *StructType, fields []*Field) {
	s.Fields = fields
	s.Complete = true
}
```

Update `TypeTable` to add `nextTagID int` and `structs map[*TagID]*StructType`; init the map in `NewTypeTable`.

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_test.go
git commit -m "feat(sema): TagID + StructType with forward / in-place completion"
```

---

## Task 9: UnionType + EnumType + Enumerator

**Files:**
- Modify: `sema/type.go`, `sema/type_table.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestUnionAndEnum(t *testing.T) {
	tt := NewTypeTable()

	uTag := tt.NewTagID()
	u := tt.Union(uTag)
	if u.Complete {
		t.Fatalf("forward union should be incomplete")
	}
	intT := tt.Builtin(Int)
	tt.CompleteUnion(u, []*Field{{Name: "i", T: intT}})
	if !u.Complete || len(u.Fields) != 1 {
		t.Fatalf("CompleteUnion failed: %+v", u)
	}

	eTag := tt.NewTagID()
	e := tt.Enum(eTag)
	tt.CompleteEnum(e, intT, []*Enumerator{{Name: "RED", Value: 0}})
	if !e.Complete || e.Underlying != intT || len(e.Enumerators) != 1 {
		t.Fatalf("CompleteEnum failed: %+v", e)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestUnionAndEnum`
Expected: FAIL.

- [ ] **Step 3: Implement Union + Enum**

Append to `sema/type.go`:
```go
type UnionType struct {
	Tag      *TagID
	Fields   []*Field
	Complete bool
}

func (*UnionType) isType() {}
func (u *UnionType) String() string {
	if !u.Complete {
		return "union<incomplete>"
	}
	return "union{...}"
}

type Enumerator struct {
	Name  string
	Value int64
}

type EnumType struct {
	Tag         *TagID
	Underlying  Type
	Enumerators []*Enumerator
	Complete    bool
}

func (*EnumType) isType() {}
func (e *EnumType) String() string {
	if !e.Complete {
		return "enum<incomplete>"
	}
	return "enum{...}"
}
```

Append to `sema/type_table.go`:
```go
func (tt *TypeTable) Union(tag *TagID) *UnionType {
	if u, ok := tt.unions[tag]; ok {
		return u
	}
	u := &UnionType{Tag: tag}
	tt.unions[tag] = u
	return u
}

func (tt *TypeTable) CompleteUnion(u *UnionType, fields []*Field) {
	u.Fields = fields
	u.Complete = true
}

func (tt *TypeTable) Enum(tag *TagID) *EnumType {
	if e, ok := tt.enums[tag]; ok {
		return e
	}
	e := &EnumType{Tag: tag}
	tt.enums[tag] = e
	return e
}

func (tt *TypeTable) CompleteEnum(e *EnumType, underlying Type, enumerators []*Enumerator) {
	e.Underlying = underlying
	e.Enumerators = enumerators
	e.Complete = true
}
```

Add `unions map[*TagID]*UnionType` and `enums map[*TagID]*EnumType` to `TypeTable`; init both in `NewTypeTable`.

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_table.go sema/type_test.go
git commit -m "feat(sema): UnionType, EnumType, and Enumerator"
```

---

## Task 10: ErrorType singleton + IsError helper

**Files:**
- Modify: `sema/type.go`, `sema/type_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/type_test.go`:
```go
func TestErrorTypeSingleton(t *testing.T) {
	if !IsError(ErrorTypeSingleton) {
		t.Fatalf("IsError(ErrorTypeSingleton) = false, want true")
	}
	tt := NewTypeTable()
	if IsError(tt.Builtin(Int)) {
		t.Fatalf("IsError(int) = true, want false")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestErrorTypeSingleton`
Expected: FAIL.

- [ ] **Step 3: Implement**

Append to `sema/type.go`:
```go
type ErrorType struct{}

func (*ErrorType) isType()        {}
func (*ErrorType) String() string { return "<error-type>" }

var ErrorTypeSingleton Type = &ErrorType{}

func IsError(t Type) bool {
	return t == ErrorTypeSingleton
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/type.go sema/type_test.go
git commit -m "feat(sema): ErrorType sentinel and IsError helper"
```

---

## Task 11: Symbol, TagInfo, namespaces

**Files:**
- Create: `sema/scope.go`
- Create: `sema/scope_test.go`

- [ ] **Step 1: Write failing test**

`sema/scope_test.go`:
```go
package sema

import (
	"shinya.click/cvm/entity"
	"testing"
)

func TestSymbolFields(t *testing.T) {
	pos := entity.SourcePos{Line: 1, Column: 2}
	sym := &Symbol{
		Name:    "x",
		Kind:    SymVar,
		T:       NewTypeTable().Builtin(Int),
		Storage: StorageAuto,
		Linkage: LinkageNone,
		Pos:     pos,
	}
	if sym.Name != "x" || sym.Kind != SymVar {
		t.Fatalf("symbol fields not stored: %+v", sym)
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestSymbolFields`
Expected: FAIL.

- [ ] **Step 3: Implement scope.go skeleton**

`sema/scope.go`:
```go
package sema

import "shinya.click/cvm/entity"

type SymbolNamespace int

const (
	NSOrdinary SymbolNamespace = iota
	NSTag
	NSLabel
	NSMember
)

type SymbolKind int

const (
	SymVar SymbolKind = iota
	SymFunc
	SymTypedef
	SymEnumerator
	SymParam
)

type StorageClass int

const (
	StorageNone StorageClass = iota
	StorageAuto
	StorageRegister
	StorageStatic
	StorageExtern
	StorageTypedef
)

type Linkage int

const (
	LinkageNone Linkage = iota
	LinkageInternal
	LinkageExternal
)

// Decl is the IR declaration interface defined in Plan B (sema/ir.go).
// In Plan A we leave it as `any` so Symbol can carry the declaration list
// without depending on the IR layer that does not yet exist.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	T        Type
	Storage  StorageClass
	Linkage  Linkage
	Decl     any        // *Decl in Plan B
	Defs     []any      // []Decl in Plan B
	Pos      entity.SourcePos
	SlotID   int
	GlobalID int
}

type TagInfo struct {
	Tag      *TagID
	T        Type
	Decl     any
	Complete bool
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/scope.go sema/scope_test.go
git commit -m "feat(sema): Symbol, TagInfo, namespace and storage-class enums"
```

---

## Task 12: Scope struct + Lookup chain

**Files:**
- Modify: `sema/scope.go`, `sema/scope_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/scope_test.go`:
```go
func TestScopeLookup(t *testing.T) {
	root := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	root.Insert("x", &Symbol{Name: "x", Kind: SymVar, T: intT})

	inner := NewScope(ScopeBlock, root)
	if got := inner.Lookup("x", NSOrdinary); got == nil {
		t.Fatalf("inner.Lookup did not walk parent chain")
	}
	if got := inner.LookupCurrent("x", NSOrdinary); got != nil {
		t.Fatalf("inner.LookupCurrent returned outer symbol")
	}
}

func TestScopeNamespaceIsolation(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	s.Insert("foo", &Symbol{Name: "foo", Kind: SymVar, T: intT})
	tagId := NewTypeTable().NewTagID()
	s.InsertTag("foo", &TagInfo{Tag: tagId})
	if s.Lookup("foo", NSOrdinary) == nil {
		t.Fatalf("ordinary 'foo' lost")
	}
	if s.LookupTag("foo") == nil {
		t.Fatalf("tag 'foo' lost")
	}
}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `go test ./sema/ -run TestScope`
Expected: FAIL.

- [ ] **Step 3: Implement Scope**

Append to `sema/scope.go`:
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

func NewScope(kind ScopeKind, parent *Scope) *Scope {
	return &Scope{
		Kind:     kind,
		Parent:   parent,
		Ordinary: map[string]*Symbol{},
		Tags:     map[string]*TagInfo{},
	}
}

func (s *Scope) Lookup(name string, ns SymbolNamespace) *Symbol {
	if ns != NSOrdinary {
		return nil
	}
	for cur := s; cur != nil; cur = cur.Parent {
		if sym, ok := cur.Ordinary[name]; ok {
			return sym
		}
	}
	return nil
}

func (s *Scope) LookupCurrent(name string, ns SymbolNamespace) *Symbol {
	if ns != NSOrdinary {
		return nil
	}
	return s.Ordinary[name]
}

func (s *Scope) LookupTag(name string) *TagInfo {
	for cur := s; cur != nil; cur = cur.Parent {
		if t, ok := cur.Tags[name]; ok {
			return t
		}
	}
	return nil
}

func (s *Scope) Insert(name string, sym *Symbol) {
	s.Ordinary[name] = sym
}

func (s *Scope) InsertTag(name string, info *TagInfo) {
	s.Tags[name] = info
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/scope.go sema/scope_test.go
git commit -m "feat(sema): Scope with Lookup chain and namespace isolation"
```

---

## Task 13: Insert with redeclaration error checking

**Files:**
- Modify: `sema/scope.go`, `sema/scope_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/scope_test.go`:
```go
func TestScopeInsertRedeclarationError(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	intT := NewTypeTable().Builtin(Int)
	pos1 := entity.SourcePos{Line: 1}
	pos2 := entity.SourcePos{Line: 2}
	if err := s.InsertChecked("x", &Symbol{Name: "x", Kind: SymVar, T: intT, Pos: pos1}); err != nil {
		t.Fatalf("first insert errored: %v", err)
	}
	if err := s.InsertChecked("x", &Symbol{Name: "x", Kind: SymTypedef, T: intT, Pos: pos2}); err == nil {
		t.Fatalf("second insert with different kind should error")
	}
}
```

- [ ] **Step 2: Run test to verify fail**

Run: `go test ./sema/ -run TestScopeInsertRedeclarationError`
Expected: FAIL.

- [ ] **Step 3: Implement InsertChecked**

Append to `sema/scope.go`:
```go
// InsertChecked inserts a symbol; if a same-name symbol of a different
// SymbolKind already exists in the current scope, returns RedefinitionSymbol.
// Same-kind redeclarations (e.g. tentative function decls) are accepted by
// merging into the existing symbol's Defs slice. Type-compatibility checks
// are NOT performed here; that lives in Plan B's decl walker, which has
// the full Type information.
func (s *Scope) InsertChecked(name string, sym *Symbol) error {
	if existing, ok := s.Ordinary[name]; ok {
		if existing.Kind != sym.Kind {
			return RedefinitionSymbol(sym.Pos, existing.Pos, name)
		}
		existing.Defs = append(existing.Defs, sym.Decl)
		return nil
	}
	s.Ordinary[name] = sym
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/scope.go sema/scope_test.go
git commit -m "feat(sema): InsertChecked with same-kind merging and cross-kind error"
```

---

## Task 14: SymbolTable wrapper + tag InsertChecked

**Files:**
- Modify: `sema/scope.go`, `sema/scope_test.go`

- [ ] **Step 1: Write failing test**

Append to `sema/scope_test.go`:
```go
func TestSymbolTableFileScope(t *testing.T) {
	st := NewSymbolTable()
	if st.File == nil {
		t.Fatalf("SymbolTable.File not initialized")
	}
	if st.File.Kind != ScopeFile {
		t.Fatalf("file scope wrong kind: %v", st.File.Kind)
	}
}

func TestScopeInsertTagChecked(t *testing.T) {
	s := NewScope(ScopeFile, nil)
	tag1 := &TagID{}
	tag2 := &TagID{}
	pos1 := entity.SourcePos{Line: 1}
	pos2 := entity.SourcePos{Line: 2}
	st1 := &StructType{Tag: tag1}
	st2 := &StructType{Tag: tag2}
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag1, T: st1}, pos1); err != nil {
		t.Fatalf("first tag insert errored: %v", err)
	}
	// Same-tag re-insertion (forward decl + same forward decl in same scope) is OK.
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag1, T: st1}, pos2); err != nil {
		t.Fatalf("same-tag re-insert errored: %v", err)
	}
	// Different tag with same name in same scope -> error.
	if err := s.InsertTagChecked("S", &TagInfo{Tag: tag2, T: st2}, pos2); err == nil {
		t.Fatalf("conflicting tag should error")
	}
}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `go test ./sema/ -run "TestSymbolTableFileScope|TestScopeInsertTagChecked"`
Expected: FAIL.

- [ ] **Step 3: Implement**

Append to `sema/scope.go`:
```go
type SymbolTable struct {
	File *Scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{File: NewScope(ScopeFile, nil)}
}

func (s *Scope) InsertTagChecked(name string, info *TagInfo, pos entity.SourcePos) error {
	if existing, ok := s.Tags[name]; ok {
		if existing.Tag != info.Tag {
			return RedefinitionSymbol(pos, entity.SourcePos{}, name)
		}
		return nil
	}
	s.Tags[name] = info
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/scope.go sema/scope_test.go
git commit -m "feat(sema): SymbolTable wrapper and tag InsertChecked"
```

---

## Task 15: Migrate TimberSaw to sema/prefilter.go

**Files:**
- Create: `sema/prefilter.go`

- [ ] **Step 1: Inspect current TimberSaw**

Run: `cat parser/timbersaw.go`

Note the algorithm: walk AST tracking typedef-vs-variable scope, fail (return error) when a `PrimaryExpression { IDENTIFIER }` is a typedef name or a `TypedefName { IDENTIFIER }` is a variable name.

- [ ] **Step 2: Copy to sema/prefilter.go with package + import changes**

`sema/prefilter.go`:
```go
package sema

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

// PreFilter walks each candidate parse tree, drops candidates whose
// IDENTIFIER usages are inconsistent with declared typedefs (the residual
// ambiguity left by parser fork pruning). Errors from culled candidates
// are returned alongside survivors so the caller can feed them into the
// best-error pool.
func PreFilter(forest []*entity.AstNode) (survivors []*entity.AstNode, errs []*common.CvmError) {
	for _, root := range forest {
		err := newTimberSaw().chop(root)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		survivors = append(survivors, root)
	}
	return survivors, errs
}

type timberSaw struct {
	symbolStack *symbolStack
}

type symbolStack struct {
	stack []*scopeSymbols
	depth int
}

type scopeSymbols struct {
	TypeNames map[string]*entity.Token
	VarNames  map[string]*entity.Token
}

func newScopeSymbols() *scopeSymbols {
	return &scopeSymbols{
		TypeNames: map[string]*entity.Token{},
		VarNames:  map[string]*entity.Token{},
	}
}

func newTimberSaw() *timberSaw {
	s := &timberSaw{symbolStack: &symbolStack{}}
	s.symbolStack.stack = append(s.symbolStack.stack, newScopeSymbols())
	return s
}

func (s *symbolStack) resize(length int) {
	for len(s.stack) < length {
		s.stack = append(s.stack, newScopeSymbols())
	}
	if len(s.stack) > length {
		s.stack = s.stack[:length]
	}
}

func (s *symbolStack) enterScope(depth int) {
	s.resize(depth + 1)
	s.depth = depth
}

func (s *symbolStack) switchScope(depth int) {
	if len(s.stack) < depth+1 {
		s.resize(depth + 1)
	}
	s.depth = depth
}

func (s *symbolStack) pushVar(token *entity.Token) error {
	name := token.Lexeme
	if previous, ok := s.stack[s.depth].TypeNames[name]; ok {
		return invalidSymbolKind(token.SourceStart, previous.SourceStart, name)
	}
	s.stack[s.depth].VarNames[name] = token
	return nil
}

func (s *symbolStack) pushType(token *entity.Token) error {
	name := token.Lexeme
	if previous, ok := s.stack[s.depth].VarNames[name]; ok {
		return invalidSymbolKind(token.SourceStart, previous.SourceStart, name)
	}
	s.stack[s.depth].TypeNames[name] = token
	return nil
}

// isTypeName: true if name is a typedef in any visible scope.
// Used to validate PrimaryExpression { IDENTIFIER } and TypedefName { IDENTIFIER }.
func (s *symbolStack) isTypeName(name string) (bool, *entity.Token) {
	for i := s.depth; i >= 0; i-- {
		if t, ok := s.stack[i].TypeNames[name]; ok {
			return true, t
		}
		if _, ok := s.stack[i].VarNames[name]; ok {
			return false, nil
		}
	}
	return false, nil
}

func (s *timberSaw) chop(node *entity.AstNode) error {
	switch node.Typ {
	case parser.FunctionDefinition:
		declarator := node.Children[1]
		for _, token := range declarator.DeclaratorID {
			if err := s.symbolStack.pushVar(token); err != nil {
				return err
			}
		}
		for i := 0; i < 2; i++ {
			if err := s.chop(node.Children[i]); err != nil {
				return err
			}
		}
		for _, param := range getParametersFromFunctionDeclarator(declarator) {
			depth := s.symbolStack.depth
			s.symbolStack.enterScope(depth + 1)
			if err := s.symbolStack.pushVar(param); err != nil {
				return err
			}
			s.symbolStack.enterScope(depth)
		}
		current := s.symbolStack.depth
		if err := s.chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
		s.symbolStack.enterScope(current)

	case parser.Declaration:
		typedef := node.Children[0].TypeDef
		ids := node.Children[0].DeclaratorID
		if node.ReducedBy(parser.Declaration, 2) {
			ids = append(ids, node.Children[1].DeclaratorID...)
		}
		for _, id := range ids {
			if typedef {
				if err := s.symbolStack.pushType(id); err != nil {
					return err
				}
			} else {
				if err := s.symbolStack.pushVar(id); err != nil {
					return err
				}
			}
		}
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}

	case parser.IterationStatement:
		if node.ReducedBy(parser.IterationStatement, 1) || node.ReducedBy(parser.IterationStatement, 2) {
			for _, child := range node.Children {
				if err := s.chop(child); err != nil {
					return err
				}
			}
			break
		}
		current := s.symbolStack.depth
		s.symbolStack.switchScope(current + 1)
		for i := 0; i < len(node.Children)-1; i++ {
			if err := s.chop(node.Children[i]); err != nil {
				return err
			}
		}
		s.symbolStack.switchScope(current)
		if err := s.chop(node.Children[len(node.Children)-1]); err != nil {
			return err
		}
		s.symbolStack.enterScope(current)

	case entity.LEFT_BRACES:
		s.symbolStack.enterScope(s.symbolStack.depth + 1)
	case entity.RIGHT_BRACES:
		s.symbolStack.enterScope(s.symbolStack.depth - 1)

	case parser.PrimaryExpression:
		if node.ReducedBy(parser.PrimaryExpression, 1) {
			name := node.Children[0].Terminal.Lexeme
			isType, prev := s.symbolStack.isTypeName(name)
			if isType {
				return invalidSymbolKind(node.Children[0].Terminal.SourceStart,
					prev.SourceStart, name)
			}
			// undeclared identifiers are NOT culled here; Sema reports them later.
		}
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}

	case parser.TypedefName:
		name := node.Children[0].Terminal.Lexeme
		isType, _ := s.symbolStack.isTypeName(name)
		if !isType {
			// At this point parser produced a typedef_name reduction but the
			// identifier is not a typedef in any visible scope. Cull this
			// candidate.
			return UndeclaredIdentifier(node.Children[0].Terminal.SourceStart, name)
		}

	default:
		for _, child := range node.Children {
			if err := s.chop(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func getParametersFromFunctionDeclarator(declarator *entity.AstNode) []*entity.Token {
	direct := declarator.Children[len(declarator.Children)-1]
	switch {
	case direct.ReducedBy(parser.DirectDeclarator, 12):
		return direct.Children[2].DeclaratorID
	case direct.ReducedBy(parser.DirectDeclarator, 13):
		return nil
	case direct.ReducedBy(parser.DirectDeclarator, 14):
		return direct.Children[2].DeclaratorID
	}
	panic(fmt.Sprintf("unexpected direct_declarator: %s", direct.Production))
}

func invalidSymbolKind(pos, previous entity.SourcePos, name string) *common.CvmError {
	return InvalidSymbolKind(pos, previous, name)
}
```

- [ ] **Step 3: Add prefilter test**

`sema/prefilter_test.go`:
```go
package sema

import (
	"shinya.click/cvm/lexer"
	"shinya.click/cvm/parser"
	"testing"
)

func TestPreFilterTypedefShadow(t *testing.T) {
	// Outer typedef shadowed by inner variable. Parser produces multiple
	// candidates; PreFilter should return at least one survivor (the
	// variable interpretation).
	src := `typedef int a;
int main() {
    int a;
    int b;
    a*b;
}`
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("parser returned no candidates")
	}
	survivors, _ := PreFilter(candidates)
	if len(survivors) == 0 {
		t.Fatal("PreFilter culled all candidates")
	}
}

func TestPreFilterDoesNotCullUndeclared(t *testing.T) {
	// `scanf` is not declared. PreFilter should NOT cull this candidate;
	// the undeclared error is Sema's responsibility (Plan B).
	src := `int main() { scanf("%d", &x); }`
	tokens, err := lexer.NewLexer(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	survivors, _ := PreFilter(candidates)
	if len(survivors) == 0 {
		t.Fatal("PreFilter incorrectly culled undeclared-identifier candidate")
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sema/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sema/prefilter.go sema/prefilter_test.go
git commit -m "feat(sema): migrate TimberSaw to sema/prefilter.go (typedef-only check)"
```

---

## Task 16: Remove chopForest call from parser.Parse

**Files:**
- Modify: `parser/parser.go`

- [ ] **Step 1: Inspect current Parse**

Run: `sed -n '32,180p' parser/parser.go`

Note the `chopForest` block at the end of `Parse()`.

- [ ] **Step 2: Remove chopForest block**

In `parser/parser.go`, replace:
```go
	if len(p.CandidateASTs) == 0 {
		return nil, p.bestError
	}

	// eliminate the wrong tree
	common.DebugPrintf("Chop Start: %d candidates\n", len(p.CandidateASTs))
	candidates, chopErrs := chopForest(p.CandidateASTs)
	for _, err := range chopErrs {
		p.recordError(err)
	}
	if len(candidates) == 0 {
		return nil, p.bestError
	}
	for _, tree := range candidates {
		fillAstParent(tree, nil)
	}
	common.DebugPrintf("Chop Result: %d candidates\n", len(candidates))
	for i, candidate := range candidates {
		common.DebugPrintf("Tree %d\n", i)
		printAST(candidate, 0)
		common.DebugPrintln()
		common.DebugPrintln()
	}
	return candidates, nil
}
```

with:
```go
	if len(p.CandidateASTs) == 0 {
		return nil, p.bestError
	}
	for _, tree := range p.CandidateASTs {
		fillAstParent(tree, nil)
	}
	return p.CandidateASTs, nil
}
```

- [ ] **Step 3: Build to verify`chopForest` references are gone**

Run: `go build ./...`
Expected: error pointing to the `chopForest` declaration in `parser/timbersaw.go` being unused — that's expected; we delete it next task.

- [ ] **Step 4: Commit**

```bash
git add parser/parser.go
git commit -m "refactor(parser): Parse() no longer calls chopForest; returns raw candidate forest"
```

---

## Task 17: Delete parser/timbersaw.go

**Files:**
- Delete: `parser/timbersaw.go`

- [ ] **Step 1: Delete the file**

Run: `rm parser/timbersaw.go`

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Run all tests except the parser end-to-end ones that depend on chop**

Run: `go test ./parser/ ./sema/ ./lexer/`
Expected: parser tests that previously relied on chop (e.g. `TestSimpleFib` which failed on `scanf`) now pass at the parser level — but we have not yet wired PreFilter into the compiler, so the chop-equivalent step is not happening for parser tests. That's intentional; tests that need full pipeline live in `sema/` (Plan C migrates them).

If any parser test fails because it called `parser.Parse()` and expected an error from chop, that's a regression to investigate. As of B-stage, the parser tests do not assert chop behavior, so they should all pass.

- [ ] **Step 4: Commit**

```bash
git add -u parser/timbersaw.go
git commit -m "refactor(parser): delete timbersaw.go (logic now in sema/prefilter.go)"
```

---

## Task 18: Wire compiler.go to call sema.PreFilter

**Files:**
- Modify: `compiler.go`

- [ ] **Step 1: Inspect current compiler**

Run: `cat compiler.go`

- [ ] **Step 2: Update RunSource to invoke PreFilter**

Replace the body of `RunSource`:
```go
func (c *Compiler) RunSource(source string) {
	if c.FileName == "" {
		c.FileName = "main.c"
	}
	c.Source = source
	c.Lines = strings.Split(source, "\n")
	tokens, err := lexer.NewLexer(c.Source).ScanTokens()
	if err != nil {
		c.handleError(err)
		return
	}
	candidates, err := parser.NewParser(tokens).Parse()
	if err != nil {
		c.handleError(err)
		return
	}
	survivors, prefilterErrs := sema.PreFilter(candidates)
	if len(survivors) == 0 {
		if len(prefilterErrs) > 0 {
			c.handleError(prefilterErrs[0])
		}
		return
	}
	// Plan B/C wires the actual Sema walker here. For now just exit success
	// once we have at least one survivor.
}
```

Add the import: `"shinya.click/cvm/sema"`.

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 4: Run all tests**

Run: `go test ./...`
Expected: same PASS/FAIL set as before Plan A (Plan A is purely refactor + new infrastructure; no behavior change in the integrated pipeline).

- [ ] **Step 5: Commit**

```bash
git add compiler.go
git commit -m "refactor(compiler): wire sema.PreFilter into RunSource (Plan A integration)"
```

---

## Task 19: Verify no test regressions across packages

**Files:** none

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -count=1`
Expected: same PASS/FAIL count as the start of Plan A (TestSimpleFib still fails on `scanf` chop — that error now comes from sema.PreFilter via compiler.go, not parser).

- [ ] **Step 2: Run `go vet`**

Run: `go vet ./...`
Expected: clean.

- [ ] **Step 3: Run `gofmt -l`**

Run: `gofmt -l ./sema/ ./parser/ ./compiler.go`
Expected: no output (all files formatted).

- [ ] **Step 4: Confirm package shapes**

Run: `ls sema/`
Expected files: `doc.go`, `error.go`, `type.go`, `type_table.go`, `type_uintptr.go`, `scope.go`, `prefilter.go`, `type_test.go`, `scope_test.go`, `prefilter_test.go`.

Run: `ls parser/timbersaw.go 2>&1 | head -1`
Expected: `ls: parser/timbersaw.go: No such file or directory`.

- [ ] **Step 5: Final commit (if any cleanup)**

If gofmt found anything, run `gofmt -w` and commit:
```bash
gofmt -w ./sema/ ./parser/ ./compiler.go
git diff --stat
git add -A
git commit -m "style: gofmt sweep at end of Plan A"
```

Otherwise nothing to commit.

---

## Plan A Acceptance

After Task 19:
- `sema/` package exists with Type system (BuiltinType, PointerType, ArrayType with conditional interning, FunctionType, QualType, StructType / UnionType / EnumType with forward / in-place completion, ErrorType singleton).
- Symbol / Scope / SymbolTable infrastructure in place with namespace separation, lookup chain, redeclaration error.
- `sema/prefilter.go` ports TimberSaw with the `undeclared identifier` check stripped (downstream to Sema).
- `parser/timbersaw.go` deleted; `parser.Parse()` returns raw forest.
- `compiler.go` calls `sema.PreFilter` between parse and (future) Sema walk.
- All Plan A tests pass; existing parser/lexer test PASS/FAIL set unchanged.

Plan B picks up here to add IR nodes, constant evaluator, and the Sema walkers themselves.
