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

// ArtifactStorageProvider provides and interface to work with particular artifact storage.
type ArtifactStorageProvider interface {
	// List lists all artifact object under provided path.
	List(artifactURI, path string) (string, []ArtifactObject, error)
}

// ArtifactStorageFactoryProvider provides an interface to work with Artifact Storage.
type ArtifactStorageFactoryProvider interface {
	// GetStorage returns Artifact storage based on provided runArtifactPath.
	GetStorage(runArtifactPath string) (ArtifactStorageProvider, error)
}

// ArtifactStorageFactory represents Artifact Storage .
type ArtifactStorageFactory struct {
	storageList map[string]ArtifactStorageProvider
}

// NewArtifactStorageFactory creates new Artifact Storage Factory instance.
func NewArtifactStorageFactory(config *config.ServiceConfig) (*ArtifactStorageFactory, error) {
	s3Storage, err := NewS3(config)
	if err != nil {
		return nil, eris.Wrap(err, "error initializing s3 artifact storage")
	}
	localStorage, err := NewLocal(config)
	if err != nil {
		return nil, eris.Wrap(err, "error initializing local artifact storage")
	}
	return &ArtifactStorageFactory{
		storageList: map[string]ArtifactStorageProvider{
			S3StorageName:    s3Storage,
			LocalStorageName: localStorage,
		},
	}, nil
}

// GetStorage returns Artifact storage based on provided runArtifactPath.
func (s ArtifactStorageFactory) GetStorage(runArtifactPath string) (ArtifactStorageProvider, error) {
	u, err := url.Parse(runArtifactPath)
	if err != nil {
		return nil, eris.Wrap(err, "error parsing artifact root")
	}
	switch u.Scheme {
	case S3StorageName:
		return s.storageList[S3StorageName], nil
	case "", LocalStorageName:
		return s.storageList[LocalStorageName], nil
	default:
		return nil, eris.Errorf("unsupported schema has been provided: %s", u.Scheme)
	}
}
