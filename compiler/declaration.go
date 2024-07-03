package compiler

import (
	"shinya.click/cvm/compiler/entity"
	parser_entity "shinya.click/cvm/parser/entity"
)

func (c *Compiler) handleDeclaration(ctx entity.SemanticsContext, declaration *parser_entity.Declaration) error {
	if declaration.Specifiers.TypeDef {
		return c.handleTypeDef(ctx, declaration)
	}

	return nil
}
