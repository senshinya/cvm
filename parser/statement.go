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
		statement, err := ParseStatement(root.Children[1])
		if err != nil {
			return nil, err
		}
		return statement, nil
	default:
		panic("unreachable")
	}
}

func ParseStatement(root *entity.AstNode) (res *entity.Statement, err error) {
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
		stmt, err := ParseStatement(root.Children[2])
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
		stmt, err := ParseStatement(root.Children[3])
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
		stmt, err := ParseStatement(root.Children[2])
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

	res := &entity.Statement{StmtType: entity.StmtTypeExpression}

	switch {
	case root.ReducedBy(glr.ExpressionStatement, 1):
		// expression_statement := SEMICOLON
		res.ExpressionStmtMetaInfo = &entity.ExpressionStmtMetaInfo{}
	case root.ReducedBy(glr.ExpressionStatement, 2):
		// expression_statement := expression SEMICOLON
		exp, err := ParseExpressionNode(root.Children[0])
		if err != nil {
			return nil, err
		}
		res.ExpressionStmtMetaInfo = &entity.ExpressionStmtMetaInfo{
			Expression: exp,
		}
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
		// selection_statement := IF LEFT_PARENTHESES expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeIf
		cond, err := ParseExpressionNode(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[4])
		if err != nil {
			return nil, err
		}
		res.IfStmtMetaInfo = &entity.IfStmtMetaInfo{
			Condition: cond,
			Body:      body,
		}
	case root.ReducedBy(glr.SelectionStatement, 2):
		// selection_statement := IF LEFT_PARENTHESES expression RIGHT_PARENTHESES statement ELSE statement
		res.StmtType = entity.StmtTypeIfElse
		cond, err := ParseExpressionNode(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[4])
		if err != nil {
			return nil, err
		}
		elseBody, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.IfElseStmtMetaInfo = &entity.IfElseStmtMetaInfo{
			Condition: cond,
			Body:      body,
			ElseBody:  elseBody,
		}
	case root.ReducedBy(glr.SelectionStatement, 3):
		// selection_statement := SWITCH LEFT_PARENTHESES expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeSwitch
		cond, err := ParseExpressionNode(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[4])
		if err != nil {
			return nil, err
		}
		res.SwitchStmtMetaInfo = &entity.SwitchStmtMetaInfo{
			Condition: cond,
			Body:      body,
		}
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
		// iteration_statement := WHILE LEFT_PARENTHESES expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeWhile
		cond, err := ParseExpressionNode(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[4])
		if err != nil {
			return nil, err
		}
		res.WhileStmtMetaInfo = &entity.WhileStmtMetaInfo{
			Condition: cond,
			Body:      body,
		}
	case root.ReducedBy(glr.IterationStatement, 2):
		// iteration_statement := DO statement WHILE LEFT_PARENTHESES expression RIGHT_PARENTHESES SEMICOLON
		res.StmtType = entity.StmtTypeDoWhile
		body, err := ParseStatement(root.Children[1])
		if err != nil {
			return nil, err
		}
		cond, err := ParseExpressionNode(root.Children[4])
		if err != nil {
			return nil, err
		}
		res.DoWhileStmtMetaInfo = &entity.DoWhileStmtMetaInfo{
			Body:      body,
			Condition: cond,
		}
	case root.ReducedBy(glr.IterationStatement, 3):
		// iteration_statement := FOR LEFT_PARENTHESES SEMICOLON SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		body, err := ParseStatement(root.Children[5])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Body: body,
		}
	case root.ReducedBy(glr.IterationStatement, 4):
		// iteration_statement := FOR LEFT_PARENTHESES expression SEMICOLON SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := ParseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 5):
		// iteration_statement := FOR LEFT_PARENTHESES SEMICOLON expression SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		cond, err := ParseExpressionNode(root.Children[3])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Condition: cond,
			Body:      body,
		}
	case root.ReducedBy(glr.IterationStatement, 6):
		// iteration_statement := FOR LEFT_PARENTHESES SEMICOLON SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		afterThought, err := ParseExpressionNode(root.Children[4])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			AfterThought: afterThought,
			Body:         body,
		}
	case root.ReducedBy(glr.IterationStatement, 7):
		// iteration_statement := FOR LEFT_PARENTHESES expression SEMICOLON expression SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := ParseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		cond, err := ParseExpressionNode(root.Children[4])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[7])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Condition:      cond,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 8):
		// iteration_statement := FOR LEFT_PARENTHESES expression SEMICOLON SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := ParseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		afterThought, err := ParseExpressionNode(root.Children[5])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[7])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			AfterThought:   afterThought,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 9):
		// iteration_statement := FOR LEFT_PARENTHESES SEMICOLON expression SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		cond, err := ParseExpressionNode(root.Children[3])
		if err != nil {
			return nil, err
		}
		afterThought, err := ParseExpressionNode(root.Children[5])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[7])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Condition:    cond,
			AfterThought: afterThought,
			Body:         body,
		}
	case root.ReducedBy(glr.IterationStatement, 10):
		// iteration_statement := FOR LEFT_PARENTHESES expression SEMICOLON expression SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := ParseStatement(root.Children[2])
		if err != nil {
			return nil, err
		}
		cond, err := ParseExpressionNode(root.Children[4])
		if err != nil {
			return nil, err
		}
		afterThought, err := ParseExpressionNode(root.Children[6])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[8])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Condition:      cond,
			AfterThought:   afterThought,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 11):
		// iteration_statement := FOR LEFT_PARENTHESES declaration SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := parseDeclaration(root.Children[2])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[5])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 12):
		// iteration_statement := FOR LEFT_PARENTHESES declaration expression SEMICOLON RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := parseDeclaration(root.Children[2])
		if err != nil {
			return nil, err
		}
		cond, err := ParseExpressionNode(root.Children[3])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Condition:      cond,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 13):
		// iteration_statement := FOR LEFT_PARENTHESES declaration SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := parseDeclaration(root.Children[2])
		if err != nil {
			return nil, err
		}
		afterThought, err := ParseExpressionNode(root.Children[4])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[6])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			AfterThought:   afterThought,
			Body:           body,
		}
	case root.ReducedBy(glr.IterationStatement, 14):
		// iteration_statement := FOR LEFT_PARENTHESES declaration expression SEMICOLON expression RIGHT_PARENTHESES statement
		res.StmtType = entity.StmtTypeFor
		init, err := parseDeclaration(root.Children[2])
		if err != nil {
			return nil, err
		}
		cond, err := ParseExpressionNode(root.Children[3])
		if err != nil {
			return nil, err
		}
		afterThought, err := ParseExpressionNode(root.Children[5])
		if err != nil {
			return nil, err
		}
		body, err := ParseStatement(root.Children[7])
		if err != nil {
			return nil, err
		}
		res.ForStmtMetaInfo = &entity.ForStmtMetaInfo{
			Initialization: init,
			Condition:      cond,
			AfterThought:   afterThought,
			Body:           body,
		}
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
		// jump_statement := GOTO IDENTIFIER SEMICOLON
		res.StmtType = entity.StmtTypeGoto
		res.GotoStmtMetaInfo = &entity.GotoStmtMetaInfo{
			Identifier: root.Children[1].Terminal.Lexeme,
		}
	case root.ReducedBy(glr.JumpStatement, 2):
		// jump_statement := CONTINUE SEMICOLON
		res.StmtType = entity.StmtTypeContinue
	case root.ReducedBy(glr.JumpStatement, 3):
		// jump_statement := BREAK SEMICOLON
		res.StmtType = entity.StmtTypeBreak
	case root.ReducedBy(glr.JumpStatement, 4):
		// jump_statement := RETURN SEMICOLON
		res.StmtType = entity.StmtTypeReturn
		res.ReturnStmtMetaInfo = &entity.ReturnStmtMetaInfo{}
	case root.ReducedBy(glr.JumpStatement, 5):
		// jump_statement := RETURN expression SEMICOLON
		res.StmtType = entity.StmtTypeReturn
		returnValue, err := ParseExpressionNode(root.Children[1])
		if err != nil {
			return nil, err
		}
		res.ReturnStmtMetaInfo = &entity.ReturnStmtMetaInfo{
			Value: returnValue,
		}
	default:
		panic("unreachable")
	}

	return res, nil
}
