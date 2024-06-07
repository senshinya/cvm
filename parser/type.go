package parser

import (
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
		parseStruct(root, typ)
	default:
		// enum
		parseUnion(root, typ)
	}
}

func parseStruct(root *AstNode, typ *syntax.Type) {
	typ.MetaType = syntax.MetaTypeStruct
}

func parseUnion(root *AstNode, typ *syntax.Type) {
	typ.MetaType = syntax.MetaTypeUnion
}
