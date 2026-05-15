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
