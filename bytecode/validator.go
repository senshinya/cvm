package bytecode

import "fmt"

func ValidateModule(m *Module) error {
	if m == nil {
		return fmt.Errorf("module is nil")
	}
	if m.Version == "" {
		return fmt.Errorf("module version is empty")
	}
	if m.Entry == nil {
		return fmt.Errorf("module entry metadata is missing")
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
		if err := validateGlobalBinding(g); err != nil {
			return err
		}
		if err := validateGlobalInit(m, g); err != nil {
			return err
		}
		if g.Kind == GlobalFunc {
			if g.Func < 0 || g.Func >= funcCount {
				return fmt.Errorf("global %q references invalid function index %d", g.Name, g.Func)
			}
			if err := validateFunctionLikeGlobalSig(m, g); err != nil {
				return err
			}
			if m.Functions[g.Func].Sig != g.Sig {
				return fmt.Errorf("global %q signature %d does not match function signature %d", g.Name, g.Sig, m.Functions[g.Func].Sig)
			}
			if m.Functions[g.Func].GlobalID != g.ID {
				return fmt.Errorf("global %q points to function %d with global id %d", g.Name, g.Func, m.Functions[g.Func].GlobalID)
			}
		}
		if g.Kind == GlobalExtern && isExternFunctionGlobal(g) {
			if err := validateFunctionLikeGlobalSig(m, g); err != nil {
				return err
			}
		}
	}
	if err := validateEntryPoint(m); err != nil {
		return err
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
		if err := validateLayout(m, l); err != nil {
			return err
		}
	}
	for i := range m.Functions {
		if err := validateFunction(m, i, &m.Functions[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateFunctionLikeGlobalSig(m *Module, g Global) error {
	if g.Sig < 0 || g.Sig >= len(m.Sigs) {
		return fmt.Errorf("function global %q references invalid signature %d", g.Name, g.Sig)
	}
	return nil
}

func validateEntryPoint(m *Module) error {
	if m.Entry.Global == NoEntryGlobal {
		if m.Entry.Name != "" {
			return fmt.Errorf("module entry name %q is set for no-entry module", m.Entry.Name)
		}
		return nil
	}
	if m.Entry.Global < 0 || m.Entry.Global >= len(m.Globals) {
		return fmt.Errorf("module entry references invalid global %d", m.Entry.Global)
	}
	g := m.Globals[m.Entry.Global]
	if g.Kind != GlobalFunc {
		return fmt.Errorf("module entry global %d is %q, not a defined function", m.Entry.Global, g.Name)
	}
	if m.Entry.Name == "" {
		return fmt.Errorf("module entry global %d has empty entry name", m.Entry.Global)
	}
	if m.Entry.Name != g.Name {
		return fmt.Errorf("module entry name %q does not match global %q", m.Entry.Name, g.Name)
	}
	return nil
}

func validateGlobalBinding(g Global) error {
	switch g.Kind {
	case GlobalExtern:
		if g.Extern.Name == "" {
			return fmt.Errorf("extern global %q has empty import name", g.Name)
		}
		if g.Extern.ABI == "" {
			return fmt.Errorf("extern global %q has empty ABI", g.Name)
		}
	case GlobalFunc, GlobalVar:
		if g.Extern != (ExternRef{}) {
			return fmt.Errorf("non-extern global %q has extern binding metadata", g.Name)
		}
	default:
		return fmt.Errorf("global %q has invalid kind %d", g.Name, g.Kind)
	}
	return nil
}

func validateGlobalInit(m *Module, g Global) error {
	if g.Kind == GlobalFunc {
		if len(g.Init.Bytes) != 0 || g.Init.ZeroFill != 0 || len(g.Init.Relocations) != 0 {
			return fmt.Errorf("function global %q has data initializer", g.Name)
		}
		return nil
	}
	if g.Init.ZeroFill < 0 {
		return fmt.Errorf("global %q has negative zero-fill %d", g.Name, g.Init.ZeroFill)
	}
	if g.Size < 0 {
		return fmt.Errorf("global %q has negative size %d", g.Name, g.Size)
	}
	if int64(len(g.Init.Bytes))+g.Init.ZeroFill > g.Size {
		return fmt.Errorf("global %q initializer size %d exceeds global size %d", g.Name, int64(len(g.Init.Bytes))+g.Init.ZeroFill, g.Size)
	}
	for _, r := range g.Init.Relocations {
		if r.Offset < 0 || r.Offset+m.Target.PointerSize > g.Size {
			return fmt.Errorf("global %q relocation offset %d does not fit in size %d", g.Name, r.Offset, g.Size)
		}
		switch r.Kind {
		case RelocGlobal:
			if r.Target < 0 || r.Target >= len(m.Globals) {
				return fmt.Errorf("global %q relocation references invalid global %d", g.Name, r.Target)
			}
		case RelocFunc:
			if r.Target < 0 || r.Target >= len(m.Globals) || !isFunctionGlobal(m.Globals[r.Target]) {
				return fmt.Errorf("global %q relocation references invalid function global %d", g.Name, r.Target)
			}
		case RelocString:
			if r.Target < 0 || r.Target >= len(m.Strings) {
				return fmt.Errorf("global %q relocation references invalid string %d", g.Name, r.Target)
			}
		default:
			return fmt.Errorf("global %q relocation has invalid kind %d", g.Name, r.Kind)
		}
	}
	return nil
}

func isFunctionGlobal(g Global) bool {
	return g.Kind == GlobalFunc || isExternFunctionGlobal(g)
}

func isExternFunctionGlobal(g Global) bool {
	return g.Kind == GlobalExtern && g.Size == 0 && g.Align == 0
}

func validateLayout(m *Module, l ObjectLayout) error {
	if l.Size < 0 {
		return fmt.Errorf("layout %d has negative size %d", l.ID, l.Size)
	}
	for _, f := range l.Fields {
		if f.ID < 0 || f.ID >= len(l.Fields) || l.Fields[f.ID].ID != f.ID {
			return fmt.Errorf("layout %d has invalid field id %d", l.ID, f.ID)
		}
		if f.Offset < 0 || f.Offset > l.Size {
			return fmt.Errorf("layout %d field %d offset %d outside size %d", l.ID, f.ID, f.Offset, l.Size)
		}
	}
	for _, bf := range l.Bit {
		if bf.ID < 0 || bf.ID >= len(l.Bit) || l.Bit[bf.ID].ID != bf.ID {
			return fmt.Errorf("layout %d has invalid bit-field id %d", l.ID, bf.ID)
		}
		containerBits := valueTypeSize(m.Target, bf.Container) * 8
		if containerBits <= 0 {
			return fmt.Errorf("layout %d bit-field %d has invalid container %s", l.ID, bf.ID, bf.Container)
		}
		if bf.BitOffset < 0 || bf.Width < 0 || int64(bf.BitOffset+bf.Width) > containerBits {
			return fmt.Errorf("layout %d bit-field %d width %d at bit offset %d exceeds container %s", l.ID, bf.ID, bf.Width, bf.BitOffset, bf.Container)
		}
		if bf.ByteOffset < 0 || bf.ByteOffset+containerBits/8 > l.Size {
			return fmt.Errorf("layout %d bit-field %d container at byte offset %d exceeds size %d", l.ID, bf.ID, bf.ByteOffset, l.Size)
		}
	}
	return nil
}

func valueTypeSize(t TargetInfo, vt ValueType) int64 {
	switch vt {
	case TypeBool:
		return t.BoolSize
	case TypeI8, TypeU8:
		return 1
	case TypeI16, TypeU16:
		return 2
	case TypeI32, TypeU32, TypeF32:
		return 4
	case TypeI64, TypeU64, TypeF64:
		return 8
	case TypeFLong:
		return 16
	case TypePtr, TypeObjectAddr:
		return t.PointerSize
	default:
		return 0
	}
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
	sig := m.Sigs[f.Sig]
	ret := sig.Ret
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
		next, err := validateInstrStack(m, stack, ins, ret, sig.Variadic, labels)
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

func validateInstrStack(m *Module, stack []ValueType, ins Instr, ret ValueType, variadic bool, labels map[int]Label) ([]ValueType, error) {
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
	case OpVaStart, OpVaEnd:
		if ins.Slot < 0 {
			return nil, fmt.Errorf("%v references negative va_list slot %d", ins.Op, ins.Slot)
		}
		if ins.Op == OpVaStart && !variadic {
			return nil, fmt.Errorf("%v in non-variadic function", ins.Op)
		}
	case OpVaArg:
		push(ins.Type)
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
