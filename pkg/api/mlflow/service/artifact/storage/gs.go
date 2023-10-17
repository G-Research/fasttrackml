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

// GCStorageName is a GCP storage name.
const (
	GCStorageName = "gc"
)

// GC represents GCP adapter to work with artifacts.
type GC struct {
	client *storage.Client
}

// NewGC creates new GC instance.
func NewGC(ctx context.Context, config *config.ServiceConfig) (*GC, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error creating GCP storage client")
	}
	return &GC{
		client: client,
	}, nil
}

// List implements ArtifactStorageProvider interface.
func (s GC) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	// 1. create s3 request input.
	bucket, rootPrefix, err := ExtractS3BucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}
	query := storage.Query{
		Delimiter: "/",
	}

	// 2. process search `path` parameter.
	prefix := filepath.Join(rootPrefix, path)
	if prefix != "" {
		query.Prefix = prefix + "/"
	}
	query.Prefix = prefix

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
func (s GC) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
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
