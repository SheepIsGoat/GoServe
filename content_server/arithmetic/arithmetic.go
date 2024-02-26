package arithmetic

func Add[T Numeric](a T, b T) T {
	return a + b
}
func Sub[T Numeric](a T, b T) T {
	return a - b
}

type Ordered interface {
	Numeric | ~string
}

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
