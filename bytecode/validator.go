package bytecode

import "fmt"

func ValidateModule(m *Module) error {
	if m == nil {
		return fmt.Errorf("module is nil")
	}
	if m.Target.Name == "" {
		return fmt.Errorf("target name is empty")
	}
	funcCount := len(m.Functions)
	for i, g := range m.Globals {
		if g.ID != i {
			return fmt.Errorf("global index %d has id %d", i, g.ID)
		}
		if g.Name == "" {
			return fmt.Errorf("global %d has empty name", i)
		}
		if g.Kind == GlobalFunc && (g.Func < 0 || g.Func >= funcCount) {
			return fmt.Errorf("global %q references invalid function index %d", g.Name, g.Func)
		}
	}
	for i, sig := range m.Sigs {
		if sig.ID != i {
			return fmt.Errorf("signature index %d has id %d", i, sig.ID)
		}
	}
	for i := range m.Functions {
		if err := validateFunction(m, i, &m.Functions[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateFunction(m *Module, index int, f *Function) error {
	if f.ID != index {
		return fmt.Errorf("function index %d has id %d", index, f.ID)
	}
	if f.Sig < 0 || f.Sig >= len(m.Sigs) {
		return fmt.Errorf("function %q references invalid signature %d", f.Name, f.Sig)
	}
	if f.GlobalID < 0 || f.GlobalID >= len(m.Globals) {
		return fmt.Errorf("function %q references invalid global %d", f.Name, f.GlobalID)
	}
	g := m.Globals[f.GlobalID]
	if g.Kind != GlobalFunc {
		return fmt.Errorf("function %q references non-function global %d", f.Name, f.GlobalID)
	}
	if g.Func != f.ID {
		return fmt.Errorf("function %q global %d points to function %d", f.Name, f.GlobalID, g.Func)
	}
	ret := m.Sigs[f.Sig].Ret
	stack := []ValueType{}
	for pc, ins := range f.Instrs {
		next, err := validateInstrStack(stack, ins, ret)
		if err != nil {
			return fmt.Errorf("function %q pc %d: %w", f.Name, pc, err)
		}
		stack = next
	}
	return nil
}

func validateInstrStack(stack []ValueType, ins Instr, ret ValueType) ([]ValueType, error) {
	pop := func(want ValueType) error {
		if len(stack) == 0 {
			return fmt.Errorf("%v stack underflow", ins.Op)
		}
		got := stack[len(stack)-1]
		if want != TypeVoid && got != want {
			return fmt.Errorf("%v wants %s, got %s", ins.Op, want, got)
		}
		stack = stack[:len(stack)-1]
		return nil
	}
	push := func(t ValueType) {
		if t != TypeVoid {
			stack = append(stack, t)
		}
	}
	switch ins.Op {
	case OpConst, OpAddrString, OpAddrGlobal, OpAddrFunc, OpLoadConst, OpLoadLocal, OpAddrLocalObject:
		push(ins.Type)
	case OpStoreLocal:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
	case OpPop:
		if err := pop(TypeVoid); err != nil {
			return nil, err
		}
	case OpSwap:
		if len(stack) < 2 {
			return nil, fmt.Errorf("swap stack underflow")
		}
		stack[len(stack)-1], stack[len(stack)-2] = stack[len(stack)-2], stack[len(stack)-1]
	case OpBinary:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if isCompare(ins.Binary) {
			push(TypeBool)
		} else {
			push(ins.Type)
		}
	case OpReturn:
		if ret == TypeVoid {
			return nil, fmt.Errorf("return value in void function")
		}
		if ins.Type != ret {
			return nil, fmt.Errorf("return type %s does not match signature return %s", ins.Type, ret)
		}
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
	case OpReturnVoid:
		if ret != TypeVoid {
			return nil, fmt.Errorf("return void in %s function", ret)
		}
		if len(stack) != 0 {
			return nil, fmt.Errorf("return void with non-empty stack")
		}
	case OpDup:
		if len(stack) == 0 {
			return nil, fmt.Errorf("dup stack underflow")
		}
		push(stack[len(stack)-1])
	}
	return stack, nil
}
