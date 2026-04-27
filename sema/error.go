package sema

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func UndeclaredIdentifier(pos entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("use of undeclared identifier '%s'", name)),
	)
}

func RedefinitionSymbol(pos, previous entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("redefinition of '%s' as different kind of symbol", name)),
		common.NewNoteMessage(previous, "previous definition is here"),
	)
}

func InvalidSymbolKind(pos, previous entity.SourcePos, name string) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(pos, fmt.Sprintf("invalid symbol kind '%s'", name)),
		common.NewNoteMessage(previous, "symbol definition is here"),
	)
}
