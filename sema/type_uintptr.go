package sema

import "unsafe"

func uintptrOf(t Type) uintptr {
	type iface struct {
		_   *struct{}
		ptr unsafe.Pointer
	}
	i := *(*iface)(unsafe.Pointer(&t))
	return uintptr(i.ptr)
}
