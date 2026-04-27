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

type Symbol struct {
	Name     string
	Kind     SymbolKind
	T        Type
	Storage  StorageClass
	Linkage  Linkage
	Decl     any
	Defs     []any
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
