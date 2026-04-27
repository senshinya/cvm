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

// InsertChecked 插入普通命名空间符号；如果当前作用域里同名符号的 kind 不同，
// 返回重定义错误。同 kind 的再次声明暂时只合并到 Defs，占位给 Plan B 的完整类型
// 兼容性检查使用。
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

type SymbolTable struct {
	File *Scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{File: NewScope(ScopeFile, nil)}
}

func (s *Scope) InsertTagChecked(name string, info *TagInfo, pos entity.SourcePos) error {
	if existing, ok := s.Tags[name]; ok {
		if existing.Tag != info.Tag || !sameTagTypeKind(existing.T, info.T) {
			return RedefinitionSymbol(pos, entity.SourcePos{}, name)
		}
		return nil
	}
	s.Tags[name] = info
	return nil
}

func sameTagTypeKind(a, b Type) bool {
	switch a.(type) {
	case *StructType:
		_, ok := b.(*StructType)
		return ok
	case *UnionType:
		_, ok := b.(*UnionType)
		return ok
	case *EnumType:
		_, ok := b.(*EnumType)
		return ok
	default:
		return a == nil || b == nil || a == b
	}
}
