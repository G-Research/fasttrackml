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

type S3TestSuite struct {
	BaseTestSuite
	Client      *s3.Client
	testBuckets []string
}

// NewS3TestSuite creates a new instance of S3TestSuite.
func NewS3TestSuite(testBuckets ...string) S3TestSuite {
	return S3TestSuite{
		testBuckets: testBuckets,
	}
}

func (s *S3TestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	client, err := NewS3Client(GetS3EndpointUri())
	s.Require().Nil(err)
	s.Client = client

	s.AddSetupHook(func() {
		s.Require().Nil(s.CreateTestBuckets())
	})
	s.AddTearDownHook(func() {
		s.Require().Nil(s.DeleteTestBuckets())
	})
}

// NewS3Client creates a new instance of S3 client.
func NewS3Client(endpoint string) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, eris.Wrap(err, "error loading configuration for S3 client")
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	}), nil
}

// CreateTestBuckets creates the test buckets.
func (s *S3TestSuite) CreateTestBuckets() error {
	for _, bucket := range s.testBuckets {
		_, err := s.Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket %q", bucket)
		}
	}
	return nil
}

// DeleteTestBuckets deletes the test buckets.
func (s *S3TestSuite) DeleteTestBuckets() error {
	for _, bucket := range s.testBuckets {
		if err := s.deleteBucket(bucket); err != nil {
			return eris.Wrapf(err, "failed to delete bucket %q", bucket)
		}
	}
	return nil
}

// deleteBucket deletes a bucket and its objects.
func (s *S3TestSuite) deleteBucket(bucket string) error {
	// Delete all objects in the bucket
	var objectIDs []types.ObjectIdentifier
	paginator := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return eris.Wrapf(err, "failed to list objects in bucket %q", bucket)
		}
		for _, object := range page.Contents {
			objectIDs = append(objectIDs, types.ObjectIdentifier{Key: object.Key})
		}
	}
	if len(objectIDs) > 0 {
		_, err := s.Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{Objects: objectIDs},
		})
		if err != nil {
			return eris.Wrapf(err, "failed to delete objects in bucket %q", bucket)
		}
	}

	// Delete the bucket
	if _, err := s.Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return eris.Wrapf(err, "failed to delete bucket %q", bucket)
	}
	waiter := s3.NewBucketNotExistsWaiter(s.Client)
	if err := waiter.Wait(
		context.Background(),
		&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		},
		time.Second*10,
	); err != nil {
		return eris.Wrapf(err, "failed to wait for bucket %q deletion", bucket)
	}
	return nil
}
