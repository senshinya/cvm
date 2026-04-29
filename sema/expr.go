package sema

import (
	"strconv"
	"strings"

	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

func (s *Sema) typeExpr(node *entity.AstNode, scope *Scope) Expr {
	if node == nil {
		return s.errorExpr(entity.SourceRange{})
	}
	for len(node.Children) == 1 && node.Terminal == nil && !isMeaningfulSingleChildExpr(node.Typ) {
		node = node.Children[0]
	}
	switch node.Typ {
	case parser.PrimaryExpression:
		return s.typePrimary(node, scope)
	case parser.PostfixExpression:
		return s.typePostfix(node, scope)
	case parser.UnaryExpression:
		return s.typeUnary(node, scope)
	case parser.CastExpression:
		return s.typeCast(node, scope)
	case parser.MultiplicativeExpression, parser.AdditiveExpression, parser.ShiftExpression,
		parser.RelationalExpression, parser.EqualityExpression, parser.AndExpression,
		parser.ExclusiveOrExpression, parser.InclusiveOrExpression, parser.LogicalAndExpression,
		parser.LogicalOrExpression:
		return s.typeBinaryExpression(node, scope)
	case parser.ConditionalExpression:
		return s.typeConditional(node, scope)
	case parser.AssignmentExpression:
		return s.typeAssignment(node, scope)
	case parser.Expression:
		return s.typeComma(node, scope)
	case parser.ConstantExpression:
		return s.typeExpr(node.Children[0], scope)
	case entity.INTEGER_CONSTANT:
		return s.makeIntLit(node)
	case entity.FLOATING_CONSTANT:
		return s.makeFloatLit(node)
	case entity.CHARACTER:
		return s.makeCharLit(node)
	case entity.STRING:
		return s.makeStringLit(node)
	case entity.IDENTIFIER:
		return s.lookupVar(node, scope)
	}
	if len(node.Children) == 1 {
		return s.typeExpr(node.Children[0], scope)
	}
	s.report(InvalidTypeSpec(node.SourceStart, "unsupported expression form"))
	return s.errorExpr(node.SourceRange)
}

func isMeaningfulSingleChildExpr(t entity.TokenType) bool {
	switch t {
	case parser.PrimaryExpression, parser.PostfixExpression, parser.UnaryExpression, parser.CastExpression,
		parser.AssignmentExpression, parser.ConditionalExpression, parser.Expression:
		return true
	}
	return false
}

func (s *Sema) errorExpr(r entity.SourceRange) Expr {
	return &IntLit{T: ErrorTypeSingleton, Range: r}
}

func (s *Sema) makeIntLit(node *entity.AstNode) Expr {
	lexeme := node.Terminal.Lexeme
	if s.Options.PedanticErrors && signedIntegerLiteralOverflow(lexeme) {
		s.report(InvalidTypeSpec(node.SourceStart, "integer constant is too large for signed type"))
	}
	s.validateIntegerLiteralPrefix(lexeme, node.SourceStart)
	return &IntLit{Value: parseIntLiteral(lexeme), T: integerLiteralType(s, lexeme), Range: node.SourceRange}
}

func (s *Sema) validateIntegerLiteralPrefix(lexeme string, pos entity.SourcePos) {
	lower := strings.ToLower(lexeme)
	switch {
	case strings.HasPrefix(lower, "0b"):
		if s.Options.PedanticErrors && !s.Options.GNUExtensions {
			s.report(InvalidTypeSpec(pos, "binary integer constants require GNU C mode"))
			return
		}
		if !validBasedIntegerLiteral(lower[2:], 2) {
			s.report(InvalidTypeSpec(pos, "invalid binary integer constant"))
		}
	case strings.HasPrefix(lower, "0o"):
		if s.Options.PedanticErrors && !s.Options.GNUExtensions {
			s.report(InvalidTypeSpec(pos, "0o integer constants require GNU C mode"))
			return
		}
		if !validBasedIntegerLiteral(lower[2:], 8) {
			s.report(InvalidTypeSpec(pos, "invalid octal integer constant"))
		}
	}
}

func validBasedIntegerLiteral(body string, base int) bool {
	digits := 0
	for len(body) > 0 {
		c := body[0]
		if c == 'u' || c == 'l' {
			break
		}
		switch base {
		case 2:
			if c != '0' && c != '1' {
				return false
			}
		case 8:
			if c < '0' || c > '7' {
				return false
			}
		}
		digits++
		body = body[1:]
	}
	if digits == 0 {
		return false
	}
	for _, c := range body {
		if c != 'u' && c != 'l' {
			return false
		}
	}
	return true
}

func (s *Sema) makeFloatLit(node *entity.AstNode) Expr {
	if invalidFloatSuffix(node.Terminal.Lexeme, s.Options.PedanticErrors) {
		s.report(InvalidTypeSpec(node.SourceStart, "invalid floating constant suffix"))
	}
	return &FloatLit{Value: parseFloatLiteral(node.Terminal.Lexeme), T: s.Types.Builtin(Double), Range: node.SourceRange}
}

func invalidFloatSuffix(lexeme string, pedanticErrors bool) bool {
	i := len(lexeme)
	for i > 0 {
		c := lexeme[i-1]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			i--
			continue
		}
		break
	}
	suffix := strings.ToLower(lexeme[i:])
	if suffix == "" || suffix == "f" || suffix == "l" {
		return false
	}
	if pedanticErrors && strings.ContainsAny(suffix, "dij") {
		return true
	}
	if suffix == "d" || suffix == "i" || suffix == "j" {
		return false
	}
	if len(suffix) == 2 && strings.ContainsAny(suffix, "fld") && (strings.Contains(suffix, "i") || strings.Contains(suffix, "j")) {
		return false
	}
	return true
}

func (s *Sema) makeCharLit(node *entity.AstNode) Expr {
	return &CharLit{Value: parseCharLiteral(node.Terminal.Lexeme), T: s.Types.Builtin(Int), Range: node.SourceRange}
}

func (s *Sema) makeStringLit(node *entity.AstNode) Expr {
	v := parseStringLiteral(node.Terminal.Lexeme)
	return &StringLit{Value: v, T: s.Types.ArrayConstant(s.Types.Builtin(Char), int64(len(v)+1)), Range: node.SourceRange}
}

func parseIntLiteral(lexeme string) int64 {
	body := integerLiteralBody(lexeme)
	v, _ := parseBasedIntLiteral(body)
	return int64(v)
}

func integerLiteralType(s *Sema, lexeme string) Type {
	suffix := integerSuffix(lexeme)
	body := integerLiteralBody(lexeme)
	value, err := parseBasedIntLiteral(body)
	if err != nil {
		return s.Types.Builtin(ULongLong)
	}
	for _, k := range integerLiteralCandidates(body, suffix) {
		if uintValueFitsBuiltin(value, k) {
			return s.Types.Builtin(k)
		}
	}
	return s.Types.Builtin(ULongLong)
}

func parseBasedIntLiteral(body string) (uint64, error) {
	lower := strings.ToLower(body)
	if strings.HasPrefix(lower, "0b") {
		return strconv.ParseUint(lower[2:], 2, 64)
	}
	if strings.HasPrefix(lower, "0o") {
		return strconv.ParseUint(lower[2:], 8, 64)
	}
	return strconv.ParseUint(body, 0, 64)
}

func integerLiteralBody(lexeme string) string {
	return lexeme[:len(lexeme)-len(integerSuffix(lexeme))]
}

func integerSuffix(lexeme string) string {
	i := len(lexeme)
	for i > 0 {
		c := lexeme[i-1]
		if c == 'u' || c == 'U' || c == 'l' || c == 'L' {
			i--
			continue
		}
		break
	}
	return strings.ToLower(lexeme[i:])
}

func integerLiteralCandidates(body, suffix string) []BuiltinKind {
	unsigned := strings.Contains(suffix, "u")
	longs := strings.Count(suffix, "l")
	decimal := integerLiteralIsDecimal(body)
	switch {
	case unsigned && longs >= 2:
		return []BuiltinKind{ULongLong}
	case unsigned && longs == 1:
		return []BuiltinKind{ULong, ULongLong}
	case unsigned:
		return []BuiltinKind{UInt, ULong, ULongLong}
	case longs >= 2 && decimal:
		return []BuiltinKind{LongLong}
	case longs >= 2:
		return []BuiltinKind{LongLong, ULongLong}
	case longs == 1 && decimal:
		return []BuiltinKind{Long, LongLong}
	case longs == 1:
		return []BuiltinKind{Long, ULong, LongLong, ULongLong}
	case decimal:
		return []BuiltinKind{Int, Long, LongLong}
	default:
		return []BuiltinKind{Int, UInt, Long, ULong, LongLong, ULongLong}
	}
}

func uintValueFitsBuiltin(v uint64, k BuiltinKind) bool {
	if isSignedIntegerKind(k) {
		return v <= uint64MaxForSignedKind(k)
	}
	return integerValueBits(k) >= 64 || v < (uint64(1)<<uint(integerValueBits(k)))
}

func uint64MaxForSignedKind(k BuiltinKind) uint64 {
	bits := integerValueBits(k)
	if bits >= 64 {
		return uint64(int64Max)
	}
	return (uint64(1) << uint(bits-1)) - 1
}

func signedIntegerLiteralOverflow(lexeme string) bool {
	suffix := integerSuffix(lexeme)
	if strings.Contains(suffix, "u") {
		return false
	}
	body := integerLiteralBody(lexeme)
	value, err := strconv.ParseUint(body, 0, 64)
	if err != nil {
		return true
	}
	for _, k := range integerLiteralCandidates(body, suffix) {
		if uintValueFitsBuiltin(value, k) {
			return false
		}
	}
	return true
}

func integerLiteralIsDecimal(body string) bool {
	return !strings.HasPrefix(body, "0x") && !strings.HasPrefix(body, "0X") && !(len(body) > 1 && body[0] == '0')
}

func parseFloatLiteral(lexeme string) float64 {
	s := strings.TrimRight(lexeme, "fFlLdDiIjJ")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseCharLiteral(lexeme string) int32 {
	s := strings.Trim(lexeme, "'")
	if strings.HasPrefix(s, "\\") {
		switch s {
		case "\\n":
			return '\n'
		case "\\t":
			return '\t'
		case "\\r":
			return '\r'
		case "\\0":
			return 0
		case "\\\\":
			return '\\'
		case "\\'":
			return '\''
		case "\\\"":
			return '"'
		}
	}
	if len(s) == 0 {
		return 0
	}
	return int32(s[0])
}

func parseStringLiteral(lexeme string) string {
	if len(lexeme) < 2 {
		return ""
	}
	return strings.ReplaceAll(lexeme[1:len(lexeme)-1], `\n`, "\n")
}

func (s *Sema) typePrimary(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.PrimaryExpression, 1):
		return s.lookupVar(node.Children[0], scope)
	case node.ReducedBy(parser.PrimaryExpression, 2):
		return s.makeStringLit(node.Children[0])
	case node.ReducedBy(parser.PrimaryExpression, 3):
		return s.makeCharLit(node.Children[0])
	case node.ReducedBy(parser.PrimaryExpression, 4):
		return s.makeIntLit(node.Children[0])
	case node.ReducedBy(parser.PrimaryExpression, 5):
		return s.makeFloatLit(node.Children[0])
	case node.ReducedBy(parser.PrimaryExpression, 6):
		return s.typeExpr(node.Children[1], scope)
	case node.ReducedBy(parser.PrimaryExpression, 7):
		if !s.Options.GNUExtensions {
			s.report(InvalidTypeSpec(node.SourceStart, "statement expression requires GNU C mode"))
		}
		block := s.typeBlock(node.Children[1], scope, &funcCtx{})
		return &StmtExpr{Block: block, T: blockResultType(s, block), Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func blockResultType(s *Sema, block *Block) Type {
	if block == nil || len(block.Items) == 0 {
		return s.Types.Builtin(Void)
	}
	if stmt, ok := block.Items[len(block.Items)-1].(*ExprStmt); ok && stmt.Expr != nil {
		return stmt.Expr.GetType()
	}
	return s.Types.Builtin(Void)
}

func (s *Sema) lookupVar(node *entity.AstNode, scope *Scope) Expr {
	name := node.Terminal.Lexeme
	sym := scope.Lookup(name, NSOrdinary)
	if sym == nil {
		if builtin := s.lookupBuiltin(name, node.SourceStart); builtin != nil {
			return &VarRef{Sym: builtin, T: builtin.T, Range: node.SourceRange}
		}
		if s.Options.Permissive {
			ft := s.Types.Function(s.Types.Builtin(Int), nil, false, false)
			sym = &Symbol{Name: name, Kind: SymFunc, T: ft, Storage: StorageExtern, Linkage: LinkageExternal, Pos: node.SourceStart}
			s.SymTab.File.Insert(name, sym)
			return &VarRef{Sym: sym, T: sym.T, Range: node.SourceRange}
		}
		s.report(UndeclaredIdentifier(node.SourceStart, name))
		return s.errorExpr(node.SourceRange)
	}
	switch sym.Kind {
	case SymVar, SymParam, SymFunc:
		return &VarRef{Sym: sym, T: sym.T, Range: node.SourceRange}
	case SymEnumerator:
		var enumerator *Enumerator
		if et, ok := unqual(sym.T).(*EnumType); ok {
			for _, e := range et.Enumerators {
				if e.Name == name {
					enumerator = e
					break
				}
			}
		}
		if enumerator == nil {
			enumerator = &Enumerator{Name: name}
		}
		return &EnumRef{Enumerator: enumerator, T: sym.T, Range: node.SourceRange}
	case SymTypedef:
		s.report(InvalidTypeSpec(node.SourceStart, "typedef name used as expression"))
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) typeBinaryExpression(node *entity.AstNode, scope *Scope) Expr {
	if len(node.Children) != 3 {
		return s.typeExpr(node.Children[0], scope)
	}
	op := s.binaryOpFromOperator(node.Children[1])
	l := s.typeExpr(node.Children[0], scope)
	r := s.typeExpr(node.Children[2], scope)
	l = s.castArrayDecay(s.castLValueToRValue(l))
	r = s.castArrayDecay(s.castLValueToRValue(r))
	switch op {
	case OpLAnd, OpLOr:
		l = s.castBoolConversion(l)
		r = s.castBoolConversion(r)
		return &BinOp{Op: op, L: l, R: r, T: s.Types.Builtin(Int), Range: node.SourceRange}
	case OpEq, OpNe, OpLt, OpLe, OpGt, OpGe:
		l, r = s.balanceComparison(op, l, r, node.SourceStart)
		return &BinOp{Op: op, L: l, R: r, T: s.Types.Builtin(Int), Range: node.SourceRange}
	case OpShl, OpShr:
		l = s.castIntegerPromotion(l)
		r = s.castIntegerPromotion(r)
		return &BinOp{Op: op, L: l, R: r, T: l.GetType(), Range: node.SourceRange}
	default:
		if isPointer(l.GetType()) || isPointer(r.GetType()) {
			return s.typePointerArithmetic(op, l, r, node.SourceRange)
		}
		l, r, common := s.castUsualArithmetic(l, r)
		return &BinOp{Op: op, L: l, R: r, T: common, Range: node.SourceRange}
	}
}

func (s *Sema) binaryOpFromOperator(opNode *entity.AstNode) BinaryOp {
	switch opNode.Typ {
	case entity.PLUS:
		return OpAdd
	case entity.MINUS:
		return OpSub
	case entity.ASTERISK:
		return OpMul
	case entity.SLASH:
		return OpDiv
	case entity.PERCENT:
		return OpMod
	case entity.AND:
		return OpAnd
	case entity.OR:
		return OpOr
	case entity.XOR:
		return OpXor
	case entity.LEFT_SHIFT:
		return OpShl
	case entity.RIGHT_SHIFT:
		return OpShr
	case entity.AND_AND:
		return OpLAnd
	case entity.OR_OR:
		return OpLOr
	case entity.EQUAL_EQUAL:
		return OpEq
	case entity.NOT_EQUAL:
		return OpNe
	case entity.LESS:
		return OpLt
	case entity.LESS_EQUAL:
		return OpLe
	case entity.GREATER:
		return OpGt
	case entity.GREATER_EQUAL:
		return OpGe
	}
	return OpAdd
}

func (s *Sema) balanceComparison(op BinaryOp, l, r Expr, pos entity.SourcePos) (Expr, Expr) {
	if isArithmetic(l.GetType()) && isArithmetic(r.GetType()) {
		l, r, _ = s.castUsualArithmetic(l, r)
		return l, r
	}
	if op == OpEq || op == OpNe {
		if isPointer(l.GetType()) && s.isComparisonNullPointerConstant(r) {
			r = s.castNullPointerConstant(r, l.GetType())
			return l, r
		}
		if isPointer(r.GetType()) && s.isComparisonNullPointerConstant(l) {
			l = s.castNullPointerConstant(l, r.GetType())
			return l, r
		}
	}
	if isPointer(l.GetType()) && isPointer(r.GetType()) {
		if comparisonPointerCompatible(op, l.GetType(), r.GetType()) {
			if s.Options.PedanticErrors && isRelationalOp(op) {
				lp := unqual(l.GetType()).(*PointerType)
				rp := unqual(r.GetType()).(*PointerType)
				if completeIncompleteArrayMismatch(lp.Pointee, rp.Pointee) {
					s.report(InvalidTypeSpec(pos, "comparison of complete and incomplete pointer types"))
				}
			}
			if l.GetType() != r.GetType() {
				r = s.castPointerConversion(r, l.GetType())
			}
			return l, r
		}
		s.report(InvalidTypeSpec(pos, "invalid operands for comparison"))
		return l, r
	}
	s.report(InvalidTypeSpec(pos, "invalid operands for comparison"))
	return l, r
}

func isRelationalOp(op BinaryOp) bool {
	return op == OpLt || op == OpLe || op == OpGt || op == OpGe
}

func (s *Sema) isComparisonNullPointerConstant(e Expr) bool {
	return s.isNullPointerConstant(e) || s.isVoidPointerZero(e)
}

func comparisonPointerCompatible(op BinaryOp, left, right Type) bool {
	lp, lok := unqual(left).(*PointerType)
	rp, rok := unqual(right).(*PointerType)
	if !lok || !rok {
		return false
	}
	if op == OpEq || op == OpNe {
		return equalityPointerCompatible(lp, rp)
	}
	return relationalPointerCompatible(lp, rp)
}

func equalityPointerCompatible(left, right *PointerType) bool {
	leftFunc := isFunctionPointer(left)
	rightFunc := isFunctionPointer(right)
	if leftFunc || rightFunc {
		return leftFunc && rightFunc && compatibleType(left.Pointee, right.Pointee)
	}
	if isVoidPointer(left) || isVoidPointer(right) {
		return true
	}
	return compatibleTypeIgnoringTopLevelQualifiers(left.Pointee, right.Pointee)
}

func relationalPointerCompatible(left, right *PointerType) bool {
	if isFunctionPointer(left) || isFunctionPointer(right) || isVoidPointer(left) || isVoidPointer(right) {
		return false
	}
	return compatibleTypeIgnoringTopLevelQualifiers(left.Pointee, right.Pointee)
}

func (s *Sema) typePointerArithmetic(op BinaryOp, l, r Expr, srcRange entity.SourceRange) Expr {
	switch op {
	case OpAdd, OpSub:
		if isPointer(l.GetType()) && isArithmetic(r.GetType()) {
			s.validatePointerArithmeticOperand(l.GetType(), srcRange.SourceStart)
			return &BinOp{Op: op, L: l, R: s.castIntegerPromotion(r), T: l.GetType(), Range: srcRange}
		}
		if op == OpAdd && isArithmetic(l.GetType()) && isPointer(r.GetType()) {
			s.validatePointerArithmeticOperand(r.GetType(), srcRange.SourceStart)
			return &BinOp{Op: op, L: s.castIntegerPromotion(l), R: r, T: r.GetType(), Range: srcRange}
		}
		if op == OpSub && isPointer(l.GetType()) && isPointer(r.GetType()) {
			s.validatePointerArithmeticOperand(l.GetType(), srcRange.SourceStart)
			s.validatePointerArithmeticOperand(r.GetType(), srcRange.SourceStart)
			return &BinOp{Op: op, L: l, R: r, T: s.Types.Builtin(Long), Range: srcRange}
		}
	}
	s.report(InvalidTypeSpec(srcRange.SourceStart, "invalid pointer arithmetic"))
	return &BinOp{Op: op, L: l, R: r, T: ErrorTypeSingleton, Range: srcRange}
}

func (s *Sema) validatePointerArithmeticOperand(t Type, pos entity.SourcePos) {
	pt, ok := unqual(t).(*PointerType)
	if !ok {
		return
	}
	if !isObjectType(pt.Pointee) {
		s.report(InvalidTypeSpec(pos, "pointer arithmetic requires pointer to complete object type"))
	}
}

func (s *Sema) typeUnary(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.UnaryExpression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.UnaryExpression, 2):
		return s.typeIncDec(node, scope, UnIncPre)
	case node.ReducedBy(parser.UnaryExpression, 3):
		return s.typeIncDec(node, scope, UnDecPre)
	case node.ReducedBy(parser.UnaryExpression, 4):
		return s.typeUnaryOperator(node, scope)
	case node.ReducedBy(parser.UnaryExpression, 5):
		x := s.typeExpr(node.Children[1], scope)
		return &SizeofExpr{Operand: SizeofOperand{Expr: x}, T: s.Types.Builtin(ULong), Range: node.SourceRange}
	case node.ReducedBy(parser.UnaryExpression, 6):
		return &SizeofExpr{Operand: SizeofOperand{Type: s.parseTypeName(node.Children[2])}, T: s.Types.Builtin(ULong), Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) typeIncDec(node *entity.AstNode, scope *Scope, op UnaryOp) Expr {
	x := s.typeExpr(node.Children[1], scope)
	return s.buildIncDec(node, x, op)
}

func (s *Sema) buildIncDec(node *entity.AstNode, x Expr, op UnaryOp) Expr {
	if x.GetCategory() != LValue {
		s.report(InvalidTypeSpec(node.SourceStart, "operand of ++/-- must be lvalue"))
		return &UnOp{Op: op, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
	}
	if isComplexType(x.GetType()) {
		s.report(InvalidTypeSpec(node.SourceStart, "operand of ++/-- cannot have complex type"))
		return &UnOp{Op: op, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
	}
	if isPointer(x.GetType()) {
		s.validatePointerArithmeticOperand(x.GetType(), node.SourceStart)
	}
	return &UnOp{Op: op, X: x, T: x.GetType(), Category: RValue, Range: node.SourceRange}
}

func (s *Sema) typeUnaryOperator(node *entity.AstNode, scope *Scope) Expr {
	opNode := node.Children[0].Children[0]
	x := s.typeExpr(node.Children[1], scope)
	switch opNode.Typ {
	case entity.AND:
		if x.GetCategory() != LValue {
			s.report(InvalidTypeSpec(node.SourceStart, "cannot take address of rvalue"))
			return &UnOp{Op: UnAddr, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
		}
		if vr, ok := x.(*VarRef); ok && vr.Sym != nil && vr.Sym.Storage == StorageRegister {
			s.report(InvalidTypeSpec(node.SourceStart, "cannot take address of register variable"))
			return &UnOp{Op: UnAddr, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
		}
		if vr, ok := x.(*VarRef); ok && vr.Sym != nil && isPlainVoidType(vr.Sym.T) {
			s.report(InvalidTypeSpec(node.SourceStart, "cannot take address of void object"))
			return &UnOp{Op: UnAddr, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
		}
		return &UnOp{Op: UnAddr, X: x, T: s.Types.Pointer(x.GetType()), Category: RValue, Range: node.SourceRange}
	case entity.ASTERISK:
		x = s.castArrayDecay(s.castLValueToRValue(x))
		pt, ok := unqual(x.GetType()).(*PointerType)
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "cannot dereference non-pointer"))
			return &UnOp{Op: UnDeref, X: x, T: ErrorTypeSingleton, Range: node.SourceRange}
		}
		return &UnOp{Op: UnDeref, X: x, T: pt.Pointee, Category: LValue, Range: node.SourceRange}
	case entity.PLUS:
		x = s.castIntegerPromotion(s.castLValueToRValue(x))
		return &UnOp{Op: UnPlus, X: x, T: x.GetType(), Category: RValue, Range: node.SourceRange}
	case entity.MINUS:
		x = s.castIntegerPromotion(s.castLValueToRValue(x))
		return &UnOp{Op: UnMinus, X: x, T: x.GetType(), Category: RValue, Range: node.SourceRange}
	case entity.TILDE:
		x = s.castIntegerPromotion(s.castLValueToRValue(x))
		return &UnOp{Op: UnBitNot, X: x, T: x.GetType(), Category: RValue, Range: node.SourceRange}
	case entity.EXCLAMATION:
		x = s.castBoolConversion(s.castLValueToRValue(x))
		return &UnOp{Op: UnLogNot, X: x, T: s.Types.Builtin(Int), Category: RValue, Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func isPlainVoidType(t Type) bool {
	if _, qualified := t.(*QualType); qualified {
		return false
	}
	bt, ok := unqual(t).(*BuiltinType)
	return ok && bt.Kind == Void
}

func (s *Sema) typeAssignment(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.AssignmentExpression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.AssignmentExpression, 2):
		l := s.typeExpr(node.Children[0], scope)
		r := s.typeExpr(node.Children[2], scope)
		if l.GetCategory() != LValue {
			s.report(InvalidTypeSpec(node.SourceStart, "left operand of assignment must be lvalue"))
		}
		opTyp := node.Children[1].Children[0].Typ
		r = s.castFunctionDecay(s.castArrayDecay(s.castLValueToRValue(r)))
		if opTyp == entity.EQUAL {
			r = s.assignmentConversion(r, l.GetType(), node.SourceStart)
			return &AssignExpr{L: l, R: r, T: l.GetType(), Range: node.SourceRange}
		}
		op := s.compoundAssignOp(opTyp)
		if (op == OpAdd || op == OpSub) && isPointer(l.GetType()) && isArithmetic(r.GetType()) {
			s.validatePointerArithmeticOperand(l.GetType(), node.SourceStart)
		}
		return &CompoundAssign{Op: op, L: l, R: r, T: l.GetType(), Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) compoundAssignOp(typ entity.TokenType) BinaryOp {
	switch typ {
	case entity.PLUS_EQUAL:
		return OpAdd
	case entity.MINUS_EQUAL:
		return OpSub
	case entity.MULTIPLY_EQUAL:
		return OpMul
	case entity.DIVIDE_EQUAL:
		return OpDiv
	case entity.MOD_EQUAL:
		return OpMod
	case entity.LEFT_SHIFT_EQUAL:
		return OpShl
	case entity.RIGHT_SHIFT_EQUAL:
		return OpShr
	case entity.AND_EQUAL:
		return OpAnd
	case entity.XOR_EQUAL:
		return OpXor
	case entity.OR_EQUAL:
		return OpOr
	}
	return OpAdd
}

func (s *Sema) typeConditional(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.ConditionalExpression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.ConditionalExpression, 2):
		cond := s.castBoolConversion(s.castLValueToRValue(s.typeExpr(node.Children[0], scope)))
		then := s.castLValueToRValue(s.typeExpr(node.Children[2], scope))
		els := s.castLValueToRValue(s.typeExpr(node.Children[4], scope))
		then, els, common := s.balanceConditionalOperands(then, els, node.SourceStart)
		return &CondExpr{Cond: cond, Then: then, Else: els, T: common, Range: node.SourceRange}
	case node.ReducedBy(parser.ConditionalExpression, 3):
		if !s.Options.GNUExtensions || s.Options.PedanticErrors {
			s.report(InvalidTypeSpec(node.SourceStart, "omitted middle operand requires GNU C mode"))
		}
		then := s.castLValueToRValue(s.typeExpr(node.Children[0], scope))
		cond := s.castBoolConversion(then)
		els := s.castLValueToRValue(s.typeExpr(node.Children[3], scope))
		then, els, common := s.balanceConditionalOperands(then, els, node.SourceStart)
		return &CondExpr{Cond: cond, Then: then, Else: els, T: common, Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) balanceConditionalOperands(then, els Expr, pos entity.SourcePos) (Expr, Expr, Type) {
	var common Type
	if isArithmetic(then.GetType()) && isArithmetic(els.GetType()) {
		then, els, common = s.castUsualArithmetic(then, els)
	} else if isPointer(then.GetType()) && s.isConditionalNullPointerConstant(els) {
		els = s.castNullPointerConstant(els, then.GetType())
		common = then.GetType()
	} else if isPointer(els.GetType()) && s.isConditionalNullPointerConstant(then) {
		then = s.castNullPointerConstant(then, els.GetType())
		common = els.GetType()
	} else if isPointer(then.GetType()) && isPointer(els.GetType()) {
		then, els, common = s.balanceConditionalPointer(then, els, pos)
	} else {
		common = then.GetType()
		if common != els.GetType() {
			els = s.assignmentConversion(els, common, pos)
		}
	}
	return then, els, common
}

func (s *Sema) isConditionalNullPointerConstant(e Expr) bool {
	return s.isNullPointerConstant(e) || s.isVoidPointerZero(e)
}

func (s *Sema) balanceConditionalPointer(then, els Expr, pos entity.SourcePos) (Expr, Expr, Type) {
	tp := unqual(then.GetType()).(*PointerType)
	ep := unqual(els.GetType()).(*PointerType)
	commonArrayPointee := compositeConditionalArrayPointee(tp.Pointee, ep.Pointee)
	switch {
	case !isFunctionPointer(tp) && !isFunctionPointer(ep) && isVoidPointer(tp):
		common := s.qualifiedVoidPointer(tp.Pointee, ep.Pointee)
		if then.GetType() != common {
			then = s.castVoidPointerConversion(then, common)
		}
		if els.GetType() != common {
			els = s.castVoidPointerConversion(els, common)
		}
		return then, els, common
	case !isFunctionPointer(tp) && !isFunctionPointer(ep) && isVoidPointer(ep):
		common := s.qualifiedVoidPointer(tp.Pointee, ep.Pointee)
		if then.GetType() != common {
			then = s.castVoidPointerConversion(then, common)
		}
		if els.GetType() != common {
			els = s.castVoidPointerConversion(els, common)
		}
		return then, els, common
	case commonArrayPointee != nil:
		common := s.Types.Pointer(commonArrayPointee)
		if then.GetType() != common {
			then = s.castPointerConversion(then, common)
		}
		if els.GetType() != common {
			els = s.castPointerConversion(els, common)
		}
		return then, els, common
	case pointerAssignmentCompatible(ep, tp):
		if els.GetType() != then.GetType() {
			els = s.castPointerConversion(els, then.GetType())
		}
		return then, els, then.GetType()
	case pointerAssignmentCompatible(tp, ep):
		if then.GetType() != els.GetType() {
			then = s.castPointerConversion(then, els.GetType())
		}
		return then, els, els.GetType()
	case compatibleTypeIgnoringTopLevelQualifiers(tp.Pointee, ep.Pointee):
		common := s.Types.Pointer(mergePointeeQualifiers(s, tp.Pointee, ep.Pointee))
		if then.GetType() != common {
			then = s.castPointerConversion(then, common)
		}
		if els.GetType() != common {
			els = s.castPointerConversion(els, common)
		}
		return then, els, common
	default:
		s.report(InvalidTypeSpec(pos, "incompatible pointer types in conditional expression"))
		return then, els, then.GetType()
	}
}

func mergePointeeQualifiers(s *Sema, a, b Type) Type {
	c, v, r := qualifierUnion(a, b)
	base := unqual(a)
	if c || v || r {
		return s.Types.Qualified(base, c, v, r)
	}
	return base
}

func compositeConditionalArrayPointee(a, b Type) Type {
	aa, aok := unqual(a).(*ArrayType)
	ba, bok := unqual(b).(*ArrayType)
	if !aok || !bok || !compatibleTypeIgnoringTopLevelQualifiers(aa.Elem, ba.Elem) {
		return nil
	}
	switch {
	case aa.SizeKind == ArrayConstantSize && ba.SizeKind == ArrayConstantSize:
		if aa.Size == ba.Size {
			return aa
		}
		return nil
	case aa.SizeKind == ArrayConstantSize:
		return aa
	case ba.SizeKind == ArrayConstantSize:
		return ba
	case aa.SizeKind == ArrayVLA:
		return aa
	case ba.SizeKind == ArrayVLA:
		return ba
	default:
		return aa
	}
}

func (s *Sema) qualifiedVoidPointer(a, b Type) Type {
	c, v, r := qualifierUnion(a, b)
	var voidT Type = s.Types.Builtin(Void)
	if c || v || r {
		voidT = s.Types.Qualified(voidT, c, v, r)
	}
	return s.Types.Pointer(voidT)
}

func qualifierUnion(a, b Type) (bool, bool, bool) {
	aq, _ := a.(*QualType)
	bq, _ := b.(*QualType)
	var c, v, r bool
	if aq != nil {
		c = c || aq.Const
		v = v || aq.Volatile
		r = r || aq.Restrict
	}
	if bq != nil {
		c = c || bq.Const
		v = v || bq.Volatile
		r = r || bq.Restrict
	}
	return c, v, r
}

func (s *Sema) typeComma(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.Expression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.Expression, 2):
		l := s.castArrayDecay(s.castLValueToRValue(s.typeExpr(node.Children[0], scope)))
		r := s.castArrayDecay(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		return &CommaExpr{L: l, R: r, T: r.GetType(), Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) typeCast(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.CastExpression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.CastExpression, 2):
		t := s.parseTypeName(node.Children[1])
		x := s.castFunctionDecay(s.castArrayDecay(s.castLValueToRValue(s.typeExpr(node.Children[3], scope))))
		if !s.castAllowed(x.GetType(), t) {
			s.report(InvalidTypeSpec(node.SourceStart, "invalid explicit cast"))
		}
		return &ExplicitCast{To: t, X: x, Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func (s *Sema) typePostfix(node *entity.AstNode, scope *Scope) Expr {
	switch {
	case node.ReducedBy(parser.PostfixExpression, 1):
		return s.typeExpr(node.Children[0], scope)
	case node.ReducedBy(parser.PostfixExpression, 2):
		rawBase := s.typeExpr(node.Children[0], scope)
		category := indexExprCategory(rawBase)
		base := s.castArrayDecay(s.castLValueToRValue(rawBase))
		idx := s.castIntegerPromotion(s.castLValueToRValue(s.typeExpr(node.Children[2], scope)))
		pt, ok := unqual(base.GetType()).(*PointerType)
		if !ok {
			if idxPtr, idxOK := unqual(idx.GetType()).(*PointerType); idxOK && isInteger(base.GetType()) {
				base, idx = idx, base
				pt, ok = idxPtr, true
			}
		}
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "subscript on non-pointer/non-array"))
			return &IndexExpr{Base: base, Index: idx, T: ErrorTypeSingleton, Category: category, Range: node.SourceRange}
		}
		return &IndexExpr{Base: base, Index: idx, T: pt.Pointee, Category: category, Range: node.SourceRange}
	case node.ReducedBy(parser.PostfixExpression, 3):
		return s.typeCall(node, scope, nil)
	case node.ReducedBy(parser.PostfixExpression, 4):
		return s.typeCall(node, scope, node.Children[2])
	case node.ReducedBy(parser.PostfixExpression, 5):
		return s.typeMember(node, scope, false)
	case node.ReducedBy(parser.PostfixExpression, 6):
		return s.typeMember(node, scope, true)
	case node.ReducedBy(parser.PostfixExpression, 7):
		x := s.typeExpr(node.Children[0], scope)
		return s.buildIncDec(node, x, UnIncPost)
	case node.ReducedBy(parser.PostfixExpression, 8):
		x := s.typeExpr(node.Children[0], scope)
		return s.buildIncDec(node, x, UnDecPost)
	case node.ReducedBy(parser.PostfixExpression, 9), node.ReducedBy(parser.PostfixExpression, 10):
		t := s.parseTypeName(node.Children[1])
		return &CompoundLit{T: t, Init: s.typeInitListForType(node.Children[4], t), Range: node.SourceRange}
	case node.ReducedBy(parser.PostfixExpression, 11):
		s.reportEmptyInitializerExtension(node.SourceStart)
		t := s.parseTypeName(node.Children[1])
		return &CompoundLit{T: t, Init: &InitList{T: t, Range: node.SourceRange}, Range: node.SourceRange}
	}
	return s.errorExpr(node.SourceRange)
}

func indexExprCategory(base Expr) ValueCategory {
	if base == nil {
		return LValue
	}
	if _, ok := unqual(base.GetType()).(*ArrayType); ok && base.GetCategory() != LValue {
		return RValue
	}
	return LValue
}

func (s *Sema) typeCall(node *entity.AstNode, scope *Scope, argList *entity.AstNode) Expr {
	callee := s.castFunctionDecay(s.castLValueToRValue(s.typeExpr(node.Children[0], scope)))
	pt, ok := unqual(callee.GetType()).(*PointerType)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "called object is not a function"))
		return &CallExpr{Callee: callee, T: ErrorTypeSingleton, Range: node.SourceRange}
	}
	ft, ok := unqual(pt.Pointee).(*FunctionType)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "called object is not a function"))
		return &CallExpr{Callee: callee, T: ErrorTypeSingleton, Range: node.SourceRange}
	}
	args := s.collectCallArgs(argList, scope)
	for i, arg := range args {
		arg = s.castFunctionDecay(s.castArrayDecay(s.castLValueToRValue(arg)))
		if ft.HasProto && i < len(ft.Params) {
			arg = s.assignmentConversion(arg, ft.Params[i], node.SourceStart)
		} else {
			arg = s.defaultArgPromotion(arg)
		}
		args[i] = arg
	}
	if ft.HasProto && !ft.Variadic && len(args) != len(ft.Params) {
		s.report(InvalidTypeSpec(node.SourceStart, "wrong number of arguments"))
	}
	s.validateCallReturnType(ft.Ret, node.SourceStart)
	return &CallExpr{Callee: callee, Args: args, T: ft.Ret, Range: node.SourceRange}
}

func (s *Sema) validateCallReturnType(t Type, pos entity.SourcePos) {
	if bt, ok := unqual(t).(*BuiltinType); ok && bt.Kind == Void {
		if _, qualified := t.(*QualType); qualified {
			s.report(InvalidTypeSpec(pos, "function call returns qualified void"))
		}
		return
	}
	if !isObjectType(t) {
		s.report(InvalidTypeSpec(pos, "function call returns incomplete type"))
	}
}

func (s *Sema) collectCallArgs(node *entity.AstNode, scope *Scope) []Expr {
	if node == nil {
		return nil
	}
	switch {
	case node.ReducedBy(parser.ArgumentExpressionList, 1):
		return []Expr{s.typeExpr(node.Children[0], scope)}
	case node.ReducedBy(parser.ArgumentExpressionList, 2):
		out := s.collectCallArgs(node.Children[0], scope)
		return append(out, s.typeExpr(node.Children[2], scope))
	}
	return nil
}

func (s *Sema) defaultArgPromotion(e Expr) Expr {
	bt, ok := unqualifiedBuiltin(e.GetType())
	if !ok {
		return e
	}
	switch bt.Kind {
	case Bool, Char, SChar, UChar, Short, UShort:
		return &ImplicitCast{From: e.GetType(), To: s.Types.Builtin(Int), X: e, Kind: DefaultArgPromotion, Range: e.Pos()}
	case Float:
		return &ImplicitCast{From: e.GetType(), To: s.Types.Builtin(Double), X: e, Kind: DefaultArgPromotion, Range: e.Pos()}
	}
	return e
}

func (s *Sema) typeMember(node *entity.AstNode, scope *Scope, arrow bool) Expr {
	base := s.typeExpr(node.Children[0], scope)
	name := node.Children[2].Terminal.Lexeme
	var fields []*Field
	var baseQual *QualType
	category := base.GetCategory()
	if arrow {
		base = s.castArrayDecay(s.castLValueToRValue(base))
		pt, ok := unqual(base.GetType()).(*PointerType)
		if !ok {
			s.report(InvalidTypeSpec(node.SourceStart, "operand of -> must be pointer"))
			return &MemberExpr{Base: base, Arrow: true, T: ErrorTypeSingleton, Range: node.SourceRange}
		}
		baseQual, _ = pt.Pointee.(*QualType)
		switch st := unqual(pt.Pointee).(type) {
		case *StructType:
			fields = st.Fields
		case *UnionType:
			fields = st.Fields
		}
		category = LValue
	} else {
		baseQual, _ = base.GetType().(*QualType)
		switch st := unqual(base.GetType()).(type) {
		case *StructType:
			fields = st.Fields
		case *UnionType:
			fields = st.Fields
		}
	}
	for _, f := range fields {
		if f.Name == name {
			return &MemberExpr{Base: base, Field: f, Arrow: arrow, T: qualifyMemberType(f.T, baseQual), Category: category, Range: node.SourceRange}
		}
	}
	s.report(UndeclaredIdentifier(node.SourceStart, name))
	return &MemberExpr{Base: base, Arrow: arrow, T: ErrorTypeSingleton, Range: node.SourceRange}
}

func qualifyMemberType(t Type, q *QualType) Type {
	if q == nil || (!q.Const && !q.Volatile && !q.Restrict) {
		return t
	}
	if at, ok := unqual(t).(*ArrayType); ok {
		return &ArrayType{
			Elem:          qualifyMemberType(at.Elem, q),
			Size:          at.Size,
			SizeExpr:      at.SizeExpr,
			SizeKind:      at.SizeKind,
			VMFromTypedef: at.VMFromTypedef,
			ParamConst:    at.ParamConst,
			ParamVolatile: at.ParamVolatile,
			ParamRestrict: at.ParamRestrict,
		}
	}
	base := t
	constQ, volatileQ, restrictQ := q.Const, q.Volatile, q.Restrict
	if tq, ok := t.(*QualType); ok {
		base = tq.Base
		constQ = constQ || tq.Const
		volatileQ = volatileQ || tq.Volatile
		restrictQ = restrictQ || tq.Restrict
	}
	return &QualType{Base: base, Const: constQ, Volatile: volatileQ, Restrict: restrictQ}
}
