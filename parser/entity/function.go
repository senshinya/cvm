package entity

import (
	"shinya.click/cvm/common"
)

type FunctionDef struct {
	Specifiers   Specifiers
	MidType      Type
	Declarator   Declarator
	Declarations []*Declaration
	BlockItems   []BlockItem
}

func (d *FunctionDef) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}

type BlockItem interface {
	GetBlockItemType() common.TokenType
}
