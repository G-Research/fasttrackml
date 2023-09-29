package common

// GetPointer returns pointer for provided value.
func GetPointer[T any](str T) *T {
	return &str
}
