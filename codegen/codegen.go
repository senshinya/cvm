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

func (fg *funcGen) emitStmt(s sema.Stmt) error {
	switch x := s.(type) {
	case *sema.Block:
		for _, item := range x.Items {
			if err := fg.emitStmt(item); err != nil {
				return err
			}
		}
	case *sema.ReturnStmt:
		if x.Value == nil {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpReturnVoid})
			return nil
		}
		if err := fg.emitValue(x.Value); err != nil {
			return err
		}
		t, err := fg.g.lowerValueType(x.Value.GetType())
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Return(t))
	default:
		return &Error{Pos: s.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "no statement lowering registered before scalar dispatcher task"}
	}
	return nil
}

func (fg *funcGen) emitValue(e sema.Expr) error {
	switch x := e.(type) {
	case *sema.IntLit:
		t, err := fg.g.lowerValueType(x.T)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(t, x.Value))
	case *sema.ImplicitCast:
		if err := fg.emitValue(x.X); err != nil {
			return err
		}
		from, err := fg.g.lowerValueType(x.From)
		if err != nil {
			return err
		}
		to, err := fg.g.lowerValueType(x.To)
		if err != nil {
			return err
		}
		if from != to {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Cast(from, to, castOpFor(x.Kind, from, to)))
		}
	default:
		return &Error{Pos: e.Pos().SourceStart, Node: fmt.Sprintf("%T", e), Op: "emitValue", Reason: "no expression lowering registered before scalar dispatcher task"}
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
