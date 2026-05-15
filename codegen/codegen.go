package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

func Generate(prog *sema.Program) (*bytecode.Module, error) {
	if err := sema.ValidateProgramInvariants(prog); err != nil {
		return nil, err
	}
	g := &generator{
		prog:      prog,
		mod:       &bytecode.Module{Target: bytecode.DefaultTarget()},
		globalMap: map[*sema.Symbol]int{},
		sigMap:    map[string]int{},
		layoutMap: map[sema.Type]int{},
	}
	if err := g.emitModule(); err != nil {
		return nil, err
	}
	if err := bytecode.ValidateModule(g.mod); err != nil {
		return nil, err
	}
	return g.mod, nil
}

type generator struct {
	prog      *sema.Program
	mod       *bytecode.Module
	globalMap map[*sema.Symbol]int
	sigMap    map[string]int
	layoutMap map[sema.Type]int
	fn        *funcGen
}

type funcGen struct {
	g         *generator
	fn        *sema.FuncDef
	out       *bytecode.Function
	nextLabel int
}

func (g *generator) emitModule() error {
	if err := g.collectGlobals(); err != nil {
		return err
	}
	for _, fn := range g.prog.Funcs {
		if err := g.emitFunction(fn); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) collectGlobals() error {
	for _, d := range g.prog.Globals {
		switch x := d.(type) {
		case *sema.VarDecl:
			g.addGlobal(x.Sym, bytecode.GlobalVar, -1)
		case *sema.FuncDecl:
			g.addGlobal(x.Sym, bytecode.GlobalExtern, -1)
		}
	}
	for i, fn := range g.prog.Funcs {
		g.addGlobal(fn.Sym, bytecode.GlobalFunc, i)
		for _, local := range fn.Locals {
			if local != nil && local.Storage == sema.StorageStatic {
				g.addGlobal(local.Sym, bytecode.GlobalVar, -1)
			}
		}
	}
	return nil
}

func (g *generator) addGlobal(sym *sema.Symbol, kind bytecode.GlobalKind, fnIndex int) int {
	if id, ok := g.globalMap[sym]; ok {
		if kind == bytecode.GlobalFunc {
			g.mod.Globals[id].Kind = kind
			g.mod.Globals[id].Func = fnIndex
		}
		return id
	}
	id := sym.GlobalID
	for len(g.mod.Globals) <= id {
		g.mod.Globals = append(g.mod.Globals, bytecode.Global{ID: len(g.mod.Globals), Func: -1})
	}
	g.mod.Globals[id] = bytecode.Global{ID: id, Name: sym.Name, Kind: kind, Func: fnIndex}
	g.globalMap[sym] = id
	return id
}

func (g *generator) internSig(ret bytecode.ValueType, params []bytecode.ValueType, variadic bool) int {
	key := fmt.Sprintf("%s/%v/%v", ret, params, variadic)
	if id, ok := g.sigMap[key]; ok {
		return id
	}
	id := len(g.mod.Sigs)
	g.mod.Sigs = append(g.mod.Sigs, bytecode.FuncSig{ID: id, Ret: ret, Params: append([]bytecode.ValueType(nil), params...), Variadic: variadic})
	g.sigMap[key] = id
	return id
}

func (g *generator) emitFunction(fn *sema.FuncDef) error {
	ret, err := g.lowerValueType(fn.T.Ret)
	if err != nil {
		return err
	}
	params := make([]bytecode.ValueType, 0, len(fn.T.Params))
	for _, p := range fn.T.Params {
		pt, err := g.lowerValueType(p)
		if err != nil {
			return err
		}
		params = append(params, pt)
	}
	sig := g.internSig(ret, params, fn.T.Variadic)
	f := bytecode.Function{
		ID:       len(g.mod.Functions),
		GlobalID: fn.Sym.GlobalID,
		Name:     fn.Sym.Name,
		Sig:      sig,
	}
	for _, p := range fn.Params {
		pt, err := g.lowerValueType(p.T)
		if err != nil {
			return err
		}
		f.Params = append(f.Params, bytecode.Param{Name: p.Sym.Name, Type: pt, Slot: p.Sym.SlotID})
	}
	seenSlots := map[int]bool{}
	for _, p := range f.Params {
		seenSlots[p.Slot] = true
	}
	for _, local := range fn.Locals {
		if local == nil || local.Sym == nil || local.Storage == sema.StorageStatic || local.Storage == sema.StorageExtern {
			continue
		}
		vt, err := g.lowerValueType(local.T)
		if err != nil {
			return err
		}
		if local.Sym.SlotID >= 0 && isSlotType(vt) {
			if seenSlots[local.Sym.SlotID] {
				continue
			}
			seenSlots[local.Sym.SlotID] = true
			f.Locals = append(f.Locals, bytecode.LocalSlot{ID: local.Sym.SlotID, Name: local.Sym.Name, Type: vt})
			continue
		}
		layout := g.placeholderLayout(local.T)
		f.Objects = append(f.Objects, bytecode.LocalObject{ID: len(f.Objects), Name: local.Sym.Name, Size: layout.Size, Align: layout.Align, Layout: layout.ID})
	}
	fg := &funcGen{g: g, fn: fn, out: &f}
	if err := fg.emitStmt(fn.Body); err != nil {
		return err
	}
	g.mod.Functions = append(g.mod.Functions, *fg.out)
	if fn.Sym.GlobalID >= 0 && fn.Sym.GlobalID < len(g.mod.Globals) {
		g.mod.Globals[fn.Sym.GlobalID].Func = f.ID
		g.mod.Globals[fn.Sym.GlobalID].Kind = bytecode.GlobalFunc
	}
	return nil
}

func castOpFor(kind sema.CastKind, from, to bytecode.ValueType) bytecode.CastOp {
	switch kind {
	case sema.BoolConversion:
		return bytecode.CastBool
	case sema.IntToFloat:
		return bytecode.CastIntToFloat
	case sema.FloatToInt:
		return bytecode.CastFloatToInt
	case sema.IntToPointer:
		return bytecode.CastIntToPtr
	case sema.PointerToInt:
		return bytecode.CastPtrToInt
	case sema.PointerConversion, sema.VoidPointerConversion, sema.ArrayDecay, sema.FunctionDecay, sema.NullPointerConstant:
		return bytecode.CastBit
	}
	switch {
	case isIntegerType(from) && to == bytecode.TypeBool:
		return bytecode.CastBool
	case isIntegerType(from) && isIntegerType(to):
		return integerCastOp(from, to)
	case isIntegerType(from) && isFloatType(to):
		return bytecode.CastIntToFloat
	case isFloatType(from) && isIntegerType(to):
		return bytecode.CastFloatToInt
	case isFloatType(from) && isFloatType(to):
		if typeSize(from) < typeSize(to) {
			return bytecode.CastFExt
		}
		if typeSize(from) > typeSize(to) {
			return bytecode.CastFTrunc
		}
		return bytecode.CastBit
	case isPointerType(from) && isIntegerType(to):
		return bytecode.CastPtrToInt
	case isIntegerType(from) && isPointerType(to):
		return bytecode.CastIntToPtr
	case isPointerType(from) && isPointerType(to):
		return bytecode.CastBit
	default:
		return bytecode.CastBit
	}
}

func (g *generator) placeholderLayout(t sema.Type) bytecode.ObjectLayout {
	id, ok := g.layoutMap[t]
	if ok {
		return g.mod.Layouts[id]
	}
	id = len(g.mod.Layouts)
	layout := bytecode.ObjectLayout{ID: id, Name: t.String(), Size: sizeofType(t), Align: alignofType(t)}
	g.layoutMap[t] = id
	g.mod.Layouts = append(g.mod.Layouts, layout)
	return layout
}

func sizeofType(t sema.Type) int64 {
	switch x := sema.Unqual(t).(type) {
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.Void:
			return 1
		case sema.Bool, sema.Char, sema.SChar, sema.UChar:
			return 1
		case sema.Short, sema.UShort:
			return 2
		case sema.Int, sema.UInt, sema.Float:
			return 4
		case sema.Long, sema.ULong, sema.LongLong, sema.ULongLong, sema.Double:
			return 8
		case sema.LongDouble:
			return 16
		}
	case *sema.PointerType, *sema.FunctionType:
		return 8
	case *sema.ArrayType:
		if x.SizeKind == sema.ArrayConstantSize {
			return x.Size * sizeofType(x.Elem)
		}
	case *sema.StructType:
		var end int64
		for _, f := range x.Fields {
			if n := f.Offset + sizeofType(f.T); n > end {
				end = n
			}
		}
		return end
	case *sema.UnionType:
		var max int64
		for _, f := range x.Fields {
			if n := sizeofType(f.T); n > max {
				max = n
			}
		}
		return max
	case *sema.EnumType:
		return sizeofType(x.Underlying)
	}
	return 0
}

func alignofType(t sema.Type) int64 {
	switch x := sema.Unqual(t).(type) {
	case *sema.BuiltinType:
		switch x.Kind {
		case sema.Void, sema.Bool, sema.Char, sema.SChar, sema.UChar:
			return 1
		case sema.Short, sema.UShort:
			return 2
		case sema.Int, sema.UInt, sema.Float:
			return 4
		case sema.Long, sema.ULong, sema.LongLong, sema.ULongLong, sema.Double:
			return 8
		case sema.LongDouble:
			return 16
		}
	case *sema.PointerType, *sema.FunctionType:
		return 8
	case *sema.ArrayType:
		return alignofType(x.Elem)
	case *sema.StructType, *sema.UnionType:
		return 1
	case *sema.EnumType:
		return alignofType(x.Underlying)
	}
	return 1
}

func integerCastOp(from, to bytecode.ValueType) bytecode.CastOp {
	fromSize, toSize := typeSize(from), typeSize(to)
	switch {
	case fromSize > toSize:
		return bytecode.CastTrunc
	case fromSize < toSize && isUnsignedType(from):
		return bytecode.CastZExt
	case fromSize < toSize:
		return bytecode.CastSExt
	default:
		return bytecode.CastBit
	}
}

func isIntegerType(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64, bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64:
		return true
	default:
		return false
	}
}

func isUnsignedType(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeBool, bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64:
		return true
	default:
		return false
	}
}

func isFloatType(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return true
	default:
		return false
	}
}

func isPointerType(t bytecode.ValueType) bool {
	return t == bytecode.TypePtr || t == bytecode.TypeObjectAddr
}

func typeSize(t bytecode.ValueType) int {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		return 1
	case bytecode.TypeI16, bytecode.TypeU16:
		return 2
	case bytecode.TypeI32, bytecode.TypeU32, bytecode.TypeF32:
		return 4
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypeF64, bytecode.TypePtr, bytecode.TypeObjectAddr:
		return 8
	case bytecode.TypeFLong:
		return 16
	default:
		return 0
	}
}
