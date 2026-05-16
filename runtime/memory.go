package runtime

import (
	"encoding/binary"
	"fmt"
	"math"

	"shinya.click/cvm/bytecode"
)

type blockKind int

const (
	blockGlobal blockKind = iota
	blockString
	blockLocal
	blockDynamic
	blockHostHandle
)

type memoryBlock struct {
	id       int
	name     string
	base     uint64
	data     []byte
	align    int64
	readonly bool
	freed    bool
	kind     blockKind
}

type Memory struct {
	target bytecode.TargetInfo
	next   uint64
	blocks []*memoryBlock
}

func NewMemory(target bytecode.TargetInfo) *Memory {
	return &Memory{target: target, next: 0x1000}
}

func (m *Memory) Alloc(name string, size, align int64, readonly bool, kind blockKind) uint64 {
	addr, err := m.TryAlloc(name, size, align, readonly, kind)
	if err != nil {
		panic(err)
	}
	return addr
}

func (m *Memory) TryAlloc(name string, size, align int64, readonly bool, kind blockKind) (uint64, error) {
	if align <= 0 {
		align = 1
	}
	if size < 0 {
		return 0, fmt.Errorf("negative memory allocation size %d", size)
	}
	if uint64(size) > uint64(maxInt()) {
		return 0, fmt.Errorf("memory allocation size %d exceeds make limit", size)
	}
	base, err := alignUpChecked(m.next, uint64(align))
	if err != nil {
		return 0, err
	}
	if uint64(size) > math.MaxUint64-base {
		return 0, fmt.Errorf("memory allocation %q overflows address space", name)
	}
	end := base + uint64(size)
	if 0x10 > math.MaxUint64-end {
		return 0, fmt.Errorf("memory allocation %q overflows next address", name)
	}
	data, err := makeBlockData(size)
	if err != nil {
		return 0, err
	}
	b := &memoryBlock{id: len(m.blocks), name: name, base: base, data: data, align: align, readonly: readonly, kind: kind}
	m.blocks = append(m.blocks, b)
	m.next = end + 0x10
	return base, nil
}

func (m *Memory) AllocBytes(name string, data []byte, readonly bool, kind blockKind) uint64 {
	addr := m.Alloc(name, int64(len(data)), 1, readonly, kind)
	copy(m.blocks[len(m.blocks)-1].data, data)
	return addr
}

func (m *Memory) Load(addr uint64, t bytecode.ValueType, align int64) (Value, error) {
	if t == bytecode.TypeFLong {
		return Value{}, fmt.Errorf("unsupported long double memory load")
	}
	b, off, size, err := m.access(addr, t, align, false)
	if err != nil {
		return Value{}, err
	}
	raw := b.data[off : off+size]
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		return UIntValue(t, uint64(raw[0])), nil
	case bytecode.TypeI16, bytecode.TypeU16:
		order, err := m.byteOrder()
		if err != nil {
			return Value{}, err
		}
		return UIntValue(t, uint64(order.Uint16(raw))), nil
	case bytecode.TypeI32, bytecode.TypeU32:
		order, err := m.byteOrder()
		if err != nil {
			return Value{}, err
		}
		return UIntValue(t, uint64(order.Uint32(raw))), nil
	case bytecode.TypeF32:
		order, err := m.byteOrder()
		if err != nil {
			return Value{}, err
		}
		return FloatValue(t, float64(math.Float32frombits(order.Uint32(raw)))), nil
	case bytecode.TypeI64, bytecode.TypeU64:
		order, err := m.byteOrder()
		if err != nil {
			return Value{}, err
		}
		return UIntValue(t, order.Uint64(raw)), nil
	case bytecode.TypeF64:
		order, err := m.byteOrder()
		if err != nil {
			return Value{}, err
		}
		return FloatValue(t, math.Float64frombits(order.Uint64(raw))), nil
	case bytecode.TypePtr, bytecode.TypeObjectAddr:
		return m.loadPointer(raw, t)
	default:
		return Value{}, fmt.Errorf("unsupported load type %s", t)
	}
}

func (m *Memory) Store(addr uint64, t bytecode.ValueType, align int64, v Value) error {
	if t == bytecode.TypeFLong {
		return fmt.Errorf("unsupported long double memory store")
	}
	b, off, size, err := m.access(addr, t, align, true)
	if err != nil {
		return err
	}
	raw := b.data[off : off+size]
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		raw[0] = byte(v.Int)
	case bytecode.TypeI16, bytecode.TypeU16:
		order, err := m.byteOrder()
		if err != nil {
			return err
		}
		order.PutUint16(raw, uint16(v.Int))
	case bytecode.TypeI32, bytecode.TypeU32:
		order, err := m.byteOrder()
		if err != nil {
			return err
		}
		order.PutUint32(raw, uint32(v.Int))
	case bytecode.TypeF32:
		order, err := m.byteOrder()
		if err != nil {
			return err
		}
		order.PutUint32(raw, math.Float32bits(float32(v.Float)))
	case bytecode.TypeI64, bytecode.TypeU64:
		order, err := m.byteOrder()
		if err != nil {
			return err
		}
		order.PutUint64(raw, v.Int)
	case bytecode.TypeF64:
		order, err := m.byteOrder()
		if err != nil {
			return err
		}
		order.PutUint64(raw, math.Float64bits(v.Float))
	case bytecode.TypePtr, bytecode.TypeObjectAddr:
		return m.storePointer(raw, v.Int)
	default:
		return fmt.Errorf("unsupported store type %s", t)
	}
	return nil
}

func (m *Memory) Copy(dst, src uint64, size int64) error {
	db, doff, err := m.rangeAccess(dst, size, true)
	if err != nil {
		return err
	}
	sb, soff, err := m.rangeAccess(src, size, false)
	if err != nil {
		return err
	}
	copy(db.data[doff:doff+int(size)], sb.data[soff:soff+int(size)])
	return nil
}

func (m *Memory) Set(dst uint64, value byte, size int64) error {
	b, off, err := m.rangeAccess(dst, size, true)
	if err != nil {
		return err
	}
	for i := 0; i < int(size); i++ {
		b.data[off+i] = value
	}
	return nil
}

func (m *Memory) ReadCString(addr uint64) (string, error) {
	b, off, err := m.rangeAccess(addr, 1, false)
	if err != nil {
		return "", err
	}
	end := off
	for end < len(b.data) && b.data[end] != 0 {
		end++
	}
	if end >= len(b.data) {
		return "", fmt.Errorf("unterminated C string at %#x", addr)
	}
	return string(b.data[off:end]), nil
}

func (m *Memory) WritePointer(addr uint64, ptr uint64) error {
	return m.Store(addr, bytecode.TypePtr, m.target.PointerAlign, PtrValue(ptr))
}

func (m *Memory) access(addr uint64, t bytecode.ValueType, align int64, write bool) (*memoryBlock, int, int, error) {
	size := int(valueSize(m.target, t))
	if size <= 0 {
		return nil, 0, 0, fmt.Errorf("unsupported memory type %s", t)
	}
	b, off, err := m.rangeAccess(addr, int64(size), write)
	if err != nil {
		return nil, 0, 0, err
	}
	if align > 1 && addr%uint64(align) != 0 {
		return nil, 0, 0, fmt.Errorf("unaligned access at %#x align=%d", addr, align)
	}
	return b, off, size, nil
}

func (m *Memory) rangeAccess(addr uint64, size int64, write bool) (*memoryBlock, int, error) {
	if size < 0 {
		return nil, 0, fmt.Errorf("negative memory size %d", size)
	}
	uSize := uint64(size)
	if uSize > math.MaxUint64-addr {
		return nil, 0, fmt.Errorf("invalid memory access at %#x size=%d", addr, size)
	}
	end := addr + uSize
	for _, b := range m.blocks {
		blockEnd := b.base + uint64(len(b.data))
		if addr < b.base || end > blockEnd {
			continue
		}
		if b.freed {
			return nil, 0, fmt.Errorf("use after free at %#x", addr)
		}
		if write && b.readonly {
			return nil, 0, fmt.Errorf("readonly memory write at %#x", addr)
		}
		return b, int(addr - b.base), nil
	}
	return nil, 0, fmt.Errorf("invalid memory access at %#x size=%d", addr, size)
}

func (m *Memory) byteOrder() (binary.ByteOrder, error) {
	switch m.target.Endian {
	case "", "little":
		return binary.LittleEndian, nil
	case "big":
		return binary.BigEndian, nil
	default:
		return nil, fmt.Errorf("unsupported endian %q", m.target.Endian)
	}
}

func (m *Memory) loadPointer(raw []byte, t bytecode.ValueType) (Value, error) {
	order, err := m.byteOrder()
	if err != nil {
		return Value{}, err
	}
	switch m.target.PointerSize {
	case 4:
		return UIntValue(t, uint64(order.Uint32(raw))), nil
	case 8:
		return UIntValue(t, order.Uint64(raw)), nil
	default:
		return Value{}, fmt.Errorf("unsupported pointer size %d", m.target.PointerSize)
	}
}

func (m *Memory) storePointer(raw []byte, ptr uint64) error {
	order, err := m.byteOrder()
	if err != nil {
		return err
	}
	switch m.target.PointerSize {
	case 4:
		if ptr > math.MaxUint32 {
			return fmt.Errorf("pointer value %#x exceeds 32-bit pointer size", ptr)
		}
		order.PutUint32(raw, uint32(ptr))
	case 8:
		order.PutUint64(raw, ptr)
	default:
		return fmt.Errorf("unsupported pointer size %d", m.target.PointerSize)
	}
	return nil
}

func valueSize(target bytecode.TargetInfo, t bytecode.ValueType) int64 {
	switch t {
	case bytecode.TypeBool:
		return target.BoolSize
	case bytecode.TypeI8, bytecode.TypeU8:
		return 1
	case bytecode.TypeI16, bytecode.TypeU16:
		return 2
	case bytecode.TypeI32, bytecode.TypeU32, bytecode.TypeF32:
		return 4
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypeF64:
		return 8
	case bytecode.TypePtr, bytecode.TypeObjectAddr:
		return target.PointerSize
	default:
		return 0
	}
}

func alignUpChecked(v, align uint64) (uint64, error) {
	if align <= 1 {
		return v, nil
	}
	rem := v % align
	if rem == 0 {
		return v, nil
	}
	delta := align - rem
	if delta > math.MaxUint64-v {
		return 0, fmt.Errorf("aligned address overflows for %#x align=%d", v, align)
	}
	return v + delta, nil
}

func makeBlockData(size int64) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("memory allocation size %d exceeds make limit: %v", size, r)
		}
	}()
	return make([]byte, int(size)), nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}
