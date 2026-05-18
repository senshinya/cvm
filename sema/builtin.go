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
	intT := s.Types.Builtin(Int)
	longT := s.Types.Builtin(Long)
	unsignedLongT := s.Types.Builtin(ULong)
	longLongT := s.Types.Builtin(LongLong)
	unsignedLongLongT := s.Types.Builtin(ULongLong)
	doubleT := s.Types.Builtin(Double)
	floatT := s.Types.Builtin(Float)
	longDoubleT := s.Types.Builtin(LongDouble)
	voidT := s.Types.Builtin(Void)
	voidPtr := s.Types.Pointer(s.Types.Builtin(Void))
	constVoidPtr := s.Types.Pointer(s.Types.Qualified(s.Types.Builtin(Void), true, false, false))
	constCharPtr := s.Types.Pointer(s.Types.Qualified(s.Types.Builtin(Char), true, false, false))
	charPtr := s.Types.Pointer(s.Types.Builtin(Char))
	charPtrPtr := s.Types.Pointer(charPtr)
	vaList := voidPtr

	switch name {
	case "__builtin_malloc", "malloc":
		return s.Types.Function(voidPtr, []Type{sizeT}, false, true)
	case "__builtin_calloc", "calloc":
		return s.Types.Function(voidPtr, []Type{sizeT, sizeT}, false, true)
	case "realloc":
		return s.Types.Function(voidPtr, []Type{voidPtr, sizeT}, false, true)
	case "__builtin_strdup", "strdup":
		return s.Types.Function(charPtr, []Type{constCharPtr}, false, true)
	case "free":
		return s.Types.Function(voidT, []Type{voidPtr}, false, true)
	case "atof":
		return s.Types.Function(doubleT, []Type{constCharPtr}, false, true)
	case "strtol":
		return s.Types.Function(longT, []Type{constCharPtr, charPtrPtr, intT}, false, true)
	case "strtoul":
		return s.Types.Function(unsignedLongT, []Type{constCharPtr, charPtrPtr, intT}, false, true)
	case "strtoll":
		return s.Types.Function(longLongT, []Type{constCharPtr, charPtrPtr, intT}, false, true)
	case "strtoull":
		return s.Types.Function(unsignedLongLongT, []Type{constCharPtr, charPtrPtr, intT}, false, true)
	case "strtod":
		return s.Types.Function(doubleT, []Type{constCharPtr, charPtrPtr}, false, true)
	case "strtof":
		return s.Types.Function(floatT, []Type{constCharPtr, charPtrPtr}, false, true)
	case "strtold":
		return s.Types.Function(longDoubleT, []Type{constCharPtr, charPtrPtr}, false, true)
	case "__builtin_memcpy", "__builtin_memmove", "__builtin_mempcpy":
		return s.Types.Function(voidPtr, []Type{voidPtr, constVoidPtr, sizeT}, false, true)
	case "memcpy", "memmove", "mempcpy":
		return s.Types.Function(voidPtr, []Type{voidPtr, constVoidPtr, sizeT}, false, true)
	case "__builtin_memset":
		return s.Types.Function(voidPtr, []Type{voidPtr, intT, sizeT}, false, true)
	case "memset":
		return s.Types.Function(voidPtr, []Type{voidPtr, intT, sizeT}, false, true)
	case "__builtin_bzero", "bzero":
		return s.Types.Function(voidT, []Type{voidPtr, sizeT}, false, true)
	case "bcmp":
		return s.Types.Function(intT, []Type{constVoidPtr, constVoidPtr, sizeT}, false, true)
	case "bcopy":
		return s.Types.Function(voidT, []Type{constVoidPtr, voidPtr, sizeT}, false, true)
	case "__builtin_strcpy", "__builtin_stpcpy", "__builtin_strcat":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr}, false, true)
	case "strcpy", "stpcpy", "strcat":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr}, false, true)
	case "__builtin_strncpy", "__builtin_stpncpy", "__builtin_strncat":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr, sizeT}, false, true)
	case "strncpy", "stpncpy", "strncat":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr, sizeT}, false, true)
	case "__builtin_strlen":
		return s.Types.Function(sizeT, []Type{constCharPtr}, false, true)
	case "strlen":
		return s.Types.Function(sizeT, []Type{constCharPtr}, false, true)
	case "strnlen":
		return s.Types.Function(sizeT, []Type{constCharPtr, sizeT}, false, true)
	case "strerror":
		return s.Types.Function(charPtr, []Type{intT}, false, true)
	case "strncmp":
		return s.Types.Function(intT, []Type{constCharPtr, constCharPtr, sizeT}, false, true)
	case "__builtin_strchr":
		return s.Types.Function(charPtr, []Type{constCharPtr, intT}, false, true)
	case "strchr":
		return s.Types.Function(charPtr, []Type{constCharPtr, intT}, false, true)
	case "strrchr":
		return s.Types.Function(charPtr, []Type{constCharPtr, intT}, false, true)
	case "__builtin_strstr":
		return s.Types.Function(charPtr, []Type{constCharPtr, constCharPtr}, false, true)
	case "strstr":
		return s.Types.Function(charPtr, []Type{constCharPtr, constCharPtr}, false, true)
	case "strpbrk":
		return s.Types.Function(charPtr, []Type{constCharPtr, constCharPtr}, false, true)
	case "strspn", "strcspn":
		return s.Types.Function(sizeT, []Type{constCharPtr, constCharPtr}, false, true)
	case "memchr":
		return s.Types.Function(voidPtr, []Type{constVoidPtr, intT, sizeT}, false, true)
	case "__builtin_object_size", "__builtin_dynamic_object_size":
		return s.Types.Function(sizeT, []Type{constVoidPtr, intT}, false, true)
	case "__builtin___memcpy_chk", "__builtin___memmove_chk", "__builtin___mempcpy_chk":
		return s.Types.Function(voidPtr, []Type{voidPtr, constVoidPtr, sizeT, sizeT}, false, true)
	case "__builtin___memset_chk":
		return s.Types.Function(voidPtr, []Type{voidPtr, intT, sizeT, sizeT}, false, true)
	case "__builtin___strcpy_chk", "__builtin___stpcpy_chk", "__builtin___strcat_chk":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr, sizeT}, false, true)
	case "__builtin___strncpy_chk", "__builtin___stpncpy_chk", "__builtin___strncat_chk":
		return s.Types.Function(charPtr, []Type{charPtr, constCharPtr, sizeT, sizeT}, false, true)
	case "__builtin_printf", "__builtin_printf_unlocked":
		return s.Types.Function(intT, []Type{constCharPtr}, true, true)
	case "printf", "printf_unlocked":
		return s.Types.Function(intT, []Type{constCharPtr}, true, true)
	case "__builtin_fprintf", "__builtin_fprintf_unlocked":
		return s.Types.Function(intT, []Type{voidPtr, constCharPtr}, true, true)
	case "fprintf", "fprintf_unlocked":
		return s.Types.Function(intT, []Type{voidPtr, constCharPtr}, true, true)
	case "__builtin_sprintf":
		return s.Types.Function(intT, []Type{charPtr, constCharPtr}, true, true)
	case "sprintf":
		return s.Types.Function(intT, []Type{charPtr, constCharPtr}, true, true)
	case "__builtin_snprintf":
		return s.Types.Function(intT, []Type{charPtr, sizeT, constCharPtr}, true, true)
	case "snprintf":
		return s.Types.Function(intT, []Type{charPtr, sizeT, constCharPtr}, true, true)
	case "__builtin_vprintf":
		return s.Types.Function(intT, []Type{constCharPtr, vaList}, false, true)
	case "vprintf":
		return s.Types.Function(intT, []Type{constCharPtr, vaList}, false, true)
	case "__builtin_vfprintf":
		return s.Types.Function(intT, []Type{voidPtr, constCharPtr, vaList}, false, true)
	case "vfprintf":
		return s.Types.Function(intT, []Type{voidPtr, constCharPtr, vaList}, false, true)
	case "__builtin_vsprintf":
		return s.Types.Function(intT, []Type{charPtr, constCharPtr, vaList}, false, true)
	case "vsprintf":
		return s.Types.Function(intT, []Type{charPtr, constCharPtr, vaList}, false, true)
	case "__builtin_vsnprintf":
		return s.Types.Function(intT, []Type{charPtr, sizeT, constCharPtr, vaList}, false, true)
	case "vsnprintf":
		return s.Types.Function(intT, []Type{charPtr, sizeT, constCharPtr, vaList}, false, true)
	case "__builtin___printf_chk":
		return s.Types.Function(intT, []Type{intT, constCharPtr}, true, true)
	case "__builtin___fprintf_chk":
		return s.Types.Function(intT, []Type{voidPtr, intT, constCharPtr}, true, true)
	case "__builtin___sprintf_chk":
		return s.Types.Function(intT, []Type{charPtr, intT, sizeT, constCharPtr}, true, true)
	case "__builtin___snprintf_chk":
		return s.Types.Function(intT, []Type{charPtr, sizeT, intT, sizeT, constCharPtr}, true, true)
	case "__builtin___vprintf_chk":
		return s.Types.Function(intT, []Type{intT, constCharPtr, vaList}, false, true)
	case "__builtin___vfprintf_chk":
		return s.Types.Function(intT, []Type{voidPtr, intT, constCharPtr, vaList}, false, true)
	case "__builtin___vsprintf_chk":
		return s.Types.Function(intT, []Type{charPtr, intT, sizeT, constCharPtr, vaList}, false, true)
	case "__builtin___vsnprintf_chk":
		return s.Types.Function(intT, []Type{charPtr, sizeT, intT, sizeT, constCharPtr, vaList}, false, true)
	case "__builtin_va_start", "__builtin_va_end":
		return s.Types.Function(voidT, nil, true, true)
	case "__builtin_pow":
		return s.Types.Function(doubleT, []Type{doubleT, doubleT}, false, true)
	case "__builtin_cabs":
		return s.Types.Function(doubleT, []Type{s.Types.Builtin(DoubleComplex)}, false, true)
	case "__builtin_cabsf":
		return s.Types.Function(floatT, []Type{s.Types.Builtin(FloatComplex)}, false, true)
	case "__builtin_cabsl":
		return s.Types.Function(longDoubleT, []Type{s.Types.Builtin(LongDoubleComplex)}, false, true)
	case "__builtin_huge_val":
		return s.Types.Function(doubleT, nil, false, true)
	case "__builtin_huge_valf":
		return s.Types.Function(floatT, nil, false, true)
	case "__builtin_huge_vall":
		return s.Types.Function(longDoubleT, nil, false, true)
	case "__builtin_nan":
		return s.Types.Function(doubleT, []Type{constCharPtr}, false, true)
	case "nan":
		return s.Types.Function(doubleT, []Type{constCharPtr}, false, true)
	case "__builtin_abort":
		return s.Types.Function(voidT, nil, false, true)
	case "__builtin_shuffle":
		return s.Types.Function(intT, nil, true, false)
	case "__builtin_complex":
		return s.Types.Function(s.Types.Builtin(DoubleComplex), nil, true, false)
	default:
		return nil
	}
}
