package storage

import (
	"context"
	"io"
	"net/url"
	"sync"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/common/config"
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

// ArtifactStorageProvider provides an interface to work with artifact storage.
type ArtifactStorageProvider interface {
	// Get returns an io.ReadCloser for specific artifact.
	Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error)
	// List lists all artifact objects under a provided path.
	List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error)
}

// ArtifactStorageFactoryProvider provides an interface provider to work with Artifact Storage.
type ArtifactStorageFactoryProvider interface {
	// GetStorage returns Artifact storage based on provided runArtifactPath.
	GetStorage(ctx context.Context, runArtifactPath string) (ArtifactStorageProvider, error)
}

// ArtifactStorageFactory represents Artifact Storage.
type ArtifactStorageFactory struct {
	config      *config.Config
	storageList sync.Map
}

// NewArtifactStorageFactory creates new Artifact Storage Factory instance.
func NewArtifactStorageFactory(config *config.Config) (*ArtifactStorageFactory, error) {
	return &ArtifactStorageFactory{
		config:      config,
		storageList: sync.Map{},
	}, nil
}

// GetStorage returns Artifact storage based on provided runArtifactPath.
func (s *ArtifactStorageFactory) GetStorage(
	ctx context.Context,
	runArtifactPath string,
) (ArtifactStorageProvider, error) {
	u, err := url.Parse(runArtifactPath)
	if err != nil {
		return nil, eris.Wrap(err, "error parsing artifact root")
	}

	storageName := u.Scheme
	if storage, ok := s.storageList.Load(storageName); ok {
		return storage.(ArtifactStorageProvider), nil
	}

	var storage ArtifactStorageProvider
	switch storageName {
	case GSStorageName:
		var err error
		storage, err = NewGS(ctx, s.config)
		if err != nil {
			return nil, eris.Wrap(err, "error initializing gs artifact storage")
		}
	case S3StorageName:
		var err error
		storage, err = NewS3(ctx, s.config)
		if err != nil {
			return nil, eris.Wrap(err, "error initializing s3 artifact storage")
		}
	case "", LocalStorageName:
		var err error
		storage, err = NewLocal(s.config)
		if err != nil {
			return nil, eris.Wrap(err, "error initializing local artifact storage")
		}
	default:
		return nil, eris.Errorf("unsupported schema has been provided: %s", u.Scheme)
	}

	s.storageList.Store(storageName, storage)
	return storage, nil
}
