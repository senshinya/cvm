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
		if param.Slot > maxSlot {
			maxSlot = param.Slot
		}
	}
	for _, local := range fn.Locals {
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
	if err := ctx.Err(); err != nil {
		return ExitStatus{}, true, vm.trapWithCause("context canceled", err)
	}
	if vm.limit > 0 && vm.steps >= vm.limit {
		return ExitStatus{}, true, vm.trap("step limit exceeded")
	}
	if len(vm.frames) == 0 {
		return ExitStatus{}, true, vm.trap("empty call stack")
	}

	fr := &vm.frames[len(vm.frames)-1]
	if fr.pc < 0 || fr.pc >= len(fr.fn.Instrs) {
		return ExitStatus{}, true, vm.trap("program counter out of range")
	}

	ins := fr.fn.Instrs[fr.pc]
	fr.pc++
	vm.steps++

	switch ins.Op {
	case bytecode.OpConst:
		vm.stack = append(vm.stack, constValue(ins))
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
		vm.frames = vm.frames[:len(vm.frames)-1]
		if len(vm.frames) == 0 {
			return ExitStatus{}, true, nil
		}
	case bytecode.OpLabel:
		// Labels are markers for control-flow instructions; execution falls through.
	default:
		return ExitStatus{}, true, vm.trap(fmt.Sprintf("unsupported opcode %s", ins.Op))
	}
	return ExitStatus{}, false, nil
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
	pc := fr.pc
	if pc > 0 {
		pc--
	}
	err.PC = pc
	if pc >= 0 && pc < len(fr.fn.Instrs) {
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
