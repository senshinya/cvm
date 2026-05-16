package sema

import "testing"

func TestValidateProgramInvariantsAcceptsAnalyzedProgram(t *testing.T) {
	prog := mustAnalyze(t, `
		int f(int a) {
			int x = a + 1;
			return x;
		}
	`)
	if err := ValidateProgramInvariants(prog); err != nil {
		t.Fatalf("valid program failed invariant check: %v", err)
	}
}

func TestValidateProgramInvariantsRejectsMissingVarRefSymbol(t *testing.T) {
	prog := mustAnalyze(t, `int f(int a) { return a; }`)
	ref := firstVarRefInStmt(prog.Funcs[0].Body)
	if ref == nil {
		t.Fatal("test source did not produce a VarRef")
	}
	ref.Sym = nil
	if err := ValidateProgramInvariants(prog); err == nil {
		t.Fatal("invariant check accepted VarRef with nil symbol")
	}
}

func TestValidateProgramInvariantsRejectsDuplicateLocalSlots(t *testing.T) {
	prog := mustAnalyze(t, `
		int f(void) {
			int x;
			int y;
			return x + y;
		}
	`)
	fn := prog.Funcs[0]
	if len(fn.Locals) < 2 {
		t.Fatalf("test source produced %d locals, want at least 2", len(fn.Locals))
	}
	fn.Locals[1].Sym.SlotID = fn.Locals[0].Sym.SlotID
	if err := ValidateProgramInvariants(prog); err == nil {
		t.Fatal("invariant check accepted duplicate local slot IDs")
	}
}

func firstVarRefInStmt(stmt Stmt) *VarRef {
	switch x := stmt.(type) {
	case *Block:
		for _, item := range x.Items {
			if ref := firstVarRefInStmt(item); ref != nil {
				return ref
			}
		}
	case *ReturnStmt:
		return firstVarRefInExpr(x.Value)
	case *ExprStmt:
		return firstVarRefInExpr(x.Expr)
	case *DeclStmt:
		for _, decl := range x.Decls {
			if vd, ok := decl.(*VarDecl); ok {
				if ref := firstVarRefInExpr(vd.Init); ref != nil {
					return ref
				}
			}
		}
	}
	return nil
}

func firstVarRefInExpr(expr Expr) *VarRef {
	switch x := expr.(type) {
	case nil:
		return nil
	case *VarRef:
		return x
	case *ImplicitCast:
		return firstVarRefInExpr(x.X)
	case *ExplicitCast:
		return firstVarRefInExpr(x.X)
	case *BinOp:
		if ref := firstVarRefInExpr(x.L); ref != nil {
			return ref
		}
		return firstVarRefInExpr(x.R)
	case *UnOp:
		return firstVarRefInExpr(x.X)
	}
	return nil
}
