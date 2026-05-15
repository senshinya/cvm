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
		if err := validateFunction(m, &m.Functions[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateFunction(m *Module, f *Function) error {
	if f.ID < 0 || f.ID >= len(m.Functions) {
		return fmt.Errorf("function %q has invalid id %d", f.Name, f.ID)
	}
	if f.Sig < 0 || f.Sig >= len(m.Sigs) {
		return fmt.Errorf("function %q references invalid signature %d", f.Name, f.Sig)
	}
	if f.GlobalID < 0 || f.GlobalID >= len(m.Globals) {
		return fmt.Errorf("function %q references invalid global %d", f.Name, f.GlobalID)
	}
	stack := []ValueType{}
	for pc, ins := range f.Instrs {
		next, err := validateInstrStack(stack, ins)
		if err != nil {
			return fmt.Errorf("function %q pc %d: %w", f.Name, pc, err)
		}
		stack = next
	}
	return nil
}

func validateInstrStack(stack []ValueType, ins Instr) ([]ValueType, error) {
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
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
	case OpReturnVoid:
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
