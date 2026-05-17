package codegen

import (
	"encoding/binary"
	"fmt"
	"math"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

type address struct {
	emit      func() error
	bit       bool
	layout    int
	field     int
	valueType bytecode.ValueType
	volatile  bool
}

type initLeaf struct {
	typ         sema.Type
	designators []sema.Designator
}

type initSpan struct {
	typ         sema.Type
	designators []sema.Designator
	start       int
	end         int
}

func (g *generator) internString(value string) int {
	bytes := append([]byte(value), 0)
	return g.internStringBytes(value, bytes)
}

func (g *generator) internStringLit(lit *sema.StringLit) int {
	if lit == nil {
		return g.internString("")
	}
	bytes := append([]byte(nil), []byte(lit.Value)...)
	bytes = append(bytes, 0)
	if at, ok := sema.Unqual(lit.T).(*sema.ArrayType); ok {
		elemSize := g.sizeof(at.Elem)
		if elemSize > 1 {
			bytes = make([]byte, int64(len(lit.Value)+1)*elemSize)
			for i := 0; i < len(lit.Value); i++ {
				writeStringElement(bytes, int64(i)*elemSize, elemSize, int64(lit.Value[i]))
			}
		}
	}
	return g.internStringBytes(lit.Value, bytes)
}

func (g *generator) internStringBytes(value string, bytes []byte) int {
	key := fmt.Sprintf("%s/%x", value, bytes)
	if id, ok := g.stringMap[key]; ok {
		return id
	}
	id := len(g.mod.Strings)
	g.mod.Strings = append(g.mod.Strings, bytecode.StringConst{ID: id, Value: value, Bytes: bytes})
	g.stringMap[key] = id
	return id
}

func writeStringElement(buf []byte, offset int64, size int64, value int64) {
	switch size {
	case 1:
		buf[offset] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(buf[offset:offset+8], uint64(value))
	default:
		for i := int64(0); i < size; i++ {
			buf[offset+i] = byte(uint64(value) >> uint(8*i))
		}
	}
}

func (g *generator) emitStaticInitializers() error {
	for _, d := range g.prog.Globals {
		vd, ok := d.(*sema.VarDecl)
		if !ok || vd.Sym == nil || vd.Sym.GlobalID < 0 || vd.Storage == sema.StorageExtern {
			continue
		}
		if err := g.emitStaticVarInitializer(vd); err != nil {
			return err
		}
	}
	for _, fn := range g.prog.Funcs {
		for _, vd := range fn.Locals {
			if vd == nil || vd.Sym == nil || vd.Storage != sema.StorageStatic {
				continue
			}
			if err := g.emitStaticVarInitializer(vd); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) emitStaticVarInitializer(vd *sema.VarDecl) error {
	if vd.Sym.GlobalID < 0 || vd.Sym.GlobalID >= len(g.mod.Globals) {
		return fmt.Errorf("static initializer for %q references invalid global %d", vd.Sym.Name, vd.Sym.GlobalID)
	}
	globalID := vd.Sym.GlobalID
	g.mod.Globals[globalID].Readonly = isConst(vd.T)
	if vd.Init == nil {
		g.mod.Globals[globalID].Init = bytecode.InitData{ZeroFill: g.mod.Globals[globalID].Size}
		return nil
	}
	init, err := g.emitStaticInitializer(vd.Init, vd.T)
	if err != nil {
		return fmt.Errorf("static initializer for %q: %w", vd.Sym.Name, err)
	}
	g.mod.Globals[globalID].Init = init
	return nil
}

func (g *generator) emitStaticInitializer(init sema.Expr, typ sema.Type) (bytecode.InitData, error) {
	size := g.sizeof(typ)
	if size < 0 {
		return bytecode.InitData{}, fmt.Errorf("negative initializer size for %s", typ)
	}
	buf := make([]byte, size)
	relocs := []bytecode.Relocation{}
	if init != nil {
		if err := g.writeStaticInitializer(buf, &relocs, 0, init, typ); err != nil {
			return bytecode.InitData{}, err
		}
	}
	return bytecode.InitData{Bytes: buf, Relocations: relocs}, nil
}

func (g *generator) writeStaticInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ sema.Type) error {
	if init == nil {
		return nil
	}
	if sl, ok := stringLiteralInitializer(init); ok {
		if _, isArray := sema.Unqual(typ).(*sema.ArrayType); isArray {
			return g.writeStaticStringArray(buf, offset, sl, typ)
		}
	}
	switch x := sema.Unqual(typ).(type) {
	case *sema.ArrayType:
		return g.writeStaticArrayInitializer(buf, relocs, offset, init, x)
	case *sema.StructType:
		return g.writeStaticStructInitializer(buf, relocs, offset, init, x)
	case *sema.UnionType:
		return g.writeStaticUnionInitializer(buf, relocs, offset, init, x)
	default:
		return g.writeStaticScalarInitializer(buf, relocs, offset, init, typ)
	}
}

func (g *generator) writeStaticArrayInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ *sema.ArrayType) error {
	if sl, ok := stringLiteralInitializer(init); ok {
		return g.writeStaticStringArray(buf, offset, sl, typ)
	}
	if _, ok := init.(*sema.InitList); !ok {
		return g.writeStaticScalarInitializer(buf, relocs, offset, init, typ)
	}
	return g.writeStaticAggregateInitializer(buf, relocs, offset, init, typ)
}

func (g *generator) writeStaticStructInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ *sema.StructType) error {
	if _, ok := init.(*sema.InitList); !ok {
		return g.writeStaticScalarInitializer(buf, relocs, offset, init, typ)
	}
	return g.writeStaticAggregateInitializer(buf, relocs, offset, init, typ)
}

func (g *generator) writeStaticUnionInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ *sema.UnionType) error {
	if _, ok := init.(*sema.InitList); !ok {
		return g.writeStaticScalarInitializer(buf, relocs, offset, init, typ)
	}
	return g.writeStaticAggregateInitializer(buf, relocs, offset, init, typ)
}

func (g *generator) writeStaticAggregateInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ sema.Type) error {
	il, ok := init.(*sema.InitList)
	if !ok {
		return g.writeStaticScalarInitializer(buf, relocs, offset, init, typ)
	}
	leaves := g.initLeaves(typ, nil)
	spans := g.directInitSpans(typ, nil)
	cursor := 0
	for _, elem := range il.Elems {
		if len(elem.Designators) > 0 {
			next, err := g.writeStaticDesignatedElem(buf, relocs, offset, typ, elem.Designators, elem.Value)
			if err != nil {
				return err
			}
			cursor = next
			continue
		}
		if cursor >= len(leaves) {
			continue
		}
		span := spanContaining(spans, cursor)
		if span != nil && cursor == span.start && useWholeInitializer(elem.Value, span.typ) {
			if err := g.writeStaticDesignatedInitializer(buf, relocs, offset, typ, span.designators, elem.Value); err != nil {
				return err
			}
			cursor = span.end
			continue
		}
		leaf := leaves[cursor]
		if err := g.writeStaticDesignatedInitializer(buf, relocs, offset, typ, leaf.designators, elem.Value); err != nil {
			return err
		}
		cursor++
	}
	return nil
}

func (g *generator) writeStaticDesignatedElem(buf []byte, relocs *[]bytecode.Relocation, offset int64, typ sema.Type, ds []sema.Designator, init sema.Expr) (int, error) {
	span, err := g.designatedSpan(typ, ds)
	if err != nil {
		return 0, err
	}
	if !isObjectType(span.typ) || useWholeInitializer(init, span.typ) {
		if err := g.zeroStaticDesignatedObject(buf, offset, typ, ds); err != nil {
			return 0, err
		}
		if err := g.writeStaticDesignatedInitializer(buf, relocs, offset, typ, ds, init); err != nil {
			return 0, err
		}
		return span.end, nil
	}
	leaves := g.initLeaves(span.typ, span.designators)
	if len(leaves) == 0 {
		return 0, fmt.Errorf("designator does not name an initializable subobject")
	}
	leaf := leaves[0]
	if err := g.writeStaticDesignatedInitializer(buf, relocs, offset, typ, leaf.designators, init); err != nil {
		return 0, err
	}
	return span.start + 1, nil
}

func (g *generator) zeroStaticDesignatedObject(buf []byte, offset int64, typ sema.Type, ds []sema.Designator) error {
	cur := typ
	base := offset
	for _, d := range ds {
		switch d.Kind {
		case sema.DesigArrayIndex:
			at, ok := sema.Unqual(cur).(*sema.ArrayType)
			if !ok {
				return fmt.Errorf("array designator applied to %s", cur)
			}
			base += d.Index * g.sizeof(at.Elem)
			cur = at.Elem
		case sema.DesigFieldName:
			field := d.Field
			if field == nil {
				field = lookupField(cur, "")
			}
			if field == nil {
				return fmt.Errorf("field designator has no resolved field")
			}
			if field.IsBitField {
				return nil
			}
			base += field.Offset
			cur = field.T
		default:
			return fmt.Errorf("unsupported designator kind %d", d.Kind)
		}
	}
	size := g.sizeof(cur)
	if err := checkStaticRange(buf, base, size); err != nil {
		return err
	}
	for i := int64(0); i < size; i++ {
		buf[base+i] = 0
	}
	return nil
}

func (g *generator) writeStaticDesignatedInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, typ sema.Type, ds []sema.Designator, init sema.Expr) error {
	if len(ds) == 0 {
		return g.writeStaticInitializer(buf, relocs, offset, init, typ)
	}
	d := ds[0]
	switch d.Kind {
	case sema.DesigArrayIndex:
		at, ok := sema.Unqual(typ).(*sema.ArrayType)
		if !ok {
			return fmt.Errorf("array designator applied to %s", typ)
		}
		return g.writeStaticDesignatedInitializer(buf, relocs, offset+d.Index*g.sizeof(at.Elem), at.Elem, ds[1:], init)
	case sema.DesigFieldName:
		field := d.Field
		if field == nil {
			field = lookupField(typ, "")
		}
		if field == nil {
			return fmt.Errorf("field designator has no resolved field")
		}
		if len(ds) == 1 {
			return g.writeStaticFieldInitializer(buf, relocs, offset, typ, field, init)
		}
		if field.IsBitField {
			return fmt.Errorf("bit-field designator %q cannot contain nested designators", field.Name)
		}
		return g.writeStaticDesignatedInitializer(buf, relocs, offset+field.Offset, field.T, ds[1:], init)
	default:
		return fmt.Errorf("unsupported designator kind %d", d.Kind)
	}
}

func (g *generator) writeStaticFieldInitializer(buf []byte, relocs *[]bytecode.Relocation, base int64, container sema.Type, field *sema.Field, init sema.Expr) error {
	if field == nil {
		return nil
	}
	if field.IsBitField {
		value, ok := staticIntegerValue(init)
		if !ok {
			return fmt.Errorf("bit-field initializer for %q is not an integer constant", field.Name)
		}
		return g.writeStaticBitField(buf, base, container, field, value)
	}
	return g.writeStaticInitializer(buf, relocs, base+field.Offset, init, field.T)
}

func (g *generator) writeStaticScalarInitializer(buf []byte, relocs *[]bytecode.Relocation, offset int64, init sema.Expr, typ sema.Type) error {
	switch x := init.(type) {
	case *sema.InitList:
		if len(x.Elems) == 0 {
			return nil
		}
		return g.writeStaticInitializer(buf, relocs, offset, x.Elems[0].Value, typ)
	case *sema.ImplicitCast:
		if x.Kind == sema.ArrayDecay {
			if sl, ok := stringLiteralInitializer(x.X); ok {
				return g.writeStaticStringPointer(relocs, offset, sl)
			}
			if cl := staticCompoundLiteralAddressOperand(x.X); cl != nil {
				return g.writeStaticCompoundLiteralPointer(relocs, offset, cl)
			}
		}
		return g.writeStaticScalarInitializer(buf, relocs, offset, x.X, typ)
	case *sema.ExplicitCast:
		return g.writeStaticScalarInitializer(buf, relocs, offset, x.X, typ)
	case *sema.UnOp:
		if x.Op == sema.UnAddr {
			if cl, ok := x.X.(*sema.CompoundLit); ok {
				return g.writeStaticCompoundLiteralPointer(relocs, offset, cl)
			}
		}
		if isComplexType(typ) {
			if cv, ok := sema.NewEvaluator(nil).EvalConstant(x); ok && cv.Kind == sema.ConstComplex {
				return g.writeStaticComplexConstant(buf, offset, cv, typ)
			}
		}
	case *sema.CompoundLit:
		return g.writeStaticInitializer(buf, relocs, offset, x.Init, typ)
	case *sema.CallExpr:
		if ok, err := g.writeStaticBuiltinComplex(buf, offset, x, typ); ok || err != nil {
			return err
		}
	case *sema.BinOp:
		if isComplexType(typ) {
			if cv, ok := sema.NewEvaluator(nil).EvalConstant(x); ok && cv.Kind == sema.ConstComplex {
				return g.writeStaticComplexConstant(buf, offset, cv, typ)
			}
		}
	case *sema.CondExpr:
		if isComplexType(typ) {
			if cv, ok := sema.NewEvaluator(nil).EvalConstant(x); ok && cv.Kind == sema.ConstComplex {
				return g.writeStaticComplexConstant(buf, offset, cv, typ)
			}
		}
	case *sema.ImagLit:
		if isComplexType(typ) {
			return g.writeStaticImaginaryComplex(buf, offset, x, typ)
		}
		vt, err := g.lowerValueType(typ)
		if err != nil {
			return err
		}
		if isFloatType(vt) {
			return g.writeStaticFloat(buf, offset, typ, 0)
		}
		if isIntegerType(vt) {
			return g.writeStaticInteger(buf, offset, typ, 0)
		}
	case *sema.IntLit:
		return g.writeStaticInteger(buf, offset, typ, x.Value)
	case *sema.CharLit:
		return g.writeStaticInteger(buf, offset, typ, int64(x.Value))
	case *sema.EnumRef:
		if x.Enumerator == nil {
			return g.writeStaticInteger(buf, offset, typ, 0)
		}
		return g.writeStaticInteger(buf, offset, typ, x.Enumerator.Value)
	case *sema.FloatLit:
		return g.writeStaticFloat(buf, offset, typ, x.Value)
	case *sema.AddrConst:
		return g.writeStaticAddress(relocs, offset, x)
	case *sema.StringLit:
		if _, ok := sema.Unqual(typ).(*sema.PointerType); ok {
			return g.writeStaticStringPointer(relocs, offset, x)
		}
	}
	return fmt.Errorf("static initializer lowering is not implemented for %T into %s", init, typ)
}

func (g *generator) writeStaticBuiltinComplex(buf []byte, offset int64, call *sema.CallExpr, typ sema.Type) (bool, error) {
	if call == nil || !isComplexType(typ) {
		return false, nil
	}
	vr := functionVarRef(call.Callee)
	if vr == nil || vr.Sym == nil || vr.Sym.Name != "__builtin_complex" {
		return false, nil
	}
	if len(call.Args) != 2 {
		return true, fmt.Errorf("__builtin_complex initializer has %d args, want 2", len(call.Args))
	}
	realType, err := complexRealType(typ)
	if err != nil {
		return true, err
	}
	for i, arg := range call.Args {
		value, ok := staticFloatingValue(arg)
		if !ok {
			return true, fmt.Errorf("__builtin_complex argument %d is not a static arithmetic constant", i)
		}
		if err := g.writeStaticFloat(buf, offset+int64(i)*g.sizeof(realType), realType, value); err != nil {
			return true, err
		}
	}
	return true, nil
}

func (g *generator) writeStaticImaginaryComplex(buf []byte, offset int64, lit *sema.ImagLit, typ sema.Type) error {
	return g.writeStaticComplexParts(buf, offset, typ, 0, lit.Value)
}

func (g *generator) writeStaticComplexConstant(buf []byte, offset int64, cv sema.ConstValue, typ sema.Type) error {
	return g.writeStaticComplexParts(buf, offset, typ, cv.Float, cv.Imag)
}

func (g *generator) writeStaticComplexParts(buf []byte, offset int64, typ sema.Type, realValue, imagValue float64) error {
	realType, err := complexRealType(typ)
	if err != nil {
		return err
	}
	if err := g.writeStaticFloat(buf, offset, realType, realValue); err != nil {
		return err
	}
	return g.writeStaticFloat(buf, offset+g.sizeof(realType), realType, imagValue)
}

func staticCompoundLiteralAddressOperand(e sema.Expr) *sema.CompoundLit {
	for {
		ic, ok := e.(*sema.ImplicitCast)
		if !ok || ic.Kind != sema.LValueToRValue {
			break
		}
		e = ic.X
	}
	if cl, ok := e.(*sema.CompoundLit); ok {
		return cl
	}
	return nil
}

func (g *generator) writeStaticCompoundLiteralPointer(relocs *[]bytecode.Relocation, offset int64, cl *sema.CompoundLit) error {
	id, err := g.ensureStaticCompoundLiteral(cl)
	if err != nil {
		return err
	}
	*relocs = append(*relocs, bytecode.Relocation{Offset: offset, Kind: bytecode.RelocGlobal, Target: id})
	return nil
}

func (g *generator) ensureStaticCompoundLiteral(cl *sema.CompoundLit) (int, error) {
	if id, ok := g.staticCompoundLiteralMap[cl]; ok {
		return id, nil
	}
	id := len(g.mod.Globals)
	g.staticCompoundLiteralMap[cl] = id
	global := bytecode.Global{
		ID:       id,
		Name:     fmt.Sprintf(".compound.%d", id),
		Kind:     bytecode.GlobalVar,
		Func:     -1,
		Sig:      bytecode.NoFuncSig,
		Size:     g.sizeof(cl.T),
		Align:    g.alignof(cl.T),
		Readonly: isConst(cl.T),
	}
	global.Init = bytecode.InitData{ZeroFill: global.Size}
	g.mod.Globals = append(g.mod.Globals, global)
	if _, err := g.lowerLayout(cl.T); err != nil {
		return 0, err
	}
	init, err := g.emitStaticInitializer(cl.Init, cl.T)
	if err != nil {
		return 0, err
	}
	g.mod.Globals[id].Init = init
	return id, nil
}

func (g *generator) writeStaticStringArray(buf []byte, offset int64, lit *sema.StringLit, typ sema.Type) error {
	at, ok := sema.Unqual(typ).(*sema.ArrayType)
	if !ok {
		return fmt.Errorf("string array initializer applied to %s", typ)
	}
	bytes := append([]byte(lit.Value), 0)
	count := int64(len(bytes))
	if at.SizeKind == sema.ArrayConstantSize && at.Size < count {
		count = at.Size
	}
	elemSize := g.sizeof(at.Elem)
	for i := int64(0); i < count; i++ {
		if err := g.writeStaticInteger(buf, offset+i*elemSize, at.Elem, int64(bytes[i])); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeStaticStringPointer(relocs *[]bytecode.Relocation, offset int64, lit *sema.StringLit) error {
	*relocs = append(*relocs, bytecode.Relocation{Offset: offset, Kind: bytecode.RelocString, Target: g.internStringLit(lit)})
	return nil
}

func (g *generator) writeStaticAddress(relocs *[]bytecode.Relocation, offset int64, init *sema.AddrConst) error {
	if init.Sym == nil {
		return fmt.Errorf("address initializer has no symbol")
	}
	kind := bytecode.RelocGlobal
	if init.Sym.Kind == sema.SymFunc {
		kind = bytecode.RelocFunc
	}
	if init.Sym.GlobalID < 0 {
		return fmt.Errorf("address initializer references %q without global id", init.Sym.Name)
	}
	*relocs = append(*relocs, bytecode.Relocation{Offset: offset, Kind: kind, Target: init.Sym.GlobalID, Addend: init.Offset})
	return nil
}

func (g *generator) writeStaticInteger(buf []byte, offset int64, typ sema.Type, value int64) error {
	size := g.sizeof(typ)
	if err := checkStaticRange(buf, offset, size); err != nil {
		return err
	}
	u := uint64(value)
	switch size {
	case 1:
		buf[offset] = byte(u)
	case 2:
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(u))
	case 4:
		binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(u))
	case 8:
		binary.LittleEndian.PutUint64(buf[offset:offset+8], u)
	default:
		for i := int64(0); i < size; i++ {
			buf[offset+i] = byte(u >> uint(8*i))
		}
	}
	return nil
}

func (g *generator) writeStaticFloat(buf []byte, offset int64, typ sema.Type, value float64) error {
	size := g.sizeof(typ)
	if err := checkStaticRange(buf, offset, size); err != nil {
		return err
	}
	vt, err := g.lowerValueType(typ)
	if err != nil {
		return err
	}
	switch vt {
	case bytecode.TypeF32:
		binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(float32(value)))
	case bytecode.TypeF64:
		binary.LittleEndian.PutUint64(buf[offset:offset+8], math.Float64bits(value))
	case bytecode.TypeFLong:
		binary.LittleEndian.PutUint64(buf[offset:offset+8], math.Float64bits(value))
	default:
		return fmt.Errorf("cannot write floating initializer into %s", typ)
	}
	return nil
}

func (g *generator) writeStaticBitField(buf []byte, base int64, container sema.Type, field *sema.Field, value int64) error {
	layout, err := g.lowerLayout(container)
	if err != nil {
		return err
	}
	fieldID, err := g.bitFieldID(layout.ID, field)
	if err != nil {
		return err
	}
	bf := layout.Bit[fieldID]
	size := int64(typeSize(bf.Container))
	if err := checkStaticRange(buf, base+bf.ByteOffset, size); err != nil {
		return err
	}
	var cur uint64
	for i := int64(0); i < size; i++ {
		cur |= uint64(buf[base+bf.ByteOffset+i]) << uint(8*i)
	}
	mask := bitMask(bf.Width) << uint(bf.BitOffset)
	next := (cur &^ mask) | ((uint64(value) << uint(bf.BitOffset)) & mask)
	for i := int64(0); i < size; i++ {
		buf[base+bf.ByteOffset+i] = byte(next >> uint(8*i))
	}
	return nil
}

func (fg *funcGen) emitInitializer(dst address, init sema.Expr, typ sema.Type) error {
	if init == nil {
		return fg.emitZeroInitializer(dst, typ)
	}
	if cond, ok := init.(*sema.CondExpr); ok {
		return fg.emitConditionalInitializer(dst, cond, typ)
	}
	if sl, ok := stringLiteralInitializer(init); ok {
		if _, isArray := sema.Unqual(typ).(*sema.ArrayType); isArray {
			return fg.emitStringArrayInitializer(dst, sl, typ)
		}
	}
	switch x := sema.Unqual(typ).(type) {
	case *sema.ArrayType:
		return fg.emitArrayInitializer(dst, init, x)
	case *sema.StructType:
		return fg.emitStructInitializer(dst, init, x)
	case *sema.UnionType:
		return fg.emitUnionInitializer(dst, init, x)
	default:
		return fg.emitScalarInitializer(dst, init, typ)
	}
}

func (fg *funcGen) emitConditionalInitializer(dst address, init *sema.CondExpr, typ sema.Type) error {
	elseLabel := fg.newLabel(true, nil)
	endLabel := fg.newLabel(false, nil)
	if err := fg.emitBoolValue(init.Cond); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, elseLabel))
	if err := fg.emitInitializer(dst, init.Then, typ); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(endLabel))
	fg.mark(elseLabel)
	if err := fg.emitInitializer(dst, init.Else, typ); err != nil {
		return err
	}
	fg.mark(endLabel)
	return nil
}

func (fg *funcGen) emitArrayInitializer(dst address, init sema.Expr, typ *sema.ArrayType) error {
	if sl, ok := stringLiteralInitializer(init); ok {
		return fg.emitStringArrayInitializer(dst, sl, typ)
	}
	if _, ok := init.(*sema.InitList); !ok {
		return fg.emitObjectCopyInitializer(dst, init, typ)
	}
	if err := fg.emitZeroInitializer(dst, typ); err != nil {
		return err
	}
	return fg.emitAggregateInitializer(dst, init, typ)
}

func (fg *funcGen) emitStructInitializer(dst address, init sema.Expr, typ *sema.StructType) error {
	if _, ok := init.(*sema.InitList); !ok {
		return fg.emitObjectCopyInitializer(dst, init, typ)
	}
	if err := fg.emitZeroInitializer(dst, typ); err != nil {
		return err
	}
	return fg.emitAggregateInitializer(dst, init, typ)
}

func (fg *funcGen) emitUnionInitializer(dst address, init sema.Expr, typ *sema.UnionType) error {
	if _, ok := init.(*sema.InitList); !ok {
		return fg.emitObjectCopyInitializer(dst, init, typ)
	}
	if err := fg.emitZeroInitializer(dst, typ); err != nil {
		return err
	}
	return fg.emitAggregateInitializer(dst, init, typ)
}

func (fg *funcGen) emitAggregateInitializer(dst address, init sema.Expr, typ sema.Type) error {
	il, ok := init.(*sema.InitList)
	if !ok {
		return fg.emitObjectCopyInitializer(dst, init, typ)
	}
	leaves := fg.g.initLeaves(typ, nil)
	spans := fg.g.directInitSpans(typ, nil)
	cursor := 0
	for _, elem := range il.Elems {
		if len(elem.Designators) > 0 {
			next, err := fg.emitDesignatedElem(dst, typ, elem.Designators, elem.Value)
			if err != nil {
				return err
			}
			cursor = next
			continue
		}
		if cursor >= len(leaves) {
			continue
		}
		span := spanContaining(spans, cursor)
		if span != nil && cursor == span.start && useWholeInitializer(elem.Value, span.typ) {
			if err := fg.emitDesignatedInitializer(dst, typ, span.designators, elem.Value); err != nil {
				return err
			}
			cursor = span.end
			continue
		}
		leaf := leaves[cursor]
		if err := fg.emitDesignatedInitializer(dst, typ, leaf.designators, elem.Value); err != nil {
			return err
		}
		cursor++
	}
	return nil
}

func (fg *funcGen) emitDesignatedElem(dst address, typ sema.Type, ds []sema.Designator, init sema.Expr) (int, error) {
	span, err := fg.g.designatedSpan(typ, ds)
	if err != nil {
		return 0, err
	}
	if !isObjectType(span.typ) || useWholeInitializer(init, span.typ) {
		if err := fg.emitDesignatedInitializer(dst, typ, ds, init); err != nil {
			return 0, err
		}
		return span.end, nil
	}
	leaves := fg.g.initLeaves(span.typ, span.designators)
	if len(leaves) == 0 {
		return 0, fmt.Errorf("designator does not name an initializable subobject")
	}
	leaf := leaves[0]
	if err := fg.emitDesignatedInitializer(dst, typ, leaf.designators, init); err != nil {
		return 0, err
	}
	return span.start + 1, nil
}

func (fg *funcGen) emitDesignatedInitializer(dst address, typ sema.Type, ds []sema.Designator, init sema.Expr) error {
	sub, subType, err := fg.designatedAddress(dst, typ, ds)
	if err != nil {
		return err
	}
	return fg.emitInitializer(sub, init, subType)
}

func (fg *funcGen) designatedAddress(dst address, typ sema.Type, ds []sema.Designator) (address, sema.Type, error) {
	if len(ds) == 0 {
		return dst, typ, nil
	}
	d := ds[0]
	switch d.Kind {
	case sema.DesigArrayIndex:
		at, ok := sema.Unqual(typ).(*sema.ArrayType)
		if !ok {
			return address{}, nil, fmt.Errorf("array designator applied to %s", typ)
		}
		return fg.designatedAddress(fg.offsetAddress(dst, d.Index*fg.g.sizeof(at.Elem)), at.Elem, ds[1:])
	case sema.DesigFieldName:
		field := d.Field
		if field == nil {
			field = lookupField(typ, "")
		}
		if field == nil {
			return address{}, nil, fmt.Errorf("field designator has no resolved field")
		}
		fieldAddr, err := fg.fieldAddress(dst, typ, field)
		if err != nil {
			return address{}, nil, err
		}
		return fg.designatedAddress(fieldAddr, field.T, ds[1:])
	default:
		return address{}, nil, fmt.Errorf("unsupported designator kind %d", d.Kind)
	}
}

func (fg *funcGen) emitScalarInitializer(dst address, init sema.Expr, typ sema.Type) error {
	if il, ok := init.(*sema.InitList); ok {
		if len(il.Elems) == 0 {
			return fg.emitZeroInitializer(dst, typ)
		}
		return fg.emitScalarInitializer(dst, il.Elems[0].Value, typ)
	}
	if dst.bit {
		if err := dst.emit(); err != nil {
			return err
		}
		if err := fg.emitValue(init); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpBitFieldStore, Type: dst.valueType, Layout: dst.layout, Field: dst.field, Volatile: dst.volatile})
		return nil
	}
	vt, err := fg.g.lowerValueType(typ)
	if err != nil {
		return err
	}
	if vt == bytecode.TypeObjectAddr {
		if isComplexType(typ) {
			return fg.emitComplexInitializer(dst, init, typ)
		}
		return fg.emitObjectCopyInitializer(dst, init, typ)
	}
	if err := dst.emit(); err != nil {
		return err
	}
	if err := fg.emitValue(init); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(vt, fg.g.alignof(typ), isVolatile(typ)))
	return nil
}

func (fg *funcGen) emitComplexInitializer(dst address, init sema.Expr, typ sema.Type) error {
	if il, ok := init.(*sema.InitList); ok {
		if len(il.Elems) == 0 {
			return fg.emitZeroInitializer(dst, typ)
		}
		return fg.emitComplexInitializer(dst, il.Elems[0].Value, typ)
	}
	if call := builtinComplexCall(init); call != nil {
		return fg.emitBuiltinComplexInitializer(dst, call, typ)
	}
	if lit, ok := init.(*sema.ImagLit); ok {
		return fg.emitImaginaryComplexInitializer(dst, lit, typ)
	}
	if realInit := complexRealInitializer(init); realInit != nil {
		savedDst, err := fg.saveAddress(dst, ".complex.dst")
		if err != nil {
			return err
		}
		if err := fg.emitZeroInitializer(savedDst, typ); err != nil {
			return err
		}
		realType, err := complexRealType(typ)
		if err != nil {
			return err
		}
		realVT, err := fg.g.lowerValueType(realType)
		if err != nil {
			return err
		}
		if err := savedDst.emit(); err != nil {
			return err
		}
		if err := fg.emitValue(realInit); err != nil {
			return err
		}
		from, err := fg.g.lowerValueType(realInit.GetType())
		if err != nil {
			return err
		}
		fg.emitCast(from, realVT, sema.UsualArithmetic)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(realVT, fg.g.alignof(realType), isVolatile(typ)))
		return nil
	}
	return fg.emitComplexValueCopy(dst, init, typ)
}

func (fg *funcGen) emitImaginaryComplexInitializer(dst address, lit *sema.ImagLit, typ sema.Type) error {
	realType, err := complexRealType(typ)
	if err != nil {
		return err
	}
	realVT, err := fg.g.lowerValueType(realType)
	if err != nil {
		return err
	}
	savedDst, err := fg.saveAddress(dst, ".complex.dst")
	if err != nil {
		return err
	}
	if err := fg.emitZeroInitializer(savedDst, typ); err != nil {
		return err
	}
	if err := fg.offsetAddress(savedDst, fg.g.sizeof(realType)).emit(); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpConst, Type: realVT, Float: lit.Value})
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(realVT, fg.g.alignof(realType), isVolatile(typ)))
	return nil
}

func builtinComplexCall(init sema.Expr) *sema.CallExpr {
	if ic, ok := init.(*sema.ImplicitCast); ok && isComplexType(ic.From) && isComplexType(ic.To) {
		init = ic.X
	}
	call, ok := init.(*sema.CallExpr)
	if !ok {
		return nil
	}
	vr := functionVarRef(call.Callee)
	if vr == nil || vr.Sym == nil || vr.Sym.Name != "__builtin_complex" {
		return nil
	}
	return call
}

func (fg *funcGen) emitBuiltinComplexInitializer(dst address, call *sema.CallExpr, typ sema.Type) error {
	if len(call.Args) != 2 {
		return fmt.Errorf("__builtin_complex initializer has %d args, want 2", len(call.Args))
	}
	realType, err := complexRealType(typ)
	if err != nil {
		return err
	}
	realVT, err := fg.g.lowerValueType(realType)
	if err != nil {
		return err
	}
	savedDst, err := fg.saveAddress(dst, ".complex.dst")
	if err != nil {
		return err
	}
	for i, arg := range call.Args {
		offset := int64(0)
		if i == 1 {
			offset = fg.g.sizeof(realType)
		}
		if err := fg.offsetAddress(savedDst, offset).emit(); err != nil {
			return err
		}
		if err := fg.emitValue(arg); err != nil {
			return err
		}
		from, err := fg.g.lowerValueType(arg.GetType())
		if err != nil {
			return err
		}
		fg.emitCast(from, realVT, sema.UsualArithmetic)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(realVT, fg.g.alignof(realType), isVolatile(typ)))
	}
	return nil
}

func complexRealInitializer(init sema.Expr) sema.Expr {
	if init == nil {
		return nil
	}
	if ic, ok := init.(*sema.ImplicitCast); ok && isComplexType(ic.To) && !isComplexType(ic.From) {
		return ic.X
	}
	if ec, ok := init.(*sema.ExplicitCast); ok && isComplexType(ec.To) && !isComplexType(ec.X.GetType()) {
		return ec.X
	}
	if !isComplexType(init.GetType()) {
		return init
	}
	return nil
}

func (fg *funcGen) emitComplexValueCopy(dst address, src sema.Expr, dstType sema.Type) error {
	if ic, ok := src.(*sema.ImplicitCast); ok && isComplexType(ic.From) && isComplexType(ic.To) {
		src = ic.X
	}
	if ec, ok := src.(*sema.ExplicitCast); ok && isComplexType(ec.X.GetType()) && isComplexType(ec.To) {
		src = ec.X
	}
	srcRealType, err := complexRealType(src.GetType())
	if err != nil {
		return err
	}
	dstRealType, err := complexRealType(dstType)
	if err != nil {
		return err
	}
	srcVT, err := fg.g.lowerValueType(srcRealType)
	if err != nil {
		return err
	}
	dstVT, err := fg.g.lowerValueType(dstRealType)
	if err != nil {
		return err
	}
	savedDst, err := fg.saveAddress(dst, ".complex.dst")
	if err != nil {
		return err
	}
	srcAddrSlot := fg.allocSyntheticSlot(".complex.src", bytecode.TypeObjectAddr)
	if err := fg.emitComplexSourceAddress(src); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.StoreLocal(bytecode.TypeObjectAddr, srcAddrSlot))
	for _, offset := range []int64{0, fg.g.sizeof(srcRealType)} {
		dstOffset := int64(0)
		if offset != 0 {
			dstOffset = fg.g.sizeof(dstRealType)
		}
		if err := fg.offsetAddress(savedDst, dstOffset).emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeObjectAddr, srcAddrSlot))
		if offset != 0 {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypeObjectAddr, Int: offset})
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Load(srcVT, fg.g.alignof(srcRealType), isVolatile(src.GetType())))
		fg.emitCast(srcVT, dstVT, sema.UsualArithmetic)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(dstVT, fg.g.alignof(dstRealType), isVolatile(dstType)))
	}
	return nil
}

func (fg *funcGen) emitComplexSourceAddress(src sema.Expr) error {
	switch x := src.(type) {
	case *sema.BinOp:
		return fg.emitValue(src)
	case *sema.ImagLit:
		return fg.emitComplexRValueAddress(src)
	case *sema.UnOp:
		if isComplexType(x.GetType()) {
			return fg.emitValue(src)
		}
		return fg.emitAddress(src)
	case *sema.ImplicitCast:
		if isComplexType(x.To) {
			return fg.emitComplexRValueAddress(src)
		}
		return fg.emitAddress(src)
	case *sema.ExplicitCast:
		if isComplexType(x.To) {
			return fg.emitComplexRValueAddress(src)
		}
		return fg.emitAddress(src)
	case *sema.CallExpr:
		if builtinComplexCall(x) != nil {
			return fg.emitComplexRValueAddress(src)
		}
		if isComplexType(x.GetType()) {
			return fg.emitValue(src)
		}
		return fg.emitAddress(src)
	default:
		return fg.emitAddress(src)
	}
}

func (fg *funcGen) emitComplexSourceAddressAs(src sema.Expr, targetType sema.Type) error {
	if isComplexType(src.GetType()) {
		return fg.emitComplexSourceAddress(src)
	}
	object, err := fg.newLocalObject(".complex.rvalue", targetType)
	if err != nil {
		return err
	}
	dst := address{emit: func() error {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrLocalObject(object))
		return nil
	}}
	if err := fg.emitComplexInitializer(dst, src, targetType); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrLocalObject(object))
	return nil
}

func (fg *funcGen) emitComplexRValueAddress(src sema.Expr) error {
	object, err := fg.newLocalObject(".complex.rvalue", src.GetType())
	if err != nil {
		return err
	}
	dst := address{emit: func() error {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrLocalObject(object))
		return nil
	}}
	if err := fg.emitComplexInitializer(dst, src, src.GetType()); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.AddrLocalObject(object))
	return nil
}

func (fg *funcGen) saveAddress(src address, name string) (address, error) {
	slot := fg.allocSyntheticSlot(name, bytecode.TypeObjectAddr)
	if err := src.emit(); err != nil {
		return address{}, err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.StoreLocal(bytecode.TypeObjectAddr, slot))
	return address{emit: func() error {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeObjectAddr, slot))
		return nil
	}}, nil
}

func complexRealType(t sema.Type) (sema.Type, error) {
	bt, ok := sema.Unqual(t).(*sema.BuiltinType)
	if !ok {
		return nil, fmt.Errorf("complex real type requested for %s", t)
	}
	switch bt.Kind {
	case sema.FloatComplex:
		return &sema.BuiltinType{Kind: sema.Float}, nil
	case sema.DoubleComplex:
		return &sema.BuiltinType{Kind: sema.Double}, nil
	case sema.LongDoubleComplex:
		return &sema.BuiltinType{Kind: sema.LongDouble}, nil
	default:
		return nil, fmt.Errorf("complex real type requested for %s", t)
	}
}

func (fg *funcGen) emitObjectCopyInitializer(dst address, init sema.Expr, typ sema.Type) error {
	if err := dst.emit(); err != nil {
		return err
	}
	if err := fg.emitAddress(init); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpMemCopy, Size: fg.g.sizeof(typ), Align: fg.g.alignof(typ), Volatile: isVolatile(typ)})
	return nil
}

func (fg *funcGen) emitZeroInitializer(dst address, typ sema.Type) error {
	if dst.bit {
		if err := dst.emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(dst.valueType, 0))
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpBitFieldStore, Type: dst.valueType, Layout: dst.layout, Field: dst.field, Volatile: dst.volatile})
		return nil
	}
	if isObjectType(typ) {
		if err := dst.emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs,
			bytecode.I32Const(0),
			bytecode.Instr{Op: bytecode.OpMemSet, Size: fg.g.sizeof(typ), Align: fg.g.alignof(typ), Volatile: isVolatile(typ)},
		)
		return nil
	}
	vt, err := fg.g.lowerValueType(typ)
	if err != nil {
		return err
	}
	if err := dst.emit(); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(vt, 0), bytecode.Store(vt, fg.g.alignof(typ), isVolatile(typ)))
	return nil
}

func (fg *funcGen) emitStringArrayInitializer(dst address, lit *sema.StringLit, typ sema.Type) error {
	at, ok := sema.Unqual(typ).(*sema.ArrayType)
	if !ok {
		return fmt.Errorf("string array initializer applied to %s", typ)
	}
	if err := fg.emitZeroInitializer(dst, typ); err != nil {
		return err
	}
	bytes := append([]byte(lit.Value), 0)
	count := int64(len(bytes))
	if at.SizeKind == sema.ArrayConstantSize && at.Size < count {
		count = at.Size
	}
	elemSize := fg.g.sizeof(at.Elem)
	for i := int64(0); i < count; i++ {
		elemAddr := fg.offsetAddress(dst, i*elemSize)
		if err := fg.emitIntegerStore(elemAddr, at.Elem, int64(bytes[i])); err != nil {
			return err
		}
	}
	return nil
}

func (fg *funcGen) emitIntegerStore(dst address, typ sema.Type, value int64) error {
	if dst.bit {
		if err := dst.emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(dst.valueType, value))
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpBitFieldStore, Type: dst.valueType, Layout: dst.layout, Field: dst.field, Volatile: dst.volatile})
		return nil
	}
	vt, err := fg.g.lowerValueType(typ)
	if err != nil {
		return err
	}
	if err := dst.emit(); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Const(vt, value), bytecode.Store(vt, fg.g.alignof(typ), isVolatile(typ)))
	return nil
}

func (fg *funcGen) fieldAddress(base address, container sema.Type, field *sema.Field) (address, error) {
	if field == nil {
		return address{}, fmt.Errorf("nil field")
	}
	if field.IsBitField {
		layout, err := fg.g.lowerLayout(container)
		if err != nil {
			return address{}, err
		}
		fieldID, err := fg.g.bitFieldID(layout.ID, field)
		if err != nil {
			return address{}, err
		}
		vt, err := fg.g.lowerValueType(field.T)
		if err != nil {
			return address{}, err
		}
		return address{emit: base.emit, bit: true, layout: layout.ID, field: fieldID, valueType: vt, volatile: layout.Bit[fieldID].Volatile || isVolatile(field.T)}, nil
	}
	return fg.offsetAddress(base, field.Offset), nil
}

func (fg *funcGen) newLocalObject(name string, typ sema.Type) (int, error) {
	layout, err := fg.g.lowerLayout(typ)
	if err != nil {
		return 0, err
	}
	id := len(fg.out.Objects)
	fg.out.Objects = append(fg.out.Objects, bytecode.LocalObject{ID: id, Name: name, Size: layout.Size, Align: layout.Align, Layout: layout.ID})
	return id, nil
}

func (fg *funcGen) offsetAddress(base address, offset int64) address {
	if offset == 0 {
		return base
	}
	return address{emit: func() error {
		if err := base.emit(); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpOffset, Type: bytecode.TypeObjectAddr, Int: offset})
		return nil
	}}
}

func checkStaticRange(buf []byte, offset, size int64) error {
	if offset < 0 || size < 0 || offset+size > int64(len(buf)) {
		return fmt.Errorf("initializer write [%d,%d) outside object size %d", offset, offset+size, len(buf))
	}
	return nil
}

func stringLiteralInitializer(init sema.Expr) (*sema.StringLit, bool) {
	switch x := init.(type) {
	case *sema.StringLit:
		return x, true
	case *sema.ImplicitCast:
		if x.Kind == sema.ArrayDecay || x.Kind == sema.LValueToRValue {
			return stringLiteralInitializer(x.X)
		}
	case *sema.InitList:
		if len(x.Elems) == 1 && len(x.Elems[0].Designators) == 0 {
			return stringLiteralInitializer(x.Elems[0].Value)
		}
	}
	return nil, false
}

func staticIntegerValue(init sema.Expr) (int64, bool) {
	switch x := init.(type) {
	case *sema.InitList:
		if len(x.Elems) == 0 {
			return 0, true
		}
		return staticIntegerValue(x.Elems[0].Value)
	case *sema.ImplicitCast:
		return staticIntegerValue(x.X)
	case *sema.ExplicitCast:
		return staticIntegerValue(x.X)
	case *sema.IntLit:
		return x.Value, true
	case *sema.CharLit:
		return int64(x.Value), true
	case *sema.EnumRef:
		if x.Enumerator == nil {
			return 0, true
		}
		return x.Enumerator.Value, true
	default:
		return 0, false
	}
}

func staticFloatingValue(init sema.Expr) (float64, bool) {
	switch x := init.(type) {
	case *sema.InitList:
		if len(x.Elems) == 0 {
			return 0, true
		}
		return staticFloatingValue(x.Elems[0].Value)
	case *sema.ImplicitCast:
		return staticFloatingValue(x.X)
	case *sema.ExplicitCast:
		return staticFloatingValue(x.X)
	case *sema.FloatLit:
		return x.Value, true
	case *sema.IntLit:
		return float64(x.Value), true
	case *sema.CharLit:
		return float64(x.Value), true
	case *sema.EnumRef:
		if x.Enumerator == nil {
			return 0, true
		}
		return float64(x.Enumerator.Value), true
	default:
		return 0, false
	}
}

func lookupField(typ sema.Type, name string) *sema.Field {
	switch x := sema.Unqual(typ).(type) {
	case *sema.StructType:
		for _, f := range x.Fields {
			if f.Name == name {
				return f
			}
		}
	case *sema.UnionType:
		for _, f := range x.Fields {
			if f.Name == name {
				return f
			}
		}
	}
	return nil
}

func (g *generator) initLeaves(typ sema.Type, prefix []sema.Designator) []initLeaf {
	switch x := sema.Unqual(typ).(type) {
	case *sema.ArrayType:
		if x.SizeKind != sema.ArrayConstantSize {
			return nil
		}
		var out []initLeaf
		for i := int64(0); i < x.Size; i++ {
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigArrayIndex, Index: i})
			out = append(out, g.initLeaves(x.Elem, ds)...)
		}
		return out
	case *sema.StructType:
		var out []initLeaf
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigFieldName, Field: f})
			out = append(out, g.initLeaves(f.T, ds)...)
		}
		return out
	case *sema.UnionType:
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigFieldName, Field: f})
			return g.initLeaves(f.T, ds)
		}
		return nil
	default:
		return []initLeaf{{typ: typ, designators: copyDesignators(prefix)}}
	}
}

func (g *generator) directInitSpans(typ sema.Type, prefix []sema.Designator) []initSpan {
	switch x := sema.Unqual(typ).(type) {
	case *sema.ArrayType:
		if x.SizeKind != sema.ArrayConstantSize {
			return nil
		}
		out := make([]initSpan, 0, x.Size)
		cursor := 0
		for i := int64(0); i < x.Size; i++ {
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigArrayIndex, Index: i})
			n := len(g.initLeaves(x.Elem, ds))
			out = append(out, initSpan{typ: x.Elem, designators: ds, start: cursor, end: cursor + n})
			cursor += n
		}
		return out
	case *sema.StructType:
		out := make([]initSpan, 0, len(x.Fields))
		cursor := 0
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigFieldName, Field: f})
			n := len(g.initLeaves(f.T, ds))
			out = append(out, initSpan{typ: f.T, designators: ds, start: cursor, end: cursor + n})
			cursor += n
		}
		return out
	case *sema.UnionType:
		for _, f := range x.Fields {
			if f == nil || (f.IsBitField && f.Name == "") {
				continue
			}
			ds := appendDesignator(prefix, sema.Designator{Kind: sema.DesigFieldName, Field: f})
			n := len(g.initLeaves(f.T, ds))
			return []initSpan{{typ: f.T, designators: ds, start: 0, end: n}}
		}
		return nil
	default:
		return []initSpan{{typ: typ, designators: copyDesignators(prefix), start: 0, end: 1}}
	}
}

func (g *generator) designatedSpan(typ sema.Type, ds []sema.Designator) (initSpan, error) {
	cur := typ
	prefix := make([]sema.Designator, 0, len(ds))
	var unionSpan *initSpan
	for _, d := range ds {
		switch d.Kind {
		case sema.DesigArrayIndex:
			at, ok := sema.Unqual(cur).(*sema.ArrayType)
			if !ok {
				return initSpan{}, fmt.Errorf("array designator applied to %s", cur)
			}
			prefix = appendDesignator(prefix, d)
			cur = at.Elem
		case sema.DesigFieldName:
			field := d.Field
			if field == nil {
				field = lookupField(cur, "")
			}
			if field == nil {
				return initSpan{}, fmt.Errorf("field designator has no resolved field")
			}
			if _, ok := sema.Unqual(cur).(*sema.UnionType); ok && unionSpan == nil {
				span, err := g.spanForPrefix(typ, prefix)
				if err != nil {
					return initSpan{}, err
				}
				unionSpan = &span
			}
			prefix = appendDesignator(prefix, sema.Designator{Kind: sema.DesigFieldName, Field: field})
			cur = field.T
		default:
			return initSpan{}, fmt.Errorf("unsupported designator kind %d", d.Kind)
		}
	}
	if unionSpan != nil {
		return initSpan{typ: cur, designators: copyDesignators(prefix), start: unionSpan.start, end: unionSpan.end}, nil
	}
	span, err := g.spanForPrefix(typ, prefix)
	if err != nil {
		return initSpan{}, err
	}
	span.typ = cur
	return span, nil
}

func (g *generator) spanForPrefix(typ sema.Type, prefix []sema.Designator) (initSpan, error) {
	leaves := g.initLeaves(typ, nil)
	start := -1
	end := -1
	for i, leaf := range leaves {
		if designatorHasPrefix(leaf.designators, prefix) {
			if start < 0 {
				start = i
			}
			end = i + 1
			continue
		}
		if start >= 0 {
			break
		}
	}
	if start < 0 {
		return initSpan{}, fmt.Errorf("designator does not name an initializable subobject")
	}
	return initSpan{typ: typ, designators: copyDesignators(prefix), start: start, end: end}, nil
}

func spanContaining(spans []initSpan, cursor int) *initSpan {
	for i := range spans {
		if cursor >= spans[i].start && cursor < spans[i].end {
			return &spans[i]
		}
	}
	return nil
}

func useWholeInitializer(init sema.Expr, typ sema.Type) bool {
	if _, ok := init.(*sema.InitList); ok {
		return true
	}
	if _, ok := stringLiteralInitializer(init); ok {
		if _, isArray := sema.Unqual(typ).(*sema.ArrayType); isArray {
			return true
		}
	}
	return isObjectType(typ) && init != nil && isObjectType(init.GetType())
}

func appendDesignator(ds []sema.Designator, d sema.Designator) []sema.Designator {
	out := copyDesignators(ds)
	out = append(out, d)
	return out
}

func copyDesignators(ds []sema.Designator) []sema.Designator {
	if len(ds) == 0 {
		return nil
	}
	out := make([]sema.Designator, len(ds))
	copy(out, ds)
	return out
}

func designatorHasPrefix(ds, prefix []sema.Designator) bool {
	if len(prefix) > len(ds) {
		return false
	}
	for i := range prefix {
		if ds[i].Kind != prefix[i].Kind {
			return false
		}
		switch prefix[i].Kind {
		case sema.DesigArrayIndex:
			if ds[i].Index != prefix[i].Index {
				return false
			}
		case sema.DesigFieldName:
			if ds[i].Field != prefix[i].Field {
				return false
			}
		}
	}
	return true
}

func bitMask(width int) uint64 {
	if width <= 0 {
		return 0
	}
	if width >= 64 {
		return ^uint64(0)
	}
	return (uint64(1) << uint(width)) - 1
}

func isConst(t sema.Type) bool {
	if q, ok := t.(*sema.QualType); ok {
		return q.Const
	}
	return false
}
