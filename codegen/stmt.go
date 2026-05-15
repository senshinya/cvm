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
		condLabel := fg.newLabel(true, nil)
		bodyLabel := fg.newLabel(true, nil)
		endLabel := fg.newLabel(true, nil)
		fg.breaks = append(fg.breaks, endLabel)
		fg.continues = append(fg.continues, condLabel)
		if x.DoWhile {
			fg.mark(bodyLabel)
			if err := fg.emitStmt(x.Body); err != nil {
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
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(condLabel))
			fg.mark(endLabel)
		}
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.continues = fg.continues[:len(fg.continues)-1]
	case *sema.ForStmt:
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
		fg.mark(condLabel)
		if x.Cond != nil {
			if err := fg.emitBoolValue(x.Cond); err != nil {
				return err
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.JumpIfZero(bytecode.TypeBool, endLabel))
		}
		if err := fg.emitStmt(x.Body); err != nil {
			return err
		}
		fg.mark(postLabel)
		if x.Post != nil {
			if err := fg.emitValue(x.Post); err != nil {
				return err
			}
			if exprLeavesValue(x.Post) {
				fg.out.Instrs = append(fg.out.Instrs, bytecode.Instr{Op: bytecode.OpPop})
			}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(condLabel))
		fg.mark(endLabel)
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.continues = fg.continues[:len(fg.continues)-1]
	case *sema.SwitchStmt:
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
		if err := fg.emitStmt(x.Body); err != nil {
			return err
		}
		fg.breaks = fg.breaks[:len(fg.breaks)-1]
		fg.mark(endLabel)
	case *sema.CaseStmt:
		fg.mark(fg.caseLabel(x))
		if err := fg.emitStmt(x.Body); err != nil {
			return err
		}
	case *sema.DefaultStmt:
		fg.mark(fg.defaultLabel(x))
		if err := fg.emitStmt(x.Body); err != nil {
			return err
		}
	case *sema.BreakStmt:
		if len(fg.breaks) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "break outside breakable statement"}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(fg.breaks[len(fg.breaks)-1]))
	case *sema.ContinueStmt:
		if len(fg.continues) == 0 {
			return &Error{Pos: x.Pos().SourceStart, Node: fmt.Sprintf("%T", s), Op: "emitStmt", Reason: "continue outside loop"}
		}
		fg.out.Instrs = append(fg.out.Instrs, bytecode.Jump(fg.continues[len(fg.continues)-1]))
	case *sema.LabeledStmt:
		fg.mark(fg.namedLabel(x))
		if err := fg.emitStmt(x.Body); err != nil {
			return err
		}
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
