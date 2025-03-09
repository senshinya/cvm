package lexer

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func UnidentifiedToken(line, column int) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(entity.SourcePos{Line: line, Column: column}, "unidentified token"),
	)
}

func InvalidCharacter(line, column int) *common.CvmError {
	return common.NewCvmError(
		common.NewErrorMessage(entity.SourcePos{Line: line, Column: column}, "invalid character"),
	)
}
