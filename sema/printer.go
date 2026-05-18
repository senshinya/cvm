package sema

import (
	"fmt"
	"strings"
)

// PrintProgram 把语义 IR 打印成稳定的树形文本，供 golden 测试对比。
func PrintProgram(p *Program) string {
	if p == nil {
		return "<nil program>\n"
	}
	var b strings.Builder
	b.WriteString("Program\n")
	for _, d := range p.Globals {
		printDecl(&b, 1, d)
	}
	for _, f := range p.Funcs {
		printFuncDef(&b, 1, f)
	}
	return b.String()
}

func printIndent(b *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		b.WriteString("  ")
	}
}

func symbolName(sym *Symbol) string {
	if sym == nil {
		return ""
	}
	return sym.Name
}

func printDecl(b *strings.Builder, depth int, d Decl) {
	printIndent(b, depth)
	switch x := d.(type) {
	case *VarDecl:
		fmt.Fprintf(b, "VarDecl name=%q type=%s storage=%v%s\n", symbolName(x.Sym), x.T, x.Storage, varLayout(x))
		printTypeMetadata(b, depth+1, x.T)
		if x.Init != nil {
			printIndent(b, depth+1)
			b.WriteString("init:\n")
			printExpr(b, depth+2, x.Init)
		}
	case *FuncDecl:
		fmt.Fprintf(b, "FuncDecl name=%q type=%s storage=%v%s\n", symbolName(x.Sym), x.T, x.Storage, globalLayout(x.Sym))
	case *TypedefDecl:
		fmt.Fprintf(b, "TypedefDecl name=%q type=%s\n", symbolName(x.Sym), x.T)
	case *TagDecl:
		fmt.Fprintf(b, "TagDecl tag=%s type=%s\n", tagName(x.Tag), x.T)
	default:
		fmt.Fprintf(b, "%T\n", d)
	}
}

func tagName(tag *TagID) string {
	if tag == nil {
		return ""
	}
	return fmt.Sprintf("#%d", tag.id)
}

func printFuncDef(b *strings.Builder, depth int, f *FuncDef) {
	printIndent(b, depth)
	fmt.Fprintf(b, "FuncDef name=%q type=%s%s\n", symbolName(f.Sym), f.T, globalLayout(f.Sym))
	for _, p := range f.Params {
		printIndent(b, depth+1)
		fmt.Fprintf(b, "Param name=%q type=%s%s\n", symbolName(p.Sym), p.T, slotLayout(p.Sym))
	}
	if f.Body != nil {
		printStmt(b, depth+1, f.Body)
	}
}

func varLayout(vd *VarDecl) string {
	if vd == nil || vd.Sym == nil {
		return ""
	}
	if vd.IsParam || vd.Storage == StorageAuto || vd.Storage == StorageRegister {
		return slotLayout(vd.Sym)
	}
	if vd.Storage == StorageExtern {
		return ""
	}
	return globalLayout(vd.Sym)
}

func globalLayout(sym *Symbol) string {
	if sym == nil {
		return ""
	}
	return fmt.Sprintf(" global=%d", sym.GlobalID)
}

func slotLayout(sym *Symbol) string {
	if sym == nil {
		return ""
	}
	return fmt.Sprintf(" slot=%d", sym.SlotID)
}

func printTypeMetadata(b *strings.Builder, depth int, t Type) {
	switch x := unqual(t).(type) {
	case *ArrayType:
		if x.SizeKind == ArrayVLA {
			printIndent(b, depth)
			b.WriteString("vla-bound:\n")
			printExpr(b, depth+1, x.SizeExpr)
		}
		printTypeMetadata(b, depth, x.Elem)
	case *PointerType:
		printTypeMetadata(b, depth, x.Pointee)
	}
}

func printStmt(b *strings.Builder, depth int, s Stmt) {
	if s == nil {
		return
	}
	printIndent(b, depth)
	switch x := s.(type) {
	case *Block:
		b.WriteString("Block\n")
		for _, it := range x.Items {
			printStmt(b, depth+1, it)
		}
	case *ExprStmt:
		b.WriteString("ExprStmt\n")
		printExpr(b, depth+1, x.Expr)
	case *DeclStmt:
		b.WriteString("DeclStmt\n")
		for _, d := range x.Decls {
			printDecl(b, depth+1, d)
		}
	case *IfStmt:
		b.WriteString("IfStmt\n")
		printNamedExpr(b, depth+1, "cond", x.Cond)
		printNamedStmt(b, depth+1, "then", x.Then)
		if x.Else != nil {
			printNamedStmt(b, depth+1, "else", x.Else)
		}
	case *WhileStmt:
		if x.DoWhile {
			b.WriteString("DoWhileStmt\n")
		} else {
			b.WriteString("WhileStmt\n")
		}
		printNamedExpr(b, depth+1, "cond", x.Cond)
		printNamedStmt(b, depth+1, "body", x.Body)
	case *ForStmt:
		b.WriteString("ForStmt\n")
		printNamedStmt(b, depth+1, "init", x.Init)
		printNamedExpr(b, depth+1, "cond", x.Cond)
		printNamedExpr(b, depth+1, "post", x.Post)
		printNamedStmt(b, depth+1, "body", x.Body)
	case *SwitchStmt:
		b.WriteString("SwitchStmt\n")
		printNamedExpr(b, depth+1, "cond", x.Cond)
		printNamedStmt(b, depth+1, "body", x.Body)
	case *CaseStmt:
		fmt.Fprintf(b, "CaseStmt value=%d\n", x.Value)
		printStmt(b, depth+1, x.Body)
	case *DefaultStmt:
		b.WriteString("DefaultStmt\n")
		printStmt(b, depth+1, x.Body)
	case *ReturnStmt:
		b.WriteString("ReturnStmt\n")
		printExpr(b, depth+1, x.Value)
	case *BreakStmt:
		b.WriteString("BreakStmt\n")
	case *ContinueStmt:
		b.WriteString("ContinueStmt\n")
	case *GotoStmt:
		fmt.Fprintf(b, "GotoStmt target=%q\n", labelName(x))
	case *LabeledStmt:
		fmt.Fprintf(b, "LabeledStmt name=%q\n", x.Name)
		printStmt(b, depth+1, x.Body)
	case *EmptyStmt:
		b.WriteString("EmptyStmt\n")
	default:
		fmt.Fprintf(b, "%T\n", s)
	}
}

func printNamedExpr(b *strings.Builder, depth int, name string, e Expr) {
	if e == nil {
		return
	}
	printIndent(b, depth)
	fmt.Fprintf(b, "%s:\n", name)
	printExpr(b, depth+1, e)
}

func printNamedStmt(b *strings.Builder, depth int, name string, s Stmt) {
	if s == nil {
		return
	}
	printIndent(b, depth)
	fmt.Fprintf(b, "%s:\n", name)
	printStmt(b, depth+1, s)
}

func labelName(g *GotoStmt) string {
	if g.Target != nil {
		return g.Target.Name
	}
	return g.Name
}

func printExpr(b *strings.Builder, depth int, e Expr) {
	if e == nil {
		return
	}
	printIndent(b, depth)
	switch x := e.(type) {
	case *IntLit:
		fmt.Fprintf(b, "IntLit value=%d type=%s\n", x.Value, x.T)
	case *FloatLit:
		fmt.Fprintf(b, "FloatLit value=%v type=%s\n", x.Value, x.T)
	case *ImagLit:
		fmt.Fprintf(b, "ImagLit value=%v type=%s\n", x.Value, x.T)
	case *CharLit:
		fmt.Fprintf(b, "CharLit value=%d type=%s\n", x.Value, x.T)
	case *StringLit:
		fmt.Fprintf(b, "StringLit value=%q type=%s\n", x.Value, x.T)
	case *VarRef:
		fmt.Fprintf(b, "VarRef name=%q type=%s\n", symbolName(x.Sym), x.T)
	case *EnumRef:
		name, value := "", int64(0)
		if x.Enumerator != nil {
			name, value = x.Enumerator.Name, x.Enumerator.Value
		}
		fmt.Fprintf(b, "EnumRef name=%q value=%d type=%s\n", name, value, x.T)
	case *BinOp:
		fmt.Fprintf(b, "BinOp op=%v type=%s\n", x.Op, x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *UnOp:
		fmt.Fprintf(b, "UnOp op=%v type=%s category=%v\n", x.Op, x.T, x.Category)
		printExpr(b, depth+1, x.X)
	case *AssignExpr:
		fmt.Fprintf(b, "AssignExpr type=%s\n", x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *CompoundAssign:
		fmt.Fprintf(b, "CompoundAssign op=%v type=%s\n", x.Op, x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *CallExpr:
		fmt.Fprintf(b, "CallExpr type=%s\n", x.T)
		printNamedExpr(b, depth+1, "callee", x.Callee)
		for _, a := range x.Args {
			printExpr(b, depth+1, a)
		}
	case *MemberExpr:
		arrow := "."
		if x.Arrow {
			arrow = "->"
		}
		field := ""
		if x.Field != nil {
			field = x.Field.Name
		}
		fmt.Fprintf(b, "MemberExpr access=%s%s type=%s category=%v\n", arrow, field, x.T, x.Category)
		printExpr(b, depth+1, x.Base)
	case *IndexExpr:
		fmt.Fprintf(b, "IndexExpr type=%s\n", x.T)
		printExpr(b, depth+1, x.Base)
		printExpr(b, depth+1, x.Index)
	case *CondExpr:
		fmt.Fprintf(b, "CondExpr type=%s\n", x.T)
		printExpr(b, depth+1, x.Cond)
		printExpr(b, depth+1, x.Then)
		printExpr(b, depth+1, x.Else)
	case *SizeofExpr:
		fmt.Fprintf(b, "SizeofExpr type=%s\n", x.T)
		if x.Operand.Type != nil {
			printIndent(b, depth+1)
			fmt.Fprintf(b, "operand-type: %s\n", x.Operand.Type)
		}
		printExpr(b, depth+1, x.Operand.Expr)
	case *CommaExpr:
		fmt.Fprintf(b, "CommaExpr type=%s\n", x.T)
		printExpr(b, depth+1, x.L)
		printExpr(b, depth+1, x.R)
	case *CompoundLit:
		fmt.Fprintf(b, "CompoundLit type=%s\n", x.T)
		printExpr(b, depth+1, x.Init)
	case *InitList:
		fmt.Fprintf(b, "InitList type=%s\n", x.T)
		for _, el := range x.Elems {
			printDesignators(b, depth+1, el.Designators)
			printExpr(b, depth+1, el.Value)
		}
	case *ImplicitCast:
		fmt.Fprintf(b, "ImplicitCast kind=%v from=%s to=%s\n", x.Kind, x.From, x.To)
		printExpr(b, depth+1, x.X)
	case *ExplicitCast:
		fmt.Fprintf(b, "ExplicitCast to=%s\n", x.To)
		printExpr(b, depth+1, x.X)
	case *AddrConst:
		fmt.Fprintf(b, "AddrConst symbol=%q offset=%d type=%s\n", symbolName(x.Sym), x.Offset, x.T)
	default:
		fmt.Fprintf(b, "%T type=%s\n", e, e.GetType())
	}
}

func printDesignators(b *strings.Builder, depth int, designators []Designator) {
	for _, d := range designators {
		printIndent(b, depth)
		switch d.Kind {
		case DesigArrayIndex:
			fmt.Fprintf(b, "Designator index=%d\n", d.Index)
		case DesigFieldName:
			field := ""
			if d.Field != nil {
				field = d.Field.Name
			}
			fmt.Fprintf(b, "Designator field=%q\n", field)
		default:
			fmt.Fprintf(b, "Designator kind=%v\n", d.Kind)
		}
	}
}
