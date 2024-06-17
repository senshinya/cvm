package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
)

func ParseTypeName(typeNameNode *AstNode) entity.Type {
	specifiersQualifiers := typeNameNode.Children[0]
	specifierNodes := flattenSpecifiersQualifiers(specifiersQualifiers)

	midType := parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_specifier
		}).([]*AstNode),
		funk.Filter(specifierNodes, func(specifier *AstNode) bool {
			return specifier.Typ == type_qualifier
		}).([]*AstNode),
	)

	if len(productions[typeNameNode.ProdIndex].Right) == 1 {
		return midType
	}

	return ParseAbstractDeclarator(typeNameNode.Children[1], midType)
}

func flattenSpecifiersQualifiers(specifiersQualifiers *AstNode) []*AstNode {
	if len(specifiersQualifiers.Children) == 1 {
		return []*AstNode{specifiersQualifiers.Children[0]}
	}

	return append(flattenSpecifiersQualifiers(specifiersQualifiers.Children[1]), specifiersQualifiers.Children[0])
}

func ParseAbstractDeclarator(root *AstNode, midType entity.Type) entity.Type {
	mostInnerNode := findMostInnerNode(root)

	currentNode := mostInnerNode
	res := entity.Type{}
	currentType := &res
	for {
		// need to parse the most out node
		if currentNode == root.Parent {
			break
		}
		prod := productions[currentNode.ProdIndex]
		if currentNode.Typ == abstract_declarator {
			if prod.Right[0] == pointer {
				// abstract_declarator := pointer
				// abstract_declarator := pointer direct_abstract_declarator
				currentType = parsePointer(currentNode.Children[0], currentType).PointerInnerType
				currentNode = currentNode.Parent
				continue
			}
			// abstract_declarator := direct_abstract_declarator
			currentNode = currentNode.Parent
			continue
		}
		// current node is direct_abstract_declarator
		if prod.Right[0] == common.LEFT_BRACKETS {
			currentType.MetaType = entity.MetaTypeArray
			currentType.ArrayMetaInfo = parseArrayMetaInfo(currentNode)
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
			continue
		}
		if prod.Right[0] == common.LEFT_PARENTHESES {
			if prod.Right[1] == abstract_declarator {
				// reduced by direct_abstract_declarator := LEFT_PARENTHESES abstract_declarator RIGHT_PARENTHESES
				currentNode = currentNode.Parent
				continue
			}
			currentType.MetaType = entity.MetaTypeFunction
			currentType.FunctionMetaInfo = parseFunctionMetaInfo(currentNode)
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
			continue
		}
		if prod.Right[1] == common.LEFT_BRACKETS {
			currentType.MetaType = entity.MetaTypeArray
			currentType.ArrayMetaInfo = parseArrayMetaInfo(currentNode)
			currentType = currentType.ArrayMetaInfo.InnerType
			currentNode = currentNode.Parent
			continue
		}
		if prod.Right[1] == common.LEFT_PARENTHESES {
			currentType.MetaType = entity.MetaTypeFunction
			currentType.FunctionMetaInfo = parseFunctionMetaInfo(currentNode)
			currentType = currentType.FunctionMetaInfo.ReturnType
			currentNode = currentNode.Parent
			continue
		}
		panic("Unknown node type")
	}
	*currentType = midType
	return res
}

func findMostInnerNode(root *AstNode) *AstNode {
	current := root
	for {
		prod := productions[current.ProdIndex]
		rightLen := len(prod.Right)
		switch prod.Right[0] {
		case pointer:
			if rightLen == 1 {
				return current
			}
			current = current.Children[1]
		case direct_abstract_declarator:
			current = current.Children[0]
		case common.LEFT_BRACKETS:
			return current
		case common.LEFT_PARENTHESES:
			if prod.Right[1] == abstract_declarator {
				current = current.Children[1]
				continue
			}
			return current
		default:
			panic("never happen")
		}
	}
	return current
}
