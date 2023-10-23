package helpers

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// CreateBuckets creates the test bucekts.
func CreateS3Buckets(s3Client *s3.Client, buckets []string) error {
	for _, bucket := range buckets {
		_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket '%s'", bucket)
		}
	}
	return nil
}

// RemoveBuckets removes the test buckets.
func RemoveS3Buckets(s3Client *s3.Client, buckets []string) error {
	for _, bucket := range buckets {
		if err := removeBucket(s3Client, bucket); err != nil {
			return eris.Wrapf(err, "failed to remove bucket '%s'", bucket)
		}
	}
	return nil
}

// removeBucket removes a bucket and its objects.
func removeBucket(s3Client *s3.Client, bucket string) error {
	// Delete all objects in the bucket
	var objectIDs []types.ObjectIdentifier
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return eris.Wrapf(err, "error paging objects in bucket '%s'", bucket)
		}
		for _, object := range page.Contents {
			objectIDs = append(objectIDs, types.ObjectIdentifier{Key: object.Key})
		}
	}
	if len(objectIDs) > 0 {
		_, err := s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{Objects: objectIDs},
		})
		if err != nil {
			return eris.Wrapf(err, "failed to delete objects in bucket '%s'", bucket)
		}
	}

	// Delete the bucket
	if _, err := s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return eris.Wrapf(err, "failed to delete bucket '%s'", bucket)
	}
	waiter := s3.NewBucketNotExistsWaiter(s3Client)
	if err := waiter.Wait(
		context.Background(),
		&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		},
		time.Second*10,
	); err != nil {
		return eris.Wrapf(err, "error waiting for bucket '%s' deletion", bucket)
	}
	return nil
}
