package helpers

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/api/googleapi"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GSBucketStorageTestSuite is a test suite for Google Storage bucket storage.
type GSBucketStorageTestSuite struct {
	*BucketStorageTestSuite
	Client *storage.Client
}

// NewGSBucketStorageSuite creates a new Google Storage bucket storage test suite.
func NewGSBucketStorageSuite(endpoint string, testBuckets []string) (*GSBucketStorageTestSuite, error) {
	client, err := storage.NewClient(
		context.TODO(), option.WithEndpoint(endpoint), option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, err
	}
	return &GSBucketStorageTestSuite{
		BucketStorageTestSuite: NewBucketStorageTestSuite(&GSBucketStorageClient{client}, testBuckets),
		Client:                 client,
	}, nil
}

// GSBucketStorageClient implements BucketStorageClient for Google Storage.
type GSBucketStorageClient struct {
	*storage.Client
}

// CreateBuckets creates the tests buckets.
func (c *GSBucketStorageClient) CreateBuckets(buckets []string) error {
	for _, bucket := range buckets {
		if err := c.Bucket(bucket).Create(context.Background(), "", nil); err != nil {
			return err
		}
	}
	return nil
}

// DeleteBuckets deletes the test buckets.
func (c *GSBucketStorageClient) DeleteBuckets(buckets []string) error {
	for _, bucket := range buckets {
		it := c.Bucket(bucket).Objects(context.Background(), &storage.Query{})
		for {
			object, err := it.Next()
			if errors.Is(err, iterator.Done) || errors.Is(err, storage.ErrBucketNotExist) {
				break
			}
			if err := c.Bucket(bucket).Object(object.Name).Delete(context.Background()); err != nil {
				return err
			}
		}
		var e *googleapi.Error
		if err := c.Bucket(bucket).Delete(context.Background()); errors.As(err, &e) &&
			e.Code != http.StatusNotFound {
			return err
		}
	}
	return nil
}
