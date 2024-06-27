package entity

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/glr"
)

type Declaration struct {
	Specifiers  Specifiers
	MidType     Type
	Declarators []Declarator
	common.SourceRange
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
	common.SourceRange
}

// Specifiers without type specifiers
type Specifiers struct {
	// storage class specifier
	Extern   bool
	Static   bool
	Auto     bool
	Register bool

	// typedef
	TypeDef bool

	// function specifier
	Inline bool

	common.SourceRange
}
