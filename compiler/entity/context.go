package entity

type SemanticsContext struct {
	ConstantPool ConstantPool

	CurrentScope *Scope
}
