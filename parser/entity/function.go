package entity

import (
	"shinya.click/cvm/common"
)

type FunctionDefinition struct {
	Specifiers   Specifiers
	Declarator   Declarator
	Declarations []*Declaration
	BlockItems   []BlockItem
}

func (d *FunctionDefinition) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}

type BlockItem interface {
	GetBlockItemType() common.TokenType
}
