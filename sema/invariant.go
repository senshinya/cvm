package sema

import "fmt"

// ValidateProgramInvariants checks IR properties that codegen relies on.
func ValidateProgramInvariants(prog *Program) error {
	if prog == nil {
		return fmt.Errorf("program is nil")
	}
	v := invariantValidator{
		globalIDs:      map[int]*Symbol{},
		globalIDLabels: map[int]string{},
	}
	for _, decl := range prog.Globals {
		if err := v.checkDecl(decl); err != nil {
			return err
		}
	}
	for _, fn := range prog.Funcs {
		if err := v.checkFunc(fn); err != nil {
			return err
		}
	}
	return nil
}

type invariantValidator struct {
	globalIDs      map[int]*Symbol
	globalIDLabels map[int]string
	slotIDs        map[int]*Symbol
	slotIDLabels   map[int]string
}

func (v *invariantValidator) checkFunc(fn *FuncDef) error {
	if fn == nil {
		return fmt.Errorf("function definition is nil")
	}
	if fn.Sym == nil {
		return fmt.Errorf("function definition has nil symbol")
	}
	if fn.T == nil {
		return fmt.Errorf("function %q has nil type", fn.Sym.Name)
	}
	if err := v.checkGlobalID(fn.Sym, "function "+fn.Sym.Name); err != nil {
		return err
	}
	v.slotIDs = map[int]*Symbol{}
	v.slotIDLabels = map[int]string{}
	for _, param := range fn.Params {
		if err := v.checkDecl(param); err != nil {
			return err
		}
	}
	for _, local := range fn.Locals {
		if err := v.checkDecl(local); err != nil {
			return err
		}
	}
	if fn.Body == nil {
		return fmt.Errorf("function %q has nil body", fn.Sym.Name)
	}
	if err := v.checkStmt(fn.Body); err != nil {
		return err
	}
	v.slotIDs = nil
	return nil
}

func (v *invariantValidator) checkDecl(decl Decl) error {
	switch d := decl.(type) {
	case nil:
		return fmt.Errorf("declaration is nil")
	case *VarDecl:
		if d.Sym == nil {
			return fmt.Errorf("variable declaration has nil symbol")
		}
		if d.T == nil {
			return fmt.Errorf("variable %q has nil type", d.Sym.Name)
		}
		if v.slotIDs == nil {
			if err := v.checkGlobalID(d.Sym, "variable "+d.Sym.Name); err != nil {
				return err
			}
			return v.checkExpr(d.Init)
		}
		if d.Storage == StorageStatic {
			if err := v.checkGlobalID(d.Sym, "variable "+d.Sym.Name); err != nil {
				return err
			}
		}
		if d.Storage != StorageStatic && d.Storage != StorageExtern {
			if err := v.checkSlotID(d.Sym, "variable "+d.Sym.Name); err != nil {
				return err
			}
		}
		return v.checkExpr(d.Init)
	case *FuncDecl:
		if d.Sym == nil {
			return fmt.Errorf("function declaration has nil symbol")
		}
		if d.T == nil {
			return fmt.Errorf("function declaration %q has nil type", d.Sym.Name)
		}
		return v.checkGlobalID(d.Sym, "function declaration "+d.Sym.Name)
	case *TypedefDecl:
		if d.Sym == nil {
			return fmt.Errorf("typedef declaration has nil symbol")
		}
		if d.T == nil {
			return fmt.Errorf("typedef %q has nil type", d.Sym.Name)
		}
	case *TagDecl:
		if d.T == nil {
			return fmt.Errorf("tag declaration has nil type")
		}
	}
	return nil
}

func (v *invariantValidator) checkGlobalID(sym *Symbol, label string) error {
	if sym == nil {
		return fmt.Errorf("%s has nil symbol", label)
	}
	if sym.GlobalID < 0 {
		return fmt.Errorf("%s has no global ID", label)
	}
	if prev, ok := v.globalIDs[sym.GlobalID]; ok && prev != sym {
		return fmt.Errorf("duplicate global ID %d for %s and %s", sym.GlobalID, v.globalIDLabels[sym.GlobalID], label)
	}
	v.globalIDs[sym.GlobalID] = sym
	v.globalIDLabels[sym.GlobalID] = label
	return nil
}

func (v *invariantValidator) checkSlotID(sym *Symbol, label string) error {
	if sym == nil {
		return fmt.Errorf("%s has nil symbol", label)
	}
	if sym.SlotID < 0 {
		return fmt.Errorf("%s has no slot ID", label)
	}
	if prev, ok := v.slotIDs[sym.SlotID]; ok && prev != sym {
		return fmt.Errorf("duplicate slot ID %d for %s and %s", sym.SlotID, v.slotIDLabels[sym.SlotID], label)
	}
	v.slotIDs[sym.SlotID] = sym
	v.slotIDLabels[sym.SlotID] = label
	return nil
}

func (v *invariantValidator) checkStmt(stmt Stmt) error {
	switch s := stmt.(type) {
	case nil:
		return nil
	case *Block:
		for _, item := range s.Items {
			if err := v.checkStmt(item); err != nil {
				return err
			}
		}
	case *IfStmt:
		if err := v.checkExpr(s.Cond); err != nil {
			return err
		}
		if err := v.checkStmt(s.Then); err != nil {
			return err
		}
		return v.checkStmt(s.Else)
	case *WhileStmt:
		if err := v.checkExpr(s.Cond); err != nil {
			return err
		}
		return v.checkStmt(s.Body)
	case *ForStmt:
		if err := v.checkStmt(s.Init); err != nil {
			return err
		}
		if err := v.checkExpr(s.Cond); err != nil {
			return err
		}
		if err := v.checkExpr(s.Post); err != nil {
			return err
		}
		return v.checkStmt(s.Body)
	case *SwitchStmt:
		if err := v.checkExpr(s.Cond); err != nil {
			return err
		}
		return v.checkStmt(s.Body)
	case *CaseStmt:
		return v.checkStmt(s.Body)
	case *DefaultStmt:
		return v.checkStmt(s.Body)
	case *ReturnStmt:
		return v.checkExpr(s.Value)
	case *GotoStmt:
		if s.Target == nil {
			return fmt.Errorf("goto %q has nil target", s.Name)
		}
	case *LabeledStmt:
		return v.checkStmt(s.Body)
	case *ExprStmt:
		return v.checkExpr(s.Expr)
	case *DeclStmt:
		for _, decl := range s.Decls {
			if err := v.checkDecl(decl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *invariantValidator) checkExpr(expr Expr) error {
	if expr == nil {
		return nil
	}
	if expr.GetType() == nil {
		return fmt.Errorf("%T has nil type", expr)
	}
	switch e := expr.(type) {
	case *VarRef:
		if e.Sym == nil {
			return fmt.Errorf("VarRef has nil symbol")
		}
	case *EnumRef:
		if e.Enumerator == nil {
			return fmt.Errorf("EnumRef has nil enumerator")
		}
	case *BinOp:
		if err := v.checkExpr(e.L); err != nil {
			return err
		}
		return v.checkExpr(e.R)
	case *UnOp:
		return v.checkExpr(e.X)
	case *AssignExpr:
		if e.L == nil || e.L.GetCategory() != LValue {
			return fmt.Errorf("assignment left operand is not an lvalue")
		}
		if err := v.checkExpr(e.L); err != nil {
			return err
		}
		return v.checkExpr(e.R)
	case *CompoundAssign:
		if e.L == nil || e.L.GetCategory() != LValue {
			return fmt.Errorf("compound assignment left operand is not an lvalue")
		}
		if err := v.checkExpr(e.L); err != nil {
			return err
		}
		return v.checkExpr(e.R)
	case *CallExpr:
		if e.Callee == nil {
			return fmt.Errorf("call has nil callee")
		}
		if !isFunctionPointerType(e.Callee.GetType()) {
			return fmt.Errorf("call callee has non-function-pointer type %s", e.Callee.GetType())
		}
		if err := v.checkExpr(e.Callee); err != nil {
			return err
		}
		for _, arg := range e.Args {
			if err := v.checkExpr(arg); err != nil {
				return err
			}
		}
	case *MemberExpr:
		if e.Field == nil {
			return fmt.Errorf("MemberExpr has nil field")
		}
		return v.checkExpr(e.Base)
	case *IndexExpr:
		if err := v.checkExpr(e.Base); err != nil {
			return err
		}
		return v.checkExpr(e.Index)
	case *CondExpr:
		if err := v.checkExpr(e.Cond); err != nil {
			return err
		}
		if err := v.checkExpr(e.Then); err != nil {
			return err
		}
		return v.checkExpr(e.Else)
	case *SizeofExpr:
		return v.checkExpr(e.Operand.Expr)
	case *CommaExpr:
		if err := v.checkExpr(e.L); err != nil {
			return err
		}
		return v.checkExpr(e.R)
	case *CompoundLit:
		return v.checkExpr(e.Init)
	case *InitList:
		for _, elem := range e.Elems {
			if err := v.checkExpr(elem.Value); err != nil {
				return err
			}
		}
	case *ImplicitCast:
		if e.From == nil || e.To == nil {
			return fmt.Errorf("ImplicitCast has nil type")
		}
		return v.checkExpr(e.X)
	case *ExplicitCast:
		return v.checkExpr(e.X)
	case *AddrConst:
		if e.Sym == nil {
			return fmt.Errorf("AddrConst has nil symbol")
		}
	case *StmtExpr:
		return v.checkStmt(e.Block)
	}
	return nil
}

func isFunctionPointerType(t Type) bool {
	ptr, ok := unqual(t).(*PointerType)
	if !ok {
		return false
	}
	_, ok = unqual(ptr.Pointee).(*FunctionType)
	return ok
}
