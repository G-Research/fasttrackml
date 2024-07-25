package storage

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/common/config"
)

// S3StorageName is a s3 storage name.
const (
	S3StorageName = "s3"
)

// S3 represents S3 adapter to work with artifacts.
type S3 struct {
	client *s3.Client
}

// NewS3 creates new S3 instance.
func NewS3(ctx context.Context, config *config.Config) (*S3, error) {
	var clientOptions []func(o *s3.Options)
	if config.S3EndpointURI != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.UsePathStyle = true
		})
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(config.S3EndpointURI)
		})
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error loading configuration for S3 client")
	}

	return &S3{
		s3.NewFromConfig(cfg, clientOptions...),
	}, nil
}

// List implements ArtifactStorageProvider interface.
func (s S3) List(ctx context.Context, artifactURI, path string) ([]ArtifactObject, error) {
	// 1. create s3 request input.
	bucket, rootPrefix, err := ExtractBucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}
	input := s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	}

	// 2. process search `path` parameter.
	prefix := filepath.Join(rootPrefix, path)
	if prefix != "" {
		prefix = prefix + "/"
	}
	input.Prefix = aws.String(prefix)

	// 3. read data from s3 storage.
	var artifactList []ArtifactObject
	paginator := s3.NewListObjectsV2Paginator(s.client, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, eris.Wrap(err, "error getting s3 page objects")
		}

		log.Debugf("got %d directories from S3 storage for bucket %q and prefix %q", len(page.CommonPrefixes), bucket, prefix)
		for _, dir := range page.CommonPrefixes {
			relPath, err := filepath.Rel(rootPrefix, *dir.Prefix)
			if err != nil {
				return nil, eris.Wrapf(err, "error getting relative path for dir: %s", *dir.Prefix)
			}
			artifactList = append(artifactList, ArtifactObject{
				Path:  relPath,
				Size:  0,
				IsDir: true,
			})
		}

		log.Debugf("got %d objects from S3 storage for bucket %q and prefix %q", len(page.Contents), bucket, prefix)
		for _, object := range page.Contents {
			relPath, err := filepath.Rel(rootPrefix, *object.Key)
			if err != nil {
				return nil, eris.Wrapf(err, "error getting relative path for object: %s", *object.Key)
			}
			artifactList = append(artifactList, ArtifactObject{
				Path:  relPath,
				Size:  *object.Size,
				IsDir: false,
			})
		}
	}

	return artifactList, nil
}

// Get returns file content at the storage location.
func (s S3) Get(ctx context.Context, artifactURI, path string) (io.ReadCloser, error) {
	// 1. create s3 request input.
	bucketName, prefix, err := ExtractBucketAndPrefix(artifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Join(prefix, path)),
	}

	// 2. get object from s3 storage.
	resp, err := s.client.GetObject(ctx, input)
	if err != nil {
		// errors.Is is not working for s3 errors, so we need to use errors.As instead.
		var s3NoSuchKey *types.NoSuchKey
		if errors.As(err, &s3NoSuchKey) {
			return nil, eris.Wrap(fs.ErrNotExist, "object does not exist")
		}
		return nil, eris.Wrap(err, "error getting object")
	}

	return resp.Body, nil
}
