package runtime

import (
	"context"
	"fmt"

	"shinya.click/cvm/bytecode"
)

type RunOptions struct {
	StepLimit int
}

type VM struct {
	program *Program
	stack   []Value
	frames  []frame
	steps   int
	limit   int
}

type frame struct {
	fn     *bytecode.Function
	pc     int
	locals []Value
	labels map[int]int
}

func Run(ctx context.Context, p *Program, opts RunOptions) (ExitStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil {
		return ExitStatus{}, &TrapError{Reason: "nil program"}
	}
	vm := &VM{
		program: p,
		limit:   opts.StepLimit,
	}
	if err := vm.pushFrame(p.entryFunc, nil); err != nil {
		return ExitStatus{}, err
	}
	for {
		st, done, err := vm.step(ctx)
		if done || err != nil {
			return st, err
		}
	}
}

func (vm *VM) pushFrame(funcID int, args []Value) error {
	if funcID < 0 || funcID >= len(vm.program.module.Functions) {
		return vm.trap("invalid function id")
	}

	fn := &vm.program.module.Functions[funcID]
	maxSlot := -1
	for _, param := range fn.Params {
		if param.Slot < 0 {
			return vm.trap(fmt.Sprintf("negative param slot %d in function %s", param.Slot, fn.Name))
		}
		if param.Slot > maxSlot {
			maxSlot = param.Slot
		}
	}
	for _, local := range fn.Locals {
		if local.ID < 0 {
			return vm.trap(fmt.Sprintf("negative local slot %d in function %s", local.ID, fn.Name))
		}
		if local.ID > maxSlot {
			maxSlot = local.ID
		}
	}

	locals := make([]Value, maxSlot+1)
	for _, param := range fn.Params {
		locals[param.Slot] = zeroValue(param.Type)
	}
	for _, local := range fn.Locals {
		locals[local.ID] = zeroValue(local.Type)
	}
	if len(args) != len(fn.Params) {
		return vm.trap(fmt.Sprintf("function %s expects %d args, got %d", fn.Name, len(fn.Params), len(args)))
	}
	for i, arg := range args {
		param := fn.Params[i]
		if arg.Type != param.Type {
			return vm.trap(fmt.Sprintf("argument %d has type %s, want %s", i, arg.Type, param.Type))
		}
		locals[param.Slot] = arg
	}

	labels := make(map[int]int)
	for pc, ins := range fn.Instrs {
		if ins.Op == bytecode.OpLabel {
			labels[ins.Label] = pc
		}
	}

	vm.frames = append(vm.frames, frame{
		fn:     fn,
		locals: locals,
		labels: labels,
	})
	return nil
}

func (vm *VM) step(ctx context.Context) (ExitStatus, bool, error) {
	if len(vm.frames) == 0 {
		return ExitStatus{}, true, vm.trap("empty call stack")
	}
	fr := &vm.frames[len(vm.frames)-1]

	if err := ctx.Err(); err != nil {
		return ExitStatus{}, true, vm.trapAtPCWithCause("context canceled", fr.pc, false, err)
	}
	if vm.limit > 0 && vm.steps >= vm.limit {
		return ExitStatus{}, true, vm.trapAtPC("step limit exceeded", fr.pc, false)
	}

	if fr.pc < 0 || fr.pc >= len(fr.fn.Instrs) {
		return ExitStatus{}, true, vm.trapAtPC("program counter out of range", fr.pc, false)
	}

	ins := fr.fn.Instrs[fr.pc]
	fr.pc++
	vm.steps++

	switch ins.Op {
	case bytecode.OpConst:
		vm.stack = append(vm.stack, constValue(ins))
	case bytecode.OpDup:
		if len(vm.stack) == 0 {
			return ExitStatus{}, true, vm.trap("stack underflow")
		}
		vm.stack = append(vm.stack, vm.stack[len(vm.stack)-1])
	case bytecode.OpPop:
		if _, err := vm.pop(bytecode.TypeVoid); err != nil {
			return ExitStatus{}, true, err
		}
	case bytecode.OpSwap:
		if len(vm.stack) < 2 {
			return ExitStatus{}, true, vm.trap("stack underflow")
		}
		vm.stack[len(vm.stack)-1], vm.stack[len(vm.stack)-2] = vm.stack[len(vm.stack)-2], vm.stack[len(vm.stack)-1]
	case bytecode.OpLoadLocal:
		if ins.Slot < 0 || ins.Slot >= len(fr.locals) {
			return ExitStatus{}, true, vm.trap("local slot out of range")
		}
		v := fr.locals[ins.Slot]
		if v.Type != ins.Type {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("local slot %d has type %s, want %s", ins.Slot, v.Type, ins.Type))
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpStoreLocal:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if ins.Slot < 0 || ins.Slot >= len(fr.locals) {
			return ExitStatus{}, true, vm.trap("local slot out of range")
		}
		if fr.locals[ins.Slot].Type != bytecode.TypeVoid && fr.locals[ins.Slot].Type != ins.Type {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("local slot %d has type %s, want %s", ins.Slot, fr.locals[ins.Slot].Type, ins.Type))
		}
		fr.locals[ins.Slot] = v
	case bytecode.OpBinary:
		if err := vm.binary(ins); err != nil {
			return ExitStatus{}, true, err
		}
	case bytecode.OpUnary:
		if err := vm.unary(ins); err != nil {
			return ExitStatus{}, true, err
		}
	case bytecode.OpCast:
		if err := vm.cast(ins); err != nil {
			return ExitStatus{}, true, err
		}
	case bytecode.OpJump:
		if err := fr.jump(ins.Label); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid jump", err)
		}
	case bytecode.OpJumpIfZero:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if v.IsZero() {
			if err := fr.jump(ins.Label); err != nil {
				return ExitStatus{}, true, vm.trapWithCause("invalid jump", err)
			}
		}
	case bytecode.OpJumpIfNonZero:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if !v.IsZero() {
			if err := fr.jump(ins.Label); err != nil {
				return ExitStatus{}, true, vm.trapWithCause("invalid jump", err)
			}
		}
	case bytecode.OpSwitch:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if !isIntegerLike(ins.Type) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("unsupported switch type %s", ins.Type))
		}
		target := ins.Label
		for _, c := range ins.Labels {
			if switchCaseMatches(ins.Type, v, c.Value) {
				target = c.Label
				break
			}
		}
		if err := fr.jump(target); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid jump", err)
		}
	case bytecode.OpReturn:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if len(vm.frames) == 1 {
			code, err := v.ExitCode()
			if err != nil {
				return ExitStatus{}, true, vm.trapWithCause("invalid exit code", err)
			}
			vm.frames = vm.frames[:len(vm.frames)-1]
			return ExitStatus{Code: code}, true, nil
		}
		vm.frames = vm.frames[:len(vm.frames)-1]
		vm.stack = append(vm.stack, v)
	case bytecode.OpReturnVoid:
		if len(vm.frames) == 1 {
			return ExitStatus{}, true, vm.trap("void return from entry function")
		}
		vm.frames = vm.frames[:len(vm.frames)-1]
	case bytecode.OpLabel:
		// Labels are markers for control-flow instructions; execution falls through.
	default:
		return ExitStatus{}, true, vm.trap(fmt.Sprintf("unsupported opcode %s", ins.Op))
	}
	return ExitStatus{}, false, nil
}

func (fr *frame) jump(label int) error {
	pc, ok := fr.labels[label]
	if !ok {
		return fmt.Errorf("missing label L%d", label)
	}
	fr.pc = pc
	return nil
}

func (vm *VM) binary(ins bytecode.Instr) error {
	r, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	l, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}

	if isFloatType(ins.Type) {
		return vm.floatBinary(ins, l, r)
	}
	if !isIntegerLike(ins.Type) {
		return vm.trap(fmt.Sprintf("unsupported binary type %s", ins.Type))
	}

	width := bitWidth(ins.Type)
	var out Value
	switch ins.Binary {
	case bytecode.BinAdd:
		out = UIntValue(ins.Type, maskToWidth(l.Int+r.Int, width))
	case bytecode.BinSub:
		out = UIntValue(ins.Type, maskToWidth(l.Int-r.Int, width))
	case bytecode.BinMul:
		out = UIntValue(ins.Type, maskToWidth(l.Int*r.Int, width))
	case bytecode.BinDivS:
		rs := signedInt(r)
		if rs == 0 {
			return vm.trap("division by zero")
		}
		ls := signedInt(l)
		if ls == minSigned(width) && rs == -1 {
			return vm.trap("signed division overflow")
		}
		out = IntValue(ins.Type, ls/rs)
	case bytecode.BinDivU:
		ru := unsignedInt(r)
		if ru == 0 {
			return vm.trap("division by zero")
		}
		out = UIntValue(ins.Type, unsignedInt(l)/ru)
	case bytecode.BinRemS:
		rs := signedInt(r)
		if rs == 0 {
			return vm.trap("division by zero")
		}
		ls := signedInt(l)
		if ls == minSigned(width) && rs == -1 {
			return vm.trap("signed remainder overflow")
		}
		out = IntValue(ins.Type, ls%rs)
	case bytecode.BinRemU:
		ru := unsignedInt(r)
		if ru == 0 {
			return vm.trap("division by zero")
		}
		out = UIntValue(ins.Type, unsignedInt(l)%ru)
	case bytecode.BinAnd:
		out = UIntValue(ins.Type, maskToWidth(l.Int&r.Int, width))
	case bytecode.BinOr:
		out = UIntValue(ins.Type, maskToWidth(l.Int|r.Int, width))
	case bytecode.BinXor:
		out = UIntValue(ins.Type, maskToWidth(l.Int^r.Int, width))
	case bytecode.BinShl:
		n, err := shiftCount(r, width)
		if err != nil {
			return vm.trapWithCause("invalid shift count", err)
		}
		out = UIntValue(ins.Type, shiftLeft(unsignedInt(l), n, width))
	case bytecode.BinShrS:
		n, err := shiftCount(r, width)
		if err != nil {
			return vm.trapWithCause("invalid shift count", err)
		}
		out = UIntValue(ins.Type, uint64(shiftRightSigned(signedInt(l), n)))
	case bytecode.BinShrU:
		n, err := shiftCount(r, width)
		if err != nil {
			return vm.trapWithCause("invalid shift count", err)
		}
		out = UIntValue(ins.Type, shiftRightUnsigned(unsignedInt(l), n))
	case bytecode.BinEq:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) == unsignedInt(r))))
	case bytecode.BinNe:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) != unsignedInt(r))))
	case bytecode.BinLtS:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(signedInt(l) < signedInt(r))))
	case bytecode.BinLtU:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) < unsignedInt(r))))
	case bytecode.BinLeS:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(signedInt(l) <= signedInt(r))))
	case bytecode.BinLeU:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) <= unsignedInt(r))))
	case bytecode.BinGtS:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(signedInt(l) > signedInt(r))))
	case bytecode.BinGtU:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) > unsignedInt(r))))
	case bytecode.BinGeS:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(signedInt(l) >= signedInt(r))))
	case bytecode.BinGeU:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(unsignedInt(l) >= unsignedInt(r))))
	default:
		return vm.trap(fmt.Sprintf("unsupported binary op %s", ins.Binary))
	}
	vm.stack = append(vm.stack, normalizeInt(out))
	return nil
}

func (vm *VM) floatBinary(ins bytecode.Instr, l, r Value) error {
	var out Value
	switch ins.Binary {
	case bytecode.BinEq:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float == r.Float)))
	case bytecode.BinNe:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float != r.Float)))
	case bytecode.BinLtF:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float < r.Float)))
	case bytecode.BinLeF:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float <= r.Float)))
	case bytecode.BinGtF:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float > r.Float)))
	case bytecode.BinGeF:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Float >= r.Float)))
	default:
		return vm.trap(fmt.Sprintf("unsupported float binary op %s", ins.Binary))
	}
	vm.stack = append(vm.stack, out)
	return nil
}

func (vm *VM) unary(ins bytecode.Instr) error {
	v, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	switch ins.Unary {
	case bytecode.UnaryNeg:
		if isFloatType(ins.Type) {
			vm.stack = append(vm.stack, FloatValue(ins.Type, -v.Float))
			return nil
		}
		if !isIntegerLike(ins.Type) {
			return vm.trap(fmt.Sprintf("unsupported unary type %s", ins.Type))
		}
		vm.stack = append(vm.stack, normalizeInt(IntValue(ins.Type, -signedInt(v))))
		return nil
	default:
		return vm.trap(fmt.Sprintf("unsupported unary op %d", int(ins.Unary)))
	}
}

func (vm *VM) cast(ins bytecode.Instr) error {
	v, err := vm.pop(ins.Type)
	if err != nil {
		return err
	}
	if ins.Cast == bytecode.CastBool {
		if ins.Type2 != bytecode.TypeBool {
			return vm.trap(fmt.Sprintf("bool cast result type is %s, want %s", ins.Type2, bytecode.TypeBool))
		}
		vm.stack = append(vm.stack, UIntValue(bytecode.TypeBool, uint64(boolInt(!v.IsZero()))))
		return nil
	}
	if ins.Type == ins.Type2 {
		vm.stack = append(vm.stack, v)
		return nil
	}
	if ins.Cast == bytecode.CastBit {
		if !canBitCast(ins.Type, ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported bit cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, bitCast(v, ins.Type2))
		return nil
	}

	switch ins.Cast {
	case bytecode.CastTrunc, bytecode.CastZExt:
		if !isIntegerLike(ins.Type) || !isIntegerLike(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported integer cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, normalizeInt(UIntValue(ins.Type2, unsignedInt(v))))
	case bytecode.CastSExt:
		if !isIntegerLike(ins.Type) || !isIntegerLike(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported integer cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, normalizeInt(IntValue(ins.Type2, signedInt(v))))
	case bytecode.CastPtrToInt:
		if !isPointerType(ins.Type) || !isIntegerLike(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported pointer-to-int cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, normalizeInt(UIntValue(ins.Type2, v.Int)))
	case bytecode.CastIntToPtr:
		if !isIntegerLike(ins.Type) || !isPointerType(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported int-to-pointer cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, bitCast(v, ins.Type2))
	case bytecode.CastFExt, bytecode.CastFTrunc, bytecode.CastIntToFloat, bytecode.CastFloatToInt:
		return vm.trap(fmt.Sprintf("unsupported float cast %s->%s", ins.Type, ins.Type2))
	default:
		return vm.trap(fmt.Sprintf("unsupported cast op %d", int(ins.Cast)))
	}
	return nil
}

func constValue(ins bytecode.Instr) Value {
	switch ins.Type {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return FloatValue(ins.Type, ins.Float)
	default:
		return UIntValue(ins.Type, uint64(ins.Int))
	}
}

func zeroValue(t bytecode.ValueType) Value {
	switch t {
	case bytecode.TypeF32, bytecode.TypeF64, bytecode.TypeFLong:
		return FloatValue(t, 0)
	default:
		return UIntValue(t, 0)
	}
}

func boolInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func switchCaseMatches(t bytecode.ValueType, v Value, c int64) bool {
	if isUnsignedIntegerType(t) {
		return c >= 0 && unsignedInt(v) == uint64(c)
	}
	return signedInt(v) == c
}

func bitCast(v Value, to bytecode.ValueType) Value {
	if isFloatType(to) {
		return FloatValue(to, v.Float)
	}
	return normalizeInt(UIntValue(to, v.Int))
}

func normalizeInt(v Value) Value {
	if isIntegerLike(v.Type) || isPointerType(v.Type) {
		v.Int = maskToWidth(v.Int, bitWidth(v.Type))
	}
	return v
}

func signedInt(v Value) int64 {
	width := bitWidth(v.Type)
	if width == 0 {
		return int64(v.Int)
	}
	u := maskToWidth(v.Int, width)
	if width == 64 {
		return int64(u)
	}
	sign := uint64(1) << (width - 1)
	if u&sign == 0 {
		return int64(u)
	}
	return int64(u | ^maskToWidth(^uint64(0), width))
}

func unsignedInt(v Value) uint64 {
	return maskToWidth(v.Int, bitWidth(v.Type))
}

func maskToWidth(v uint64, width uint) uint64 {
	if width >= 64 {
		return v
	}
	return v & ((uint64(1) << width) - 1)
}

func minSigned(width uint) int64 {
	if width >= 64 {
		return -1 << 63
	}
	return -(int64(1) << (width - 1))
}

func shiftCount(v Value, width uint) (uint, error) {
	if isSignedIntegerType(v.Type) {
		n := signedInt(v)
		if n < 0 || uint64(n) >= uint64(width) {
			return 0, fmt.Errorf("%d outside [0,%d)", n, width)
		}
		return uint(n), nil
	}
	n := unsignedInt(v)
	if n >= uint64(width) {
		return 0, fmt.Errorf("%d outside [0,%d)", n, width)
	}
	return uint(n), nil
}

func shiftLeft(v uint64, n, width uint) uint64 {
	return maskToWidth(v<<n, width)
}

func shiftRightUnsigned(v uint64, n uint) uint64 {
	return v >> n
}

func shiftRightSigned(v int64, n uint) int64 {
	return v >> n
}

func bitWidth(t bytecode.ValueType) uint {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeU8:
		return 8
	case bytecode.TypeI16, bytecode.TypeU16:
		return 16
	case bytecode.TypeI32, bytecode.TypeU32:
		return 32
	case bytecode.TypeI64, bytecode.TypeU64, bytecode.TypePtr, bytecode.TypeObjectAddr:
		return 64
	default:
		return 0
	}
}

func isIntegerLike(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeBool, bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64, bytecode.TypeU8, bytecode.TypeU16, bytecode.TypeU32, bytecode.TypeU64:
		return true
	default:
		return false
	}
}

func isSignedIntegerType(t bytecode.ValueType) bool {
	switch t {
	case bytecode.TypeI8, bytecode.TypeI16, bytecode.TypeI32, bytecode.TypeI64:
		return true
	default:
		return false
	}
}

func isUnsignedIntegerType(t bytecode.ValueType) bool {
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

func canBitCast(from, to bytecode.ValueType) bool {
	return (isIntegerLike(from) || isPointerType(from)) && (isIntegerLike(to) || isPointerType(to))
}

func (vm *VM) pop(t bytecode.ValueType) (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, vm.trap("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	if t != bytecode.TypeVoid && v.Type != t {
		return Value{}, vm.trap(fmt.Sprintf("stack value has type %s, want %s", v.Type, t))
	}
	return v, nil
}

func (vm *VM) trap(reason string) *TrapError {
	return vm.trapWithCause(reason, nil)
}

func (vm *VM) trapWithCause(reason string, cause error) *TrapError {
	if len(vm.frames) == 0 {
		return &TrapError{
			Reason: reason,
			Cause:  cause,
			Stack:  vm.stackTrace(),
		}
	}
	fr := vm.frames[len(vm.frames)-1]
	pc := fr.pc
	if pc > 0 {
		pc--
	}
	return vm.trapAtPCWithCause(reason, pc, true, cause)
}

func (vm *VM) trapAtPC(reason string, pc int, includeOpcode bool) *TrapError {
	return vm.trapAtPCWithCause(reason, pc, includeOpcode, nil)
}

func (vm *VM) trapAtPCWithCause(reason string, pc int, includeOpcode bool, cause error) *TrapError {
	err := &TrapError{
		Reason: reason,
		Cause:  cause,
		Stack:  vm.stackTrace(),
	}
	if len(vm.frames) == 0 {
		return err
	}
	fr := vm.frames[len(vm.frames)-1]
	err.Function = fr.fn.Name
	err.HasLocation = true
	err.FunctionID = fr.fn.ID
	err.PC = pc
	err.Opcode = bytecode.Opcode(-1)
	if includeOpcode && pc >= 0 && pc < len(fr.fn.Instrs) {
		err.Opcode = fr.fn.Instrs[pc].Op
	}
	return err
}

func (vm *VM) stackTrace() []string {
	if len(vm.frames) == 0 {
		return nil
	}
	stack := make([]string, len(vm.frames))
	for i := range vm.frames {
		fr := vm.frames[len(vm.frames)-1-i]
		name := fr.fn.Name
		if name == "" {
			name = "fn"
		}
		stack[i] = fmt.Sprintf("%s#%d", name, fr.fn.ID)
	}
	return stack
}
