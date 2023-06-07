package common

// GetPointer returns pointer for provided string.
func GetPointer[T any](str T) *T {
	return &str
}
