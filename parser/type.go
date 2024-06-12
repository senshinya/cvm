package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
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

func parseTypeSpecifiersAndQualifiers(specifiers, qualifiers []*AstNode) syntax.Type {
	typ := syntax.Type{}
	if len(specifiers) == 0 {
		panic("need type specifiers")
	}
	parseTypeQualifiers(qualifiers, &typ.TypeQualifiers)
	parseTypeSpecifiers(specifiers, &typ)
	return typ
}

func parseTypeSpecifiers(specifiers []*AstNode, typ *syntax.Type) {
	var numRec *numSpecifierRecorder
	for _, specifier := range specifiers {
		n := specifier.Children[0]
		switch n.Typ {
		case common.VOID:
			if typ.MetaType != syntax.MetaTypeUnknown {
				panic("conflict type declaration")
			}
			typ.MetaType = syntax.MetaTypeVoid
		case common.CHAR, common.SHORT, common.INT, common.LONG, common.FLOAT,
			common.DOUBLE, common.SIGNED, common.UNSIGNED, common.BOOL:
			// count keywords
			if typ.MetaType != syntax.MetaTypeUnknown &&
				typ.MetaType != syntax.MetaTypeNumber {
				panic("conflict type declaration")
			}
			if typ.MetaType == syntax.MetaTypeUnknown {
				typ.MetaType = syntax.MetaTypeNumber
				numRec = &numSpecifierRecorder{}
			}
			countNumberTypeSpecifiers(n.Typ, numRec)
		case common.COMPLEX:
			// support complex?
		case struct_or_union_specifier:
			if typ.MetaType != syntax.MetaTypeUnknown {
				panic("conflict type declaration")
			}
			parseStructOrUnion(n, typ)
		case enum_specifier:
			// TODO enum declare
		case typedef_name:
			// TODO need a symbol table!
			typ.MetaType = syntax.MetaTypeUserDefined
		}
	}
	switch typ.MetaType {
	case syntax.MetaTypeVoid:
	case syntax.MetaTypeNumber:
		typ.NumberMetaInfo = parseNumberRec(numRec)
	case syntax.MetaTypeStruct:
		// TODO struct declare
	case syntax.MetaTypeUnion:
		// TODO union declare
	case syntax.MetaTypeEnum:
		// TODO enum declare
	default:

	}
}

func parseNumberRec(numRec *numSpecifierRecorder) *syntax.NumberMetaInfo {
	res := &syntax.NumberMetaInfo{}

	// base type specifier
	res.BaseNumType = syntax.BaseNumTypeInt
	if numRec.char+numRec.int_+numRec.float+numRec.double+numRec.bool_ > 1 {
		panic("invalid number type combination")
	}
	switch {
	case numRec.char == 1:
		res.BaseNumType = syntax.BaseNumTypeChar
	case numRec.int_ == 1:
		res.BaseNumType = syntax.BaseNumTypeInt
	case numRec.float == 1:
		res.BaseNumType = syntax.BaseNumTypeFloat
	case numRec.double == 1:
		res.BaseNumType = syntax.BaseNumTypeDouble
	case numRec.bool_ == 1:
		res.BaseNumType = syntax.BaseNumTypeBool
	}

	// extend type specifier
	if numRec.short != 0 {
		if res.BaseNumType != syntax.BaseNumTypeInt {
			panic("invalid number type combination")
		}
		res.BaseNumType = syntax.BaseNumTypeShort
	}
	if numRec.long != 0 {
		switch {
		case numRec.long == 1:
			if res.BaseNumType != syntax.BaseNumTypeInt &&
				res.BaseNumType != syntax.BaseNumTypeDouble {
				panic("invalid number type combination")
			}
			if res.BaseNumType == syntax.BaseNumTypeInt {
				res.BaseNumType = syntax.BaseNumTypeLong
			}
			if res.BaseNumType == syntax.BaseNumTypeDouble {
				res.BaseNumType = syntax.BaseNumTypeLongDouble
			}
		case numRec.long == 2:
			if res.BaseNumType != syntax.BaseNumTypeInt {
				panic("invalid number type combination")
			}
			res.BaseNumType = syntax.BaseNumTypeLongLong
		default:
			panic("invalid number type combination")
		}
	}

	// signed or unsigned
	if numRec.signed+numRec.unsigned > 1 {
		panic("invalid number type combination")
	}
	if numRec.signed+numRec.unsigned == 1 {
		if res.BaseNumType == syntax.BaseNumTypeFloat ||
			res.BaseNumType == syntax.BaseNumTypeDouble ||
			res.BaseNumType == syntax.BaseNumTypeBool ||
			res.BaseNumType == syntax.BaseNumTypeLongDouble {
			panic("invalid number type combination")
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
		case syntax.BaseNumTypeChar:
			// default char is unsigned char
			res.Unsigned = true
		case syntax.BaseNumTypeShort, syntax.BaseNumTypeInt,
			syntax.BaseNumTypeLong, syntax.BaseNumTypeLongLong:
			res.Signed = true
		}
	}

	return res
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

func parseTypeQualifiers(qualifiers []*AstNode, typ *syntax.TypeQualifiers) {
	for _, qualifier := range qualifiers {
		n := qualifier.Children[0]
		switch n.Typ {
		case common.CONST:
			typ.Const = true
		case common.RESTRICT:
			typ.Restrict = true
		case common.VOLATILE:
			typ.Volatile = true
		}
	}
}

func parseStructOrUnion(root *AstNode, typ *syntax.Type) {
	structOrUnion := root.Children[0]
	switch structOrUnion.Children[0].Typ {
	case common.STRUCT:
		// struct
		typ.MetaType = syntax.MetaTypeStruct
		typ.StructMetaInfo = parseStructUnionMeta(root)
	default:
		// union
		typ.MetaType = syntax.MetaTypeUnion
		typ.UnionMetaInfo = parseStructUnionMeta(root)
	}
}

func parseStructUnionMeta(root *AstNode) *syntax.StructUnionMetaInfo {
	meta := &syntax.StructUnionMetaInfo{}

	prod := productions[root.ProdIndex]
	switch {
	case len(prod.Right) == 2:
		// struct_or_union_specifier := struct_or_union IDENTIFIER
		meta.Identifier = root.Children[1].Terminal.Lexeme
		meta.Incomplete = true
	case len(prod.Right) == 4:
		// struct_or_union_specifier := struct_or_union LEFT_BRACES struct_declaration_list RIGHT_BRACES
		meta.FieldMetaInfo = parseStructDeclarationList(root.Children[2])
	case len(prod.Right) == 5:
		// struct_or_union_specifier := struct_or_union IDENTIFIER LEFT_BRACES struct_declaration_list RIGHT_BRACES
		meta.Identifier = root.Children[1].Terminal.Lexeme
		meta.FieldMetaInfo = parseStructDeclarationList(root.Children[3])
	}

	return meta
}

func parseStructDeclarationList(root *AstNode) []*syntax.FieldMetaInfo {
	structDeclarations := flattenStructDeclarationList(root)

	var res []*syntax.FieldMetaInfo
	for _, structDeclaration := range structDeclarations {
		// struct_declaration := specifier_qualifier_list struct_declarator_list SEMICOLON
		specifiersQualifiers := flattenSpecifiersQualifiers(structDeclaration.Children[0])
		midType := parseTypeSpecifiersAndQualifiers(
			funk.Filter(specifiersQualifiers, func(specifier *AstNode) bool {
				return specifier.Typ == type_specifier
			}).([]*AstNode),
			funk.Filter(specifiersQualifiers, func(specifier *AstNode) bool {
				return specifier.Typ == type_qualifier
			}).([]*AstNode),
		)

		structDeclarators := flattenStructDeclaratorList(root.Children[1])
		for _, structDeclarator := range structDeclarators {
			prod := productions[structDeclarator.ProdIndex]
			switch len(prod.Right) {
			case 1:
				// struct_declarator := declarator
				declare, err := parseDeclarator(structDeclarator.Children[0], midType)
				if err != nil {
					panic(err)
				}
				res = append(res, &syntax.FieldMetaInfo{
					Type:       declare.Type,
					Identifier: &declare.Identifier,
				})
			case 2:
				// struct_declarator := COLON constant_expression
				res = append(res, &syntax.FieldMetaInfo{
					Type:     midType,
					BitWidth: ParseExpressionNode(structDeclarator.Children[1]),
				})
			case 3:
				// struct_declarator := declarator COLON constant_expression
				declare, err := parseDeclarator(structDeclarator.Children[0], midType)
				if err != nil {
					panic(err)
				}
				res = append(res, &syntax.FieldMetaInfo{
					Type:       declare.Type,
					Identifier: &declare.Identifier,
					BitWidth:   ParseExpressionNode(structDeclarator.Children[2]),
				})
			}
		}
	}

	return res
}

func flattenStructDeclaratorList(root *AstNode) []*AstNode {
	if len(root.Children) == 1 {
		return []*AstNode{root.Children[0]}
	}

	return append(flattenStructDeclaratorList(root.Children[0]), root.Children[2])
}

func flattenStructDeclarationList(root *AstNode) []*AstNode {
	if len(root.Children) == 1 {
		return []*AstNode{root.Children[0]}
	}

	return append(flattenStructDeclarationList(root.Children[0]), root.Children[1])
}

func parseUnion(root *AstNode, typ *syntax.Type) {
	typ.MetaType = syntax.MetaTypeUnion
}
