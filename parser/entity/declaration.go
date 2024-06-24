package entity

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/glr"
)

type Declaration struct {
	Specifiers  Specifiers
	MidType     Type
	Declarators []Declarator
}

func (d *Declaration) GetUnitType() TranslationUnitType {
	return UnitTypeDeclaration
}

func (d *Declaration) GetBlockItemType() common.TokenType {
	return glr.Declaration
}

type Declarator struct {
	Identifier  string
	Type        Type
	Initializer *Initializer
}
