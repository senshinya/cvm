package parser

import (
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func ParseCompoundStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.CompoundStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{StmtType: entity.StmtTypeCompound}
	meta := &entity.CompoundStmtMetaInfo{}

	switch {
	case root.ReducedBy(glr.CompoundStatement, 1):
		// compound_statement := LEFT_BRACES RIGHT_BRACES
	case root.ReducedBy(glr.CompoundStatement, 2):
		// compound_statement := LEFT_BRACES block_item_list RIGHT_BRACES
		blockItemNodes := flattenBlockItemList(root.Children[1])
		for _, blockItemNode := range blockItemNodes {
			blockItem, err := parseBlockItem(blockItemNode)
			if err != nil {
				return nil, err
			}
			meta.BlockItems = append(meta.BlockItems, blockItem)
		}
	default:
		panic("unreachable")
	}

	res.CompoundStmtMetaInfo = meta
	return res, nil
}

func parseBlockItem(root *entity.AstNode) (entity.BlockItem, error) {
	switch {
	case root.ReducedBy(glr.BlockItem, 1):
		// block_item := declaration
		declaration, err := parseDeclaration(root.Children[1])
		if err != nil {
			return nil, err
		}
		return declaration, nil
	case root.ReducedBy(glr.BlockItem, 2):
		// block_item := statement
		statement, err := parseStatement(root.Children[1])
		if err != nil {
			return nil, err
		}
		return statement, nil
	default:
		panic("unreachable")
	}
}

func parseStatement(root *entity.AstNode) (res *entity.Statement, err error) {
	if err = root.AssertNonTerminal(glr.Statement); err != nil {
		panic(err)
	}

	switch {
	case root.ReducedBy(glr.Statement, 1):
		// statement := labeled_statement
		res, err = ParseLabeledStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Statement, 2):
		// statement := compound_statement
		res, err = ParseCompoundStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Statement, 3):
		// statement := expression_statement
		res, err = ParseExpressionStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Statement, 4):
		// statement := selection_statement
		res, err = ParseSelectionStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Statement, 5):
		// statement := iteration_statement
		res, err = ParseIterationStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	case root.ReducedBy(glr.Statement, 6):
		// statement := jump_statement
		res, err = ParseJumpStatement(root.Children[0])
		if err != nil {
			return nil, err
		}
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseLabeledStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.LabeledStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{}

	switch {
	case root.ReducedBy(glr.LabeledStatement, 1):
		// labeled_statement := IDENTIFIER COLON statement
		res.StmtType = entity.StmtTypeLabeled
		stmt, err := parseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		res.LabeledStmtMetaInfo = &entity.LabeledStmtMetaInfo{
			Identifier: root.Children[0].Terminal.Lexeme,
			Body:       stmt,
		}
	case root.ReducedBy(glr.LabeledStatement, 2):
		// labeled_statement := CASE constant_expression COLON statement
		res.StmtType = entity.StmtTypeCase
		cond, err := ParseExpressionNode(root.Children[1])
		if err != nil {
			return nil, err
		}
		stmt, err := parseStatement(root.Children[3])
		if err != nil {
			return nil, err
		}
		res.CaseStmtMetaInfo = &entity.CaseStmtMetaInfo{
			Condition: cond,
			Body:      stmt,
		}
	case root.ReducedBy(glr.LabeledStatement, 3):
		// labeled_statement := DEFAULT COLON statement
		res.StmtType = entity.StmtTypeDefault
		stmt, err := parseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		res.DefaultStmtMetaInfo = &entity.DefaultStmtMetaInfo{
			Body: stmt,
		}
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseExpressionStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.ExpressionStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{}

	switch {
	case root.ReducedBy(glr.ExpressionStatement, 1):
	case root.ReducedBy(glr.ExpressionStatement, 2):
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseSelectionStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.SelectionStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{}

	switch {
	case root.ReducedBy(glr.SelectionStatement, 1):
	case root.ReducedBy(glr.SelectionStatement, 2):
	case root.ReducedBy(glr.SelectionStatement, 3):
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseIterationStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.IterationStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{}

	switch {
	case root.ReducedBy(glr.IterationStatement, 1):
	case root.ReducedBy(glr.IterationStatement, 2):
	case root.ReducedBy(glr.IterationStatement, 3):
	case root.ReducedBy(glr.IterationStatement, 4):
	case root.ReducedBy(glr.IterationStatement, 5):
	case root.ReducedBy(glr.IterationStatement, 6):
	case root.ReducedBy(glr.IterationStatement, 7):
	case root.ReducedBy(glr.IterationStatement, 8):
	case root.ReducedBy(glr.IterationStatement, 9):
	case root.ReducedBy(glr.IterationStatement, 10):
	case root.ReducedBy(glr.IterationStatement, 11):
	case root.ReducedBy(glr.IterationStatement, 12):
	case root.ReducedBy(glr.IterationStatement, 13):
	case root.ReducedBy(glr.IterationStatement, 14):
	default:
		panic("unreachable")
	}

	return res, nil
}

func ParseJumpStatement(root *entity.AstNode) (*entity.Statement, error) {
	if err := root.AssertNonTerminal(glr.JumpStatement); err != nil {
		panic(err)
	}

	res := &entity.Statement{}

	switch {
	case root.ReducedBy(glr.JumpStatement, 1):
	case root.ReducedBy(glr.JumpStatement, 2):
	case root.ReducedBy(glr.JumpStatement, 3):
	case root.ReducedBy(glr.JumpStatement, 4):
	case root.ReducedBy(glr.JumpStatement, 5):
	default:
		panic("unreachable")
	}

	return res, nil
}
