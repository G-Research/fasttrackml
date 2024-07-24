package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/config"
)

// LocalStorageName is a file storage name.
const (
	LocalStorageName = "file"
)

// Local represents local file storage adapter to work with artifacts.
type Local struct{}

// NewLocal creates new Local storage instance.
func NewLocal(config *config.Config) (*Local, error) {
	return &Local{}, nil
}

// List implements ArtifactStorageProvider interface.
func (s Local) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	// 1. trim the `file://` prefix if it exists.
	artifactURI = strings.TrimPrefix(artifactURI, "file://")

	// 2. process search `path` parameter.
	absPath := filepath.Join(artifactURI, path)

	// 3. read data from local storage.
	objects, err := os.ReadDir(absPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []ArtifactObject{}, nil
		}
		return nil, eris.Wrapf(err, "error reading object from local storage")
	}

	log.Debugf("got %d objects from local storage for path %q", len(objects), absPath)
	artifactList := make([]ArtifactObject, len(objects))
	for i, object := range objects {
		info, err := object.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// the file has been removed since we read the directory
				continue
			}
			return nil, eris.Wrapf(err, "error getting info for object: %s", object.Name())
		}
		artifactList[i] = ArtifactObject{
			Path:  filepath.Join(path, info.Name()),
			IsDir: object.IsDir(),
		}
		if !object.IsDir() {
			artifactList[i].Size = info.Size()
		}
	}

	return artifactList, nil
}

// Get returns actual file content at the storage location.
func (s Local) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
	// 1. trim the `file://` prefix if it exists.
	artifactURI = strings.TrimPrefix(artifactURI, "file://")

	// 2. process `path` parameter.
	absPath := filepath.Join(artifactURI, path)

	// 3. checks that the file exists and is not a directory.
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, eris.Wrap(err, "path could not be opened")
	}
	if fileInfo.IsDir() {
		return nil, eris.Wrap(fs.ErrNotExist, "path is a directory")
	}

	// 4. open the file.
	// artifactURI and path are validated by the caller
	// #nosec G304
	file, err := os.Open(absPath)
	if err != nil {
		return nil, eris.Wrap(err, "unable to open file")
	}

	return file, nil
}
