package pkgptr

func Ptr[T any](v T) *T {
	return &v
}

func Value[T any](v *T) T {
	var zero T
	if v == nil {
		return zero
	}

	return *v
}
