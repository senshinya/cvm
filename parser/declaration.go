package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

func parseDeclaration(root *AstNode) (syntax.TranslationUnit, error) {
	res := &syntax.Declaration{
		Specifiers: syntax.Specifiers{},
		MidType:    syntax.Type{},
	}
	isTypeDef, err := parseDeclarationSpecifiers(root.Children[0], res)
	if err != nil {
		return nil, err
	}
	if isTypeDef {
		return parseTypeDef(root)
	}
	return res, nil
}

func parseDeclarationSpecifiers(specifiersNode *AstNode, rec *syntax.Declaration) (bool, error) {
	specifierNodes := flattenDeclarationSpecifier(specifiersNode)

	// parse storage class specifier
	storagesSpecifiers := funk.Filter(specifierNodes, func(specifier *AstNode) bool {
		return specifier.Typ == storage_class_specifier
	}).([]*AstNode)
	for _, storagesSpecifier := range storagesSpecifiers {
		isTypeDef := parseStorageClassSpecifier(storagesSpecifier, rec)
		if isTypeDef {
			return true, nil
		}
	}
	if err := checkStorageClassSpecifiers(rec); err != nil {
		return false, err
	}

	// parse type specifier
	typeSpecifiers := funk.Filter(specifierNodes, func(specifier *AstNode) bool {
		return specifier.Typ == type_specifier
	}).([]*AstNode)
	recorder := &numSpecifierRecorder{}
	for _, typeSpecifier := range typeSpecifiers {
		parseDeclarationTypeSpecifier(typeSpecifier, rec, recorder)
	}
	parseNumMidType(rec, recorder)

	// parse type qualifier

	// parse function specifier

	return false, nil
}

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

func checkStorageClassSpecifiers(rec *syntax.Declaration) error {
	// TODO
	return nil
}

func parseDeclarationTypeSpecifier(typeSpecifier *AstNode, rec *syntax.Declaration, numRec *numSpecifierRecorder) {
	n := typeSpecifier.Children[0]
	switch n.Typ {
	case common.VOID:
		if rec.MidType.MetaType != syntax.MetaTypeUnknown {
			panic("conflict type declaration")
		}
		rec.MidType.MetaType = syntax.MetaTypeVoid
	case common.CHAR, common.SHORT, common.INT, common.LONG, common.FLOAT,
		common.DOUBLE, common.SIGNED, common.UNSIGNED, common.BOOL:
		if rec.MidType.MetaType != 0 && rec.MidType.MetaType != syntax.MetaTypeNumber {
			panic("conflict type declaration")
		}
		if rec.MidType.MetaType == syntax.MetaTypeUnknown {
			rec.MidType.MetaType = syntax.MetaTypeNumber
			rec.MidType.NumberMetaInfo = &syntax.NumberMetaInfo{}
		}
		parseDeclarationNumber(n, numRec)
	case common.COMPLEX:
		// support complex?
	case struct_or_union_specifier:
		// TODO struct or union declare
	case enum_specifier:
		// TODO enum declare
	case common.TYPE_NAME:
		// TODO TYPE_NAME declare, need a symbol table!
	}
}

func parseDeclarationNumber(n *AstNode, numRec *numSpecifierRecorder) {
	switch n.Typ {
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

func parseNumMidType(rec *syntax.Declaration, numRec *numSpecifierRecorder) {

}

func parseStorageClassSpecifier(storageSpecifier *AstNode, rec *syntax.Declaration) bool {
	n := storageSpecifier.Children[0]
	switch n.Typ {
	case common.TYPEDEF:
		return true
	case common.EXTERN:
		rec.Specifiers.Extern = true
	case common.STATIC:
		rec.Specifiers.Static = true
	case common.AUTO:
		rec.Specifiers.Auto = true
	case common.REGISTER:
		rec.Specifiers.Register = true
	}
	return false
}

func flattenDeclarationSpecifier(specifiers *AstNode) []*AstNode {
	if len(productions[specifiers.ProdIndex].Right) == 1 {
		return []*AstNode{specifiers.Children[0]}
	}

	return append(flattenDeclarationSpecifier(specifiers.Children[1]), specifiers.Children[0])
}
