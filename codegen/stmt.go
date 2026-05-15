package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

func (fg *funcGen) emitStmt(s sema.Stmt) error {
	switch x := s.(type) {
	case *sema.Block:
		scopeMark := len(fg.activeDynamicObjects)
		for _, item := range x.Items {
			if err := fg.emitStmt(item); err != nil {
				return err
			}
		}
		if fg.lastInstrTerminal() {
			fg.activeDynamicObjects = fg.activeDynamicObjects[:scopeMark]
			return nil
		}
		fg.popDynamicObjectScope(scopeMark)
	case *sema.DeclStmt:
		for _, d := range x.Decls {
			vd, ok := d.(*sema.VarDecl)
			if !ok || vd.Storage == sema.StorageStatic || vd.Storage == sema.StorageExtern {
				continue
			}
			if isVLAType(vd.T) {
				if err := fg.emitVLADecl(vd); err != nil {
					return err
				}
				continue
			}
			if vd.Init == nil {
				continue
			}
			st, err := fg.storageForVar(vd.Sym, vd.T)
			if err != nil {
				return err
			}
			if st.kind != storageLocalSlot {
				if err := fg.emitInitStore(vd); err != nil {
					return err
				}
				continue
			}
			if err := fg.emitValue(vd.Init); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.StoreLocal(st.typ, st.slot))
		}
	case *sema.IfStmt:
		elseLabel := fg.newLabel(true, nil)
		endLabel := fg.newLabel(true, nil)
		if err := fg.emitBoolValue(x.Cond); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, elseLabel))
		if err := fg.emitStmt(x.Then); err != nil {
			return err
		}
		if x.Else != nil {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(endLabel))
			fg.mark(elseLabel)
			if err := fg.emitStmt(x.Else); err != nil {
				return err
			}
			fg.mark(endLabel)
		} else {
			fg.mark(elseLabel)
		}
	case *sema.WhileStmt:
		pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
		pendingContinueNames := append([]string(nil), fg.pendingContinueNames...)
		fg.pendingBreakNames = nil
		fg.pendingContinueNames = nil
		defer func() {
			fg.pendingBreakNames = pendingBreakNames
			fg.pendingContinueNames = pendingContinueNames
		}()
		condLabel := fg.newLabel(true, nil)
		bodyLabel := fg.newLabel(true, nil)
		endLabel := fg.newLabel(true, nil)
		fg.breaks = append(fg.breaks, endLabel)
		fg.continues = append(fg.continues, condLabel)
		loopCleanupMark := len(fg.activeDynamicObjects)
		fg.breakCleanupMarks = append(fg.breakCleanupMarks, loopCleanupMark)
		fg.continueCleanupMarks = append(fg.continueCleanupMarks, loopCleanupMark)
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel, loopCleanupMark)
		popNamedContinues := fg.pushNamedContinues(pendingContinueNames, condLabel, loopCleanupMark)
		if x.DoWhile {
			fg.mark(bodyLabel)
			if err := fg.emitStmt(x.Body); err != nil {
				popNamedContinues()
				popNamedBreaks()
				return err
			}
			fg.mark(condLabel)
			if err := fg.emitBoolValue(x.Cond); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfNonZero(bytecode.TypeBool, bodyLabel))
			fg.mark(endLabel)
		} else {
			fg.mark(condLabel)
			if err := fg.emitBoolValue(x.Cond); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, endLabel))
			if err := fg.emitStmt(x.Body); err != nil {
				popNamedContinues()
				popNamedBreaks()
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(condLabel))
			fg.mark(endLabel)
		}
		popNamedContinues()
		popNamedBreaks()
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.continues = fg.continues[:len(fg.continues)-1]
		fg.breakCleanupMarks = fg.breakCleanupMarks[:len(fg.breakCleanupMarks)-1]
		fg.continueCleanupMarks = fg.continueCleanupMarks[:len(fg.continueCleanupMarks)-1]
	case *sema.ForStmt:
		pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
		pendingContinueNames := append([]string(nil), fg.pendingContinueNames...)
		fg.pendingBreakNames = nil
		fg.pendingContinueNames = nil
		defer func() {
			fg.pendingBreakNames = pendingBreakNames
			fg.pendingContinueNames = pendingContinueNames
		}()
		loopCleanupMark := len(fg.activeDynamicObjects)
		if x.Init != nil {
			if err := fg.emitStmt(x.Init); err != nil {
				return err
			}
		}
		condLabel := fg.newLabel(true, nil)
		postLabel := fg.newLabel(true, nil)
		endLabel := fg.newLabel(true, nil)
		fg.breaks = append(fg.breaks, endLabel)
		fg.continues = append(fg.continues, postLabel)
		fg.breakCleanupMarks = append(fg.breakCleanupMarks, loopCleanupMark)
		fg.continueCleanupMarks = append(fg.continueCleanupMarks, loopCleanupMark)
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel, loopCleanupMark)
		popNamedContinues := fg.pushNamedContinues(pendingContinueNames, postLabel, loopCleanupMark)
		fg.mark(condLabel)
		if x.Cond != nil {
			if err := fg.emitBoolValue(x.Cond); err != nil {
				popNamedContinues()
				popNamedBreaks()
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, endLabel))
		}
		if err := fg.emitStmt(x.Body); err != nil {
			popNamedContinues()
			popNamedBreaks()
			return err
		}
		fg.mark(postLabel)
		if x.Post != nil {
			if err := fg.emitValue(x.Post); err != nil {
				popNamedContinues()
				popNamedBreaks()
				return err
			}
			if exprLeavesValue(x.Post) {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
			}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(condLabel))
		fg.mark(endLabel)
		if !fg.lastInstrTerminal() {
			fg.popDynamicObjectScope(loopCleanupMark)
		}
		popNamedContinues()
		popNamedBreaks()
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.continues = fg.continues[:len(fg.continues)-1]
		fg.breakCleanupMarks = fg.breakCleanupMarks[:len(fg.breakCleanupMarks)-1]
		fg.continueCleanupMarks = fg.continueCleanupMarks[:len(fg.continueCleanupMarks)-1]
	case *sema.SwitchStmt:
		pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
		fg.pendingBreakNames = nil
		defer func() {
			fg.pendingBreakNames = pendingBreakNames
		}()
		endLabel := fg.newLabel(true, nil)
		defaultLabel := endLabel
		if x.Default != nil {
			defaultLabel = fg.newLabel(true, nil)
			fg.defaultLabels[x.Default] = defaultLabel
		}
		cases := make([]bytecode.SwitchCase, 0, len(x.Cases))
		for _, c := range x.Cases {
			label := fg.newLabel(true, nil)
			fg.caseLabels[c] = label
			cases = append(cases, bytecode.SwitchCase{Value: c.Value, Label: label})
		}
		t, err := fg.g.lowerValueType(x.Cond.GetType())
		if err != nil {
			return err
		}
		if err := fg.emitValue(x.Cond); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpSwitch, Type: t, Label: defaultLabel, Labels: cases})
		fg.breaks = append(fg.breaks, endLabel)
		switchCleanupMark := len(fg.activeDynamicObjects)
		fg.breakCleanupMarks = append(fg.breakCleanupMarks, switchCleanupMark)
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel, switchCleanupMark)
		if err := fg.emitStmt(x.Body); err != nil {
			popNamedBreaks()
			return err
		}
		popNamedBreaks()
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.breakCleanupMarks = fg.breakCleanupMarks[:len(fg.breakCleanupMarks)-1]
		fg.mark(endLabel)
	case *sema.CaseStmt:
		fg.mark(fg.caseLabel(x))
		popNamedBreaks := func() {}
		if len(fg.pendingBreakNames) > 0 && len(fg.breaks) > 0 {
			pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
			fg.pendingBreakNames = nil
			cleanupMark := fg.currentBreakCleanupMark()
			popNamedBreaks = fg.pushNamedBreaks(pendingBreakNames, fg.breaks[len(fg.breaks)-1], cleanupMark)
		}
		if err := fg.emitStmt(x.Body); err != nil {
			popNamedBreaks()
			return err
		}
		popNamedBreaks()
	case *sema.DefaultStmt:
		fg.mark(fg.defaultLabel(x))
		popNamedBreaks := func() {}
		if len(fg.pendingBreakNames) > 0 && len(fg.breaks) > 0 {
			pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
			fg.pendingBreakNames = nil
			cleanupMark := fg.currentBreakCleanupMark()
			popNamedBreaks = fg.pushNamedBreaks(pendingBreakNames, fg.breaks[len(fg.breaks)-1], cleanupMark)
		}
		if err := fg.emitStmt(x.Body); err != nil {
			popNamedBreaks()
			return err
		}
		popNamedBreaks()
	case *sema.BreakStmt:
		target := 0
		if x.Name != "" {
			stack := fg.namedBreaks[x.Name]
			if len(stack) == 0 {
				return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: fmt.Sprintf("named break target %q is not active", x.Name)}
			}
			target = stack[len(stack)-1]
			cleanupStack := fg.namedBreakCleanups[x.Name]
			if len(cleanupStack) > 0 {
				fg.emitDynamicObjectCleanups(cleanupStack[len(cleanupStack)-1])
			}
		} else if len(fg.breaks) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "break outside breakable statement"}
		} else {
			target = fg.breaks[len(fg.breaks)-1]
			fg.emitDynamicObjectCleanups(fg.currentBreakCleanupMark())
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(target))
	case *sema.ContinueStmt:
		target := 0
		if x.Name != "" {
			stack := fg.namedContinues[x.Name]
			if len(stack) == 0 {
				return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: fmt.Sprintf("named continue target %q is not active", x.Name)}
			}
			target = stack[len(stack)-1]
			cleanupStack := fg.namedContinueCleanups[x.Name]
			if len(cleanupStack) > 0 {
				fg.emitDynamicObjectCleanups(cleanupStack[len(cleanupStack)-1])
			}
		} else if len(fg.continues) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "continue outside loop"}
		} else {
			target = fg.continues[len(fg.continues)-1]
			fg.emitDynamicObjectCleanups(fg.currentContinueCleanupMark())
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(target))
	case *sema.LabeledStmt:
		fg.mark(fg.namedLabel(x))
		target := labeledStmtTargetKind(x.Body)
		prevBreakNames := fg.pendingBreakNames
		prevContinueNames := fg.pendingContinueNames
		if target == namedTargetLoop || target == namedTargetSwitch {
			fg.pendingBreakNames = append(fg.pendingBreakNames, x.Name)
		}
		if target == namedTargetLoop {
			fg.pendingContinueNames = append(fg.pendingContinueNames, x.Name)
		}
		if err := fg.emitStmt(x.Body); err != nil {
			fg.pendingBreakNames = prevBreakNames
			fg.pendingContinueNames = prevContinueNames
			return err
		}
		fg.pendingBreakNames = prevBreakNames
		fg.pendingContinueNames = prevContinueNames
	case *sema.GotoStmt:
		if x.Target == nil {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "goto target is unresolved"}
		}
		fg.emitDynamicObjectCleanups(fg.labelCleanupMark(x.Target))
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(fg.namedLabel(x.Target)))
	case *sema.EmptyStmt:
	case *sema.ExprStmt:
		if x.Expr == nil {
			return nil
		}
		if err := fg.emitValue(x.Expr); err != nil {
			return err
		}
		if exprLeavesValue(x.Expr) {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
		}
	case *sema.ReturnStmt:
		if x.Value == nil {
			fg.emitDynamicObjectCleanups(0)
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpReturnVoid})
			return nil
		}
		if err := fg.emitValue(x.Value); err != nil {
			return err
		}
		t, err := fg.g.lowerValueType(x.Value.GetType())
		if err != nil {
			return err
		}
		fg.emitDynamicObjectCleanups(0)
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Return(t))
	default:
		return &Error{Pos: s.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "statement lowering is not implemented for this node"}
	}
	return nil
}

func (fg *funcGen) newLabel(statement bool, stack []bytecode.ValueType) int {
	id := fg.nextLabel
	fg.nextLabel++
	fg.out.Labels = append(fg.out.Labels, bytecode.Label{ID: id, Stack: append([]bytecode.ValueType(nil), stack...), Statement: statement})
	return id
}

func (fg *funcGen) mark(label int) {
	fg.out.Instrs = append(fg.out.Instrs, bytecode.LabelInstr(label))
}

func (fg *funcGen) namedLabel(s *sema.LabeledStmt) int {
	if label, ok := fg.labels[s]; ok {
		return label
	}
	label := fg.newLabel(true, nil)
	fg.labels[s] = label
	if label >= 0 && label < len(fg.out.Labels) {
		fg.out.Labels[label].Name = s.Name
	}
	return label
}

func (fg *funcGen) caseLabel(s *sema.CaseStmt) int {
	if label, ok := fg.caseLabels[s]; ok {
		return label
	}
	label := fg.newLabel(true, nil)
	fg.caseLabels[s] = label
	return label
}

func (fg *funcGen) defaultLabel(s *sema.DefaultStmt) int {
	if label, ok := fg.defaultLabels[s]; ok {
		return label
	}
	label := fg.newLabel(true, nil)
	fg.defaultLabels[s] = label
	return label
}

func (fg *funcGen) labelCleanupMark(label *sema.LabeledStmt) int {
	if mark, ok := fg.labelCleanupMarks[label]; ok {
		return mark
	}
	return len(fg.activeDynamicObjects)
}

func labelCleanupMarks(root sema.Stmt) map[*sema.LabeledStmt]int {
	marks := map[*sema.LabeledStmt]int{}
	walkLabelCleanupMarks(root, 0, marks)
	return marks
}

func walkLabelCleanupMarks(s sema.Stmt, depth int, marks map[*sema.LabeledStmt]int) int {
	switch x := s.(type) {
	case *sema.Block:
		current := depth
		for _, item := range x.Items {
			current = walkLabelCleanupMarks(item, current, marks)
		}
		return depth
	case *sema.DeclStmt:
		current := depth
		for _, d := range x.Decls {
			vd, ok := d.(*sema.VarDecl)
			if ok && vd.Storage != sema.StorageStatic && vd.Storage != sema.StorageExtern && isVLAType(vd.T) {
				current++
			}
		}
		return current
	case *sema.IfStmt:
		walkLabelCleanupMarks(x.Then, depth, marks)
		if x.Else != nil {
			walkLabelCleanupMarks(x.Else, depth, marks)
		}
		return depth
	case *sema.WhileStmt:
		walkLabelCleanupMarks(x.Body, depth, marks)
		return depth
	case *sema.ForStmt:
		bodyDepth := depth
		if x.Init != nil {
			bodyDepth = walkLabelCleanupMarks(x.Init, depth, marks)
		}
		walkLabelCleanupMarks(x.Body, bodyDepth, marks)
		return depth
	case *sema.SwitchStmt:
		walkLabelCleanupMarks(x.Body, depth, marks)
		return depth
	case *sema.CaseStmt:
		walkLabelCleanupMarks(x.Body, depth, marks)
		return depth
	case *sema.DefaultStmt:
		walkLabelCleanupMarks(x.Body, depth, marks)
		return depth
	case *sema.LabeledStmt:
		marks[x] = depth
		return walkLabelCleanupMarks(x.Body, depth, marks)
	default:
		return depth
	}
}

func (fg *funcGen) pushNamedBreaks(names []string, label, cleanupMark int) func() {
	for _, name := range names {
		fg.namedBreaks[name] = append(fg.namedBreaks[name], label)
		fg.namedBreakCleanups[name] = append(fg.namedBreakCleanups[name], cleanupMark)
	}
	return func() {
		for i := len(names) - 1; i >= 0; i-- {
			name := names[i]
			stack := fg.namedBreaks[name]
			if len(stack) <= 1 {
				delete(fg.namedBreaks, name)
				delete(fg.namedBreakCleanups, name)
				continue
			}
			fg.namedBreaks[name] = stack[:len(stack)-1]
			cleanups := fg.namedBreakCleanups[name]
			if len(cleanups) <= 1 {
				delete(fg.namedBreakCleanups, name)
			} else {
				fg.namedBreakCleanups[name] = cleanups[:len(cleanups)-1]
			}
		}
	}
}

func (fg *funcGen) pushNamedContinues(names []string, label, cleanupMark int) func() {
	for _, name := range names {
		fg.namedContinues[name] = append(fg.namedContinues[name], label)
		fg.namedContinueCleanups[name] = append(fg.namedContinueCleanups[name], cleanupMark)
	}
	return func() {
		for i := len(names) - 1; i >= 0; i-- {
			name := names[i]
			stack := fg.namedContinues[name]
			if len(stack) <= 1 {
				delete(fg.namedContinues, name)
				delete(fg.namedContinueCleanups, name)
				continue
			}
			fg.namedContinues[name] = stack[:len(stack)-1]
			cleanups := fg.namedContinueCleanups[name]
			if len(cleanups) <= 1 {
				delete(fg.namedContinueCleanups, name)
			} else {
				fg.namedContinueCleanups[name] = cleanups[:len(cleanups)-1]
			}
		}
	}
}

func (fg *funcGen) currentBreakCleanupMark() int {
	if len(fg.breakCleanupMarks) == 0 {
		return len(fg.activeDynamicObjects)
	}
	return fg.breakCleanupMarks[len(fg.breakCleanupMarks)-1]
}

func (fg *funcGen) currentContinueCleanupMark() int {
	if len(fg.continueCleanupMarks) == 0 {
		return len(fg.activeDynamicObjects)
	}
	return fg.continueCleanupMarks[len(fg.continueCleanupMarks)-1]
}

func (fg *funcGen) emitDynamicObjectCleanups(mark int) {
	if mark < 0 {
		mark = 0
	}
	if mark > len(fg.activeDynamicObjects) {
		mark = len(fg.activeDynamicObjects)
	}
	for i := len(fg.activeDynamicObjects) - 1; i >= mark; i-- {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpFreeDynamicObject, Object: fg.activeDynamicObjects[i]})
	}
}

func (fg *funcGen) popDynamicObjectScope(mark int) {
	fg.emitDynamicObjectCleanups(mark)
	fg.activeDynamicObjects = fg.activeDynamicObjects[:mark]
}

func (fg *funcGen) lastInstrTerminal() bool {
	if len(fg.out.Instrs) == 0 {
		return false
	}
	switch fg.out.Instrs[len(fg.out.Instrs)-1].Op {
	case bytecode.OpReturn, bytecode.OpReturnVoid, bytecode.OpReturnObject, bytecode.OpJump, bytecode.OpUnreachable:
		return true
	default:
		return false
	}
}

func (fg *funcGen) emitVLADecl(vd *sema.VarDecl) error {
	if vd == nil || vd.Sym == nil {
		return nil
	}
	layout, err := fg.g.lowerLayout(vd.T)
	if err != nil {
		return err
	}
	slot := vd.Sym.SlotID
	if slot < 0 {
		return &Error{Pos: vd.Pos().SourceStart, Node: fmt.Sprintf("%T", vd), Op: "emitStmt", Reason: "VLA declaration has no size slot"}
	}
	if !fg.hasLocalSlot(slot) {
		fg.out.Locals = append(fg.out.Locals, bytecode.LocalSlot{ID: slot, Name: vd.Sym.Name + "$size", Type: bytecode.TypeI64})
	}
	objectID := len(fg.out.DynamicObjects)
	fg.out.DynamicObjects = append(fg.out.DynamicObjects, bytecode.DynamicObject{ID: objectID, Name: vd.Sym.Name, Align: layout.Align, Layout: layout.ID})
	fg.dynamicObjectMap[vd.Sym] = objectID
	fg.dynamicSizeSlotMap[vd.Sym] = slot
	if err := fg.emitRuntimeSizeof(vd.T); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.Instr{Op: bytecode.OpDup},
		bytecode.StoreLocal(bytecode.TypeI64, slot),
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: objectID, Type: bytecode.TypeI64, Align: layout.Align, Layout: layout.ID},
	)
	fg.activeDynamicObjects = append(fg.activeDynamicObjects, objectID)
	return nil
}

func (fg *funcGen) hasLocalSlot(slot int) bool {
	for _, p := range fg.out.Params {
		if p.Slot == slot {
			return true
		}
	}
	for _, l := range fg.out.Locals {
		if l.ID == slot {
			return true
		}
	}
	return false
}

type namedTargetKind int

const (
	namedTargetNone namedTargetKind = iota
	namedTargetLoop
	namedTargetSwitch
)

func labeledStmtTargetKind(s sema.Stmt) namedTargetKind {
	switch x := s.(type) {
	case *sema.LabeledStmt:
		return labeledStmtTargetKind(x.Body)
	case *sema.ForStmt, *sema.WhileStmt:
		return namedTargetLoop
	case *sema.SwitchStmt, *sema.CaseStmt, *sema.DefaultStmt:
		return namedTargetSwitch
	default:
		return namedTargetNone
	}
}

func (fg *funcGen) emitInitStore(vd *sema.VarDecl) error {
	lhs := &sema.VarRef{Sym: vd.Sym, T: vd.T, Range: vd.Range}
	vt, err := fg.g.lowerValueType(vd.T)
	if err != nil {
		return err
	}
	if vt == bytecode.TypeObjectAddr {
		if err := fg.emitAddress(lhs); err != nil {
			return err
		}
		if err := fg.emitAddress(vd.Init); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpMemCopy, Size: fg.g.sizeof(vd.T), Align: fg.g.alignof(vd.T), Volatile: isVolatile(vd.T)})
		return nil
	}
	if err := fg.emitAddress(lhs); err != nil {
		return err
	}
	if err := fg.emitValue(vd.Init); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs, bytecode.Store(vt, fg.g.alignof(vd.T), isVolatile(vd.T)))
	return nil
}
