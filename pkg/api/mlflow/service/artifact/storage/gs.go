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
	"google.golang.org/api/option"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// GSStorageName is a GS storage name.
const (
	GSStorageName = "gs"
)

// GS represents adapter to work with GS storage artifacts.
type GS struct {
	client *storage.Client
}

// NewGS creates new GC instance.
func NewGS(ctx context.Context, config *config.ServiceConfig) (*GS, error) {
	var options []option.ClientOption
	if config.GSEndpointURI != "" {
		// include option.WithoutAuthentication() option, because otherwise standard GCP DSK won't work properly.
		// make it configurable via ENV if it's really needed.
		options = append(options, option.WithEndpoint(config.GSEndpointURI), option.WithoutAuthentication())
	}
	client, err := storage.NewClient(ctx, options...)
	if err != nil {
		return nil, eris.Wrap(err, "error creating GS storage client")
	}
	return &GS{
		client: client,
	}, nil
}

// List implements ArtifactStorageProvider interface.
func (s GS) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	// 1. process input parameters.
	bucket, rootPrefix, err := ExtractBucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}
	prefix := filepath.Join(rootPrefix, path)
	if prefix != "" {
		prefix = prefix + "/"
	}

	// 2. read data from gs storage.
	var artifactList []ArtifactObject
	it := s.client.Bucket(bucket).Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})
	for {
		object, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, eris.Wrap(err, "error getting object information")
		}

		objectName := object.Name
		if object.Name == "" {
			objectName = object.Prefix
		}

		relPath, err := filepath.Rel(rootPrefix, objectName)
		if err != nil {
			return nil, eris.Wrapf(err, "error getting relative path for object: %s", object.Name)
		}

		// filter current directory from the result set.
		if relPath == path {
			continue
		}
		artifactList = append(artifactList, ArtifactObject{
			Path:  relPath,
			Size:  object.Size,
			IsDir: object.Size == 0,
		})
	}

	return artifactList, nil
}

// Get returns file content at the storage location.
func (s GS) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
	// 1. create s3 request input.
	bucketName, prefix, err := ExtractBucketAndPrefix(artifactURI)
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
