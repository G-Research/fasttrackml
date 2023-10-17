package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/rotisserie/eris"
	"google.golang.org/api/iterator"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// GSStorageName is a GCP storage name.
const (
	GSStorageName = "gs"
)

// GS represents GCP adapter to work with artifacts.
type GS struct {
	client *storage.Client
}

// NewGS creates new GC instance.
func NewGS(ctx context.Context, config *config.ServiceConfig) (*GS, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error creating GCP storage client")
	}
	return &GS{
		client: client,
	}, nil
}

// List implements ArtifactStorageProvider interface.
func (s GS) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	// 1. create s3 request input.
	bucket, rootPrefix, err := ExtractS3BucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}
	prefix := filepath.Join(rootPrefix, path)
	if prefix != "" {
		prefix = prefix + "/"
	}
	query := storage.Query{
		Prefix: prefix,
	}
	// 2. process search `path` parameter.

	// 3. read data from gcp storage.
	var artifactList []ArtifactObject
	it := s.client.Bucket(bucket).Objects(ctx, &query)
	for {
		object, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, eris.Wrap(err, "error getting object information")
		}

		relPath, err := filepath.Rel(rootPrefix, object.Name)
		if err != nil {
			return nil, eris.Wrapf(err, "error getting relative path for object: %s", object.Name)
		}
		artifactObject := ArtifactObject{
			Path:  relPath,
			Size:  object.Size,
			IsDir: false,
		}
		if object.Size == 0 {
			artifactObject.IsDir = true
		}
		artifactList = append(artifactList, artifactObject)
	}

	return artifactList, nil
}

// Get returns file content at the storage location.
func (s GS) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
	// 1. create s3 request input.
	bucketName, prefix, err := ExtractS3BucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}

	// 2. get object from gcp storage.
	reader, err := s.client.Bucket(bucketName).Object(filepath.Join(prefix, path)).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, eris.Wrap(fs.ErrNotExist, "object does not exist")
		}
		return nil, eris.Wrap(err, "error getting object")
	}

	return reader, nil
}
