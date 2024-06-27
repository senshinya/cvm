package parser

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
	"shinya.click/cvm/parser/glr"
)

func parseArrayMetaInfo(arrayNode *entity.RawAstNode) (*entity.ArrayMetaInfo, error) {
	if err := arrayNode.AssertNonTerminal(glr.DirectAbstractDeclarator, glr.DirectDeclarator); err != nil {
		panic(err)
	}

	res := &entity.ArrayMetaInfo{InnerType: &entity.Type{}}
	for i := 0; i < len(arrayNode.Children); i++ {
		child := arrayNode.Children[i]
		if child.Typ == glr.DirectAbstractDeclarator ||
			child.Typ == glr.DirectDeclarator {
			continue
		}
		if child.Typ == common.LEFT_BRACKETS ||
			child.Typ == common.RIGHT_BRACKETS {
			continue
		}
		if child.Typ == common.STATIC {
			res.Static = true
			continue
		}
		if child.Typ == common.ASTERISK {
			res.Asterisk = true
			continue
		}
		if child.Typ == glr.TypeQualifierList {
			parseTypeQualifiers(flattenTypeQualifierList(child), &res.TypeQualifiers)
			continue
		}
		// assignment_expression
		var err error
		res.Size, err = ParseExpressionNode(child)
		if err != nil {
			return nil, err
		}
	}
	if res.Size == nil {
		res.Incomplete = true
	}
	return res, nil
}
