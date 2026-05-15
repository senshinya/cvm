package codegen

import "shinya.click/cvm/sema"

func findAddressTaken(fn *sema.FuncDef) map[*sema.Symbol]bool {
	out := map[*sema.Symbol]bool{}
	if fn == nil {
		return out
	}
	walkStmtForAddressTaken(fn.Body, out)
	return out
}

func walkStmtForAddressTaken(s sema.Stmt, out map[*sema.Symbol]bool) {
	switch x := s.(type) {
	case *sema.Block:
		for _, item := range x.Items {
			walkStmtForAddressTaken(item, out)
		}
	case *sema.DeclStmt:
		for _, d := range x.Decls {
			if vd, ok := d.(*sema.VarDecl); ok {
				walkExprForAddressTaken(vd.Init, out)
			}
		}
	case *sema.ExprStmt:
		walkExprForAddressTaken(x.Expr, out)
	case *sema.ReturnStmt:
		walkExprForAddressTaken(x.Value, out)
	case *sema.IfStmt:
		walkExprForAddressTaken(x.Cond, out)
		walkStmtForAddressTaken(x.Then, out)
		walkStmtForAddressTaken(x.Else, out)
	case *sema.WhileStmt:
		walkExprForAddressTaken(x.Cond, out)
		walkStmtForAddressTaken(x.Body, out)
	case *sema.ForStmt:
		walkStmtForAddressTaken(x.Init, out)
		walkExprForAddressTaken(x.Cond, out)
		walkExprForAddressTaken(x.Post, out)
		walkStmtForAddressTaken(x.Body, out)
	case *sema.SwitchStmt:
		walkExprForAddressTaken(x.Cond, out)
		walkStmtForAddressTaken(x.Body, out)
	case *sema.CaseStmt:
		walkStmtForAddressTaken(x.Body, out)
	case *sema.DefaultStmt:
		walkStmtForAddressTaken(x.Body, out)
	case *sema.LabeledStmt:
		walkStmtForAddressTaken(x.Body, out)
	}
}

func walkExprForAddressTaken(e sema.Expr, out map[*sema.Symbol]bool) {
	switch x := e.(type) {
	case nil:
	case *sema.UnOp:
		if x.Op == sema.UnAddr {
			markAddressRoot(x.X, out)
		}
		walkExprForAddressTaken(x.X, out)
	case *sema.BinOp:
		walkExprForAddressTaken(x.L, out)
		walkExprForAddressTaken(x.R, out)
	case *sema.AssignExpr:
		walkExprForAddressTaken(x.L, out)
		walkExprForAddressTaken(x.R, out)
	case *sema.CompoundAssign:
		walkExprForAddressTaken(x.L, out)
		walkExprForAddressTaken(x.R, out)
	case *sema.CallExpr:
		walkExprForAddressTaken(x.Callee, out)
		for _, arg := range x.Args {
			walkExprForAddressTaken(arg, out)
		}
	case *sema.MemberExpr:
		walkExprForAddressTaken(x.Base, out)
	case *sema.IndexExpr:
		walkExprForAddressTaken(x.Base, out)
		walkExprForAddressTaken(x.Index, out)
	case *sema.CondExpr:
		walkExprForAddressTaken(x.Cond, out)
		walkExprForAddressTaken(x.Then, out)
		walkExprForAddressTaken(x.Else, out)
	case *sema.SizeofExpr:
		walkExprForAddressTaken(x.Operand.Expr, out)
	case *sema.CommaExpr:
		walkExprForAddressTaken(x.L, out)
		walkExprForAddressTaken(x.R, out)
	case *sema.CompoundLit:
		walkExprForAddressTaken(x.Init, out)
	case *sema.InitList:
		for _, elem := range x.Elems {
			walkExprForAddressTaken(elem.Value, out)
		}
	case *sema.ImplicitCast:
		walkExprForAddressTaken(x.X, out)
	case *sema.ExplicitCast:
		walkExprForAddressTaken(x.X, out)
	}
}

func markAddressRoot(e sema.Expr, out map[*sema.Symbol]bool) {
	switch x := e.(type) {
	case *sema.VarRef:
		if x.Sym != nil {
			out[x.Sym] = true
		}
	case *sema.MemberExpr:
		markAddressRoot(x.Base, out)
	case *sema.IndexExpr:
		markAddressRoot(x.Base, out)
	case *sema.ImplicitCast:
		markAddressRoot(x.X, out)
	case *sema.UnOp:
		if x.Op == sema.UnAddr {
			markAddressRoot(x.X, out)
		}
	}
}
