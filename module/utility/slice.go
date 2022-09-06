// Package utility
package utility

// SliceUnique Удаление дублей из среза простых типов данных.
func SliceUnique[T uint | uintptr | uint8 | uint16 | uint32 | uint64 | int | int8 | int32 | int16 | int64 |
	float32 | float64 | string | complex64 | complex128](ss []T) (ret []T) {
	var (
		m map[T]bool
		s T
		n int
	)

	m = make(map[T]bool)
	for n = range ss {
		m[ss[n]] = true
	}
	ret = make([]T, 0, len(m))
	for s = range m {
		ret = append(ret, s)
	}

	return
}
