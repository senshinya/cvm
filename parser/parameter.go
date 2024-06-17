package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
)

func parseFunctionMetaInfo(node *AstNode) *entity.FunctionMetaInfo {
	res := &entity.FunctionMetaInfo{ReturnType: &entity.Type{}}
	prod := productions[node.ProdIndex]
	switch node.Typ {
	case direct_declarator:
		if len(prod.Right) == 3 {
			// reduced by direct_declarator := direct_declarator LEFT_PARENTHESES RIGHT_PARENTHESES
			return res
		}
		if prod.Right[2] == parameter_type_list {
			params, variadic := parseParameterTypeList(node.Children[2])
			res.Parameters = params
			res.Variadic = variadic
			return res
		}
		// identifier_list
		res.IdentifierList = parseIdentifierList(node.Children[2])
		return res
	case direct_abstract_declarator:
		if prod.Right[0] == common.LEFT_PARENTHESES {
			if len(prod.Right) == 2 {
				// reduced by direct_abstract_declarator := LEFT_PARENTHESES RIGHT_PARENTHESES
				return res
			}
			params, variadic := parseParameterTypeList(node.Children[1])
			res.Parameters = params
			res.Variadic = variadic
			return res
		}
		if len(prod.Right) == 3 {
			return res
		}
		params, variadic := parseParameterTypeList(node.Children[2])
		res.Parameters = params
		res.Variadic = variadic
		return res
	}
	panic("unreachable")
}

func parseParameterTypeList(node *AstNode) ([]*entity.FunctionParameter, bool) {
	variadic := false

	prod := productions[node.ProdIndex]
	if len(prod.Right) == 3 {
		variadic = true
	}
	parameterList := node.Children[0]

	parameterDeclarations := flattenParameterList(parameterList)
	var params []*entity.FunctionParameter
	for _, paramDeclare := range parameterDeclarations {
		params = append(params, parseParamDeclare(paramDeclare))
	}

	if len(params) == 1 && params[0].Type.MetaType == entity.MetaTypeVoid {
		params = nil
	}

	return params, variadic
}

func parseParamDeclare(paramDeclare *AstNode) *entity.FunctionParameter {
	res := &entity.FunctionParameter{}

	tmp := &entity.Declaration{}
	isTypeDef, err := parseDeclarationSpecifiers(paramDeclare.Children[0], tmp)
	if isTypeDef {
		panic("type def should not appear in parameter declaration")
	}
	if err != nil {
		panic(err)
	}
	res.Specifiers = tmp.Specifiers

	prod := productions[paramDeclare.ProdIndex]
	if len(prod.Right) == 1 {
		// parameter_declaration := declaration_specifiers
		res.Type = tmp.MidType
		return res
	}

	if prod.Right[1] == declarator {
		// parameter_declaration := declaration_specifiers declarator
		declare, err := parseDeclarator(paramDeclare.Children[1], tmp.MidType)
		if err != nil {
			panic(err)
		}
		res.Identifier = &declare.Identifier
		res.Type = declare.Type
		return res
	}
	// parameter_declaration := declaration_specifiers abstract_declarator
	res.Type = ParseAbstractDeclarator(paramDeclare.Children[1], tmp.MidType)
	return res
}

func flattenParameterList(parameterList *AstNode) []*AstNode {
	prod := productions[parameterList.ProdIndex]
	if len(prod.Right) == 1 {
		return []*AstNode{parameterList.Children[0]}
	}

	return append(flattenParameterList(parameterList.Children[0]), parameterList.Children[2])
}

func parseIdentifierList(node *AstNode) []string {
	prod := productions[node.ProdIndex]
	if len(prod.Right) == 1 {
		return []string{node.Children[0].Terminal.Lexeme}
	}

	return append(parseIdentifierList(node.Children[0]), node.Children[2].Terminal.Lexeme)
}
