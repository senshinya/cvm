package codegen

import (
	"fmt"

	"shinya.click/cvm/entity"
)

type Error struct {
	Pos    entity.SourcePos
	Node   string
	Op     string
	Reason string
}

func (e *Error) Error() string {
	if e.Pos.Line > 0 {
		return fmt.Sprintf("%d:%d: codegen %s %s: %s", e.Pos.Line, e.Pos.Column, e.Node, e.Op, e.Reason)
	}
	return fmt.Sprintf("codegen %s %s: %s", e.Node, e.Op, e.Reason)
}
