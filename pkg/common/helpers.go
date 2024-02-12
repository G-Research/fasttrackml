package common

// Constants to represent non-usual numbers.
const (
	NANValue            = "NaN"
	NANPositiveInfinity = "Infinity"
	NANNegativeInfinity = "-Infinity"
)

// Constants for experiment tags keys.
const (
	DescriptionTagKey = "mlflow.note.content"
)

// GetPointer returns pointer for provided string.
func GetPointer[T any](str T) *T {
	return &str
}
