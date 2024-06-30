package parser

import (
	"shinya.click/cvm/parser/glr"
)

func flattenDeclarationSpecifier(specifiers *glr.RawAstNode) []*glr.RawAstNode {
	if specifiers.ReducedBy(glr.DeclarationSpecifiers, 1, 2, 3, 4) {
		return []*glr.RawAstNode{specifiers.Children[0]}
	}

	return append(flattenDeclarationSpecifier(specifiers.Children[1]), specifiers.Children[0])
}

func flattenInitDeclarators(declarators *glr.RawAstNode) []*glr.RawAstNode {
	if declarators.ReducedBy(glr.InitDeclaratorList, 1) {
		return []*glr.RawAstNode{declarators.Children[0]}
	}

	return append(flattenInitDeclarators(declarators.Children[0]), declarators.Children[2])
}

func flattenTypeQualifierList(listNode *glr.RawAstNode) []*glr.RawAstNode {
	if listNode.ReducedBy(glr.TypeQualifierList, 1) {
		return []*glr.RawAstNode{listNode.Children[0]}
	}
	return append(flattenTypeQualifierList(listNode.Children[0]), listNode.Children[1])
}

func flattenExpression(node *glr.RawAstNode) []*glr.RawAstNode {
	if node.ReducedBy(glr.Expression, 1) {
		return []*glr.RawAstNode{node.Children[0]}
	}

	return append(flattenExpression(node.Children[0]), node.Children[2])
}

func flattenParameterList(parameterList *glr.RawAstNode) []*glr.RawAstNode {
	if parameterList.ReducedBy(glr.ParameterList, 1) {
		return []*glr.RawAstNode{parameterList.Children[0]}
	}

	return append(flattenParameterList(parameterList.Children[0]), parameterList.Children[2])
}

func flattenStructDeclaratorList(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.StructDeclarationList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenStructDeclaratorList(root.Children[0]), root.Children[2])
}

func flattenStructDeclarationList(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.StructDeclaratorList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenStructDeclarationList(root.Children[0]), root.Children[1])
}

func flattenSpecifiersQualifiers(specifiersQualifiers *glr.RawAstNode) []*glr.RawAstNode {
	if specifiersQualifiers.ReducedBy(glr.SpecifierQualifierList, 1, 3) {
		return []*glr.RawAstNode{specifiersQualifiers.Children[0]}
	}

	return append(flattenSpecifiersQualifiers(specifiersQualifiers.Children[1]), specifiersQualifiers.Children[0])
}

func flattenDesignatorList(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.DesignatorList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenDesignatorList(root.Children[0]), root.Children[1])
}

func flattenEnumerators(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.EnumeratorList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenEnumerators(root.Children[0]), root.Children[2])
}

func flattenArgumentExpressions(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.ArgumentExpressionList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenArgumentExpressions(root.Children[0]), root.Children[2])
}

func flattenDeclarationList(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.DeclarationList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenDeclarationList(root.Children[0]), root.Children[1])
}

func flattenBlockItemList(root *glr.RawAstNode) []*glr.RawAstNode {
	if root.ReducedBy(glr.BlockItemList, 1) {
		return []*glr.RawAstNode{root.Children[0]}
	}

	return append(flattenBlockItemList(root.Children[0]), root.Children[1])
}
