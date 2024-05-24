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
		isTypeDef := parseStorageClassSpecifier(storagesSpecifier, &rec.Specifiers)
		if isTypeDef {
			return true, nil
		}
	}
	if err := checkStorageClassSpecifiers(rec); err != nil {
		return false, err
	}

	// parse type specifier and qualifiers
	midType := parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_specifier
		}).([]*AstNode),
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_qualifier
		}).([]*AstNode),
	)
	rec.MidType = midType

	// parse function specifier

	return false, nil
}

func checkStorageClassSpecifiers(rec *syntax.Declaration) error {
	// TODO check storage class specifiers conflict
	return nil
}

func parseStorageClassSpecifier(storageSpecifier *AstNode, spe *syntax.Specifiers) bool {
	n := storageSpecifier.Children[0]
	switch n.Typ {
	case common.TYPEDEF:
		return true
	case common.EXTERN:
		spe.Extern = true
	case common.STATIC:
		spe.Static = true
	case common.AUTO:
		spe.Auto = true
	case common.REGISTER:
		spe.Register = true
	}
	return false
}

func flattenDeclarationSpecifier(specifiers *AstNode) []*AstNode {
	if len(productions[specifiers.ProdIndex].Right) == 1 {
		return []*AstNode{specifiers.Children[0]}
	}

	return append(flattenDeclarationSpecifier(specifiers.Children[1]), specifiers.Children[0])
}
