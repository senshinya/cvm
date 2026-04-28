package preprocessor

import (
	"fmt"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

func ppError(pos entity.SourcePos, format string, args ...any) *common.CvmError {
	return common.NewCvmError(common.NewErrorMessage(pos, fmt.Sprintf(format, args...)))
}
