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

type PointerType struct {
	Pointee Type
}

func (*PointerType) isType() {}

func (p *PointerType) String() string {
	return p.Pointee.String() + "*"
}

type ArraySizeKind int

const (
	ArrayUnsized ArraySizeKind = iota
	ArrayConstantSize
	ArrayVLA
	ArrayStarSize
)

type ArrayType struct {
	Elem     Type
	Size     int64
	SizeExpr any
	SizeKind ArraySizeKind
}

func (*ArrayType) isType() {}

func (a *ArrayType) String() string {
	switch a.SizeKind {
	case ArrayConstantSize:
		return a.Elem.String() + "[" + itoa(a.Size) + "]"
	case ArrayUnsized:
		return a.Elem.String() + "[]"
	case ArrayVLA:
		return a.Elem.String() + "[<vla>]"
	case ArrayStarSize:
		return a.Elem.String() + "[*]"
	}
	return a.Elem.String() + "[?]"
}

func itoa(n int64) string {
	// 这里只需要一个很小的整数转字符串工具，避免为了单点用途引入额外 import。
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

type FunctionType struct {
	Ret      Type
	Params   []Type
	Variadic bool
	HasProto bool
}

func (*FunctionType) isType() {}

func (f *FunctionType) String() string {
	s := f.Ret.String() + " ("
	for i, p := range f.Params {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	if f.Variadic {
		if len(f.Params) > 0 {
			s += ", "
		}
		s += "..."
	}
	if !f.HasProto {
		s += "<noproto>"
	}
	s += ")"
	return s
}

type QualType struct {
	Base                      Type
	Const, Volatile, Restrict bool
}

func (*QualType) isType() {}

func (q *QualType) String() string {
	s := ""
	if q.Const {
		s += "const "
	}
	if q.Volatile {
		s += "volatile "
	}
	if q.Restrict {
		s += "restrict "
	}
	return s + q.Base.String()
}
