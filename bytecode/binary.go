package bytecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

var binaryMagic = [8]byte{'C', 'V', 'M', 'B', 'C', 0, 0, 1}

const (
	binaryFormatVersion = uint16(1)
	binarySectionModule = uint16(1)
	maxBinaryCount      = uint32(1 << 24)
	maxBinaryPayload    = uint64(1 << 32)
)

func EncodeModule(w io.Writer, m *Module) error {
	if err := ValidateModule(m); err != nil {
		return err
	}
	var payload bytes.Buffer
	bw := binaryModuleWriter{w: &payload}
	bw.writeModule(m)
	if bw.err != nil {
		return bw.err
	}
	data := payload.Bytes()
	if _, err := w.Write(binaryMagic[:]); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, binaryFormatVersion); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, binarySectionModule); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint64(len(data))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, crc32.ChecksumIEEE(data)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

func DecodeModule(r io.Reader) (*Module, error) {
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if magic != binaryMagic {
		return nil, fmt.Errorf("invalid bytecode magic")
	}
	br := binaryHeaderReader{r: r}
	formatVersion := br.u16()
	sectionCount := br.u16()
	_ = br.u32()
	if br.err != nil {
		return nil, br.err
	}
	if formatVersion != binaryFormatVersion {
		return nil, fmt.Errorf("unsupported bytecode binary version %d", formatVersion)
	}
	if sectionCount != 1 {
		return nil, fmt.Errorf("unsupported bytecode section count %d", sectionCount)
	}
	sectionID := br.u16()
	sectionVersion := br.u16()
	length := br.u64()
	checksum := br.u32()
	_ = br.u32()
	if br.err != nil {
		return nil, br.err
	}
	if sectionID != binarySectionModule || sectionVersion != 1 {
		return nil, fmt.Errorf("unsupported bytecode section id=%d version=%d", sectionID, sectionVersion)
	}
	if length > maxBinaryPayload {
		return nil, fmt.Errorf("bytecode module payload length %d exceeds limit", length)
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	if got := crc32.ChecksumIEEE(data); got != checksum {
		return nil, fmt.Errorf("bytecode module checksum mismatch: got %#x want %#x", got, checksum)
	}
	payload := bytes.NewReader(data)
	mr := binaryModuleReader{r: payload}
	mod := mr.module()
	if mr.err != nil {
		return nil, mr.err
	}
	if payload.Len() != 0 {
		return nil, fmt.Errorf("bytecode module payload has %d trailing bytes", payload.Len())
	}
	var trailing [1]byte
	n, err := r.Read(trailing[:])
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n != 0 {
		return nil, fmt.Errorf("bytecode file has trailing bytes")
	}
	if err := ValidateModule(mod); err != nil {
		return nil, err
	}
	return mod, nil
}

type binaryModuleWriter struct {
	w   io.Writer
	err error
}

func (w *binaryModuleWriter) writeModule(m *Module) {
	w.str(m.Version)
	w.i32(m.Entry.Global)
	w.str(m.Entry.Name)
	w.target(m.Target)
	w.globals(m.Globals)
	w.strings(m.Strings)
	w.layouts(m.Layouts)
	w.sigs(m.Sigs)
	w.functions(m.Functions)
}

func (w *binaryModuleWriter) target(t TargetInfo) {
	w.str(t.Name)
	w.str(t.Endian)
	w.i64(t.PointerSize)
	w.i64(t.PointerAlign)
	w.i64(t.BoolSize)
	w.i64(t.BoolAlign)
	w.str(t.BitFieldPolicy)
	w.str(t.LayoutVersion)
}

func (w *binaryModuleWriter) globals(gs []Global) {
	w.count(len(gs))
	for _, g := range gs {
		w.i32(g.ID)
		w.str(g.Name)
		w.i32(int(g.Kind))
		w.i32(g.Func)
		w.i32(g.Sig)
		w.str(g.Extern.Module)
		w.str(g.Extern.Name)
		w.str(g.Extern.ABI)
		w.i64(g.Size)
		w.i64(g.Align)
		w.bool(g.Readonly)
		w.init(g.Init)
	}
}

func (w *binaryModuleWriter) init(init InitData) {
	w.i64(init.ZeroFill)
	w.bytes(init.Bytes)
	w.count(len(init.Relocations))
	for _, r := range init.Relocations {
		w.i64(r.Offset)
		w.i32(int(r.Kind))
		w.i32(r.Target)
		w.i64(r.Addend)
	}
}

func (w *binaryModuleWriter) strings(ss []StringConst) {
	w.count(len(ss))
	for _, s := range ss {
		w.i32(s.ID)
		w.str(s.Value)
		w.bytes(s.Bytes)
	}
}

func (w *binaryModuleWriter) layouts(ls []ObjectLayout) {
	w.count(len(ls))
	for _, l := range ls {
		w.i32(l.ID)
		w.str(l.Name)
		w.i64(l.Size)
		w.i64(l.Align)
		w.count(len(l.Fields))
		for _, f := range l.Fields {
			w.i32(f.ID)
			w.str(f.Name)
			w.i64(f.Offset)
			w.i32(int(f.Type))
		}
		w.count(len(l.Bit))
		for _, bf := range l.Bit {
			w.i32(bf.ID)
			w.str(bf.Name)
			w.i32(int(bf.Container))
			w.i64(bf.ByteOffset)
			w.i32(bf.BitOffset)
			w.i32(bf.Width)
			w.bool(bf.Signed)
			w.bool(bf.Volatile)
			w.str(bf.LayoutPolicy)
		}
		w.i64(l.ElemSize)
	}
}

func (w *binaryModuleWriter) sigs(sigs []FuncSig) {
	w.count(len(sigs))
	for _, sig := range sigs {
		w.i32(sig.ID)
		w.i32(int(sig.Ret))
		w.count(len(sig.Params))
		for _, p := range sig.Params {
			w.i32(int(p))
		}
		w.bool(sig.Variadic)
	}
}

func (w *binaryModuleWriter) functions(fs []Function) {
	w.count(len(fs))
	for _, f := range fs {
		w.i32(f.ID)
		w.i32(f.GlobalID)
		w.str(f.Name)
		w.i32(f.Sig)
		w.count(len(f.Params))
		for _, p := range f.Params {
			w.str(p.Name)
			w.i32(int(p.Type))
			w.i32(p.Slot)
		}
		w.count(len(f.Locals))
		for _, l := range f.Locals {
			w.i32(l.ID)
			w.str(l.Name)
			w.i32(int(l.Type))
		}
		w.count(len(f.Objects))
		for _, o := range f.Objects {
			w.i32(o.ID)
			w.str(o.Name)
			w.i64(o.Size)
			w.i64(o.Align)
			w.i32(o.Layout)
		}
		w.count(len(f.DynamicObjects))
		for _, o := range f.DynamicObjects {
			w.i32(o.ID)
			w.str(o.Name)
			w.i64(o.Align)
			w.i32(o.Layout)
		}
		w.count(len(f.Labels))
		for _, l := range f.Labels {
			w.i32(l.ID)
			w.str(l.Name)
			w.count(len(l.Stack))
			for _, vt := range l.Stack {
				w.i32(int(vt))
			}
			w.bool(l.Statement)
		}
		w.count(len(f.Instrs))
		for _, ins := range f.Instrs {
			w.instr(ins)
		}
		w.i32(f.MaxStack)
	}
}

func (w *binaryModuleWriter) instr(ins Instr) {
	w.i32(int(ins.Op))
	w.i32(int(ins.Type))
	w.i32(int(ins.Type2))
	w.i64(ins.Int)
	w.f64(ins.Float)
	w.str(ins.String)
	w.i32(ins.Slot)
	w.i32(ins.Global)
	w.i32(ins.Func)
	w.i32(ins.Sig)
	w.i32(ins.Layout)
	w.i32(ins.Field)
	w.i32(ins.Object)
	w.i32(ins.Label)
	w.count(len(ins.Labels))
	for _, c := range ins.Labels {
		w.i64(c.Value)
		w.i32(c.Label)
	}
	w.i64(ins.Align)
	w.i64(ins.Size)
	w.bool(ins.Volatile)
	w.i32(int(ins.Binary))
	w.i32(int(ins.Unary))
	w.i32(int(ins.Cast))
	w.i32(ins.Argc)
}

func (w *binaryModuleWriter) str(s string) {
	w.bytes([]byte(s))
}

func (w *binaryModuleWriter) bytes(b []byte) {
	w.count(len(b))
	if w.err != nil {
		return
	}
	_, w.err = w.w.Write(b)
}

func (w *binaryModuleWriter) count(n int) {
	if n < 0 || uint64(n) > uint64(maxBinaryCount) {
		w.err = fmt.Errorf("bytecode binary count %d exceeds limit", n)
		return
	}
	w.u32(uint32(n))
}

func (w *binaryModuleWriter) bool(v bool) {
	if v {
		w.u8(1)
		return
	}
	w.u8(0)
}

func (w *binaryModuleWriter) u8(v uint8) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.LittleEndian, v)
}

func (w *binaryModuleWriter) u32(v uint32) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.LittleEndian, v)
}

func (w *binaryModuleWriter) i32(v int) {
	if w.err != nil {
		return
	}
	if v < -1<<31 || v > 1<<31-1 {
		w.err = fmt.Errorf("bytecode binary int32 overflow: %d", v)
		return
	}
	w.err = binary.Write(w.w, binary.LittleEndian, int32(v))
}

func (w *binaryModuleWriter) i64(v int64) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.LittleEndian, v)
}

func (w *binaryModuleWriter) f64(v float64) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.LittleEndian, v)
}

type binaryHeaderReader struct {
	r   io.Reader
	err error
}

func (r *binaryHeaderReader) u16() uint16 {
	var v uint16
	r.read(&v)
	return v
}

func (r *binaryHeaderReader) u32() uint32 {
	var v uint32
	r.read(&v)
	return v
}

func (r *binaryHeaderReader) u64() uint64 {
	var v uint64
	r.read(&v)
	return v
}

func (r *binaryHeaderReader) read(v any) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, binary.LittleEndian, v)
}

type binaryModuleReader struct {
	r   io.Reader
	err error
}

func (r *binaryModuleReader) module() *Module {
	m := &Module{}
	m.Version = r.str()
	m.Entry = &EntryPoint{Global: r.i32(), Name: r.str()}
	m.Target = r.target()
	m.Globals = r.globals()
	m.Strings = r.strings()
	m.Layouts = r.layouts()
	m.Sigs = r.sigs()
	m.Functions = r.functions()
	return m
}

func (r *binaryModuleReader) target() TargetInfo {
	return TargetInfo{
		Name:           r.str(),
		Endian:         r.str(),
		PointerSize:    r.i64(),
		PointerAlign:   r.i64(),
		BoolSize:       r.i64(),
		BoolAlign:      r.i64(),
		BitFieldPolicy: r.str(),
		LayoutVersion:  r.str(),
	}
}

func (r *binaryModuleReader) globals() []Global {
	n := r.count()
	if n == 0 {
		return nil
	}
	gs := make([]Global, n)
	for i := range gs {
		gs[i] = Global{
			ID:   r.i32(),
			Name: r.str(),
			Kind: GlobalKind(r.i32()),
			Func: r.i32(),
			Sig:  r.i32(),
			Extern: ExternRef{
				Module: r.str(),
				Name:   r.str(),
				ABI:    r.str(),
			},
			Size:     r.i64(),
			Align:    r.i64(),
			Readonly: r.bool(),
			Init:     r.init(),
		}
	}
	return gs
}

func (r *binaryModuleReader) init() InitData {
	init := InitData{ZeroFill: r.i64(), Bytes: r.bytes()}
	relocCount := r.count()
	if relocCount > 0 {
		init.Relocations = make([]Relocation, relocCount)
	}
	for i := range init.Relocations {
		init.Relocations[i] = Relocation{
			Offset: r.i64(),
			Kind:   RelocationKind(r.i32()),
			Target: r.i32(),
			Addend: r.i64(),
		}
	}
	return init
}

func (r *binaryModuleReader) strings() []StringConst {
	n := r.count()
	if n == 0 {
		return nil
	}
	ss := make([]StringConst, n)
	for i := range ss {
		ss[i] = StringConst{ID: r.i32(), Value: r.str(), Bytes: r.bytes()}
	}
	return ss
}

func (r *binaryModuleReader) layouts() []ObjectLayout {
	n := r.count()
	if n == 0 {
		return nil
	}
	ls := make([]ObjectLayout, n)
	for i := range ls {
		ls[i] = ObjectLayout{ID: r.i32(), Name: r.str(), Size: r.i64(), Align: r.i64()}
		fieldCount := r.count()
		if fieldCount > 0 {
			ls[i].Fields = make([]FieldLayout, fieldCount)
		}
		for j := range ls[i].Fields {
			ls[i].Fields[j] = FieldLayout{ID: r.i32(), Name: r.str(), Offset: r.i64(), Type: ValueType(r.i32())}
		}
		bitCount := r.count()
		if bitCount > 0 {
			ls[i].Bit = make([]BitFieldLayout, bitCount)
		}
		for j := range ls[i].Bit {
			ls[i].Bit[j] = BitFieldLayout{
				ID:           r.i32(),
				Name:         r.str(),
				Container:    ValueType(r.i32()),
				ByteOffset:   r.i64(),
				BitOffset:    r.i32(),
				Width:        r.i32(),
				Signed:       r.bool(),
				Volatile:     r.bool(),
				LayoutPolicy: r.str(),
			}
		}
		ls[i].ElemSize = r.i64()
	}
	return ls
}

func (r *binaryModuleReader) sigs() []FuncSig {
	n := r.count()
	if n == 0 {
		return nil
	}
	sigs := make([]FuncSig, n)
	for i := range sigs {
		sigs[i] = FuncSig{ID: r.i32(), Ret: ValueType(r.i32())}
		paramCount := r.count()
		if paramCount > 0 {
			sigs[i].Params = make([]ValueType, paramCount)
		}
		for j := range sigs[i].Params {
			sigs[i].Params[j] = ValueType(r.i32())
		}
		sigs[i].Variadic = r.bool()
	}
	return sigs
}

func (r *binaryModuleReader) functions() []Function {
	n := r.count()
	if n == 0 {
		return nil
	}
	fs := make([]Function, n)
	for i := range fs {
		fs[i] = Function{ID: r.i32(), GlobalID: r.i32(), Name: r.str(), Sig: r.i32()}
		paramCount := r.count()
		if paramCount > 0 {
			fs[i].Params = make([]Param, paramCount)
		}
		for j := range fs[i].Params {
			fs[i].Params[j] = Param{Name: r.str(), Type: ValueType(r.i32()), Slot: r.i32()}
		}
		localCount := r.count()
		if localCount > 0 {
			fs[i].Locals = make([]LocalSlot, localCount)
		}
		for j := range fs[i].Locals {
			fs[i].Locals[j] = LocalSlot{ID: r.i32(), Name: r.str(), Type: ValueType(r.i32())}
		}
		objectCount := r.count()
		if objectCount > 0 {
			fs[i].Objects = make([]LocalObject, objectCount)
		}
		for j := range fs[i].Objects {
			fs[i].Objects[j] = LocalObject{ID: r.i32(), Name: r.str(), Size: r.i64(), Align: r.i64(), Layout: r.i32()}
		}
		dynamicCount := r.count()
		if dynamicCount > 0 {
			fs[i].DynamicObjects = make([]DynamicObject, dynamicCount)
		}
		for j := range fs[i].DynamicObjects {
			fs[i].DynamicObjects[j] = DynamicObject{ID: r.i32(), Name: r.str(), Align: r.i64(), Layout: r.i32()}
		}
		labelCount := r.count()
		if labelCount > 0 {
			fs[i].Labels = make([]Label, labelCount)
		}
		for j := range fs[i].Labels {
			fs[i].Labels[j] = Label{ID: r.i32(), Name: r.str()}
			stackCount := r.count()
			if stackCount > 0 {
				fs[i].Labels[j].Stack = make([]ValueType, stackCount)
			}
			for k := range fs[i].Labels[j].Stack {
				fs[i].Labels[j].Stack[k] = ValueType(r.i32())
			}
			fs[i].Labels[j].Statement = r.bool()
		}
		instrCount := r.count()
		if instrCount > 0 {
			fs[i].Instrs = make([]Instr, instrCount)
		}
		for j := range fs[i].Instrs {
			fs[i].Instrs[j] = r.instr()
		}
		fs[i].MaxStack = r.i32()
	}
	return fs
}

func (r *binaryModuleReader) instr() Instr {
	ins := Instr{
		Op:     Opcode(r.i32()),
		Type:   ValueType(r.i32()),
		Type2:  ValueType(r.i32()),
		Int:    r.i64(),
		Float:  r.f64(),
		String: r.str(),
		Slot:   r.i32(),
		Global: r.i32(),
		Func:   r.i32(),
		Sig:    r.i32(),
		Layout: r.i32(),
		Field:  r.i32(),
		Object: r.i32(),
		Label:  r.i32(),
	}
	labelCount := r.count()
	if labelCount > 0 {
		ins.Labels = make([]SwitchCase, labelCount)
	}
	for i := range ins.Labels {
		ins.Labels[i] = SwitchCase{Value: r.i64(), Label: r.i32()}
	}
	ins.Align = r.i64()
	ins.Size = r.i64()
	ins.Volatile = r.bool()
	ins.Binary = BinaryOp(r.i32())
	ins.Unary = UnaryOp(r.i32())
	ins.Cast = CastOp(r.i32())
	ins.Argc = r.i32()
	return ins
}

func (r *binaryModuleReader) str() string {
	return string(r.bytes())
}

func (r *binaryModuleReader) bytes() []byte {
	n := r.count()
	if n == 0 {
		return nil
	}
	b := make([]byte, n)
	if r.err != nil {
		return nil
	}
	_, r.err = io.ReadFull(r.r, b)
	return b
}

func (r *binaryModuleReader) count() int {
	n := r.u32()
	if n > maxBinaryCount {
		r.err = fmt.Errorf("bytecode binary count %d exceeds limit", n)
		return 0
	}
	return int(n)
}

func (r *binaryModuleReader) bool() bool {
	v := r.u8()
	if v > 1 && r.err == nil {
		r.err = fmt.Errorf("invalid bytecode boolean value %d", v)
	}
	return v == 1
}

func (r *binaryModuleReader) u8() uint8 {
	var v uint8
	r.read(&v)
	return v
}

func (r *binaryModuleReader) u32() uint32 {
	var v uint32
	r.read(&v)
	return v
}

func (r *binaryModuleReader) i32() int {
	var v int32
	r.read(&v)
	return int(v)
}

func (r *binaryModuleReader) i64() int64 {
	var v int64
	r.read(&v)
	return v
}

func (r *binaryModuleReader) f64() float64 {
	var v float64
	r.read(&v)
	return v
}

func (r *binaryModuleReader) read(v any) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, binary.LittleEndian, v)
}
