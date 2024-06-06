package parser

import (
	"errors"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/syntax"
)

func parseDeclaration(root *AstNode) (syntax.TranslationUnit, error) {
	res := &syntax.Declaration{}

	// parse specifiers
	specifiers, midType, err := parseDeclarationSpecifiers(root.Children[0])
	if err != nil {
		return nil, err
	}
	if specifiers.TypeDef {
		return parseTypeDef(root, specifiers, midType)
	}
	res.Specifiers = specifiers
	res.MidType = midType

	if len(productions[root.ProdIndex].Right) == 2 {
		// reduced by declaration := declaration_specifiers SEMICOLON
		// this production can only declare struct, union or enum
		// otherwise "declaration does not declare anything" occurs
		// we treat it as error
		if res.MidType.MetaType != syntax.MetaTypeStruct &&
			res.MidType.MetaType != syntax.MetaTypeUnion &&
			res.MidType.MetaType != syntax.MetaTypeEnum {
			return nil, errors.New("declaration does not declare anything")
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

func parseDeclarationSpecifiers(specifiersNode *AstNode) (syntax.Specifiers, syntax.Type, error) {
	specifiers, midType := syntax.Specifiers{}, syntax.Type{}

	specifierNodes := flattenDeclarationSpecifier(specifiersNode)

	// parse storage class specifier
	storagesSpecifiers := funk.Filter(specifierNodes, func(specifier *AstNode) bool {
		return specifier.Typ == storage_class_specifier
	}).([]*AstNode)
	for _, storagesSpecifier := range storagesSpecifiers {
		parseStorageClassSpecifier(storagesSpecifier, &specifiers)
	}
	if err := checkStorageClassSpecifiers(specifiers); err != nil {
		return specifiers, midType, err
	}

	// parse type specifier and qualifiers
	midType = parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_specifier
		}).([]*AstNode),
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_qualifier
		}).([]*AstNode),
	)

	// parse function specifier
	functionSpecifiers := funk.Filter(specifierNodes, func(specifier *AstNode) bool {
		return specifier.Typ == function_specifier
	}).([]*AstNode)
	if len(functionSpecifiers) != 0 {
		specifiers.Inline = true
	}

	return specifiers, midType, nil
}

func checkStorageClassSpecifiers(specifiers syntax.Specifiers) error {
	// TODO check storage class specifiers conflict
	return nil
}

func parseStorageClassSpecifier(storageSpecifier *AstNode, spe *syntax.Specifiers) {
	n := storageSpecifier.Children[0]
	switch n.Typ {
	case common.TYPEDEF:
		spe.TypeDef = true
	case common.EXTERN:
		spe.Extern = true
	case common.STATIC:
		spe.Static = true
	case common.AUTO:
		spe.Auto = true
	case common.REGISTER:
		spe.Register = true
	}
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
		// need to parse the most out node
		if currentNode == root.Parent {
			break
		}
		prod := productions[currentNode.ProdIndex]
		if currentNode.Typ == declarator {
			if len(prod.Right) == 1 {
				// declarator := direct_declarator
				currentNode = currentNode.Parent
				continue
			}
			// reduced by declarator := pointer direct_declarator
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
			currentType.MetaType = syntax.MetaTypeArray
			currentType.ArrayMetaInfo = parseArrayMetaInfo(currentNode)
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
			continue
		}
		if currentNode.Children[1].Typ == common.LEFT_PARENTHESES {
			currentType.MetaType = syntax.MetaTypeFunction
			currentType.FunctionMetaInfo = parseFunctionMetaInfo(currentNode)
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
			continue
		}
		return res, errors.New("unknown current node type")
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
		parseTypeQualifiers(typeQualifiers, &currentType.TypeQualifiers)
		currentType = currentType.PointerInnerType
		currentPointer = currentPointer.Parent
	}
	currentType.MetaType = syntax.MetaTypePointer
	currentType.PointerInnerType = &syntax.Type{}
	if len(rootPointer.Children) == 2 {
		// parse the root qualifiers
		typeQualifiers := flattenTypeQualifierList(rootPointer.Children[1])
		parseTypeQualifiers(typeQualifiers, &currentType.TypeQualifiers)
	}
	return currentType
}

func flattenTypeQualifierList(listNode *AstNode) []*AstNode {
	if len(listNode.Children) == 1 {
		return []*AstNode{listNode.Children[0]}
	}
	return append(flattenTypeQualifierList(listNode.Children[0]), listNode.Children[1])
}

func parseArrayMetaInfo(arrayNode *AstNode) *syntax.ArrayMetaInfo {
	res := &syntax.ArrayMetaInfo{InnerType: &syntax.Type{}}
	prod := productions[arrayNode.ProdIndex]
	for i := 1; i < len(prod.Right); i++ {
		child := arrayNode.Children[i]
		if child.Typ == common.LEFT_BRACKETS ||
			child.Typ == common.RIGHT_BRACKETS {
			continue
		}
		if child.Typ == common.STATIC {
			res.Static = true
			continue
		}
		if child.Typ == common.ASTERISK {
			res.Asterisk = true
		}
		if child.Typ == type_qualifier_list {
			parseTypeQualifiers(flattenTypeQualifierList(child), &res.TypeQualifiers)
			continue
		}
		// assignment_expression
		res.Size = ParseExpressionNode(child)
	}
	// TODO Check MetaInfo
	return res
}
