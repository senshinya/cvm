package runtime

import (
	"context"
	"fmt"
	"math"

	"shinya.click/cvm/bytecode"
)

type RunOptions struct {
	StepLimit int
}

type VM struct {
	program  *Program
	stack    []Value
	frames   []frame
	closures map[uint64]closure
	steps    int
	limit    int
}

type frame struct {
	fn             *bytecode.Function
	pc             int
	locals         []Value
	variadicArgs   []Value
	vaLists        map[int]int
	activeVaList   int
	hasActiveVa    bool
	labels         map[int]int
	localObjects   map[int]uint64
	dynamicObjects map[int]uint64
	closures       []uint64
}

type closure struct {
	global   int
	sig      int
	captures []Value
}

func Run(ctx context.Context, p *Program, opts RunOptions) (ExitStatus, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil {
		return ExitStatus{}, &TrapError{Reason: "nil program"}
	}
	if p.module == nil {
		return ExitStatus{}, &TrapError{Reason: "program module is nil"}
	}
	if p.memory == nil {
		return ExitStatus{}, &TrapError{Reason: "program memory is nil"}
	}
	vm := &VM{
		program:  p,
		closures: make(map[uint64]closure),
		limit:    opts.StepLimit,
	}
	if err := vm.pushFrame(p.entryFunc, p.entryArgs); err != nil {
		return ExitStatus{}, err
	}
	for {
		st, done, err := vm.step(ctx)
		if done || err != nil {
			cleanupErr := vm.cleanupFrames()
			if err != nil {
				return st, err
			}
			if cleanupErr != nil {
				return st, cleanupErr
			}
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
	sig := vm.program.module.Sigs[fn.Sig]
	if !sig.Variadic && len(args) != len(fn.Params) {
		return vm.trap(fmt.Sprintf("function %s expects %d args, got %d", fn.Name, len(fn.Params), len(args)))
	}
	if sig.Variadic && len(args) < len(fn.Params) {
		return vm.trap(fmt.Sprintf("variadic function %s expects at least %d args, got %d", fn.Name, len(fn.Params), len(args)))
	}
	for i, arg := range args[:len(fn.Params)] {
		param := fn.Params[i]
		if arg.Type != param.Type {
			return vm.trap(fmt.Sprintf("argument %d has type %s, want %s", i, arg.Type, param.Type))
		}
		locals[param.Slot] = arg
	}
	var variadicArgs []Value
	if sig.Variadic {
		variadicArgs = append([]Value(nil), args[len(fn.Params):]...)
	}

	labels := make(map[int]int)
	for pc, ins := range fn.Instrs {
		if ins.Op == bytecode.OpLabel {
			labels[ins.Label] = pc
		}
	}

	localObjects := make(map[int]uint64, len(fn.Objects))
	for _, object := range fn.Objects {
		if _, exists := localObjects[object.ID]; exists {
			return vm.trap(fmt.Sprintf("duplicate local object %d in function %s", object.ID, fn.Name))
		}
		addr, err := vm.program.Memory().TryAlloc(fmt.Sprintf("local:%s:%s", fn.Name, object.Name), object.Size, object.Align, false, blockLocal)
		if err != nil {
			return vm.trapWithCause(fmt.Sprintf("local object %d allocation failed", object.ID), err)
		}
		localObjects[object.ID] = addr
	}

	vm.frames = append(vm.frames, frame{
		fn:             fn,
		locals:         locals,
		variadicArgs:   variadicArgs,
		vaLists:        make(map[int]int),
		activeVaList:   -1,
		labels:         labels,
		localObjects:   localObjects,
		dynamicObjects: make(map[int]uint64),
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
	case bytecode.OpAddrGlobal:
		addr, err := vm.program.TryGlobalAddr(ins.Global)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid global address", err)
		}
		vm.stack = append(vm.stack, ObjectAddrValue(addr))
	case bytecode.OpAddrString:
		addr, err := vm.program.TryStringAddr(int(ins.Int))
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid string address", err)
		}
		vm.stack = append(vm.stack, ObjectAddrValue(addr))
	case bytecode.OpAddrFunc:
		addr, err := vm.program.TryFuncAddr(ins.Global)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid function address", err)
		}
		vm.stack = append(vm.stack, PtrValue(addr))
	case bytecode.OpAddrLocalObject:
		addr, ok := fr.localObjects[ins.Object]
		if !ok {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid local object %d", ins.Object))
		}
		vm.stack = append(vm.stack, ObjectAddrValue(addr))
	case bytecode.OpAllocDynamicObject:
		size, err := vm.pop(bytecode.TypeI64)
		if err != nil {
			return ExitStatus{}, true, err
		}
		object, ok := fr.dynamicObject(ins.Object)
		if !ok {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid dynamic object %d", ins.Object))
		}
		if _, exists := fr.dynamicObjects[ins.Object]; exists {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("dynamic object %d is already allocated", ins.Object))
		}
		if ins.Layout != object.Layout {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("dynamic object %d layout %d does not match instruction layout %d", ins.Object, object.Layout, ins.Layout))
		}
		if object.Layout < 0 || object.Layout >= len(vm.program.module.Layouts) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid dynamic object layout %d", object.Layout))
		}
		align := ins.Align
		if align <= 0 {
			align = object.Align
		}
		addr, err := vm.program.Memory().TryAlloc(fmt.Sprintf("dynamic:%s:%s", fr.fn.Name, object.Name), signedInt(size), align, false, blockDynamic)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause(fmt.Sprintf("dynamic object %d allocation failed", ins.Object), err)
		}
		fr.dynamicObjects[ins.Object] = addr
	case bytecode.OpFreeDynamicObject:
		if _, ok := fr.dynamicObject(ins.Object); !ok {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid dynamic object %d", ins.Object))
		}
		addr, ok := fr.dynamicObjects[ins.Object]
		if !ok {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("dynamic object %d is not allocated", ins.Object))
		}
		if err := vm.program.Memory().Free(addr, blockDynamic); err != nil {
			return ExitStatus{}, true, vm.trapWithCause(fmt.Sprintf("dynamic object %d free failed", ins.Object), err)
		}
		delete(fr.dynamicObjects, ins.Object)
	case bytecode.OpDynamicObjectAddr:
		if ins.Type != bytecode.TypeObjectAddr {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("dynamic object address has type %s, want %s", ins.Type, bytecode.TypeObjectAddr))
		}
		addr, ok := fr.dynamicObjects[ins.Object]
		if !ok {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("dynamic object %d is not allocated", ins.Object))
		}
		vm.stack = append(vm.stack, ObjectAddrValue(addr))
	case bytecode.OpLoadConst:
		base, err := vm.program.TryGlobalAddr(ins.Global)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid global address", err)
		}
		addr, err := addSignedOffset(base, ins.Int)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("load const offset overflow", err)
		}
		v, err := vm.program.Memory().Load(addr, ins.Type, ins.Align)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("load const failed", err)
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpLoad:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		v, err := vm.program.Memory().Load(addr.Int, ins.Type, ins.Align)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("memory load failed", err)
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpStore:
		v, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if err := vm.program.Memory().Store(addr.Int, ins.Type, ins.Align, v); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("memory store failed", err)
		}
	case bytecode.OpMemCopy:
		src, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		dst, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if err := vm.program.Memory().Copy(dst.Int, src.Int, ins.Size); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("memory copy failed", err)
		}
	case bytecode.OpMemSet:
		v, err := vm.pop(bytecode.TypeI32)
		if err != nil {
			return ExitStatus{}, true, err
		}
		dst, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if err := vm.program.Memory().Set(dst.Int, byte(v.Int), ins.Size); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("memory set failed", err)
		}
	case bytecode.OpOffset:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		out, err := addSignedOffset(addr.Int, ins.Int)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("offset overflow", err)
		}
		vm.stack = append(vm.stack, ObjectAddrValue(out))
	case bytecode.OpFieldAddr:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if ins.Layout < 0 || ins.Layout >= len(vm.program.module.Layouts) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid layout %d", ins.Layout))
		}
		layout := vm.program.module.Layouts[ins.Layout]
		if ins.Field < 0 || ins.Field >= len(layout.Fields) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("invalid field %d in layout %d", ins.Field, ins.Layout))
		}
		out, err := addSignedOffset(addr.Int, layout.Fields[ins.Field].Offset)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("field offset overflow", err)
		}
		vm.stack = append(vm.stack, ObjectAddrValue(out))
	case bytecode.OpBitFieldLoad:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		bf, err := vm.bitFieldLayout(ins)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid bit-field", err)
		}
		value, err := vm.loadBitField(addr.Int, bf, ins.Type)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("bit-field load failed", err)
		}
		vm.stack = append(vm.stack, value)
	case bytecode.OpBitFieldStore:
		value, err := vm.pop(ins.Type)
		if err != nil {
			return ExitStatus{}, true, err
		}
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		bf, err := vm.bitFieldLayout(ins)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid bit-field", err)
		}
		if err := vm.storeBitField(addr.Int, bf, value); err != nil {
			return ExitStatus{}, true, vm.trapWithCause("bit-field store failed", err)
		}
	case bytecode.OpPtrAdd:
		index, err := vm.popInteger()
		if err != nil {
			return ExitStatus{}, true, err
		}
		base, err := vm.popPointer()
		if err != nil {
			return ExitStatus{}, true, err
		}
		out, err := addPointerIndex(base.Int, index, ins.Size)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("pointer add failed", err)
		}
		vm.stack = append(vm.stack, UIntValue(base.Type, out))
	case bytecode.OpPtrAddDynamic:
		stride, err := vm.pop(bytecode.TypeI64)
		if err != nil {
			return ExitStatus{}, true, err
		}
		index, err := vm.popInteger()
		if err != nil {
			return ExitStatus{}, true, err
		}
		base, err := vm.popPointer()
		if err != nil {
			return ExitStatus{}, true, err
		}
		out, err := addPointerIndex(base.Int, index, signedInt(stride))
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("dynamic pointer add failed", err)
		}
		vm.stack = append(vm.stack, UIntValue(base.Type, out))
	case bytecode.OpPtrDiff:
		right, err := vm.popPointer()
		if err != nil {
			return ExitStatus{}, true, err
		}
		left, err := vm.popPointer()
		if err != nil {
			return ExitStatus{}, true, err
		}
		diff, err := pointerDiff(left.Int, right.Int, ins.Size)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("pointer diff failed", err)
		}
		vm.stack = append(vm.stack, IntValue(bytecode.TypeI64, diff))
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
	case bytecode.OpCall:
		args, err := vm.popCallArgs(ins.Sig, ins.Argc)
		if err != nil {
			return ExitStatus{}, true, err
		}
		st, done, err := vm.invokeGlobal(ctx, ins.Global, ins.Sig, args)
		if done || err != nil {
			return st, done, err
		}
	case bytecode.OpCallIndirect:
		args, err := vm.popCallArgs(ins.Sig, ins.Argc)
		if err != nil {
			return ExitStatus{}, true, err
		}
		callee, err := vm.popPointer()
		if err != nil {
			return ExitStatus{}, true, err
		}
		if cl, ok := vm.closures[callee.Int]; ok {
			fullArgs := append(append([]Value(nil), args...), cl.captures...)
			st, done, err := vm.invokeGlobal(ctx, cl.global, cl.sig, fullArgs)
			if done || err != nil {
				return st, done, err
			}
			break
		}
		globalID, err := vm.program.FuncGlobalByAddress(callee.Int)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid indirect call target", err)
		}
		st, done, err := vm.invokeGlobal(ctx, globalID, ins.Sig, args)
		if done || err != nil {
			return st, done, err
		}
	case bytecode.OpMakeClosure:
		captures, err := vm.popClosureCaptures(ins.Sig, ins.Argc)
		if err != nil {
			return ExitStatus{}, true, err
		}
		addr, err := vm.makeClosure(ins.Global, ins.Sig, captures)
		if err != nil {
			return ExitStatus{}, true, err
		}
		vm.stack = append(vm.stack, PtrValue(addr))
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
			if err := vm.popFrame(); err != nil {
				return ExitStatus{}, true, err
			}
			return ExitStatus{Code: code}, true, nil
		}
		if err := vm.popFrame(); err != nil {
			return ExitStatus{}, true, err
		}
		vm.stack = append(vm.stack, v)
	case bytecode.OpReturnVoid:
		if len(vm.frames) == 1 {
			return ExitStatus{}, true, vm.trap("void return from entry function")
		}
		if err := vm.popFrame(); err != nil {
			return ExitStatus{}, true, err
		}
	case bytecode.OpReturnObject:
		addr, err := vm.pop(bytecode.TypeObjectAddr)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if len(vm.frames) == 1 {
			return ExitStatus{}, true, vm.trap("object return from entry function")
		}
		retAddr, err := vm.copyReturnObject(ins.Object, addr.Int)
		if err != nil {
			return ExitStatus{}, true, err
		}
		if err := vm.popFrame(); err != nil {
			return ExitStatus{}, true, err
		}
		vm.stack = append(vm.stack, ObjectAddrValue(retAddr))
	case bytecode.OpUnreachable:
		return ExitStatus{}, true, vm.trap("unreachable")
	case bytecode.OpVaStart:
		if !vm.program.module.Sigs[fr.fn.Sig].Variadic {
			return ExitStatus{}, true, vm.trap("va_start in non-variadic function")
		}
		fr.vaLists[ins.Slot] = 0
		fr.activeVaList = ins.Slot
		fr.hasActiveVa = true
	case bytecode.OpVaArg:
		if !fr.hasActiveVa {
			return ExitStatus{}, true, vm.trap("va_arg without active va_list")
		}
		cursor := fr.vaLists[fr.activeVaList]
		if cursor < 0 || cursor >= len(fr.variadicArgs) {
			return ExitStatus{}, true, vm.trap("va_arg past end of arguments")
		}
		v := fr.variadicArgs[cursor]
		if v.Type != ins.Type {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("va_arg has type %s, want %s", v.Type, ins.Type))
		}
		fr.vaLists[fr.activeVaList] = cursor + 1
		vm.stack = append(vm.stack, v)
	case bytecode.OpVaEnd:
		delete(fr.vaLists, ins.Slot)
		if fr.hasActiveVa && fr.activeVaList == ins.Slot {
			fr.activeVaList = -1
			fr.hasActiveVa = false
		}
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

func (fr *frame) dynamicObject(objectID int) (bytecode.DynamicObject, bool) {
	for _, object := range fr.fn.DynamicObjects {
		if object.ID == objectID {
			return object, true
		}
	}
	return bytecode.DynamicObject{}, false
}

func (vm *VM) popFrame() error {
	if len(vm.frames) == 0 {
		return vm.trap("empty call stack")
	}
	fr := &vm.frames[len(vm.frames)-1]
	for objectID, addr := range fr.dynamicObjects {
		if err := vm.program.Memory().Free(addr, blockDynamic); err != nil {
			return vm.trapWithCause(fmt.Sprintf("dynamic object %d free failed", objectID), err)
		}
	}
	for _, addr := range fr.closures {
		delete(vm.closures, addr)
		if err := vm.program.Memory().Free(addr, blockLocal); err != nil {
			return vm.trapWithCause(fmt.Sprintf("closure %x free failed", addr), err)
		}
	}
	for objectID, addr := range fr.localObjects {
		if err := vm.program.Memory().Free(addr, blockLocal); err != nil {
			return vm.trapWithCause(fmt.Sprintf("local object %d free failed", objectID), err)
		}
	}
	vm.frames = vm.frames[:len(vm.frames)-1]
	return nil
}

func (vm *VM) copyReturnObject(objectID int, src uint64) (uint64, error) {
	if len(vm.frames) == 0 {
		return 0, vm.trap("empty call stack")
	}
	fr := &vm.frames[len(vm.frames)-1]
	var object bytecode.LocalObject
	found := false
	for _, candidate := range fr.fn.Objects {
		if candidate.ID == objectID {
			object = candidate
			found = true
			break
		}
	}
	if !found {
		return 0, vm.trap(fmt.Sprintf("return object %d not found in function %s", objectID, fr.fn.Name))
	}
	dst, err := vm.program.Memory().TryAlloc(fmt.Sprintf("return:%s:%s", fr.fn.Name, object.Name), object.Size, object.Align, false, blockGlobal)
	if err != nil {
		return 0, vm.trapWithCause(fmt.Sprintf("return object %d allocation failed", objectID), err)
	}
	if err := vm.program.Memory().Copy(dst, src, object.Size); err != nil {
		return 0, vm.trapWithCause(fmt.Sprintf("return object %d copy failed", objectID), err)
	}
	return dst, nil
}

func (vm *VM) cleanupFrames() error {
	for len(vm.frames) != 0 {
		if err := vm.popFrame(); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VM) popCallArgs(sigID, argc int) ([]Value, error) {
	if vm.program == nil || vm.program.module == nil {
		return nil, vm.trap("nil program")
	}
	if sigID < 0 || sigID >= len(vm.program.module.Sigs) {
		return nil, vm.trap(fmt.Sprintf("invalid call signature %d", sigID))
	}
	if argc < 0 {
		return nil, vm.trap(fmt.Sprintf("negative call argc %d", argc))
	}

	sig := vm.program.module.Sigs[sigID]
	if !sig.Variadic && argc != len(sig.Params) {
		return nil, vm.trap(fmt.Sprintf("call argc %d does not match signature parameter count %d", argc, len(sig.Params)))
	}
	if sig.Variadic && argc < len(sig.Params) {
		return nil, vm.trap(fmt.Sprintf("variadic call argc %d is less than signature parameter count %d", argc, len(sig.Params)))
	}

	args := make([]Value, argc)
	for i := argc - 1; i >= len(sig.Params); i-- {
		v, err := vm.popAnyValue()
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	for i := len(sig.Params) - 1; i >= 0; i-- {
		v, err := vm.pop(sig.Params[i])
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return args, nil
}

func (vm *VM) popClosureCaptures(sigID, argc int) ([]Value, error) {
	if vm.program == nil || vm.program.module == nil {
		return nil, vm.trap("nil program")
	}
	if sigID < 0 || sigID >= len(vm.program.module.Sigs) {
		return nil, vm.trap(fmt.Sprintf("invalid closure signature %d", sigID))
	}
	if argc < 0 {
		return nil, vm.trap(fmt.Sprintf("negative closure argc %d", argc))
	}

	sig := vm.program.module.Sigs[sigID]
	if argc > len(sig.Params) {
		return nil, vm.trap(fmt.Sprintf("closure argc %d exceeds signature parameter count %d", argc, len(sig.Params)))
	}

	args := make([]Value, argc)
	start := len(sig.Params) - argc
	for i := argc - 1; i >= 0; i-- {
		v, err := vm.pop(sig.Params[start+i])
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return args, nil
}

func (vm *VM) makeClosure(globalID, sigID int, captures []Value) (uint64, error) {
	if len(vm.frames) == 0 {
		return 0, vm.trap("empty call stack")
	}
	g, err := vm.program.global(globalID)
	if err != nil {
		return 0, vm.trapWithCause("invalid closure target", err)
	}
	if g.Sig != sigID {
		return 0, vm.trap(fmt.Sprintf("closure signature %d does not match global %d signature %d", sigID, globalID, g.Sig))
	}
	if g.Kind != bytecode.GlobalFunc {
		return 0, vm.trap(fmt.Sprintf("closure target global %d is not a function", globalID))
	}

	fr := &vm.frames[len(vm.frames)-1]
	addr, err := vm.program.Memory().TryAlloc(
		fmt.Sprintf("closure:%s:%d", fr.fn.Name, len(fr.closures)),
		vm.program.module.Target.PointerSize,
		vm.program.module.Target.PointerAlign,
		false,
		blockLocal,
	)
	if err != nil {
		return 0, vm.trapWithCause("closure allocation failed", err)
	}
	vm.closures[addr] = closure{
		global:   globalID,
		sig:      sigID,
		captures: append([]Value(nil), captures...),
	}
	fr.closures = append(fr.closures, addr)
	return addr, nil
}

func (vm *VM) invokeGlobal(ctx context.Context, globalID, sigID int, args []Value) (ExitStatus, bool, error) {
	g, err := vm.program.global(globalID)
	if err != nil {
		return ExitStatus{}, true, vm.trapWithCause("invalid call target", err)
	}
	if g.Sig != sigID {
		return ExitStatus{}, true, vm.trap(fmt.Sprintf("call signature %d does not match global %d signature %d", sigID, globalID, g.Sig))
	}

	switch g.Kind {
	case bytecode.GlobalFunc:
		if g.Func < 0 || g.Func >= len(vm.program.module.Functions) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("global %d references invalid function id %d", globalID, g.Func))
		}
		if vm.program.module.Functions[g.Func].Sig != sigID {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("call signature %d does not match function %d signature %d", sigID, g.Func, vm.program.module.Functions[g.Func].Sig))
		}
		if err := vm.pushFrame(g.Func, args); err != nil {
			return ExitStatus{}, true, err
		}
		return ExitStatus{}, false, nil
	case bytecode.GlobalExtern:
		if !isExternFunction(g) {
			return ExitStatus{}, true, vm.trap(fmt.Sprintf("global %d is not an extern function", globalID))
		}
		fn, err := vm.program.ExternByGlobal(globalID)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause("invalid extern call target", err)
		}
		ret, exit, err := fn(ctx, vm.program.ExternContext(), args)
		if err != nil {
			return ExitStatus{}, true, vm.trapWithCause(fmt.Sprintf("extern %s failed", g.Extern.Name), err)
		}
		if exit != nil {
			return *exit, true, nil
		}
		sig := vm.program.module.Sigs[sigID]
		if sig.Ret != bytecode.TypeVoid {
			if ret.Type != sig.Ret {
				return ExitStatus{}, true, vm.trap(fmt.Sprintf("extern %s returned %s, want %s", g.Extern.Name, ret.Type, sig.Ret))
			}
			vm.stack = append(vm.stack, ret)
		}
		return ExitStatus{}, false, nil
	default:
		return ExitStatus{}, true, vm.trap(fmt.Sprintf("global %d is not callable", globalID))
	}
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
	if ins.Type == bytecode.TypePtr || ins.Type == bytecode.TypeObjectAddr {
		return vm.pointerBinary(ins, l, r)
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

func (vm *VM) pointerBinary(ins bytecode.Instr, l, r Value) error {
	var out Value
	switch ins.Binary {
	case bytecode.BinEq:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Int == r.Int)))
	case bytecode.BinNe:
		out = UIntValue(bytecode.TypeBool, uint64(boolInt(l.Int != r.Int)))
	default:
		return vm.trap(fmt.Sprintf("unsupported pointer binary op %s", ins.Binary))
	}
	vm.stack = append(vm.stack, out)
	return nil
}

func (vm *VM) floatBinary(ins bytecode.Instr, l, r Value) error {
	var out Value
	switch ins.Binary {
	case bytecode.BinAdd:
		out = floatResult(ins.Type, l.Float+r.Float)
	case bytecode.BinSub:
		out = floatResult(ins.Type, l.Float-r.Float)
	case bytecode.BinMul:
		out = floatResult(ins.Type, l.Float*r.Float)
	case bytecode.BinDivS:
		out = floatResult(ins.Type, l.Float/r.Float)
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

func floatResult(t bytecode.ValueType, v float64) Value {
	if t == bytecode.TypeF32 {
		v = float64(float32(v))
	}
	return FloatValue(t, v)
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
	case bytecode.CastFExt, bytecode.CastFTrunc:
		if !isFloatType(ins.Type) || !isFloatType(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported float cast %s->%s", ins.Type, ins.Type2))
		}
		vm.stack = append(vm.stack, floatResult(ins.Type2, v.Float))
	case bytecode.CastIntToFloat:
		if !isIntegerLike(ins.Type) || !isFloatType(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported int-to-float cast %s->%s", ins.Type, ins.Type2))
		}
		f := float64(signedInt(v))
		if isUnsignedIntegerType(ins.Type) {
			f = float64(unsignedInt(v))
		}
		vm.stack = append(vm.stack, floatResult(ins.Type2, f))
	case bytecode.CastFloatToInt:
		if !isFloatType(ins.Type) || !isIntegerLike(ins.Type2) {
			return vm.trap(fmt.Sprintf("unsupported float-to-int cast %s->%s", ins.Type, ins.Type2))
		}
		if isUnsignedIntegerType(ins.Type2) {
			vm.stack = append(vm.stack, normalizeInt(UIntValue(ins.Type2, uint64(v.Float))))
		} else {
			vm.stack = append(vm.stack, normalizeInt(IntValue(ins.Type2, int64(v.Float))))
		}
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

func (vm *VM) bitFieldLayout(ins bytecode.Instr) (bytecode.BitFieldLayout, error) {
	if ins.Layout < 0 || ins.Layout >= len(vm.program.module.Layouts) {
		return bytecode.BitFieldLayout{}, fmt.Errorf("invalid layout %d", ins.Layout)
	}
	layout := vm.program.module.Layouts[ins.Layout]
	if ins.Field < 0 || ins.Field >= len(layout.Bit) {
		return bytecode.BitFieldLayout{}, fmt.Errorf("invalid bit-field %d in layout %d", ins.Field, ins.Layout)
	}
	return layout.Bit[ins.Field], nil
}

func (vm *VM) loadBitField(base uint64, bf bytecode.BitFieldLayout, outType bytecode.ValueType) (Value, error) {
	addr, err := addSignedOffset(base, bf.ByteOffset)
	if err != nil {
		return Value{}, err
	}
	raw, err := vm.program.Memory().Load(addr, bf.Container, 1)
	if err != nil {
		return Value{}, err
	}
	value := (unsignedInt(raw) >> uint(bf.BitOffset)) & bitFieldMask(bf.Width)
	if bf.Signed && bf.Width > 0 {
		sign := uint64(1) << uint(bf.Width-1)
		if value&sign != 0 {
			value |= ^bitFieldMask(bf.Width)
		}
	}
	return normalizeInt(UIntValue(outType, value)), nil
}

func (vm *VM) storeBitField(base uint64, bf bytecode.BitFieldLayout, value Value) error {
	addr, err := addSignedOffset(base, bf.ByteOffset)
	if err != nil {
		return err
	}
	raw, err := vm.program.Memory().Load(addr, bf.Container, 1)
	if err != nil {
		return err
	}
	mask := bitFieldMask(bf.Width) << uint(bf.BitOffset)
	next := (unsignedInt(raw) &^ mask) | ((unsignedInt(value) << uint(bf.BitOffset)) & mask)
	return vm.program.Memory().Store(addr, bf.Container, 1, UIntValue(bf.Container, next))
}

func bitFieldMask(width int) uint64 {
	if width <= 0 {
		return 0
	}
	if width >= 64 {
		return ^uint64(0)
	}
	return (uint64(1) << uint(width)) - 1
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

func addSignedOffset(base uint64, offset int64) (uint64, error) {
	if offset >= 0 {
		delta := uint64(offset)
		if delta > math.MaxUint64-base {
			return 0, fmt.Errorf("%#x + %d overflows", base, offset)
		}
		return base + delta, nil
	}
	delta := absInt64(offset)
	if delta > base {
		return 0, fmt.Errorf("%#x + %d underflows", base, offset)
	}
	return base - delta, nil
}

func addPointerIndex(base uint64, index Value, elemSize int64) (uint64, error) {
	neg, delta, err := scaledIndex(index, elemSize)
	if err != nil {
		return 0, err
	}
	if neg {
		if delta > base {
			return 0, fmt.Errorf("%#x - %d underflows", base, delta)
		}
		return base - delta, nil
	}
	if delta > math.MaxUint64-base {
		return 0, fmt.Errorf("%#x + %d overflows", base, delta)
	}
	return base + delta, nil
}

func scaledIndex(index Value, elemSize int64) (bool, uint64, error) {
	if elemSize <= 0 {
		return false, 0, fmt.Errorf("invalid element size %d", elemSize)
	}
	elem := uint64(elemSize)
	neg := false
	var mag uint64
	if isSignedIntegerType(index.Type) {
		n := signedInt(index)
		if n < 0 {
			neg = true
			mag = absInt64(n)
		} else {
			mag = uint64(n)
		}
	} else if isUnsignedIntegerType(index.Type) {
		mag = unsignedInt(index)
	} else {
		return false, 0, fmt.Errorf("unsupported pointer index type %s", index.Type)
	}
	if mag != 0 && elem > math.MaxUint64/mag {
		return false, 0, fmt.Errorf("pointer index %d * element size %d overflows", mag, elemSize)
	}
	return neg, mag * elem, nil
}

func pointerDiff(left, right uint64, elemSize int64) (int64, error) {
	if elemSize <= 0 {
		return 0, fmt.Errorf("invalid element size %d", elemSize)
	}
	elem := uint64(elemSize)
	if left >= right {
		delta := left - right
		if delta%elem != 0 {
			return 0, fmt.Errorf("pointer difference %d is not divisible by element size %d", delta, elemSize)
		}
		quotient := delta / elem
		if quotient > math.MaxInt64 {
			return 0, fmt.Errorf("pointer difference %d exceeds i64 range", quotient)
		}
		return int64(quotient), nil
	}
	delta := right - left
	if delta%elem != 0 {
		return 0, fmt.Errorf("pointer difference -%d is not divisible by element size %d", delta, elemSize)
	}
	quotient := delta / elem
	minInt64Magnitude := uint64(math.MaxInt64) + 1
	if quotient > minInt64Magnitude {
		return 0, fmt.Errorf("pointer difference -%d exceeds i64 range", quotient)
	}
	if quotient == minInt64Magnitude {
		return math.MinInt64, nil
	}
	return -int64(quotient), nil
}

func absInt64(v int64) uint64 {
	if v >= 0 {
		return uint64(v)
	}
	return uint64(-(v + 1)) + 1
}

func (vm *VM) popInteger() (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, vm.trap("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	if !isIntegerLike(v.Type) {
		return Value{}, vm.trap(fmt.Sprintf("stack value has type %s, want integer", v.Type))
	}
	vm.stack = vm.stack[:len(vm.stack)-1]
	return v, nil
}

func (vm *VM) popPointer() (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, vm.trap("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	if !isPointerType(v.Type) {
		return Value{}, vm.trap(fmt.Sprintf("stack value has type %s, want pointer", v.Type))
	}
	vm.stack = vm.stack[:len(vm.stack)-1]
	return v, nil
}

func (vm *VM) popAnyValue() (Value, error) {
	if len(vm.stack) == 0 {
		return Value{}, vm.trap("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	if v.Type == bytecode.TypeVoid {
		return Value{}, vm.trap("stack value has type void, want non-void")
	}
	vm.stack = vm.stack[:len(vm.stack)-1]
	return v, nil
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
