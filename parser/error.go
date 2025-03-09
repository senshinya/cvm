package parser

import (
	"fmt"
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func UndeclaredIdentifier(pos entity.SourcePos, identifier string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("use of undeclared identifier '%s'", identifier)),
	)
}

func RedefinitionSymbol(pos, previous entity.SourcePos, identifier string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("redefinition of '%s' as different kind of symbol", identifier)),
		common.NewNoteMessage(previous, "previous definition is here"),
	)
}

func InvalidSymbolKind(pos, previous entity.SourcePos, identifier string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("invalid symbol kind '%s'", identifier)),
		common.NewNoteMessage(previous, "symbol definition is here"),
	)
}

func InvalidTypeSpecifier(pos entity.SourcePos) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, "invalid type specifier"),
	)
}
