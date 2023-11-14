package common

import (
	"fmt"
	"io/fs"
	"regexp"
)

type singleFileFS struct {
	fs.FS
	path string
}

// NewSingleFileFS returns a new FS that only allows access to a single file.
func NewSingleFileFS(fs fs.FS, path string) fs.FS {
	return &singleFileFS{
		fs,
		path,
	}
}

// Open opens the named file.
func (f *singleFileFS) Open(name string) (fs.File, error) {
	return f.FS.Open(f.path)
}

type onlyRootFS struct {
	fs.FS
	path string
}

// NewOnlyRootFS returns a new FS that only allows access to the root directory.
func NewOnlyRootFS(fs fs.FS, path string) fs.FS {
	return &onlyRootFS{
		fs,
		path,
	}
}

// Open opens the named file.
func (f onlyRootFS) Open(name string) (fs.File, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.FS.Open(f.path)
}

// ErrorMessageForUI returns the error message of an error, rewritten for simplicity in the UI.
func ErrorMessageForUI(field, errMsg string) string {
	uniqueError := regexp.MustCompile("(?i)unique")
	validationError := regexp.MustCompile("INVALID_PARAMETER_VALUE")
	msg := []byte(errMsg)
	switch {
	case uniqueError.Match(msg):
		return fmt.Sprintf("The %s is already in use.", field)
	case validationError.Match(msg):
		return fmt.Sprintf("The %s is invalid.", field)
	default:
		return fmt.Sprintf("An unexepected error was encountered: %s", msg)
	}
}
