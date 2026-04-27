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
