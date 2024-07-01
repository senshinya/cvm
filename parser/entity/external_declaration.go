package entity

type ExternalDeclarationType uint8

const (
	ExternalDeclarationTypeDeclaration ExternalDeclarationType = 1
	ExternalDeclarationTypeFunctionDef ExternalDeclarationType = 2
)

type ExternalDeclaration interface {
	GetExternalDeclarationType() ExternalDeclarationType
}
