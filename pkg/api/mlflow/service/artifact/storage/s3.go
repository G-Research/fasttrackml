package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// S3 represents S3 adapter to work with artifacts.
type S3 struct {
	client *s3.Client
	config *config.ServiceConfig
}

// NewS3 creates new S3 instance.
func NewS3(config *config.ServiceConfig) (*S3, error) {
	storage := S3{
		config: config,
	}

	var clientOptions []func(o *s3.Options)
	var configOptions []func(*awsConfig.LoadOptions) error
	if config.S3EndpointURI != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.UsePathStyle = true
		})
		configOptions = append(configOptions, awsConfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == s3.ServiceID {
						return aws.Endpoint{
							URL:           config.S3EndpointURI,
							SigningRegion: region,
						}, nil
					}
					return aws.Endpoint{}, eris.Errorf("unknown endpoint requested for the service: %s", service)
				},
			),
		))
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), configOptions...)
	if err != nil {
		return nil, eris.Wrap(err, "error loading configuration for S3 client")
	}
	storage.client = s3.NewFromConfig(cfg, clientOptions...)
	return &storage, nil
}

// List implements Provider interface.
func (s S3) List(artifactURI, path string) (string, []ArtifactObject, error) {
	bucket, prefix, err := ExtractS3BucketAndPrefix(artifactURI)
	if err != nil {
		return "", nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}
	input := s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	// 1. process search `prefix` parameter.
	path, err = url.JoinPath(*input.Prefix, path)
	if err != nil {
		return "", nil, eris.Wrap(err, "error constructing s3 prefix")
	}
	input.Prefix = aws.String(path)

	paginator := s3.NewListObjectsV2Paginator(s.client, &input)
	if err != nil {
		return "", nil, eris.Wrap(err, "error creating s3 paginated request")
	}

	var artifactList []ArtifactObject
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return "", nil, eris.Wrap(err, "error getting s3 page objects")
		}
		log.Debugf("got %d objects from S3 storage for path: %s", len(page.Contents), path)
		for _, object := range page.Contents {
			artifactList = append(artifactList, ArtifactObject{
				Path:  *object.Key,
				Size:  object.Size,
				IsDir: false,
			})
		}
	}

	return fmt.Sprintf("s3://%s", bucket), artifactList, nil
}

// GetArtifact will return actual item in the storage location
func (s S3) GetArtifact(runArtifactURI, itemPath string) (io.ReadCloser, error) {
	bucketName, prefix, err := ExtractS3BucketAndPrefix(runArtifactURI)
	if err != nil {
		return nil, eris.Wrap(err, "error extracting bucket and prefix from provided uri")
	}

	// Create a GetObjectInput with the bucket name and object key
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filepath.Join(prefix, itemPath)),
	}

	// Fetch the object from S3
	resp, err := s.client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
