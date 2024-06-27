package parser

import (
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func parseFunctionDefinition(root *entity.RawAstNode) (*entity.FunctionDefinition, error) {
	if err := root.AssertNonTerminal(glr.FunctionDefinition); err != nil {
		panic(err)
	}

	res := &entity.FunctionDefinition{}

	// parse specifiers
	specifiers, midType, err := parseDeclarationSpecifiers(root.Children[0])
	if err != nil {
		return nil, err
	}
	res.Specifiers = specifiers

	// parse declarator
	res.Declarator, err = parseDeclarator(root.Children[1], midType)
	if err != nil {
		return nil, err
	}

	switch {
	case root.ReducedBy(glr.FunctionDefinition, 1):
		// function_definition := declaration_specifiers declarator compound_statement
		var compoundStmt *entity.Statement
		compoundStmt, err = ParseCompoundStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		res.BlockItems = compoundStmt.CompoundStmtMetaInfo.BlockItems
	case root.ReducedBy(glr.FunctionDefinition, 2):
		// function_definition := declaration_specifiers declarator declaration_list compound_statement
		res.Declarations, err = ParseDeclarationList(root.Children[2])
		if err != nil {
			return nil, err
		}
		var compoundStmt *entity.Statement
		compoundStmt, err = ParseCompoundStatement(root.Children[3])
		if err != nil {
			return nil, err
		}
		res.BlockItems = compoundStmt.CompoundStmtMetaInfo.BlockItems
	default:
		panic("unreachable")
	}

	return res, nil
}
