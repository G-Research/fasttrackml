package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// S3 represents S3 adapter to work with artifacts.
type S3 struct {
	bucket string
	client *s3.Client
	config *config.ServiceConfig
}

// NewS3 creates new S3 instance.
func NewS3(bucket string, config *config.ServiceConfig) (*S3, error) {
	storage := S3{
		bucket: bucket,
		config: config,
	}

	var clientOptions []func(o *s3.Options)
	var configOptions []func(*awsConfig.LoadOptions) error
	if config.S3EndpointURL != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.UsePathStyle = true
		})
		configOptions = append(configOptions, awsConfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == s3.ServiceID {
						return aws.Endpoint{
							URL:           config.S3EndpointURL,
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
func (s S3) List(artifactURI, path, nextPageToken string) (string, string, []ArtifactObject, error) {
	input := s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(ExtractS3Path(s.config.ArtifactRoot, artifactURI)),
	}
	if path != "" {
		// filter first `/` just to make sure that Prefix will be always correct.
		input.Prefix = aws.String(fmt.Sprintf("%s/%s", *input.Prefix, strings.TrimLeft(path, "/")))
	}
	if nextPageToken != "" {
		input.ContinuationToken = aws.String(nextPageToken)
	}

	output, err := s.client.ListObjectsV2(context.TODO(), &input)
	if err != nil {
		fmt.Println(err)
		return "", "", nil, eris.Wrap(err, "error getting s3 objects")
	}

	log.Debugf("got %d objects from S3 storage for path: %s", len(output.Contents), path)
	artifactList := make([]ArtifactObject, len(output.Contents))
	for i, object := range output.Contents {
		artifactList[i] = ArtifactObject{
			Path:  *object.Key,
			Size:  object.Size,
			IsDir: false,
		}
	}

	if output.NextContinuationToken != nil {
		return *output.NextContinuationToken, s.config.ArtifactRoot, artifactList, nil
	}

	return "", s.config.ArtifactRoot, artifactList, nil
}
