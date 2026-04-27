package sema

import (
	"testing"

	"shinya.click/cvm/entity"
)

func TestSymbolFields(t *testing.T) {
	pos := entity.SourcePos{Line: 1, Column: 2}
	sym := &Symbol{
		Name:    "x",
		Kind:    SymVar,
		T:       NewTypeTable().Builtin(Int),
		Storage: StorageAuto,
		Linkage: LinkageNone,
		Pos:     pos,
	}
	if sym.Name != "x" || sym.Kind != SymVar {
		t.Fatalf("symbol fields not stored: %+v", sym)
	}
}
