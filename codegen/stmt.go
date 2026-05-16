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
			if typeHasVariableSize(vd.T) {
				if err := fg.emitVLADecl(vd); err != nil {
					return err
				}
				continue
			}
			if typeHasVariablyModifiedType(vd.T) {
				if err := fg.prepareDynamicSizeTypesForSymbol(vd.Sym, vd.T, vd.Sym.Name+"$size"); err != nil {
					return err
				}
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
		cleanupLabel := fg.newLabel(true, nil)
		afterLabel := fg.newLabel(true, nil)
		fg.breaks = append(fg.breaks, afterLabel)
		fg.continues = append(fg.continues, postLabel)
		fg.breakCleanupMarks = append(fg.breakCleanupMarks, loopCleanupMark)
		continueCleanupMark := len(fg.activeDynamicObjects)
		fg.continueCleanupMarks = append(fg.continueCleanupMarks, continueCleanupMark)
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, afterLabel, loopCleanupMark)
		popNamedContinues := fg.pushNamedContinues(pendingContinueNames, postLabel, continueCleanupMark)
		fg.mark(condLabel)
		if x.Cond != nil {
			if err := fg.emitBoolValue(x.Cond); err != nil {
				popNamedContinues()
				popNamedBreaks()
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, cleanupLabel))
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
		fg.mark(cleanupLabel)
		if !fg.lastInstrTerminal() {
			fg.popDynamicObjectScope(loopCleanupMark)
		}
		fg.mark(afterLabel)
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
			if ok && vd.Storage != sema.StorageStatic && vd.Storage != sema.StorageExtern && typeHasVariableSize(vd.T) {
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
	if err := fg.emitRuntimeSizeofSavedForSymbol(vd.Sym, vd.T, slot, vd.Sym.Name+"$size"); err != nil {
		return err
	}
	fg.out.Instrs = append(fg.out.Instrs,
		bytecode.Instr{Op: bytecode.OpAllocDynamicObject, Object: objectID, Type: bytecode.TypeI64, Align: layout.Align, Layout: layout.ID},
	)
	fg.activeDynamicObjects = append(fg.activeDynamicObjects, objectID)
	return nil
}

func (fg *funcGen) emitRuntimeSizeofSaved(t sema.Type, slot int, name string) error {
	return fg.emitRuntimeSizeofSavedForSymbol(nil, t, slot, name)
}

func (fg *funcGen) emitRuntimeSizeofSavedForSymbol(sym *sema.Symbol, t sema.Type, slot int, name string) error {
	switch x := sema.Unqual(t).(type) {
	case *sema.ArrayType:
		if x.SizeKind == sema.ArrayStarSize {
			return fmt.Errorf("cannot lower runtime sizeof for star-sized array")
		}
		if x.SizeKind == sema.ArrayVLA {
			if x.SizeExpr == nil {
				return fmt.Errorf("VLA type has no bound expression")
			}
			if err := fg.emitValue(x.SizeExpr); err != nil {
				return err
			}
			from, err := fg.g.lowerValueType(x.SizeExpr.GetType())
			if err != nil {
				return err
			}
			fg.emitCast(from, bytecode.TypeI64, sema.IntegralConversion)
		} else {
			fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(x.Size))
		}
		elemSlot := -1
		if typeHasVariableSize(x.Elem) {
			elemSlot = fg.allocSyntheticI64Slot(name + "$elem")
		}
		if err := fg.emitRuntimeSizeofSavedForSymbol(sym, x.Elem, elemSlot, name+"$elem"); err != nil {
			return err
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Binary(bytecode.TypeI64, bytecode.BinMul))
		if typeHasVariableSize(x) {
			if slot < 0 {
				slot = fg.allocSyntheticI64Slot(name)
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(bytecode.TypeI64, slot))
			if sym != nil {
				fg.bindDynamicSizeSlot(sym, t, slot)
			}
		}
		return nil
	case *sema.StructType:
		return fg.emitRuntimeSizeofRecordSavedForSymbol(sym, t, x.Fields, false, slot, name)
	case *sema.UnionType:
		return fg.emitRuntimeSizeofRecordSavedForSymbol(sym, t, x.Fields, true, slot, name)
	default:
		fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(fg.g.sizeof(t)))
		return nil
	}
}

func (fg *funcGen) emitRuntimeSizeofRecordSavedForSymbol(sym *sema.Symbol, t sema.Type, fields []*sema.Field, union bool, slot int, name string) error {
	var variableFields []*sema.Field
	var staticEnd int64
	for _, f := range fields {
		if f == nil {
			continue
		}
		if typeHasVariableSize(f.T) {
			variableFields = append(variableFields, f)
			continue
		}
		end := fg.g.sizeof(f.T)
		if !union {
			end += f.Offset
		}
		if end > staticEnd {
			staticEnd = end
		}
	}
	if len(variableFields) == 0 {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(fg.g.sizeof(t)))
		return nil
	}
	if len(variableFields) != 1 {
		return fmt.Errorf("cannot lower runtime sizeof for aggregate with multiple variable-size fields")
	}
	field := variableFields[0]
	fieldName := field.Name
	if fieldName == "" {
		fieldName = "field"
	}
	if staticEnd > 0 && (union || staticEnd > field.Offset) {
		return fmt.Errorf("cannot lower runtime sizeof for aggregate mixing static and variable-size fields")
	}
	if err := fg.emitRuntimeSizeofSavedForSymbol(sym, field.T, -1, name+"$"+fieldName); err != nil {
		return err
	}
	if !union && field.Offset != 0 {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.I64Const(field.Offset), bytecode.Binary(bytecode.TypeI64, bytecode.BinAdd))
	}
	if slot >= 0 {
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpDup}, bytecode.StoreLocal(bytecode.TypeI64, slot))
		if sym != nil {
			fg.bindDynamicSizeSlot(sym, t, slot)
		}
	}
	return nil
}

func (fg *funcGen) prepareDynamicSizeTypesForSymbol(sym *sema.Symbol, t sema.Type, name string) error {
	return fg.prepareDynamicSizeTypesSeen(sym, t, name, map[sema.Type]bool{})
}

func (fg *funcGen) prepareDynamicSizeTypesSeen(sym *sema.Symbol, t sema.Type, name string, seen map[sema.Type]bool) error {
	key := sema.Unqual(t)
	if seen[key] {
		return nil
	}
	seen[key] = true
	switch x := key.(type) {
	case *sema.PointerType:
		if err := fg.prepareDynamicSizeTypesSeen(sym, x.Pointee, name+"$pointee", seen); err != nil {
			return err
		}
		if sym != nil {
			fg.dynamicPointerTypeMap[sym] = fg.dynamicSizeSymbolMap[sym]
		}
		return nil
	case *sema.ArrayType:
		if !typeHasVariableSize(x) {
			return nil
		}
		if _, ok := fg.dynamicSizeSlotForSymbol(sym, key); !ok {
			slot := fg.allocSyntheticI64Slot(name)
			if err := fg.emitRuntimeSizeofSavedForSymbol(sym, key, slot, name); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
		}
		return fg.prepareDynamicSizeTypesSeen(sym, x.Elem, name+"$elem", seen)
	default:
		return nil
	}
}

func (fg *funcGen) bindDynamicSizeSlot(sym *sema.Symbol, t sema.Type, slot int) {
	if sym == nil {
		return
	}
	m := fg.dynamicSizeSymbolMap[sym]
	if m == nil {
		m = map[string]int{}
		fg.dynamicSizeSymbolMap[sym] = m
	}
	m[dynamicSizeKey(t)] = slot
}

func (fg *funcGen) dynamicSizeSlotForSymbol(sym *sema.Symbol, t sema.Type) (int, bool) {
	if sym == nil {
		return 0, false
	}
	m := fg.dynamicSizeSymbolMap[sym]
	if m == nil {
		return 0, false
	}
	slot, ok := m[dynamicSizeKey(t)]
	return slot, ok
}

func dynamicSizeKey(t sema.Type) string {
	switch x := sema.Unqual(t).(type) {
	case *sema.PointerType:
		return "*" + dynamicSizeKey(x.Pointee)
	case *sema.ArrayType:
		var size string
		switch x.SizeKind {
		case sema.ArrayConstantSize:
			size = fmt.Sprintf("%d", x.Size)
		case sema.ArrayUnsized:
			size = ""
		case sema.ArrayVLA:
			size = exprKey(x.SizeExpr)
		case sema.ArrayStarSize:
			size = "*"
		default:
			size = "?"
		}
		return dynamicSizeKey(x.Elem) + "[" + size + "]"
	default:
		return sema.Unqual(t).String()
	}
}

func exprKey(e sema.Expr) string {
	switch x := e.(type) {
	case nil:
		return "<nil>"
	case *sema.IntLit:
		return fmt.Sprintf("int:%d", x.Value)
	case *sema.CharLit:
		return fmt.Sprintf("char:%d", x.Value)
	case *sema.VarRef:
		if x.Sym != nil {
			return fmt.Sprintf("var:%s:%d:%d", x.Sym.Name, x.Sym.GlobalID, x.Sym.SlotID)
		}
		return "var:<nil>"
	case *sema.ImplicitCast:
		return fmt.Sprintf("icast:%d:%s", x.Kind, exprKey(x.X))
	case *sema.ExplicitCast:
		return fmt.Sprintf("ecast:%s:%s", x.To, exprKey(x.X))
	case *sema.BinOp:
		return fmt.Sprintf("bin:%d:%s:%s", x.Op, exprKey(x.L), exprKey(x.R))
	case *sema.UnOp:
		return fmt.Sprintf("un:%d:%s", x.Op, exprKey(x.X))
	default:
		return fmt.Sprintf("%T:%v", e, e.Pos().SourceStart)
	}
}

func (fg *funcGen) allocSyntheticI64Slot(name string) int {
	return fg.allocSyntheticSlot(name, bytecode.TypeI64)
}

func (fg *funcGen) allocSyntheticSlot(name string, typ bytecode.ValueType) int {
	slot := fg.nextSyntheticSlot
	fg.nextSyntheticSlot++
	fg.out.Locals = append(fg.out.Locals, bytecode.LocalSlot{ID: slot, Name: name, Type: typ})
	return slot
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
	dst := address{emit: func() error {
		return fg.emitAddress(lhs)
	}}
	return fg.emitInitializer(dst, vd.Init, vd.T)
}
