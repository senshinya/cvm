package sema

import "shinya.click/cvm/entity"

func findFirstNode(node *entity.AstNode, typ entity.TokenType) *entity.AstNode {
	if node == nil {
		return nil
	}
	if node.Typ == typ {
		return node
	}
	for _, c := range node.Children {
		if got := findFirstNode(c, typ); got != nil {
			return got
		}
	}
	return nil
}
