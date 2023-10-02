package helpers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rotisserie/eris"
)

// NewS3Client creates new instance of S3 client.
func NewS3Client(endpoint string) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithEndpointResolverWithOptions(
		aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, eris.Errorf("unknown endpoint requested for the service: %s", service)
			},
		),
	))
	if err != nil {
		return nil, eris.Wrap(err, "error loading configuration for S3 client")
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}
