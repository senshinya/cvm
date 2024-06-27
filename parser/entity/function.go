package entity

import (
	"shinya.click/cvm/common"
)

type FunctionDefinition struct {
	Specifiers   Specifiers
	Declarator   Declarator
	Declarations []*Declaration
	BlockItems   []BlockItem
	common.SourceRange
}

func (d *FunctionDefinition) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}

type BlockItem interface {
	GetBlockItemType() common.TokenType
}
