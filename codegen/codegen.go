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
		mod:       bytecode.NewModule(),
		globalMap: map[*sema.Symbol]int{},
		sigMap:    map[string]int{},
		layoutMap: map[sema.Type]int{},
		stringMap: map[string]int{},
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
	stringMap map[string]int
	fn        *funcGen
}

type funcGen struct {
	g                     *generator
	fn                    *sema.FuncDef
	out                   *bytecode.Function
	nextLabel             int
	objectMap             map[*sema.Symbol]int
	dynamicObjectMap      map[*sema.Symbol]int
	dynamicSizeSlotMap    map[*sema.Symbol]int
	dynamicSizeSymbolMap  map[*sema.Symbol]map[string]int
	dynamicPointerTypeMap map[*sema.Symbol]map[string]int
	activeDynamicObjects  []int
	nextSyntheticSlot     int
	addressTaken          map[*sema.Symbol]bool
	breaks                []int
	continues             []int
	breakCleanupMarks     []int
	continueCleanupMarks  []int
	namedBreaks           map[string][]int
	namedContinues        map[string][]int
	namedBreakCleanups    map[string][]int
	namedContinueCleanups map[string][]int
	pendingBreakNames     []string
	pendingContinueNames  []string
	labels                map[*sema.LabeledStmt]int
	labelCleanupMarks     map[*sema.LabeledStmt]int
	caseLabels            map[*sema.CaseStmt]int
	defaultLabels         map[*sema.DefaultStmt]int
}

func (g *generator) emitModule() error {
	if err := g.collectGlobals(); err != nil {
		return err
	}
	if err := g.emitStaticInitializers(); err != nil {
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
			kind := bytecode.GlobalVar
			if x.Storage == sema.StorageExtern {
				kind = bytecode.GlobalExtern
			}
			if _, err := g.addGlobal(x.Sym, kind, -1); err != nil {
				return err
			}
		case *sema.FuncDecl:
			globalID, err := g.addGlobal(x.Sym, bytecode.GlobalExtern, -1)
			if err != nil {
				return err
			}
			sig, err := g.lowerFuncSig(x.T)
			if err != nil {
				return err
			}
			g.mod.Globals[globalID].Sig = sig
		}
	}
	for i, fn := range g.prog.Funcs {
		globalID, err := g.addGlobal(fn.Sym, bytecode.GlobalFunc, i)
		if err != nil {
			return err
		}
		if fn.Sym.Name == "main" {
			g.mod.Entry = &bytecode.EntryPoint{Global: globalID, Name: fn.Sym.Name}
		}
		for _, local := range fn.Locals {
			if local != nil && local.Storage == sema.StorageStatic {
				if _, err := g.addGlobal(local.Sym, bytecode.GlobalVar, -1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *generator) addGlobal(sym *sema.Symbol, kind bytecode.GlobalKind, fnIndex int) (int, error) {
	if id, ok := g.globalMap[sym]; ok {
		if kind == bytecode.GlobalFunc {
			g.mod.Globals[id].Kind = kind
			g.mod.Globals[id].Func = fnIndex
			g.mod.Globals[id].Extern = bytecode.ExternRef{}
		}
		if kind == bytecode.GlobalVar && g.mod.Globals[id].Kind == bytecode.GlobalExtern {
			g.mod.Globals[id].Kind = kind
			g.mod.Globals[id].Extern = bytecode.ExternRef{}
			g.mod.Globals[id].Size = g.sizeof(sym.T)
			g.mod.Globals[id].Align = g.alignof(sym.T)
			g.mod.Globals[id].Init.ZeroFill = g.mod.Globals[id].Size
			if _, err := g.lowerLayout(sym.T); err != nil {
				return id, err
			}
		}
		return id, nil
	}
	id := sym.GlobalID
	for len(g.mod.Globals) <= id {
		g.mod.Globals = append(g.mod.Globals, bytecode.Global{ID: len(g.mod.Globals), Func: -1, Sig: bytecode.NoFuncSig})
	}
	global := bytecode.Global{ID: id, Name: sym.Name, Kind: kind, Func: fnIndex, Sig: bytecode.NoFuncSig}
	if kind == bytecode.GlobalExtern {
		global.Extern = externRefForSymbol(sym)
	}
	if kind == bytecode.GlobalVar || (kind == bytecode.GlobalExtern && sym.Kind == sema.SymVar) {
		global.Size = g.sizeof(sym.T)
		global.Align = g.alignof(sym.T)
	}
	if kind == bytecode.GlobalVar {
		global.Init.ZeroFill = global.Size
	}
	g.mod.Globals[id] = global
	g.globalMap[sym] = id
	if kind == bytecode.GlobalVar || (kind == bytecode.GlobalExtern && sym.Kind == sema.SymVar) {
		if _, err := g.lowerLayout(sym.T); err != nil {
			return id, err
		}
	}
	return id, nil
}

func externRefForSymbol(sym *sema.Symbol) bytecode.ExternRef {
	return bytecode.ExternRef{
		Name: sym.Name,
		ABI:  bytecode.DefaultExternABI,
	}
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
	sig, err := g.lowerFuncSig(fn.T)
	if err != nil {
		return err
	}
	addressTaken := findAddressTaken(fn)
	f := bytecode.Function{
		ID:       len(g.mod.Functions),
		GlobalID: fn.Sym.GlobalID,
		Name:     fn.Sym.Name,
		Sig:      sig,
	}
	paramDeclsBySlot := map[int]*sema.VarDecl{}
	for _, p := range fn.Params {
		pt, err := g.lowerValueType(p.T)
		if err != nil {
			return err
		}
		f.Params = append(f.Params, bytecode.Param{Name: p.Sym.Name, Type: pt, Slot: p.Sym.SlotID})
		paramDeclsBySlot[p.Sym.SlotID] = p
	}
	seenSlots := map[int]bool{}
	maxSlot := -1
	for _, p := range fn.Params {
		if p.Sym != nil && p.Sym.SlotID > maxSlot {
			maxSlot = p.Sym.SlotID
		}
	}
	objectMap := map[*sema.Symbol]int{}
	for _, p := range f.Params {
		seenSlots[p.Slot] = true
		if p.Slot > maxSlot {
			maxSlot = p.Slot
		}
		if pd := paramDeclsBySlot[p.Slot]; pd != nil && (addressTaken[pd.Sym] || isVolatile(pd.T)) {
			layout, err := g.lowerLayout(pd.T)
			if err != nil {
				return err
			}
			objectID := len(f.Objects)
			objectMap[pd.Sym] = objectID
			f.Objects = append(f.Objects, bytecode.LocalObject{ID: objectID, Name: pd.Sym.Name, Size: layout.Size, Align: layout.Align, Layout: layout.ID})
		}
	}
	for _, local := range fn.Locals {
		if local == nil || local.Sym == nil || local.Storage == sema.StorageStatic || local.Storage == sema.StorageExtern {
			continue
		}
		if local.Sym.SlotID > maxSlot {
			maxSlot = local.Sym.SlotID
		}
		if isVLAType(local.T) {
			continue
		}
		vt, err := g.lowerValueType(local.T)
		if err != nil {
			return err
		}
		if local.Sym.SlotID >= 0 && isSlotType(vt) && !addressTaken[local.Sym] && !isVolatile(local.T) {
			if seenSlots[local.Sym.SlotID] {
				continue
			}
			seenSlots[local.Sym.SlotID] = true
			f.Locals = append(f.Locals, bytecode.LocalSlot{ID: local.Sym.SlotID, Name: local.Sym.Name, Type: vt})
			continue
		}
		layout, err := g.lowerLayout(local.T)
		if err != nil {
			return err
		}
		objectID := len(f.Objects)
		objectMap[local.Sym] = objectID
		f.Objects = append(f.Objects, bytecode.LocalObject{ID: objectID, Name: local.Sym.Name, Size: layout.Size, Align: layout.Align, Layout: layout.ID})
	}
	fg := &funcGen{
		g:                     g,
		fn:                    fn,
		out:                   &f,
		objectMap:             objectMap,
		dynamicObjectMap:      map[*sema.Symbol]int{},
		dynamicSizeSlotMap:    map[*sema.Symbol]int{},
		dynamicSizeSymbolMap:  map[*sema.Symbol]map[string]int{},
		dynamicPointerTypeMap: map[*sema.Symbol]map[string]int{},
		nextSyntheticSlot:     maxSlot + 1,
		addressTaken:          addressTaken,
		namedBreaks:           map[string][]int{},
		namedContinues:        map[string][]int{},
		namedBreakCleanups:    map[string][]int{},
		namedContinueCleanups: map[string][]int{},
		labels:                map[*sema.LabeledStmt]int{},
		labelCleanupMarks:     labelCleanupMarks(fn.Body),
		caseLabels:            map[*sema.CaseStmt]int{},
		defaultLabels:         map[*sema.DefaultStmt]int{},
	}
	for _, p := range fn.Params {
		objectID, ok := objectMap[p.Sym]
		if !ok {
			continue
		}
		vt, err := g.lowerValueType(p.T)
		if err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs,
			bytecode.AddrLocalObject(objectID),
			bytecode.LoadLocal(vt, p.Sym.SlotID),
			bytecode.Store(vt, g.alignof(p.T), isVolatile(p.T)),
		)
	}
	for _, p := range fn.Params {
		if p != nil && p.T != nil && typeHasVariablyModifiedType(p.T) {
			if err := fg.prepareDynamicSizeTypesForSymbol(p.Sym, p.T, p.Sym.Name+"$size"); err != nil {
				return err
			}
		}
	}
	if err := fg.emitStmt(fn.Body); err != nil {
		return err
	}
	if err := fg.emitImplicitTerminal(); err != nil {
		return err
	}
	g.mod.Functions = append(g.mod.Functions, *fg.out)
	if fn.Sym.GlobalID >= 0 && fn.Sym.GlobalID < len(g.mod.Globals) {
		g.mod.Globals[fn.Sym.GlobalID].Func = f.ID
		g.mod.Globals[fn.Sym.GlobalID].Kind = bytecode.GlobalFunc
		g.mod.Globals[fn.Sym.GlobalID].Sig = sig
		g.mod.Globals[fn.Sym.GlobalID].Extern = bytecode.ExternRef{}
	}
	return nil
}

func (fg *funcGen) emitImplicitTerminal() error {
	if len(fg.out.Instrs) > 0 && isTerminalInstr(fg.out.Instrs[len(fg.out.Instrs)-1]) {
		return nil
	}
	ret := fg.g.mod.Sigs[fg.out.Sig].Ret
	if ret == bytecode.TypeVoid {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpReturnVoid})
		return nil
	}
	if fg.out.Name == "main" {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(ret, 0), bytecode.Return(ret))
		return nil
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpUnreachable})
	return nil
}

func isTerminalInstr(ins bytecode.Instr) bool {
	switch ins.Op {
	case bytecode.OpReturn, bytecode.OpReturnVoid, bytecode.OpReturnObject, bytecode.OpJump, bytecode.OpUnreachable:
		return true
	default:
		return false
	}
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
