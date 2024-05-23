package syntax

type FunctionDef struct {
}

func (d *FunctionDef) GetUnitType() TranslationUnitType {
	return UnitTypeFunctionDef
}
