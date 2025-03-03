package common

import (
	deepcopyraw "github.com/mohae/deepcopy"
)

func DeepCopy[T any](v T) T {
	return deepcopyraw.Copy(v).(T)
}
