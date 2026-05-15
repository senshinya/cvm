package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

func (fg *funcGen) emitStmt(s sema.Stmt) error {
	switch x := s.(type) {
	case *sema.Block:
		for _, item := range x.Items {
			if err := fg.emitStmt(item); err != nil {
				return err
			}
		}
	case *sema.DeclStmt:
		for _, d := range x.Decls {
			vd, ok := d.(*sema.VarDecl)
			if !ok || vd.Init == nil || vd.Storage == sema.StorageStatic || vd.Storage == sema.StorageExtern {
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
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel)
		popNamedContinues := fg.pushNamedContinues(pendingContinueNames, condLabel)
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
	case *sema.ForStmt:
		pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
		pendingContinueNames := append([]string(nil), fg.pendingContinueNames...)
		fg.pendingBreakNames = nil
		fg.pendingContinueNames = nil
		defer func() {
			fg.pendingBreakNames = pendingBreakNames
			fg.pendingContinueNames = pendingContinueNames
		}()
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
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel)
		popNamedContinues := fg.pushNamedContinues(pendingContinueNames, postLabel)
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
		popNamedContinues()
		popNamedBreaks()
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.continues = fg.continues[:len(fg.continues)-1]
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
		popNamedBreaks := fg.pushNamedBreaks(pendingBreakNames, endLabel)
		if err := fg.emitStmt(x.Body); err != nil {
			popNamedBreaks()
			return err
		}
		popNamedBreaks()
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.mark(endLabel)
	case *sema.CaseStmt:
		fg.mark(fg.caseLabel(x))
		popNamedBreaks := func() {}
		if len(fg.pendingBreakNames) > 0 && len(fg.breaks) > 0 {
			pendingBreakNames := append([]string(nil), fg.pendingBreakNames...)
			fg.pendingBreakNames = nil
			popNamedBreaks = fg.pushNamedBreaks(pendingBreakNames, fg.breaks[len(fg.breaks)-1])
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
			popNamedBreaks = fg.pushNamedBreaks(pendingBreakNames, fg.breaks[len(fg.breaks)-1])
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
		} else if len(fg.breaks) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "break outside breakable statement"}
		} else {
			target = fg.breaks[len(fg.breaks)-1]
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
		} else if len(fg.continues) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "continue outside loop"}
		} else {
			target = fg.continues[len(fg.continues)-1]
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

func (fg *funcGen) pushNamedBreaks(names []string, label int) func() {
	for _, name := range names {
		fg.namedBreaks[name] = append(fg.namedBreaks[name], label)
	}
	return func() {
		for i := len(names) - 1; i >= 0; i-- {
			name := names[i]
			stack := fg.namedBreaks[name]
			if len(stack) <= 1 {
				delete(fg.namedBreaks, name)
				continue
			}
			fg.namedBreaks[name] = stack[:len(stack)-1]
		}
	}
}

func (fg *funcGen) pushNamedContinues(names []string, label int) func() {
	for _, name := range names {
		fg.namedContinues[name] = append(fg.namedContinues[name], label)
	}
	return func() {
		for i := len(names) - 1; i >= 0; i-- {
			name := names[i]
			stack := fg.namedContinues[name]
			if len(stack) <= 1 {
				delete(fg.namedContinues, name)
				continue
			}
			fg.namedContinues[name] = stack[:len(stack)-1]
		}
	}
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
