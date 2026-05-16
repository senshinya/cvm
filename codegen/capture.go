package codegen

import (
	"fmt"

	"shinya.click/cvm/bytecode"
	"shinya.click/cvm/sema"
)

type capture struct {
	sym       *sema.Symbol
	sizeTypes []sema.Type
}

func (g *generator) prepareNestedCaptures() {
	owner := map[*sema.Symbol]*sema.FuncDef{}
	for _, fn := range g.prog.Funcs {
		if fn == nil || fn.Sym == nil {
			continue
		}
		g.funcMap[fn.Sym] = fn
		if fn.Sym.GlobalID >= 0 {
			g.funcByGlobal[fn.Sym.GlobalID] = fn
		}
		if fn.Sym.Name != "" {
			g.funcByName[fn.Sym.Name] = fn
		}
		for _, p := range fn.Params {
			if p != nil && p.Sym != nil {
				owner[p.Sym] = fn
			}
		}
		for _, local := range fn.Locals {
			if local == nil || local.Sym == nil || local.Storage == sema.StorageStatic || local.Storage == sema.StorageExtern {
				continue
			}
			owner[local.Sym] = fn
		}
	}
	for _, fn := range g.prog.Funcs {
		caps := collectCaptures(fn, owner)
		if len(caps) == 0 {
			continue
		}
		g.nestedCaptures[fn] = caps
		for _, cap := range caps {
			own := owner[cap.sym]
			if own == nil || own == fn {
				continue
			}
			m := g.capturedByOwner[own]
			if m == nil {
				m = map[*sema.Symbol]bool{}
				g.capturedByOwner[own] = m
			}
			m[cap.sym] = true
		}
	}
}

func collectCaptures(fn *sema.FuncDef, owner map[*sema.Symbol]*sema.FuncDef) []capture {
	if fn == nil {
		return nil
	}
	seen := map[*sema.Symbol]bool{}
	var out []capture
	add := func(sym *sema.Symbol) {
		if sym == nil || seen[sym] || owner[sym] == nil || owner[sym] == fn {
			return
		}
		if sym.Storage == sema.StorageStatic || sym.Storage == sema.StorageExtern {
			return
		}
		seen[sym] = true
		out = append(out, capture{sym: sym, sizeTypes: captureSizeTypes(sym.T)})
	}
	walkStmtForCaptures(fn.Body, add)
	return out
}

func captureSizeTypes(t sema.Type) []sema.Type {
	seen := map[string]bool{}
	var out []sema.Type
	var walk func(sema.Type)
	walk = func(t sema.Type) {
		switch x := sema.Unqual(t).(type) {
		case *sema.PointerType:
			walk(x.Pointee)
		case *sema.ArrayType:
			if typeHasVariableSize(x) {
				key := dynamicSizeKey(x)
				if !seen[key] {
					seen[key] = true
					out = append(out, x)
				}
			}
			walk(x.Elem)
		case *sema.StructType:
			if typeHasVariableSize(x) {
				key := dynamicSizeKey(x)
				if !seen[key] {
					seen[key] = true
					out = append(out, x)
				}
			}
			for _, f := range x.Fields {
				if f != nil {
					walk(f.T)
				}
			}
		case *sema.UnionType:
			if typeHasVariableSize(x) {
				key := dynamicSizeKey(x)
				if !seen[key] {
					seen[key] = true
					out = append(out, x)
				}
			}
			for _, f := range x.Fields {
				if f != nil {
					walk(f.T)
				}
			}
		}
	}
	walk(t)
	return out
}

func (fg *funcGen) emitCaptureArgs(caps []capture) (int, error) {
	count := 0
	for _, cap := range caps {
		if cap.sym == nil {
			continue
		}
		if err := fg.emitAddress(&sema.VarRef{Sym: cap.sym, T: cap.sym.T}); err != nil {
			return count, fmt.Errorf("capture %q address: %w", cap.sym.Name, err)
		}
		count++
		for _, typ := range cap.sizeTypes {
			slot, ok := fg.dynamicSizeSlotForSymbol(cap.sym, typ)
			if !ok {
				return count, fmt.Errorf("capture %q missing VLA metadata for %s", cap.sym.Name, typ)
			}
			fg.out.Instrs = append(fg.out.Instrs, bytecode.LoadLocal(bytecode.TypeI64, slot))
			count++
		}
	}
	return count, nil
}

func walkStmtForCaptures(s sema.Stmt, add func(*sema.Symbol)) {
	switch x := s.(type) {
	case *sema.Block:
		for _, item := range x.Items {
			walkStmtForCaptures(item, add)
		}
	case *sema.DeclStmt:
		for _, d := range x.Decls {
			if vd, ok := d.(*sema.VarDecl); ok {
				walkExprForCaptures(vd.Init, add)
			}
		}
	case *sema.ExprStmt:
		walkExprForCaptures(x.Expr, add)
	case *sema.ReturnStmt:
		walkExprForCaptures(x.Value, add)
	case *sema.IfStmt:
		walkExprForCaptures(x.Cond, add)
		walkStmtForCaptures(x.Then, add)
		walkStmtForCaptures(x.Else, add)
	case *sema.WhileStmt:
		walkExprForCaptures(x.Cond, add)
		walkStmtForCaptures(x.Body, add)
	case *sema.ForStmt:
		walkStmtForCaptures(x.Init, add)
		walkExprForCaptures(x.Cond, add)
		walkExprForCaptures(x.Post, add)
		walkStmtForCaptures(x.Body, add)
	case *sema.SwitchStmt:
		walkExprForCaptures(x.Cond, add)
		walkStmtForCaptures(x.Body, add)
	case *sema.CaseStmt:
		walkStmtForCaptures(x.Body, add)
	case *sema.DefaultStmt:
		walkStmtForCaptures(x.Body, add)
	case *sema.LabeledStmt:
		walkStmtForCaptures(x.Body, add)
	}
}

func walkExprForCaptures(e sema.Expr, add func(*sema.Symbol)) {
	switch x := e.(type) {
	case nil:
	case *sema.VarRef:
		add(x.Sym)
	case *sema.UnOp:
		walkExprForCaptures(x.X, add)
	case *sema.BinOp:
		walkExprForCaptures(x.L, add)
		walkExprForCaptures(x.R, add)
	case *sema.AssignExpr:
		walkExprForCaptures(x.L, add)
		walkExprForCaptures(x.R, add)
	case *sema.CompoundAssign:
		walkExprForCaptures(x.L, add)
		walkExprForCaptures(x.R, add)
	case *sema.CallExpr:
		walkExprForCaptures(x.Callee, add)
		for _, arg := range x.Args {
			walkExprForCaptures(arg, add)
		}
	case *sema.MemberExpr:
		walkExprForCaptures(x.Base, add)
	case *sema.IndexExpr:
		walkExprForCaptures(x.Base, add)
		walkExprForCaptures(x.Index, add)
	case *sema.CondExpr:
		walkExprForCaptures(x.Cond, add)
		walkExprForCaptures(x.Then, add)
		walkExprForCaptures(x.Else, add)
	case *sema.SizeofExpr:
		walkExprForCaptures(x.Operand.Expr, add)
	case *sema.CommaExpr:
		walkExprForCaptures(x.L, add)
		walkExprForCaptures(x.R, add)
	case *sema.CompoundLit:
		walkExprForCaptures(x.Init, add)
	case *sema.InitList:
		for _, elem := range x.Elems {
			walkExprForCaptures(elem.Value, add)
		}
	case *sema.ImplicitCast:
		walkExprForCaptures(x.X, add)
	case *sema.ExplicitCast:
		walkExprForCaptures(x.X, add)
	case *sema.StmtExpr:
		walkStmtForCaptures(x.Block, add)
	}
}
