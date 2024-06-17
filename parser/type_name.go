package parser

import (
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseTypeName(typeNameNode *entity.AstNode) entity.Type {
	specifiersQualifiers := typeNameNode.Children[0]
	specifierNodes := flattenSpecifiersQualifiers(specifiersQualifiers)

	midType := parseTypeSpecifiersAndQualifiers(
		funk.Filter(specifierNodes, func(specifier *entity.AstNode) bool {
			return specifier.Typ == glr.TypeSpecifier
		}).([]*entity.AstNode),
		funk.Filter(specifierNodes, func(specifier *entity.AstNode) bool {
			return specifier.Typ == glr.TypeQualifier
		}).([]*entity.AstNode),
	)

	if len(glr.Productions[typeNameNode.ProdIndex].Right) == 1 {
		return midType
	}

	return ParseAbstractDeclarator(typeNameNode.Children[1], midType)
}

func flattenSpecifiersQualifiers(specifiersQualifiers *entity.AstNode) []*entity.AstNode {
	if len(specifiersQualifiers.Children) == 1 {
		return []*entity.AstNode{specifiersQualifiers.Children[0]}
	}

	return append(flattenSpecifiersQualifiers(specifiersQualifiers.Children[1]), specifiersQualifiers.Children[0])
}

func ParseAbstractDeclarator(root *entity.AstNode, midType entity.Type) entity.Type {
	mostInnerNode := findMostInnerNode(root)

	currentNode := mostInnerNode
	res := entity.Type{}
	currentType := &res
	for {
		// need to parse the most out node
		if currentNode == root.Parent {
			break
		}
		prod := glr.Productions[currentNode.ProdIndex]
		if currentNode.Typ == glr.AbstractDeclarator {
			if prod.Right[0] == glr.Pointer {
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
			if prod.Right[1] == glr.AbstractDeclarator {
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

func findMostInnerNode(root *entity.AstNode) *entity.AstNode {
	current := root
	for {
		prod := glr.Productions[current.ProdIndex]
		rightLen := len(prod.Right)
		switch prod.Right[0] {
		case glr.Pointer:
			if rightLen == 1 {
				return current
			}
			current = current.Children[1]
		case glr.DirectAbstractDeclarator:
			current = current.Children[0]
		case common.LEFT_BRACKETS:
			return current
		case common.LEFT_PARENTHESES:
			if prod.Right[1] == glr.AbstractDeclarator {
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
