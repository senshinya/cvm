package entity

import (
	"shinya.click/cvm/common"
)

type FunctionDefinition struct {
	Specifiers   Specifiers
	Declarator   *Declarator
	Declarations []*Declaration
	BlockItems   []BlockItem
	common.SourceRange
}

func (d *FunctionDefinition) GetExternalDeclarationType() ExternalDeclarationType {
	return ExternalDeclarationTypeFunctionDef
}

type BlockItem interface {
	GetBlockItemType() common.TokenType
}
