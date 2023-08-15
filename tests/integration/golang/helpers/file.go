package helpers

import (
	"path/filepath"

	"github.com/rotisserie/eris"
)

// GetAbsolutePath returns absolute path for provided path.
func GetAbsolutePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", eris.Wrapf(err, "error getting absolute path for: %s", path)
	}
	return absolutePath, nil
}
