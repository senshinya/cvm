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

	// parse specifiers
	isTypeDef, err := parseDeclarationSpecifiers(root.Children[0], res)
	if err != nil {
		return nil, err
	}
	if isTypeDef {
		return parseTypeDef(root)
	}

	if len(productions[root.ProdIndex].Right) == 2 {
		// reduced by declaration := declaration_specifiers SEMICOLON
		// this production can only declare struct, union or enum
		// otherwise "declaration does not declare anything" occurs
		// we treat it as error
		if res.MidType.MetaType != syntax.MetaTypeStruct &&
			res.MidType.MetaType != syntax.MetaTypeUnion &&
			res.MidType.MetaType != syntax.MetaTypeEnum {
			panic("declaration does not declare anything")
		}
		return res, nil
	}

	// parse init declarators
	res.Declarators, err = parseInitDeclarators(root.Children[1], res.MidType)
	if err != nil {
		return nil, err
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
	functionSpecifiers := funk.Filter(specifierNodes, func(specifier *AstNode) bool {
		return specifier.Typ == function_specifier
	}).([]*AstNode)
	if len(functionSpecifiers) != 0 {
		rec.Specifiers.Inline = true
	}

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

func parseInitDeclarators(declarators *AstNode, midType syntax.Type) ([]syntax.Declarator, error) {
	var res []syntax.Declarator
	initDeclarators := flattenInitDeclarators(declarators)
	for _, initDeclarator := range initDeclarators {
		// parse declarator
		declare, err := parseDeclarator(initDeclarator.Children[0], midType)
		if err != nil {
			return nil, err
		}
		res = append(res, declare)

		if len(productions[initDeclarator.ProdIndex].Right) != 3 {
			continue
		}

		// TODO parse initializer
	}
	return res, nil
}

func flattenInitDeclarators(declarators *AstNode) []*AstNode {
	if len(productions[declarators.ProdIndex].Right) == 1 {
		return []*AstNode{declarators.Children[0]}
	}

	return append(flattenInitDeclarators(declarators.Children[0]), declarators.Children[2])
}

func parseDeclarator(root *AstNode, midType syntax.Type) (syntax.Declarator, error) {
	res := syntax.Declarator{}

	// 1. find the most inner direct_declarator node that contains only IDENTIFIER
	currentNode := root
	for {
		if currentNode.Typ == declarator {
			if len(productions[currentNode.ProdIndex].Right) == 2 {
				// reduced by declarator := pointer direct_declarator
				currentNode = currentNode.Children[1]
				continue
			}
			currentNode = currentNode.Children[0]
			continue
		}
		// current node type is direct_declarator
		if len(productions[currentNode.ProdIndex].Right) == 1 {
			// gotcha
			break
		}
		if currentNode.Children[0].Typ == common.LEFT_PARENTHESES {
			currentNode = currentNode.Children[1]
			continue
		}
		currentNode = currentNode.Children[0]
	}
	res.Identifier = currentNode.Children[0].Terminal.Lexeme

	currentType := &res.Type
	for {
		// currentType should always be empty at the start and end of the loop
		if currentNode == root {
			break
		}
		prod := productions[currentNode.ProdIndex]
		if currentNode.Typ == declarator {
			// must be reduced by declarator := pointer direct_declarator
			// cause if be reduced by declarator := direct_declarator, should break before
			currentType = parsePointer(currentNode.Children[0], currentType).PointerInnerType
			currentNode = currentNode.Parent
			continue
		}
		// current node is direct declarator
		if len(prod.Right) == 1 {
			// reduced by direct_declarator := IDENTIFIER, do nothing
			currentNode = currentNode.Parent
			continue
		}
		if currentNode.Children[0].Typ == common.LEFT_PARENTHESES {
			// reduced by direct_declarator := LEFT_PARENTHESES declarator RIGHT_PARENTHESES, do nothing
			currentNode = currentNode.Parent
			continue
		}
		if currentNode.Children[1].Typ == common.LEFT_BRACKETS {
			// TODO parse array declarator
			currentType.MetaType = syntax.MetaTypeArray
			currentType.ArrayMetaInfo = &syntax.ArrayMetaInfo{InnerType: &syntax.Type{}}
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
			continue
		}
		if currentNode.Children[1].Typ == common.LEFT_PARENTHESES {
			// TODO parse function declarator
			currentType.MetaType = syntax.MetaTypeFunction
			currentType.FunctionMetaInfo = &syntax.FunctionMetaInfo{ReturnType: &syntax.Type{}}
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
			continue
		}
		panic("unknown current node type")
	}
	*currentType = midType
	return res, nil
}

func parsePointer(rootPointer *AstNode, currentType *syntax.Type) *syntax.Type {
	// find the most inner pointer
	currentPointer := rootPointer
	for {
		prod := productions[rootPointer.ProdIndex]
		if len(prod.Right) == 1 ||
			(len(prod.Right) == 2 && rootPointer.Children[1].Typ == type_qualifier_list) {
			// gotcha
			break
		}
		if len(prod.Right) == 2 {
			currentPointer = rootPointer.Children[1]
		} else {
			// length = 3
			currentPointer = rootPointer.Children[2]
		}
	}

	for {
		if currentPointer == rootPointer {
			break
		}
		currentType.MetaType = syntax.MetaTypePointer
		currentType.PointerInnerType = &syntax.Type{}
		prod := productions[currentPointer.ProdIndex]
		if len(prod.Right) == 1 ||
			(len(prod.Right) == 2 && currentPointer.Children[1].Typ == pointer) {
			currentType = currentType.PointerInnerType
			currentPointer = currentPointer.Parent
			continue
		}
		typeQualifiers := flattenTypeQualifierList(currentPointer.Children[1])
		parseTypeQualifiers(typeQualifiers, currentType)
		currentType = currentType.PointerInnerType
		currentPointer = currentPointer.Parent
	}
	currentType.MetaType = syntax.MetaTypePointer
	currentType.PointerInnerType = &syntax.Type{}
	if len(rootPointer.Children) == 2 {
		// parse the root qualifiers
		typeQualifiers := flattenTypeQualifierList(rootPointer.Children[1])
		parseTypeQualifiers(typeQualifiers, currentType)
	}
	return currentType
}

func flattenTypeQualifierList(listNode *AstNode) []*AstNode {
	if len(listNode.Children) == 1 {
		return []*AstNode{listNode.Children[0]}
	}
	return append(flattenTypeQualifierList(listNode.Children[0]), listNode.Children[1])
}
