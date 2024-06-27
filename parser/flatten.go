package parser

import (
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func flattenDeclarationSpecifier(specifiers *entity.RawAstNode) []*entity.RawAstNode {
	if specifiers.ReducedBy(glr.DeclarationSpecifiers, 1, 2, 3, 4) {
		return []*entity.RawAstNode{specifiers.Children[0]}
	}

	return append(flattenDeclarationSpecifier(specifiers.Children[1]), specifiers.Children[0])
}

func flattenInitDeclarators(declarators *entity.RawAstNode) []*entity.RawAstNode {
	if declarators.ReducedBy(glr.InitDeclaratorList, 1) {
		return []*entity.RawAstNode{declarators.Children[0]}
	}

	return append(flattenInitDeclarators(declarators.Children[0]), declarators.Children[2])
}

func flattenTypeQualifierList(listNode *entity.RawAstNode) []*entity.RawAstNode {
	if listNode.ReducedBy(glr.TypeQualifierList, 1) {
		return []*entity.RawAstNode{listNode.Children[0]}
	}
	return append(flattenTypeQualifierList(listNode.Children[0]), listNode.Children[1])
}

func flattenExpression(node *entity.RawAstNode) []*entity.RawAstNode {
	if node.ReducedBy(glr.Expression, 1) {
		return []*entity.RawAstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}

func flattenParameterList(parameterList *entity.RawAstNode) []*entity.RawAstNode {
	if parameterList.ReducedBy(glr.ParameterList, 1) {
		return []*entity.RawAstNode{parameterList.Children[0]}
	}

	return append(flattenParameterList(parameterList.Children[0]), parameterList.Children[2])
}

func flattenStructDeclaratorList(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.StructDeclarationList, 1) {
		return []*entity.RawAstNode{root.Children[0]}
	}

	return append(flattenStructDeclaratorList(root.Children[0]), root.Children[2])
}

func flattenStructDeclarationList(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.StructDeclaratorList, 1) {
		return []*entity.RawAstNode{root.Children[0]}
	}

	return append(flattenStructDeclarationList(root.Children[0]), root.Children[1])
}

func flattenSpecifiersQualifiers(specifiersQualifiers *entity.RawAstNode) []*entity.RawAstNode {
	if specifiersQualifiers.ReducedBy(glr.SpecifierQualifierList, 1, 3) {
		return []*entity.RawAstNode{specifiersQualifiers.Children[0]}
	}

	return append(flattenSpecifiersQualifiers(specifiersQualifiers.Children[1]), specifiersQualifiers.Children[0])
}

func flattenDesignatorList(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.DesignatorList, 1) {
		return []*entity.RawAstNode{root.Children[0]}
	}

	return append(flattenDesignatorList(root.Children[0]), root.Children[1])
}

func flattenEnumerators(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.EnumeratorList, 1) {
		return []*entity.RawAstNode{root.Children[0]}
	}

	return append(flattenEnumerators(root.Children[0]), root.Children[2])
}

func flattenArgumentExpressions(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.ArgumentExpressionList, 1) {
		return []*entity.RawAstNode{root.Children[0]}
	}

	return append(flattenArgumentExpressions(root.Children[0]), root.Children[2])
}

func flattenDeclarationList(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.DeclarationList, 1) {
		return []*entity.RawAstNode{root}
	}

	return append(flattenDeclarationList(root.Children[0]), root.Children[1])
}

func flattenBlockItemList(root *entity.RawAstNode) []*entity.RawAstNode {
	if root.ReducedBy(glr.BlockItemList, 1) {
		return []*entity.RawAstNode{root}
	}

	return append(flattenBlockItemList(root.Children[0]), root.Children[1])
}
