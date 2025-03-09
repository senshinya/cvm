package common

import (
	deepcopyraw "github.com/mohae/deepcopy"
)

func DeepCopy[T any](v T) T {
	return deepcopyraw.Copy(v).(T)
}

func IfElse[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}
