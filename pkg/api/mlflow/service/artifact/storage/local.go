package storage

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// LocalStorageName is a file storage name.
const (
	LocalStorageName = "file"
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
	// 1. trim the `file://` prefix if it exists.
	artifactURI = strings.TrimPrefix(artifactURI, "file://")

	// 2. process search `path` parameter.
	absPath, err := url.JoinPath(artifactURI, path)
	if err != nil {
		return "", nil, eris.Wrap(err, "error constructing full path")
	}

	// 3. read data from local storage.
	objects, err := os.ReadDir(absPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return artifactURI, []ArtifactObject{}, nil
		}
		return "", nil, eris.Wrapf(err, "error reading object from local storage")
	}

	log.Debugf("got %d objects from local storage for path %q", len(objects), absPath)
	artifactList := make([]ArtifactObject, len(objects))
	for i, object := range objects {
		info, err := object.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// file has been removed since we read the directory
				continue
			}
			return "", nil, eris.Wrapf(err, "error getting info for object: %s", object.Name())
		}
		artifactList[i] = ArtifactObject{
			Path:  filepath.Join(path, info.Name()),
			Size:  info.Size(),
			IsDir: object.IsDir(),
		}
	}

	return artifactURI, artifactList, nil
}
