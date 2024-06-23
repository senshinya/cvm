package parser

import (
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func flattenDeclarationSpecifier(specifiers *entity.AstNode) []*entity.AstNode {
	if specifiers.ReducedBy(glr.DeclarationSpecifiers, 1, 2, 3, 4) {
		return []*entity.AstNode{specifiers.Children[0]}
	}

	return append(flattenDeclarationSpecifier(specifiers.Children[1]), specifiers.Children[0])
}

func flattenInitDeclarators(declarators *entity.AstNode) []*entity.AstNode {
	if declarators.ReducedBy(glr.InitDeclaratorList, 1) {
		return []*entity.AstNode{declarators.Children[0]}
	}

	return append(flattenInitDeclarators(declarators.Children[0]), declarators.Children[2])
}

func flattenTypeQualifierList(listNode *entity.AstNode) []*entity.AstNode {
	if listNode.ReducedBy(glr.TypeQualifierList, 1) {
		return []*entity.AstNode{listNode.Children[0]}
	}
	return append(flattenTypeQualifierList(listNode.Children[0]), listNode.Children[1])
}

func flattenExpression(node *entity.AstNode) []*entity.AstNode {
	if node.ReducedBy(glr.Expression, 1) {
		return []*entity.AstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}

func flattenParameterList(parameterList *entity.AstNode) []*entity.AstNode {
	if parameterList.ReducedBy(glr.ParameterList, 1) {
		return []*entity.AstNode{parameterList.Children[0]}
	}

	return append(flattenParameterList(parameterList.Children[0]), parameterList.Children[2])
}

func flattenStructDeclaratorList(root *entity.AstNode) []*entity.AstNode {
	if root.ReducedBy(glr.StructDeclarationList, 1) {
		return []*entity.AstNode{root.Children[0]}
	}

	return append(flattenStructDeclaratorList(root.Children[0]), root.Children[2])
}

func flattenStructDeclarationList(root *entity.AstNode) []*entity.AstNode {
	if root.ReducedBy(glr.StructDeclaratorList, 1) {
		return []*entity.AstNode{root.Children[0]}
	}

	return append(flattenStructDeclarationList(root.Children[0]), root.Children[1])
}

func flattenSpecifiersQualifiers(specifiersQualifiers *entity.AstNode) []*entity.AstNode {
	if specifiersQualifiers.ReducedBy(glr.SpecifierQualifierList, 1, 3) {
		return []*entity.AstNode{specifiersQualifiers.Children[0]}
	}

	return append(flattenSpecifiersQualifiers(specifiersQualifiers.Children[1]), specifiersQualifiers.Children[0])
}

func flattenDesignatorList(root *entity.AstNode) []*entity.AstNode {
	if root.ReducedBy(glr.DesignatorList, 1) {
		return []*entity.AstNode{root.Children[0]}
	}

	return append(flattenDesignatorList(root.Children[0]), root.Children[1])
}

func flattenEnumerators(root *entity.AstNode) []*entity.AstNode {
	if root.ReducedBy(glr.EnumeratorList, 1) {
		return []*entity.AstNode{root.Children[0]}
	}

	return append(flattenEnumerators(root.Children[0]), root.Children[2])
}

func flattenArgumentExpressions(root *entity.AstNode) []*entity.AstNode {
	if root.ReducedBy(glr.ArgumentExpressionList, 1) {
		return []*entity.AstNode{root.Children[0]}
	}

	return append(flattenArgumentExpressions(root.Children[0]), root.Children[2])
}
