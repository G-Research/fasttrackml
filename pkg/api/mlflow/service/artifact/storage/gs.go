package storage

import (
	"context"
	"io"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// GCStorageName is a GCP storage name.
const (
	GCStorageName = "gc"
)

// GC represents GCP adapter to work with artifacts.
type GC struct{}

// NewGC creates new GC instance.
func NewGC(ctx context.Context, config *config.ServiceConfig) (*GC, error) {
	return nil, nil
}

// List implements ArtifactStorageProvider interface.
func (s GC) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	return nil, nil
}

// Get returns file content at the storage location.
func (s GC) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
	return nil, nil
}
