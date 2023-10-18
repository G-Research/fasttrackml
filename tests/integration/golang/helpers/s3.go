package helpers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rotisserie/eris"
)

var testBuckets = []string{"bucket1", "bucket2", "bucket3"}

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

func CreateBuckets(s3Client *s3.Client) error {
	for _, bucket := range testBuckets {
		_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket:                    aws.String(bucket),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{},
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket '%s': %v", bucket, err)
		}
	}
	return nil
}

func RemoveBuckets(s3Client *s3.Client) error {
	for _, bucket := range testBuckets {
		// Delete all objects in the bucket
		listObjectsOutput, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket '%s': %v", bucket, err)
		}
		var objectIds []types.ObjectIdentifier
		for _, object := range listObjectsOutput.Contents {
			objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(*object.Key)})
		}

		if len(objectIds) > 0 {
			_, err = s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket),
				Delete: &types.Delete{Objects: objectIds},
			})
			if err != nil {
				return fmt.Errorf("failed to delete objects in bucket '%s': %v", bucket, err)
			}
		}

		// Delete the bucket
		_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return fmt.Errorf("failed to delete bucket: %v", err)
		}
	}
	return nil
}
