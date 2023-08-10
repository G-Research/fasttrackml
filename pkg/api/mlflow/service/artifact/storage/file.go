package storage

import (
	"fmt"
	"net/url"
	"os"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// File represents S3 adapter to work with artifacts.
type File struct {
	config *config.ServiceConfig
}

// NewFile creates new File storage instance.
func NewFile(config *config.ServiceConfig) (*File, error) {
	return &File{
		config: config,
	}, nil
}

// List implements Provider interface.
func (s File) List(artifactURI, path, _ string) (string, string, []ArtifactObject, error) {
	// 1. process search `prefix` parameter.
	path, err := url.JoinPath(artifactURI, path)
	if err != nil {
		return "", "", nil, eris.Wrap(err, "error constructing s3 prefix")
	}

	// 2. read data from local storage.
	objects, err := os.ReadDir(path)
	if err != nil {
		return "", "", nil, eris.Wrapf(err, "error reading object from local storage")
	}

	log.Debugf("got %d objects from local storage for path: %s", len(objects), path)
	artifactList := make([]ArtifactObject, len(objects))
	for i, object := range objects {
		info, err := object.Info()
		if err != nil {
			return "", "", nil, eris.Wrapf(err, "error getting info for object: %s", object.Name())
		}
		artifactList[i] = ArtifactObject{
			Path:  fmt.Sprintf("%s/%s", path, info.Name()),
			Size:  info.Size(),
			IsDir: object.IsDir(),
		}
	}
	return "", s.config.ArtifactRoot, artifactList, nil
}
