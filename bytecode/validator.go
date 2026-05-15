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
		if g.Kind == GlobalFunc {
			if g.Func < 0 || g.Func >= funcCount {
				return fmt.Errorf("global %q references invalid function index %d", g.Name, g.Func)
			}
			if m.Functions[g.Func].GlobalID != g.ID {
				return fmt.Errorf("global %q points to function %d with global id %d", g.Name, g.Func, m.Functions[g.Func].GlobalID)
			}
		}
	}
	for i, sig := range m.Sigs {
		if sig.ID != i {
			return fmt.Errorf("signature index %d has id %d", i, sig.ID)
		}
	}
	for i, s := range m.Strings {
		if s.ID != i {
			return fmt.Errorf("string index %d has id %d", i, s.ID)
		}
	}
	for i, l := range m.Layouts {
		if l.ID != i {
			return fmt.Errorf("layout index %d has id %d", i, l.ID)
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
	labels := map[int]Label{}
	for _, l := range f.Labels {
		if _, exists := labels[l.ID]; exists {
			return fmt.Errorf("function %q duplicate label L%d", f.Name, l.ID)
		}
		labels[l.ID] = l
	}
	labelPCs := map[int]int{}
	for pc, ins := range f.Instrs {
		if ins.Op != OpLabel {
			continue
		}
		if _, ok := labels[ins.Label]; !ok {
			return fmt.Errorf("function %q pc %d: label instruction references missing label L%d", f.Name, pc, ins.Label)
		}
		if prev, exists := labelPCs[ins.Label]; exists {
			return fmt.Errorf("function %q pc %d: duplicate label instruction L%d previously at pc %d", f.Name, pc, ins.Label, prev)
		}
		labelPCs[ins.Label] = pc
	}
	locals := map[int]ValueType{}
	for _, p := range f.Params {
		locals[p.Slot] = p.Type
	}
	for _, l := range f.Locals {
		if _, exists := locals[l.ID]; exists {
			return fmt.Errorf("function %q duplicate local slot %d", f.Name, l.ID)
		}
		locals[l.ID] = l.Type
	}
	objects := map[int]LocalObject{}
	for _, o := range f.Objects {
		if _, exists := objects[o.ID]; exists {
			return fmt.Errorf("function %q duplicate object %d", f.Name, o.ID)
		}
		if o.Layout < 0 || o.Layout >= len(m.Layouts) {
			return fmt.Errorf("function %q object %d references invalid layout %d", f.Name, o.ID, o.Layout)
		}
		objects[o.ID] = o
	}
	dynamicObjects := map[int]DynamicObject{}
	for _, o := range f.DynamicObjects {
		if _, exists := dynamicObjects[o.ID]; exists {
			return fmt.Errorf("function %q duplicate dynamic object %d", f.Name, o.ID)
		}
		if o.Layout < 0 || o.Layout >= len(m.Layouts) {
			return fmt.Errorf("function %q dynamic object %d references invalid layout %d", f.Name, o.ID, o.Layout)
		}
		dynamicObjects[o.ID] = o
	}
	ret := m.Sigs[f.Sig].Ret
	stack := []ValueType{}
	terminalReturn := false
	for pc, ins := range f.Instrs {
		if err := validateInstrRefs(m, ins, labels, labelPCs, locals, objects, dynamicObjects); err != nil {
			return fmt.Errorf("function %q pc %d: %w", f.Name, pc, err)
		}
		if ins.Op == OpLabel {
			label := labels[ins.Label]
			if terminalReturn {
				stack = append([]ValueType(nil), label.Stack...)
			} else if !sameStack(stack, label.Stack) {
				return fmt.Errorf("function %q pc %d: label L%d stack %v does not match declared %v", f.Name, pc, ins.Label, stack, label.Stack)
			}
			if label.Statement && len(stack) != 0 {
				return fmt.Errorf("function %q pc %d: label L%d requires empty stack, got %d values", f.Name, pc, ins.Label, len(stack))
			}
			terminalReturn = false
		}
		next, err := validateInstrStack(m, stack, ins, ret, labels)
		if err != nil {
			return fmt.Errorf("function %q pc %d: %w", f.Name, pc, err)
		}
		stack = next
		terminalReturn = ins.Op == OpReturn || ins.Op == OpReturnVoid || ins.Op == OpReturnObject || ins.Op == OpJump || ins.Op == OpUnreachable
	}
	if !terminalReturn {
		if len(stack) != 0 {
			return fmt.Errorf("function %q ends without terminal return and has non-empty stack", f.Name)
		}
		return fmt.Errorf("function %q missing terminal return", f.Name)
	}
	return nil
}

func validateInstrRefs(m *Module, ins Instr, labels map[int]Label, labelPCs map[int]int, locals map[int]ValueType, objects map[int]LocalObject, dynamicObjects map[int]DynamicObject) error {
	requireLabel := func(label int) error {
		if _, ok := labels[label]; !ok {
			return fmt.Errorf("%v references missing label L%d", ins.Op, label)
		}
		if _, ok := labelPCs[label]; !ok {
			return fmt.Errorf("%v references unmarked label L%d", ins.Op, label)
		}
		return nil
	}
	requireGlobal := func(global int) error {
		if global < 0 || global >= len(m.Globals) {
			return fmt.Errorf("%v references invalid global %d", ins.Op, global)
		}
		return nil
	}
	requireString := func(id int64) error {
		if id < 0 || id >= int64(len(m.Strings)) {
			return fmt.Errorf("%v references invalid string %d", ins.Op, id)
		}
		return nil
	}
	requireSig := func(sig int) error {
		if sig < 0 || sig >= len(m.Sigs) {
			return fmt.Errorf("%v references invalid signature %d", ins.Op, sig)
		}
		return nil
	}
	requireObject := func(object int) error {
		if _, ok := objects[object]; !ok {
			return fmt.Errorf("%v references invalid object %d", ins.Op, object)
		}
		return nil
	}
	requireDynamicObject := func(object int) error {
		if _, ok := dynamicObjects[object]; !ok {
			return fmt.Errorf("%v references invalid dynamic object %d", ins.Op, object)
		}
		return nil
	}
	requireLocal := func(slot int, t ValueType) error {
		got, ok := locals[slot]
		if !ok {
			return fmt.Errorf("%v references invalid local slot %d", ins.Op, slot)
		}
		if got != t {
			return fmt.Errorf("%v local slot %d has type %s, got instruction type %s", ins.Op, slot, got, t)
		}
		return nil
	}
	requireField := func(layoutID, fieldID int, bit bool) error {
		if layoutID < 0 || layoutID >= len(m.Layouts) {
			return fmt.Errorf("%v references invalid layout %d", ins.Op, layoutID)
		}
		layout := m.Layouts[layoutID]
		if bit {
			if fieldID < 0 || fieldID >= len(layout.Bit) || layout.Bit[fieldID].ID != fieldID {
				return fmt.Errorf("%v references invalid bitfield %d in layout %d", ins.Op, fieldID, layoutID)
			}
			return nil
		}
		if fieldID < 0 || fieldID >= len(layout.Fields) || layout.Fields[fieldID].ID != fieldID {
			return fmt.Errorf("%v references invalid field %d in layout %d", ins.Op, fieldID, layoutID)
		}
		return nil
	}

	switch ins.Op {
	case OpAddrString:
		return requireString(ins.Int)
	case OpAddrGlobal, OpLoadConst:
		return requireGlobal(ins.Global)
	case OpAddrFunc:
		if err := requireGlobal(ins.Global); err != nil {
			return err
		}
		g := m.Globals[ins.Global]
		if g.Kind != GlobalFunc && (g.Kind != GlobalExtern || g.Size != 0 || g.Align != 0) {
			return fmt.Errorf("%v references non-function global %d", ins.Op, ins.Global)
		}
	case OpLoadLocal, OpStoreLocal:
		return requireLocal(ins.Slot, ins.Type)
	case OpAddrLocalObject:
		return requireObject(ins.Object)
	case OpAllocDynamicObject, OpFreeDynamicObject, OpDynamicObjectAddr:
		return requireDynamicObject(ins.Object)
	case OpFieldAddr:
		return requireField(ins.Layout, ins.Field, false)
	case OpBitFieldLoad, OpBitFieldStore:
		return requireField(ins.Layout, ins.Field, true)
	case OpLabel, OpJump, OpJumpIfZero, OpJumpIfNonZero:
		return requireLabel(ins.Label)
	case OpSwitch:
		if err := requireLabel(ins.Label); err != nil {
			return err
		}
		for _, c := range ins.Labels {
			if err := requireLabel(c.Label); err != nil {
				return err
			}
		}
	case OpCall:
		if err := requireSig(ins.Sig); err != nil {
			return err
		}
		if err := requireGlobal(ins.Global); err != nil {
			return err
		}
		g := m.Globals[ins.Global]
		if g.Kind != GlobalFunc && (g.Kind != GlobalExtern || g.Size != 0 || g.Align != 0) {
			return fmt.Errorf("%v references non-function global %d", ins.Op, ins.Global)
		}
		if g.Kind == GlobalFunc {
			fn := g.Func
			if fn < 0 || fn >= len(m.Functions) {
				return fmt.Errorf("%v references invalid function %d", ins.Op, fn)
			}
			if m.Functions[fn].Sig != ins.Sig {
				return fmt.Errorf("%v signature %d does not match function %d signature %d", ins.Op, ins.Sig, fn, m.Functions[fn].Sig)
			}
		}
	case OpCallIndirect:
		return requireSig(ins.Sig)
	}
	return nil
}

func validateInstrStack(m *Module, stack []ValueType, ins Instr, ret ValueType, labels map[int]Label) ([]ValueType, error) {
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
	popAnyOf := func(wants ...ValueType) (ValueType, error) {
		if len(stack) == 0 {
			return TypeVoid, fmt.Errorf("%v stack underflow", ins.Op)
		}
		got := stack[len(stack)-1]
		for _, want := range wants {
			if got == want {
				stack = stack[:len(stack)-1]
				return got, nil
			}
		}
		return TypeVoid, fmt.Errorf("%v wants one of %v, got %s", ins.Op, wants, got)
	}
	popAnyValue := func() error {
		if len(stack) == 0 {
			return fmt.Errorf("%v stack underflow", ins.Op)
		}
		got := stack[len(stack)-1]
		if got == TypeVoid {
			return fmt.Errorf("%v wants non-void value, got %s", ins.Op, got)
		}
		stack = stack[:len(stack)-1]
		return nil
	}
	push := func(t ValueType) {
		if t != TypeVoid {
			stack = append(stack, t)
		}
	}
	checkTargetStack := func(labelID int) error {
		label, ok := labels[labelID]
		if !ok {
			return fmt.Errorf("%v references missing label L%d", ins.Op, labelID)
		}
		if label.Statement && len(stack) != 0 {
			return fmt.Errorf("%v to statement label L%d with non-empty stack %v", ins.Op, labelID, stack)
		}
		if !sameStack(stack, label.Stack) {
			return fmt.Errorf("%v target L%d stack %v does not match declared %v", ins.Op, labelID, stack, label.Stack)
		}
		return nil
	}
	switch ins.Op {
	case OpConst, OpAddrString, OpAddrGlobal, OpAddrFunc, OpLoadConst, OpLoadLocal, OpAddrLocalObject, OpDynamicObjectAddr:
		push(ins.Type)
	case OpStoreLocal:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
	case OpAllocDynamicObject:
		if err := pop(TypeI64); err != nil {
			return nil, err
		}
	case OpFreeDynamicObject:
	case OpLoad:
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
		push(ins.Type)
	case OpStore:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
	case OpMemCopy:
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
	case OpMemSet:
		if err := pop(TypeI32); err != nil {
			return nil, err
		}
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
	case OpOffset, OpFieldAddr:
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
		push(TypeObjectAddr)
	case OpBitFieldLoad:
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
		push(ins.Type)
	case OpBitFieldStore:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
	case OpPtrAdd:
		if _, err := popAnyOf(TypeI32, TypeI64, TypeU32, TypeU64); err != nil {
			return nil, err
		}
		base, err := popAnyOf(TypePtr, TypeObjectAddr)
		if err != nil {
			return nil, err
		}
		push(base)
	case OpPtrAddDynamic:
		if err := pop(TypeI64); err != nil {
			return nil, err
		}
		if _, err := popAnyOf(TypeI32, TypeI64, TypeU32, TypeU64); err != nil {
			return nil, err
		}
		base, err := popAnyOf(TypePtr, TypeObjectAddr)
		if err != nil {
			return nil, err
		}
		push(base)
	case OpPtrDiff:
		if _, err := popAnyOf(TypePtr, TypeObjectAddr); err != nil {
			return nil, err
		}
		if _, err := popAnyOf(TypePtr, TypeObjectAddr); err != nil {
			return nil, err
		}
		push(TypeI64)
	case OpPop:
		if err := pop(TypeVoid); err != nil {
			return nil, err
		}
	case OpSwap:
		if len(stack) < 2 {
			return nil, fmt.Errorf("%v stack underflow", ins.Op)
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
	case OpUnary:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		push(ins.Type)
	case OpCast:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		push(ins.Type2)
	case OpLabel:
	case OpJump:
		if err := checkTargetStack(ins.Label); err != nil {
			return nil, err
		}
	case OpJumpIfZero, OpJumpIfNonZero:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if err := checkTargetStack(ins.Label); err != nil {
			return nil, err
		}
	case OpSwitch:
		if err := pop(ins.Type); err != nil {
			return nil, err
		}
		if err := checkTargetStack(ins.Label); err != nil {
			return nil, err
		}
		for _, c := range ins.Labels {
			if err := checkTargetStack(c.Label); err != nil {
				return nil, err
			}
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
		if len(stack) != 0 {
			return nil, fmt.Errorf("return with non-empty stack")
		}
	case OpReturnVoid:
		if ret != TypeVoid {
			return nil, fmt.Errorf("return void in %s function", ret)
		}
		if len(stack) != 0 {
			return nil, fmt.Errorf("return void with non-empty stack")
		}
	case OpReturnObject:
		if ret != TypeObjectAddr {
			return nil, fmt.Errorf("return object in %s function", ret)
		}
		if err := pop(TypeObjectAddr); err != nil {
			return nil, err
		}
		if len(stack) != 0 {
			return nil, fmt.Errorf("return object with non-empty stack")
		}
	case OpUnreachable:
		if len(stack) != 0 {
			return nil, fmt.Errorf("unreachable with non-empty stack")
		}
	case OpCall, OpCallIndirect:
		sig := m.Sigs[ins.Sig]
		if !sig.Variadic && ins.Argc != len(sig.Params) {
			return nil, fmt.Errorf("%v argc %d does not match signature parameter count %d", ins.Op, ins.Argc, len(sig.Params))
		}
		if sig.Variadic && ins.Argc < len(sig.Params) {
			return nil, fmt.Errorf("%v argc %d is less than signature parameter count %d", ins.Op, ins.Argc, len(sig.Params))
		}
		for i := 0; i < ins.Argc-len(sig.Params); i++ {
			if err := popAnyValue(); err != nil {
				return nil, err
			}
		}
		for i := len(sig.Params) - 1; i >= 0; i-- {
			if err := pop(sig.Params[i]); err != nil {
				return nil, err
			}
		}
		if ins.Op == OpCallIndirect {
			if _, err := popAnyOf(TypePtr, TypeObjectAddr); err != nil {
				return nil, err
			}
		}
		push(sig.Ret)
	case OpDup:
		if len(stack) == 0 {
			return nil, fmt.Errorf("%v stack underflow", ins.Op)
		}
		push(stack[len(stack)-1])
	case OpVaStart, OpVaArg, OpVaEnd:
		return nil, fmt.Errorf("unsupported opcode %v", ins.Op)
	default:
		return nil, fmt.Errorf("unsupported opcode %v", ins.Op)
	}
	return stack, nil
}

func sameStack(a, b []ValueType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
