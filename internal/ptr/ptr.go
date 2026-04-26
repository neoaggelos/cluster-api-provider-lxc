package ptr

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}

// Deref returns value from pointer or default value.
func Deref[T any](v *T, or T) T {
	if v == nil {
		return or
	}
	return *v
}
