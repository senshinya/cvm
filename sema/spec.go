package sema

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
	"shinya.click/cvm/parser"
)

type SpecResult struct {
	Type      Type
	Storage   StorageClass
	IsTypedef bool
	IsInline  bool
}

func (s *Sema) parseSpec(node *entity.AstNode) SpecResult {
	var (
		typeSpecs []*entity.AstNode
		c, v, r   bool
		storage   StorageClass
		typedef   bool
		inline    bool
	)
	s.collectSpecParts(node, &typeSpecs, &c, &v, &r, &storage, &typedef, &inline)
	t := s.buildBaseType(typeSpecs, node.SourceStart)
	if c || v || r {
		t = s.Types.Qualified(t, c, v, r)
	}
	return SpecResult{Type: t, Storage: storage, IsTypedef: typedef, IsInline: inline}
}

func (s *Sema) collectSpecParts(node *entity.AstNode, typeSpecs *[]*entity.AstNode, c, v, r *bool, storage *StorageClass, typedef *bool, inline *bool) {
	switch node.Typ {
	case parser.DeclarationSpecifiers:
		switch {
		case node.ReducedBy(parser.DeclarationSpecifiers, 1):
			s.handleStorageClass(node.Children[0], storage, typedef)
		case node.ReducedBy(parser.DeclarationSpecifiers, 2):
			*typeSpecs = append(*typeSpecs, node.Children[0])
		case node.ReducedBy(parser.DeclarationSpecifiers, 3):
			s.handleQualifier(node.Children[0], c, v, r)
		case node.ReducedBy(parser.DeclarationSpecifiers, 4):
			*inline = true
		case node.ReducedBy(parser.DeclarationSpecifiers, 5):
			s.handleStorageClass(node.Children[0], storage, typedef)
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		case node.ReducedBy(parser.DeclarationSpecifiers, 6):
			*typeSpecs = append(*typeSpecs, node.Children[0])
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		case node.ReducedBy(parser.DeclarationSpecifiers, 7):
			s.handleQualifier(node.Children[0], c, v, r)
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		case node.ReducedBy(parser.DeclarationSpecifiers, 8):
			*inline = true
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		}
	case parser.SpecifierQualifierList:
		switch {
		case node.ReducedBy(parser.SpecifierQualifierList, 1):
			*typeSpecs = append(*typeSpecs, node.Children[0])
		case node.ReducedBy(parser.SpecifierQualifierList, 2):
			*typeSpecs = append(*typeSpecs, node.Children[0])
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		case node.ReducedBy(parser.SpecifierQualifierList, 3):
			s.handleQualifier(node.Children[0], c, v, r)
		case node.ReducedBy(parser.SpecifierQualifierList, 4):
			s.handleQualifier(node.Children[0], c, v, r)
			s.collectSpecParts(node.Children[1], typeSpecs, c, v, r, storage, typedef, inline)
		}
	}
}

func (s *Sema) handleStorageClass(node *entity.AstNode, storage *StorageClass, typedef *bool) {
	switch {
	case node.ReducedBy(parser.StorageClassSpecifier, 1):
		*storage = StorageTypedef
		*typedef = true
	case node.ReducedBy(parser.StorageClassSpecifier, 2):
		*storage = StorageExtern
	case node.ReducedBy(parser.StorageClassSpecifier, 3):
		*storage = StorageStatic
	case node.ReducedBy(parser.StorageClassSpecifier, 4):
		*storage = StorageAuto
	case node.ReducedBy(parser.StorageClassSpecifier, 5):
		*storage = StorageRegister
	}
}

func (s *Sema) handleQualifier(node *entity.AstNode, c, v, r *bool) {
	switch {
	case node.ReducedBy(parser.TypeQualifier, 1):
		*c = true
	case node.ReducedBy(parser.TypeQualifier, 2):
		*r = true
	case node.ReducedBy(parser.TypeQualifier, 3):
		*v = true
	}
}

func (s *Sema) buildBaseType(specs []*entity.AstNode, pos entity.SourcePos) Type {
	if len(specs) == 0 {
		s.report(InvalidTypeSpec(pos, "missing type specifier"))
		return ErrorTypeSingleton
	}
	for _, sp := range specs {
		switch {
		case sp.ReducedBy(parser.TypeSpecifier, 12):
			if len(specs) != 1 {
				s.report(InvalidTypeSpec(pos, "struct/union cannot combine with other specifiers"))
				return ErrorTypeSingleton
			}
			return s.buildStructUnion(sp.Children[0])
		case sp.ReducedBy(parser.TypeSpecifier, 13):
			if len(specs) != 1 {
				s.report(InvalidTypeSpec(pos, "enum cannot combine with other specifiers"))
				return ErrorTypeSingleton
			}
			return s.buildEnum(sp.Children[0])
		case sp.ReducedBy(parser.TypeSpecifier, 14):
			if len(specs) != 1 {
				s.report(InvalidTypeSpec(pos, "typedef name cannot combine with other specifiers"))
				return ErrorTypeSingleton
			}
			return s.lookupTypedef(sp.Children[0])
		}
	}
	return s.combineArithmetic(specs, pos)
}

func (s *Sema) combineArithmetic(specs []*entity.AstNode, pos entity.SourcePos) Type {
	var nVoid, nBool, nChar, nShort, nInt, nLong, nFloat, nDouble, nSigned, nUnsigned, nComplex int
	for _, sp := range specs {
		switch {
		case sp.ReducedBy(parser.TypeSpecifier, 1):
			nVoid++
		case sp.ReducedBy(parser.TypeSpecifier, 2):
			nChar++
		case sp.ReducedBy(parser.TypeSpecifier, 3):
			nShort++
		case sp.ReducedBy(parser.TypeSpecifier, 4):
			nInt++
		case sp.ReducedBy(parser.TypeSpecifier, 5):
			nLong++
		case sp.ReducedBy(parser.TypeSpecifier, 6):
			nFloat++
		case sp.ReducedBy(parser.TypeSpecifier, 7):
			nDouble++
		case sp.ReducedBy(parser.TypeSpecifier, 8):
			nSigned++
		case sp.ReducedBy(parser.TypeSpecifier, 9):
			nUnsigned++
		case sp.ReducedBy(parser.TypeSpecifier, 10):
			nBool++
		case sp.ReducedBy(parser.TypeSpecifier, 11):
			nComplex++
		}
	}
	if nSigned > 0 && nUnsigned > 0 {
		s.report(InvalidTypeSpec(pos, "both signed and unsigned"))
		return ErrorTypeSingleton
	}
	switch {
	case nVoid == 1 && nSigned+nUnsigned+nBool+nChar+nShort+nInt+nLong+nFloat+nDouble+nComplex == 0:
		return s.Types.Builtin(Void)
	case nBool == 1 && nSigned+nUnsigned+nChar+nShort+nInt+nLong+nFloat+nDouble+nComplex == 0:
		return s.Types.Builtin(Bool)
	case nChar == 1:
		if nUnsigned == 1 {
			return s.Types.Builtin(UChar)
		}
		if nSigned == 1 {
			return s.Types.Builtin(SChar)
		}
		return s.Types.Builtin(Char)
	case nShort == 1:
		if nUnsigned == 1 {
			return s.Types.Builtin(UShort)
		}
		return s.Types.Builtin(Short)
	case nLong == 2:
		if nUnsigned == 1 {
			return s.Types.Builtin(ULongLong)
		}
		return s.Types.Builtin(LongLong)
	case nLong == 1:
		if nDouble == 1 {
			if nComplex == 1 {
				return s.Types.Builtin(LongDoubleComplex)
			}
			return s.Types.Builtin(LongDouble)
		}
		if nUnsigned == 1 {
			return s.Types.Builtin(ULong)
		}
		return s.Types.Builtin(Long)
	case nFloat == 1:
		if nComplex == 1 {
			return s.Types.Builtin(FloatComplex)
		}
		return s.Types.Builtin(Float)
	case nDouble == 1:
		if nComplex == 1 {
			return s.Types.Builtin(DoubleComplex)
		}
		return s.Types.Builtin(Double)
	case nInt == 1 || nSigned+nUnsigned > 0:
		if nUnsigned == 1 {
			return s.Types.Builtin(UInt)
		}
		return s.Types.Builtin(Int)
	}
	s.report(InvalidTypeSpec(pos, "unsupported type specifier combination"))
	return ErrorTypeSingleton
}

func (s *Sema) buildStructUnion(node *entity.AstNode) Type {
	isUnion := node.Children[0].ReducedBy(parser.StructOrUnion, 2)
	switch {
	case node.ReducedBy(parser.StructOrUnionSpecifier, 1):
		t := s.newAnonStructUnion(isUnion)
		fields := s.parseStructDeclList(node.Children[2])
		s.validateFlexibleArrayMembers(fields, isUnion, node.SourceStart)
		s.completeStructUnion(t, fields)
		return t
	case node.ReducedBy(parser.StructOrUnionSpecifier, 2):
		name := node.Children[1].Terminal.Lexeme
		t := s.lookupOrCreateCurrentTag(name, isUnion, node.SourceStart)
		fields := s.parseStructDeclList(node.Children[3])
		s.validateFlexibleArrayMembers(fields, isUnion, node.SourceStart)
		s.completeStructUnion(t, fields)
		return t
	case node.ReducedBy(parser.StructOrUnionSpecifier, 3):
		return s.lookupOrCreateTag(node.Children[1].Terminal.Lexeme, isUnion, node.SourceStart)
	}
	return ErrorTypeSingleton
}

func (s *Sema) newAnonStructUnion(isUnion bool) Type {
	tag := s.Types.NewTagID()
	if isUnion {
		return s.Types.Union(tag)
	}
	return s.Types.Struct(tag)
}

func (s *Sema) lookupOrCreateTag(name string, isUnion bool, pos entity.SourcePos) Type {
	if existing := s.scope.LookupTag(name); existing != nil {
		if !tagInfoMatchesStructUnion(existing.T, isUnion) {
			s.report(InvalidTypeSpec(pos, "tag defined as wrong kind"))
			return ErrorTypeSingleton
		}
		return existing.T
	}
	return s.createTag(name, isUnion, pos)
}

func (s *Sema) lookupOrCreateCurrentTag(name string, isUnion bool, pos entity.SourcePos) Type {
	if existing := s.scope.LookupCurrentTag(name); existing != nil {
		if !tagInfoMatchesStructUnion(existing.T, isUnion) {
			s.report(InvalidTypeSpec(pos, "tag defined as wrong kind"))
			return ErrorTypeSingleton
		}
		return existing.T
	}
	return s.createTag(name, isUnion, pos)
}

func tagInfoMatchesStructUnion(t Type, isUnion bool) bool {
	switch t.(type) {
	case *StructType:
		return !isUnion
	case *UnionType:
		return isUnion
	default:
		return false
	}
}

func (s *Sema) createTag(name string, isUnion bool, pos entity.SourcePos) Type {
	tag := s.Types.NewTagID()
	var t Type
	if isUnion {
		t = s.Types.Union(tag)
	} else {
		t = s.Types.Struct(tag)
	}
	if err := s.scope.InsertTagChecked(name, &TagInfo{Tag: tag, T: t}, pos); err != nil {
		s.report(err.(*common.CvmError))
	}
	return t
}

func (s *Sema) completeStructUnion(t Type, fields []*Field) {
	var offset int64
	for _, f := range fields {
		f.Offset = offset
		if _, ok := t.(*StructType); ok {
			offset += sizeofType(f.T)
		}
	}
	switch x := t.(type) {
	case *StructType:
		s.Types.CompleteStruct(x, fields)
	case *UnionType:
		s.Types.CompleteUnion(x, fields)
	}
	for cur := s.scope; cur != nil; cur = cur.Parent {
		for _, info := range cur.Tags {
			if info.T == t {
				info.Complete = true
			}
		}
	}
}

// C99 flexible array member 只能出现在结构体最后，并且前面必须有其他具名成员。
func (s *Sema) validateFlexibleArrayMembers(fields []*Field, isUnion bool, pos entity.SourcePos) {
	hasNamedField := false
	for i, f := range fields {
		if f == nil {
			continue
		}
		if isFlexibleArrayMember(f.T) {
			if isUnion {
				s.report(InvalidTypeSpec(pos, "flexible array member cannot appear in union"))
			}
			if i != len(fields)-1 {
				s.report(InvalidTypeSpec(pos, "flexible array member must be last"))
			}
			if !hasNamedField {
				s.report(InvalidTypeSpec(pos, "flexible array member requires a previous named member"))
			}
		} else if !isUnion && typeContainsFlexibleArrayMember(f.T) {
			s.report(InvalidTypeSpec(pos, "invalid use of structure containing flexible array member"))
		} else if !isObjectType(f.T) {
			s.report(InvalidTypeSpec(pos, "field type must be complete object type"))
		}
		if f.Name != "" {
			hasNamedField = true
		}
	}
}

func (s *Sema) parseStructDeclList(node *entity.AstNode) []*Field {
	var fields []*Field
	switch {
	case node.ReducedBy(parser.StructDeclarationList, 1):
		fields = append(fields, s.parseStructDeclaration(node.Children[0])...)
	case node.ReducedBy(parser.StructDeclarationList, 2):
		fields = append(fields, s.parseStructDeclList(node.Children[0])...)
		fields = append(fields, s.parseStructDeclaration(node.Children[1])...)
	}
	return fields
}

func (s *Sema) parseStructDeclaration(node *entity.AstNode) []*Field {
	spec := s.parseSpec(node.Children[0])
	s.validateRestrictType(spec.Type, node.SourceStart)
	return s.parseStructDeclaratorList(node.Children[1], spec.Type)
}

func (s *Sema) parseStructDeclaratorList(node *entity.AstNode, base Type) []*Field {
	var fields []*Field
	switch {
	case node.ReducedBy(parser.StructDeclaratorList, 1):
		fields = append(fields, s.parseStructDeclarator(node.Children[0], base))
	case node.ReducedBy(parser.StructDeclaratorList, 2):
		fields = append(fields, s.parseStructDeclaratorList(node.Children[0], base)...)
		fields = append(fields, s.parseStructDeclarator(node.Children[2], base))
	}
	return fields
}

func (s *Sema) parseStructDeclarator(node *entity.AstNode, base Type) *Field {
	switch {
	case node.ReducedBy(parser.StructDeclarator, 1):
		s.validateDeclaratorArrayQualifiers(node.Children[0], false)
		t, name := s.applyDeclarator(node.Children[0], base)
		s.validateRestrictType(t, node.SourceStart)
		return &Field{Name: name, T: t}
	case node.ReducedBy(parser.StructDeclarator, 2):
		width := s.evalBitWidth(node.Children[1])
		s.validateBitFieldWidth(base, width, node.SourceStart)
		return &Field{T: base, BitWidth: width, IsBitField: true}
	case node.ReducedBy(parser.StructDeclarator, 3):
		s.validateDeclaratorArrayQualifiers(node.Children[0], false)
		t, name := s.applyDeclarator(node.Children[0], base)
		s.validateRestrictType(t, node.SourceStart)
		width := s.evalBitWidth(node.Children[2])
		s.validateBitFieldWidth(t, width, node.SourceStart)
		return &Field{Name: name, T: t, BitWidth: width, IsBitField: true}
	}
	return &Field{T: ErrorTypeSingleton}
}

func (s *Sema) validateBitFieldWidth(t Type, width int, pos entity.SourcePos) {
	// GCC/C99 中 _Bool 只有 0 和 1 两个值，位域宽度不能超过一个值位。
	if bt, ok := unqual(t).(*BuiltinType); ok && bt.Kind == Bool && width > 1 {
		s.report(InvalidTypeSpec(pos, "_Bool bit-field width must not exceed 1"))
	}
}

func (s *Sema) evalBitWidth(node *entity.AstNode) int {
	expr := s.typeExpr(node, s.scope)
	cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr)
	if !ok {
		s.report(InvalidTypeSpec(node.SourceStart, "bit-field width must be integer constant expression"))
		return 0
	}
	return int(cv.Int)
}

func (s *Sema) buildEnum(node *entity.AstNode) Type {
	intT := s.Types.Builtin(Int)
	switch {
	case node.ReducedBy(parser.EnumSpecifier, 5):
		name := node.Children[1].Terminal.Lexeme
		if existing := s.scope.LookupTag(name); existing != nil {
			if _, ok := existing.T.(*EnumType); !ok {
				s.report(InvalidTypeSpec(node.SourceStart, "tag defined as wrong kind"))
				return ErrorTypeSingleton
			}
			return existing.T
		}
		// GCC 的 C99 warning-only 用例会接受 enum 前向声明；当前没有 warning 通道，因此按可继续分析处理。
		tag := s.Types.NewTagID()
		et := s.Types.Enum(tag)
		_ = s.scope.InsertTagChecked(name, &TagInfo{Tag: tag, T: et}, node.SourceStart)
		return et
	case node.ReducedBy(parser.EnumSpecifier, 1), node.ReducedBy(parser.EnumSpecifier, 3):
		tag := s.Types.NewTagID()
		et := s.Types.Enum(tag)
		enums := s.parseEnumeratorList(node.Children[2], intT)
		s.Types.CompleteEnum(et, intT, enums)
		s.registerEnumerators(enums, intT, et)
		return et
	case node.ReducedBy(parser.EnumSpecifier, 2), node.ReducedBy(parser.EnumSpecifier, 4):
		name := node.Children[1].Terminal.Lexeme
		var et *EnumType
		if existing := s.scope.LookupCurrentTag(name); existing != nil {
			et, _ = existing.T.(*EnumType)
		}
		if et == nil {
			tag := s.Types.NewTagID()
			et = s.Types.Enum(tag)
			_ = s.scope.InsertTagChecked(name, &TagInfo{Tag: tag, T: et}, node.SourceStart)
		}
		enums := s.parseEnumeratorList(node.Children[3], intT)
		s.Types.CompleteEnum(et, intT, enums)
		s.registerEnumerators(enums, intT, et)
		return et
	}
	return ErrorTypeSingleton
}

func (s *Sema) parseEnumeratorList(node *entity.AstNode, base Type) []*Enumerator {
	var out []*Enumerator
	switch {
	case node.ReducedBy(parser.EnumeratorList, 1):
		out = append(out, s.parseEnumerator(node.Children[0], 0, base))
	case node.ReducedBy(parser.EnumeratorList, 2):
		out = append(out, s.parseEnumeratorList(node.Children[0], base)...)
		next := int64(0)
		if len(out) > 0 {
			next = out[len(out)-1].Value + 1
		}
		out = append(out, s.parseEnumerator(node.Children[2], next, base))
	}
	return out
}

func (s *Sema) parseEnumerator(node *entity.AstNode, defaultVal int64, base Type) *Enumerator {
	name := node.Children[0].Children[0].Terminal.Lexeme
	val := defaultVal
	if node.ReducedBy(parser.Enumerator, 2) {
		expr := s.typeExpr(node.Children[2], s.scope)
		if cv, ok := NewEvaluator(s).EvalC99IntegerConstantExpression(expr); ok {
			val = cv.Int
		} else {
			s.report(InvalidTypeSpec(node.SourceStart, "enum value must be integer constant expression"))
		}
	}
	return &Enumerator{Name: name, Value: val}
}

func (s *Sema) registerEnumerators(enums []*Enumerator, base Type, et *EnumType) {
	for _, e := range enums {
		sym := &Symbol{Name: e.Name, Kind: SymEnumerator, T: et, Pos: entity.SourcePos{}}
		if err := s.scope.InsertChecked(e.Name, sym); err != nil {
			s.report(err.(*common.CvmError))
		}
		_ = base
	}
}

func (s *Sema) lookupTypedef(node *entity.AstNode) Type {
	name := node.Children[0].Terminal.Lexeme
	sym := s.scope.Lookup(name, NSOrdinary)
	if sym == nil || sym.Kind != SymTypedef {
		s.report(UndeclaredIdentifier(node.SourceStart, name))
		return ErrorTypeSingleton
	}
	return sym.T
}
