package compiler

import (
	"shinya.click/cvm/compiler/entity"
	parser_entity "shinya.click/cvm/parser/entity"
)

type Compiler struct {
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(units []parser_entity.ExternalDeclaration) error {
	ctx := entity.SemanticsContext{
		ConstantPool: entity.ConstantPool{},
		CurrentScope: entity.NewFileScope(),
	}

	for _, unit := range units {
		err := c.handleExternalDeclaration(ctx, unit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) handleExternalDeclaration(ctx entity.SemanticsContext, unit parser_entity.ExternalDeclaration) error {
	switch unit.GetExternalDeclarationType() {
	case parser_entity.ExternalDeclarationTypeDeclaration:
		return c.handleDeclaration(ctx, unit.(*parser_entity.Declaration))
	case parser_entity.ExternalDeclarationTypeFunctionDef:
		return c.handleFunctionDef(ctx, unit.(*parser_entity.FunctionDefinition))
	default:
		panic("unreachable")
	}
}
