package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func parseFunctionMetaInfo(node *glr.RawAstNode) (*entity.FunctionMetaInfo, error) {
	if err := node.AssertNonTerminal(glr.DirectAbstractDeclarator, glr.DirectDeclarator); err != nil {
		panic(err)
	}

	res := &entity.FunctionMetaInfo{ReturnType: &entity.Type{}}
	switch {
	case node.ReducedBy(glr.DirectDeclarator, 12):
		// direct_declarator := direct_declarator LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
		params, variadic, err := parseParameterTypeList(node.Children[2])
		if err != nil {
			return nil, err
		}
		res.Parameters = params
		res.Variadic = variadic
		return res, nil
	case node.ReducedBy(glr.DirectDeclarator, 13):
		// direct_declarator := direct_declarator LEFT_PARENTHESES RIGHT_PARENTHESES
		return res, nil
	case node.ReducedBy(glr.DirectDeclarator, 14):
		// direct_declarator := direct_declarator LEFT_PARENTHESES identifier_list RIGHT_PARENTHESES
		res.IdentifierList = parseIdentifierList(node.Children[2])
		return res, nil
	case node.ReducedBy(glr.DirectAbstractDeclarator, 8):
		// direct_abstract_declarator := LEFT_PARENTHESES RIGHT_PARENTHESES
		return res, nil
	case node.ReducedBy(glr.DirectAbstractDeclarator, 9):
		// direct_abstract_declarator := LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
		params, variadic, err := parseParameterTypeList(node.Children[1])
		if err != nil {
			return nil, err
		}
		res.Parameters = params
		res.Variadic = variadic
		return res, nil
	case node.ReducedBy(glr.DirectAbstractDeclarator, 19):
		// direct_abstract_declarator := direct_abstract_declarator LEFT_PARENTHESES RIGHT_PARENTHESES
		return res, nil
	case node.ReducedBy(glr.DirectAbstractDeclarator, 20):
		// direct_abstract_declarator := direct_abstract_declarator LEFT_PARENTHESES parameter_type_list RIGHT_PARENTHESES
		params, variadic, err := parseParameterTypeList(node.Children[2])
		if err != nil {
			return nil, err
		}
		res.Parameters = params
		res.Variadic = variadic
		return res, nil
	default:
		panic("unreachable")
	}
}

func parseParameterTypeList(node *glr.RawAstNode) ([]*entity.FunctionParameter, bool, error) {
	if err := node.AssertNonTerminal(glr.ParameterTypeList); err != nil {
		panic(err)
	}

	variadic := node.ReducedBy(glr.ParameterTypeList, 2)

	parameterList := node.Children[0]
	parameterDeclarations := flattenParameterList(parameterList)
	var params []*entity.FunctionParameter
	for _, paramDeclare := range parameterDeclarations {
		param, err := parseFunctionParameter(paramDeclare)
		if err != nil {
			return nil, false, err
		}
		params = append(params, param)
	}

	if len(params) == 1 && params[0].Type.MetaType == entity.MetaTypeVoid {
		params = nil
	}

	return params, variadic, nil
}

func parseFunctionParameter(paramDeclare *glr.RawAstNode) (*entity.FunctionParameter, error) {
	if err := paramDeclare.AssertNonTerminal(glr.ParameterDeclaration); err != nil {
		panic(err)
	}

	res := &entity.FunctionParameter{SourceRange: paramDeclare.GetSourceRange()}

	specifiers, midType, err := parseDeclarationSpecifiers(paramDeclare.Children[0])
	if err != nil {
		return nil, err
	}
	res.Specifiers = specifiers

	switch {
	case paramDeclare.ReducedBy(glr.ParameterDeclaration, 1):
		// parameter_declaration := declaration_specifiers
		res.Type = midType
	case paramDeclare.ReducedBy(glr.ParameterDeclaration, 2):
		// parameter_declaration := declaration_specifiers declarator
		declare, err := parseDeclarator(paramDeclare.Children[1], midType)
		if err != nil {
			return nil, err
		}
		res.Identifier = declare.Identifier
		res.Type = declare.Type
	case paramDeclare.ReducedBy(glr.ParameterDeclaration, 3):
		// parameter_declaration := declaration_specifiers abstract_declarator
		res.Type, err = ParseAbstractDeclarator(paramDeclare.Children[1], midType)
		if err != nil {
			return nil, err
		}
	default:
		panic("unreachable")
	}

	return res, nil
}

func parseIdentifierList(node *glr.RawAstNode) []*common.Token {
	if err := node.AssertNonTerminal(glr.IdentifierList); err != nil {
		panic(err)
	}
	if node.ReducedBy(glr.IdentifierList, 1) {
		return []*common.Token{node.Children[0].Terminal}
	}

	return append(parseIdentifierList(node.Children[0]), node.Children[2].Terminal)
}
