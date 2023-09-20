package storage

import (
	"io"
	"os"
)

// Noop represents empty storage adapter.
type Noop struct{}

// NewNoop creates new Noop instance.
func NewNoop() *Noop {
	return &Noop{}
}

// List implements Provider interface.
func (s Noop) List(_, _ string) (string, []ArtifactObject, error) {
	return "", make([]ArtifactObject, 0), nil
}

// GetArtifact implements Provider interface.
func (s Noop) GetArtifact(_, _ string) (io.Reader, error) {
	return &os.File{}, nil
}
