package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseTypeName(typeNameNode *entity.AstNode) (typ entity.Type, err error) {
	if err := typeNameNode.AssertNonTerminal(glr.TypeName); err != nil {
		panic(err)
	}

	specifiersQualifiers := typeNameNode.Children[0]
	specifierNodes := flattenSpecifiersQualifiers(specifiersQualifiers)

	midType, err := parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *entity.AstNode) bool {
			return specifier.Typ == glr.TypeSpecifier
		}).([]*entity.AstNode),
		funk.Filter(specifierNodes, func(specifier *entity.AstNode) bool {
			return specifier.Typ == glr.TypeQualifier
		}).([]*entity.AstNode),
	)
	if err != nil {
		return
	}

	switch {
	case typeNameNode.ReducedBy(glr.TypeName, 1):
		// type_name := specifier_qualifier_list
		return midType, nil
	case typeNameNode.ReducedBy(glr.TypeName, 2):
		// type_name := specifier_qualifier_list abstract_declarator
		return ParseAbstractDeclarator(typeNameNode.Children[1], midType), nil
	default:
		panic("unreachable")
	}
}

func ParseAbstractDeclarator(root *entity.AstNode, midType entity.Type) (res entity.Type) {
	if err := root.AssertNonTerminal(glr.AbstractDeclarator); err != nil {
		panic(err)
	}

	mostInnerNode := findMostInnerNode(root)

	currentNode := mostInnerNode
	currentType := &res
	for {
		// need to parse the most out node
		if currentNode == root.Parent {
			break
		}
		switch {
		case currentNode.ReducedBy(glr.AbstractDeclarator, 1, 3):
			// abstract_declarator := pointer
			// abstract_declarator := pointer direct_abstract_declarator
			currentType = parsePointer(currentNode.Children[0], currentType).PointerInnerType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.AbstractDeclarator, 2):
			// abstract_declarator := direct_abstract_declarator
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectAbstractDeclarator, 1, 2, 3, 4, 5, 6, 7):
			currentType.MetaType = entity.MetaTypeArray
			currentType.ArrayMetaInfo = parseArrayMetaInfo(currentNode)
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectAbstractDeclarator, 8, 9):
			// direct_abstract_declarator := LEFT_PARENTHESES RIGHT_PARENTHESES
			// direct_abstract_declarator := LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
			currentType.MetaType = entity.MetaTypeFunction
			currentType.FunctionMetaInfo = parseFunctionMetaInfo(currentNode)
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectAbstractDeclarator, 10):
			// direct_abstract_declarator := LEFT_PARENTHESES abstract_declarator RIGHT_PARENTHESES
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectAbstractDeclarator, 11, 12, 13, 14, 15, 16, 17, 18):
			currentType.MetaType = entity.MetaTypeArray
			currentType.ArrayMetaInfo = parseArrayMetaInfo(currentNode)
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
		case currentNode.ReducedBy(glr.DirectAbstractDeclarator, 19, 20):
			// direct_abstract_declarator := direct_abstract_declarator LEFT_PARENTHESES RIGHT_PARENTHESES
			// direct_abstract_declarator := direct_abstract_declarator LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
			currentType.MetaType = entity.MetaTypeFunction
			currentType.FunctionMetaInfo = parseFunctionMetaInfo(currentNode)
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
		default:
			panic("unreachable")
		}
	}
	*currentType = midType
	return res
}

func findMostInnerNode(root *entity.AstNode) *entity.AstNode {
	current := root
	for {
		switch {
		case current.ReducedBy(glr.AbstractDeclarator, 1):
			// abstract_declarator := pointer
			return current
		case current.ReducedBy(glr.AbstractDeclarator, 2):
			// abstract_declarator := direct_abstract_declarator
			current = current.Children[0]
		case current.ReducedBy(glr.AbstractDeclarator, 3):
			// abstract_declarator := pointer direct_abstract_declarator
			current = current.Children[1]
		case current.ReducedBy(glr.DirectAbstractDeclarator, 1, 2, 3, 4, 5, 6, 7):
			return current
		case current.ReducedBy(glr.DirectAbstractDeclarator, 8, 9):
			// direct_abstract_declarator := LEFT_PARENTHESES RIGHT_PARENTHESES
			// direct_abstract_declarator := LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
			return current
		case current.ReducedBy(glr.DirectAbstractDeclarator, 10):
			// // direct_abstract_declarator := LEFT_PARENTHESES abstract_declarator RIGHT_PARENTHESES
			current = current.Children[1]
		case current.ReducedBy(glr.DirectAbstractDeclarator, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20):
			current = current.Children[0]
		default:
			panic("unreachable")
		}
	}
	return current
}
