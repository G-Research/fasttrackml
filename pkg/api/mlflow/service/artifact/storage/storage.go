package storage

import (
	"io"
	"net/url"
	"sync"

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

// ArtifactStorageProvider provides and interface to work with artifact storage.
type ArtifactStorageProvider interface {
	List(runArtifactURI, path string) ([]ArtifactObject, error)
	GetArtifact(runArtifactURI, path string) (io.ReadCloser, error)
}

// ArtifactStorageFactoryProvider provides an interface to work with Artifact Storage.
type ArtifactStorageFactoryProvider interface {
	// GetStorage returns Artifact storage based on provided runArtifactPath.
	GetStorage(runArtifactPath string) (ArtifactStorageProvider, error)
}

// ArtifactStorageFactory represents Artifact Storage .
type ArtifactStorageFactory struct {
	config      *config.ServiceConfig
	storageList sync.Map
}

// NewArtifactStorageFactory creates new Artifact Storage Factory instance.
func NewArtifactStorageFactory(config *config.ServiceConfig) (*ArtifactStorageFactory, error) {
	return &ArtifactStorageFactory{
		config:      config,
		storageList: sync.Map{},
	}, nil
}

// GetStorage returns Artifact storage based on provided runArtifactPath.
func (s *ArtifactStorageFactory) GetStorage(runArtifactPath string) (ArtifactStorageProvider, error) {
	u, err := url.Parse(runArtifactPath)
	if err != nil {
		return nil, eris.Wrap(err, "error parsing artifact root")
	}

	switch u.Scheme {
	case S3StorageName:
		if storage, ok := s.storageList.Load(S3StorageName); ok {
			if localStorage, ok := storage.(*S3); ok {
				return localStorage, nil
			}
			return nil, eris.New("storage is not s3 artifact storage")
		}
		storage, err := NewS3(s.config)
		if err != nil {
			return nil, eris.Wrap(err, "error initializing s3 artifact storage")
		}
		s.storageList.Store(S3StorageName, storage)
		return storage, nil
	case "", LocalStorageName:
		if storage, ok := s.storageList.Load(LocalStorageName); ok {
			if localStorage, ok := storage.(*Local); ok {
				return localStorage, nil
			}
			return nil, eris.New("storage is not local artifact storage")
		}
		storage, err := NewLocal(s.config)
		if err != nil {
			return nil, eris.Wrap(err, "error initializing local artifact storage")
		}
		s.storageList.Store(LocalStorageName, storage)
		return storage, nil
	default:
		return nil, eris.Errorf("unsupported schema has been provided: %s", u.Scheme)
	}
}
