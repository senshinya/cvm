package sema

type Type interface {
	isType()
	String() string
}

type BuiltinKind int

const (
	Void BuiltinKind = iota
	Bool
	Char
	SChar
	UChar
	Short
	UShort
	Int
	UInt
	Long
	ULong
	LongLong
	ULongLong
	Float
	Double
	LongDouble
	FloatComplex
	DoubleComplex
	LongDoubleComplex
)

var builtinNames = [...]string{
	Void:              "void",
	Bool:              "_Bool",
	Char:              "char",
	SChar:             "signed char",
	UChar:             "unsigned char",
	Short:             "short",
	UShort:            "unsigned short",
	Int:               "int",
	UInt:              "unsigned int",
	Long:              "long",
	ULong:             "unsigned long",
	LongLong:          "long long",
	ULongLong:         "unsigned long long",
	Float:             "float",
	Double:            "double",
	LongDouble:        "long double",
	FloatComplex:      "float _Complex",
	DoubleComplex:     "double _Complex",
	LongDoubleComplex: "long double _Complex",
}

type BuiltinType struct {
	Kind BuiltinKind
}

func (*BuiltinType) isType() {}

func (b *BuiltinType) String() string {
	return builtinNames[b.Kind]
}
