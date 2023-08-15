package storage

import (
	"net/url"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// ArtifactObject represents Artifact object agnostic to selected storage.
type ArtifactObject struct {
	Path  string
	Size  int64 // artifact object size in bytes.
	IsDir bool
}

// GetPath returns Artifact Path.
func (o ArtifactObject) GetPath() string {
	return o.Path
}

// GetSize returns Artifact Size in bytes.
func (o ArtifactObject) GetSize() int64 {
	return o.Size
}

// IsDirectory show that object is directly or not.
func (o ArtifactObject) IsDirectory() bool {
	return o.IsDir
}

// Provider provides and interface to work with artifact storage.
type Provider interface {
	List(artifactURI, path, nextPageToken string) (string, string, []ArtifactObject, error)
}

// NewArtifactStorage creates new Artifact storage.
func NewArtifactStorage(config *config.ServiceConfig) (Provider, error) {
	if config.ArtifactRoot != "" {
		u, err := url.Parse(config.ArtifactRoot)
		if err != nil {
			return nil, eris.Wrap(err, "error parsing artifact root")
		}

		switch u.Scheme {
		case "s3":
			return NewS3(config)
		case "", "file":
			return NewLocal(config)
		default:
			return nil, eris.Errorf("unsupported `schema` has been provided: %s", u.Scheme)
		}
	}
	return NewNoop(), nil
}
