package parser

import (
	"errors"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

type numSpecifierRecorder struct {
	signed   int
	unsigned int
	char     int
	short    int
	int_     int
	long     int
	float    int
	double   int
	bool_    int
}

func parseTypeSpecifiersAndQualifiers(specifiers, qualifiers []*entity.RawAstNode) (typ entity.Type, err error) {
	if len(specifiers) == 0 {
		panic("need type specifiers")
	}
	parseTypeQualifiers(qualifiers, &typ.TypeQualifiers)
	err = parseTypeSpecifiers(specifiers, &typ)
	return
}

func parseTypeSpecifiers(specifiers []*entity.RawAstNode, typ *entity.Type) (err error) {
	var numRec *numSpecifierRecorder
	for _, specifier := range specifiers {
		n := specifier.Children[0]
		switch {
		case specifier.ReducedBy(glr.TypeSpecifier, 1):
			// type_specifier := VOID
			if typ.MetaType != entity.MetaTypeUnknown {
				err = errors.New("conflict type declaration")
				return
			}
			typ.MetaType = entity.MetaTypeVoid
		case specifier.ReducedBy(glr.TypeSpecifier, 2, 3, 4, 5, 6, 7, 8, 9, 10):
			if typ.MetaType != entity.MetaTypeUnknown &&
				typ.MetaType != entity.MetaTypeNumber {
				err = errors.New("conflict type declaration")
				return
			}
			if typ.MetaType == entity.MetaTypeUnknown {
				typ.MetaType = entity.MetaTypeNumber
				numRec = &numSpecifierRecorder{}
			}
			countNumberTypeSpecifiers(n.Typ, numRec)
		case specifier.ReducedBy(glr.TypeSpecifier, 11):
			// type_specifier := COMPLEX
			// support complex?
		case specifier.ReducedBy(glr.TypeSpecifier, 12):
			// type_specifier := struct_or_union_specifier
			if typ.MetaType != entity.MetaTypeUnknown {
				err = errors.New("conflict type declaration")
				return
			}
			err = parseStructOrUnion(n, typ)
			if err != nil {
				return
			}
		case specifier.ReducedBy(glr.TypeSpecifier, 13):
			// type_specifier := enum_specifier
			if typ.MetaType != entity.MetaTypeUnknown {
				err = errors.New("conflict type declaration")
				return
			}
			typ.MetaType = entity.MetaTypeEnum
			typ.EnumMetaInfo, err = parseEnumSpecifier(n.Children[0])
			if err != nil {
				return
			}
		case specifier.ReducedBy(glr.TypeSpecifier, 14):
			// type_specifier := typedef_name
			if typ.MetaType != entity.MetaTypeUnknown {
				err = errors.New("conflict type declaration")
				return
			}
			typ.MetaType = entity.MetaTypeUserDefined
			typ.UserDefinedTypeName = &specifier.Children[0].Children[0].Terminal.Lexeme
		default:
			panic("unreachable")
		}
	}
	if typ.MetaType == entity.MetaTypeNumber {
		typ.NumberMetaInfo, err = parseNumberRec(numRec)
		if err != nil {
			return
		}
	}
	return
}

func parseNumberRec(numRec *numSpecifierRecorder) (*entity.NumberMetaInfo, error) {
	res := &entity.NumberMetaInfo{}

	// base type specifier
	res.BaseNumType = entity.BaseNumTypeInt
	if numRec.char+numRec.int_+numRec.float+numRec.double+numRec.bool_ > 1 {
		return nil, errors.New("invalid number type combination")
	}
	switch {
	case numRec.char == 1:
		res.BaseNumType = entity.BaseNumTypeChar
	case numRec.int_ == 1:
		res.BaseNumType = entity.BaseNumTypeInt
	case numRec.float == 1:
		res.BaseNumType = entity.BaseNumTypeFloat
	case numRec.double == 1:
		res.BaseNumType = entity.BaseNumTypeDouble
	case numRec.bool_ == 1:
		res.BaseNumType = entity.BaseNumTypeBool
	}

	// extend type specifier
	if numRec.short != 0 {
		if res.BaseNumType != entity.BaseNumTypeInt {
			return nil, errors.New("invalid number type combination")
		}
		res.BaseNumType = entity.BaseNumTypeShort
	}
	if numRec.long != 0 {
		switch {
		case numRec.long == 1:
			if res.BaseNumType != entity.BaseNumTypeInt &&
				res.BaseNumType != entity.BaseNumTypeDouble {
				return nil, errors.New("invalid number type combination")
			}
			if res.BaseNumType == entity.BaseNumTypeInt {
				res.BaseNumType = entity.BaseNumTypeLong
			}
			if res.BaseNumType == entity.BaseNumTypeDouble {
				res.BaseNumType = entity.BaseNumTypeLongDouble
			}
		case numRec.long == 2:
			if res.BaseNumType != entity.BaseNumTypeInt {
				return nil, errors.New("invalid number type combination")
			}
			res.BaseNumType = entity.BaseNumTypeLongLong
		default:
			return nil, errors.New("invalid number type combination")
		}
	}

	// signed or unsigned
	if numRec.signed+numRec.unsigned > 1 {
		return nil, errors.New("invalid number type combination")
	}
	if numRec.signed+numRec.unsigned == 1 {
		if res.BaseNumType == entity.BaseNumTypeFloat ||
			res.BaseNumType == entity.BaseNumTypeDouble ||
			res.BaseNumType == entity.BaseNumTypeBool ||
			res.BaseNumType == entity.BaseNumTypeLongDouble {
			return nil, errors.New("invalid number type combination")
		}
		if numRec.signed == 1 {
			res.Signed = true
		} else {
			res.Unsigned = true
		}
	}
	// handle default
	if numRec.signed+numRec.unsigned == 0 {
		switch res.BaseNumType {
		case entity.BaseNumTypeChar:
			// default char is unsigned char
			res.Unsigned = true
		case entity.BaseNumTypeShort, entity.BaseNumTypeInt,
			entity.BaseNumTypeLong, entity.BaseNumTypeLongLong:
			res.Signed = true
		}
	}

	return res, nil
}

func countNumberTypeSpecifiers(typ common.TokenType, numRec *numSpecifierRecorder) {
	switch typ {
	case common.SIGNED:
		numRec.signed++
	case common.UNSIGNED:
		numRec.unsigned++
	case common.CHAR:
		numRec.char++
	case common.SHORT:
		numRec.short++
	case common.INT:
		numRec.int_++
	case common.LONG:
		numRec.long++
	case common.FLOAT:
		numRec.float++
	case common.DOUBLE:
		numRec.double++
	case common.BOOL:
		numRec.bool_++
	}
}

func parseTypeQualifiers(qualifiers []*entity.RawAstNode, typ *entity.TypeQualifiers) {
	for _, qualifier := range qualifiers {
		switch {
		case qualifier.ReducedBy(glr.TypeQualifier, 1):
			// type_qualifier := CONST
			typ.Const = true
		case qualifier.ReducedBy(glr.TypeQualifier, 2):
			// type_qualifier := RESTRICT
			typ.Restrict = true
		case qualifier.ReducedBy(glr.TypeQualifier, 3):
			// type_qualifier := VOLATILE
			typ.Volatile = true
		default:
			panic("unreachable")
		}
	}
}

func parseStructOrUnion(root *entity.RawAstNode, typ *entity.Type) error {
	if err := root.AssertNonTerminal(glr.StructOrUnionSpecifier); err != nil {
		panic(err)
	}

	structOrUnion := root.Children[0]
	var err error
	switch {
	case structOrUnion.ReducedBy(glr.StructOrUnion, 1):
		// struct_or_union := STRUCT
		typ.MetaType = entity.MetaTypeStruct
		typ.StructMetaInfo, err = parseStructUnionMeta(root)
	case structOrUnion.ReducedBy(glr.StructOrUnion, 2):
		// struct_or_union := UNION
		typ.MetaType = entity.MetaTypeUnion
		typ.UnionMetaInfo, err = parseStructUnionMeta(root)
	default:
		panic("unreachable")
	}
	return err
}

func parseStructUnionMeta(root *entity.RawAstNode) (*entity.StructUnionMetaInfo, error) {
	if err := root.AssertNonTerminal(glr.StructOrUnionSpecifier); err != nil {
		panic(err)
	}

	meta := &entity.StructUnionMetaInfo{}
	var err error
	switch {
	case root.ReducedBy(glr.StructOrUnionSpecifier, 1):
		// struct_or_union_specifier := struct_or_union LEFT_BRACES struct_declaration_list RIGHT_BRACES
		meta.FieldMetaInfo, err = parseStructDeclarationList(root.Children[2])
	case root.ReducedBy(glr.StructOrUnionSpecifier, 2):
		// struct_or_union_specifier := struct_or_union IDENTIFIER LEFT_BRACES struct_declaration_list RIGHT_BRACES
		meta.Identifier = &root.Children[1].Terminal.Lexeme
		meta.FieldMetaInfo, err = parseStructDeclarationList(root.Children[3])
	case root.ReducedBy(glr.StructOrUnionSpecifier, 3):
		// struct_or_union_specifier := struct_or_union IDENTIFIER
		meta.Identifier = &root.Children[1].Terminal.Lexeme
		meta.Incomplete = true
	default:
		panic("unreachable")
	}
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func parseStructDeclarationList(root *entity.RawAstNode) ([]*entity.FieldMetaInfo, error) {
	if err := root.AssertNonTerminal(glr.StructDeclarationList); err != nil {
		panic(err)
	}

	structDeclarations := flattenStructDeclarationList(root)

	var res []*entity.FieldMetaInfo
	for _, structDeclaration := range structDeclarations {
		// struct_declaration := specifier_qualifier_list struct_declarator_list SEMICOLON
		specifiersQualifiers := flattenSpecifiersQualifiers(structDeclaration.Children[0])
		midType, err := parseTypeSpecifiersAndQualifiers(
			funk.Filter(specifiersQualifiers, func(specifier *entity.RawAstNode) bool {
				return specifier.Typ == glr.TypeSpecifier
			}).([]*entity.RawAstNode),
			funk.Filter(specifiersQualifiers, func(specifier *entity.RawAstNode) bool {
				return specifier.Typ == glr.TypeQualifier
			}).([]*entity.RawAstNode),
		)
		if err != nil {
			return nil, err
		}

		structDeclarators := flattenStructDeclaratorList(root.Children[1])
		for _, structDeclarator := range structDeclarators {
			switch {
			case structDeclarator.ReducedBy(glr.StructDeclarator, 1):
				// struct_declarator := declarator
				declare, err := parseDeclarator(structDeclarator.Children[0], midType)
				if err != nil {
					return nil, err
				}
				res = append(res, &entity.FieldMetaInfo{
					Type:       declare.Type,
					Identifier: &declare.Identifier,
				})
			case structDeclarator.ReducedBy(glr.StructDeclarator, 2):
				// struct_declarator := COLON constant_expression
				bitWise, err := ParseExpressionNode(structDeclarator.Children[1])
				if err != nil {
					return nil, err
				}
				res = append(res, &entity.FieldMetaInfo{
					Type:     midType,
					BitWidth: bitWise,
				})
			case structDeclarator.ReducedBy(glr.StructDeclarator, 3):
				// struct_declarator := declarator COLON constant_expression
				declare, err := parseDeclarator(structDeclarator.Children[0], midType)
				if err != nil {
					return nil, err
				}
				bitWise, err := ParseExpressionNode(structDeclarator.Children[2])
				if err != nil {
					return nil, err
				}
				res = append(res, &entity.FieldMetaInfo{
					Type:       declare.Type,
					Identifier: &declare.Identifier,
					BitWidth:   bitWise,
				})
			default:
				panic("unreachable")
			}
		}
	}

	return res, nil
}

func parseEnumSpecifier(root *entity.RawAstNode) (*entity.EnumMetaInfo, error) {
	if err := root.AssertNonTerminal(glr.EnumSpecifier); err != nil {
		panic(err)
	}

	res := &entity.EnumMetaInfo{}
	var err error
	switch {
	case root.ReducedBy(glr.EnumSpecifier, 1):
		// enum_specifier := ENUM LEFT_BRACES enumerator_list RIGHT_BRACES
		res.EnumFields, err = parseEnumeratorList(root.Children[2])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.EnumSpecifier, 2):
		// enum_specifier := ENUM IDENTIFIER LEFT_BRACES enumerator_list RIGHT_BRACES
		res.Identifier = &root.Children[1].Terminal.Lexeme
		res.EnumFields, err = parseEnumeratorList(root.Children[3])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.EnumSpecifier, 3):
		// enum_specifier := ENUM LEFT_BRACES enumerator_list COMMA RIGHT_BRACES
		res.EnumFields, err = parseEnumeratorList(root.Children[2])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.EnumSpecifier, 4):
		// enum_specifier := ENUM IDENTIFIER LEFT_BRACES enumerator_list COMMA RIGHT_BRACES
		res.Identifier = &root.Children[1].Terminal.Lexeme
	case root.ReducedBy(glr.EnumSpecifier, 5):
		// enum_specifier := ENUM IDENTIFIER
		res.Incomplete = true
		res.Identifier = &root.Children[1].Terminal.Lexeme
	default:
		panic("unreachable")
	}

	return res, nil
}

func parseEnumeratorList(root *entity.RawAstNode) ([]*entity.EnumFieldMetaInfo, error) {
	if err := root.AssertNonTerminal(glr.EnumeratorList); err != nil {
		panic(err)
	}

	enumerators := flattenEnumerators(root)
	var res []*entity.EnumFieldMetaInfo
	for _, enumerator := range enumerators {
		field, err := parseEnumerator(enumerator)
		if err != nil {
			return nil, err
		}
		res = append(res, field)
	}
	return res, nil
}

func parseEnumerator(root *entity.RawAstNode) (*entity.EnumFieldMetaInfo, error) {
	if err := root.AssertNonTerminal(glr.Enumerator); err != nil {
		panic(err)
	}

	res := &entity.EnumFieldMetaInfo{}
	switch {
	case root.ReducedBy(glr.Enumerator, 1):
		// enumerator := enumeration_constant
		res.Identifier = root.Children[0].Terminal.Lexeme
	case root.ReducedBy(glr.Enumerator, 2):
		// enumerator := enumeration_constant EQUAL constant_expression
		res.Identifier = root.Children[0].Terminal.Lexeme
		var err error
		res.Value, err = ParseExpressionNode(root.Children[2])
		if err != nil {
			return nil, err
		}
	default:
		panic("unreachable")
	}

	return res, nil
}
