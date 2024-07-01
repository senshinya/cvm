package entity

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/glr"
)

type Declaration struct {
	Specifiers      Specifiers
	MidType         Type
	InitDeclarators []InitDeclarator
	common.SourceRange
}

func (d *Declaration) GetExternalDeclarationType() ExternalDeclarationType {
	return ExternalDeclarationTypeDeclaration
}

func (d *Declaration) GetBlockItemType() common.TokenType {
	return glr.Declaration
}

type Declarator struct {
	Identifier *common.Token
	Type       Type
	common.SourceRange
}

type InitDeclarator struct {
	Declarator  *Declarator
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
