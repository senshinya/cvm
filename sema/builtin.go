package sema

import "shinya.click/cvm/entity"

func (s *Sema) lookupBuiltin(name string, pos entity.SourcePos) *Symbol {
	if sym := s.SymTab.File.LookupCurrent(name, NSOrdinary); sym != nil {
		return sym
	}
	ft := s.builtinFunctionType(name)
	if ft == nil {
		return nil
	}
	sym := &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: StorageExtern, Linkage: LinkageExternal, Pos: pos}
	s.SymTab.File.Insert(name, sym)
	return sym
}

func (s *Sema) builtinFunctionType(name string) *FunctionType {
	sizeT := s.Types.Builtin(ULong)
	voidPtr := s.Types.Pointer(s.Types.Builtin(Void))
	constCharPtr := s.Types.Pointer(s.Types.Qualified(s.Types.Builtin(Char), true, false, false))
	charPtr := s.Types.Pointer(s.Types.Builtin(Char))

	switch name {
	case "__builtin_malloc":
		return s.Types.Function(voidPtr, []Type{sizeT}, false, true)
	case "__builtin_calloc":
		return s.Types.Function(voidPtr, []Type{sizeT, sizeT}, false, true)
	case "__builtin_strdup":
		return s.Types.Function(charPtr, []Type{constCharPtr}, false, true)
	default:
		return nil
	}
}
