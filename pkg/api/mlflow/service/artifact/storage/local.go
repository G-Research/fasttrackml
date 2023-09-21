package storage

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// Local represents local file storage adapter to work with artifacts.
type Local struct {
	config *config.ServiceConfig
}

// NewLocal creates new Local storage instance.
func NewLocal(config *config.ServiceConfig) (*Local, error) {
	return &Local{
		config: config,
	}, nil
}

// List implements ArtifactStorageProvider interface.
func (s Local) List(artifactURI, path string) (string, []ArtifactObject, error) {
	// 1. process search `prefix` parameter.
	path, err := url.JoinPath(artifactURI, path)
	if err != nil {
		return "", nil, eris.Wrap(err, "error constructing full path")
	}

	// 2. read data from local storage.
	objects, err := os.ReadDir(path)
	if err != nil {
		return "", nil, eris.Wrapf(err, "error reading object from local storage")
	}

	log.Debugf("got %d objects from local storage for path: %s", len(objects), path)
	artifactList := make([]ArtifactObject, len(objects))
	for i, object := range objects {
		info, err := object.Info()
		if err != nil {
			return "", nil, eris.Wrapf(err, "error getting info for object: %s", object.Name())
		}
		artifactList[i] = ArtifactObject{
			Path:  filepath.Join(path, info.Name()),
			Size:  info.Size(),
			IsDir: object.IsDir(),
		}
	}
	return s.config.DefaultArtifactRoot, artifactList, nil
}
