package ptr

// Deref returns value from pointer or default value.
func Deref[T any](v *T, or T) T {
	if v == nil {
		return or
	}
	return *v
}
