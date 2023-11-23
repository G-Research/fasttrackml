package helpers

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/api/googleapi"

	"cloud.google.com/go/storage"
	"github.com/rotisserie/eris"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GSTestSuite struct {
	BaseTestSuite
	Client      *storage.Client
	testBuckets []string
}

// NewGSTestSuite creates a new instance of GSTestSuite.
func NewGSTestSuite(testBuckets ...string) GSTestSuite {
	return GSTestSuite{
		testBuckets: testBuckets,
	}
}

func (s *GSTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	client, err := NewGSClient(GetGSEndpointUri())
	s.Require().Nil(err)
	s.Client = client

	s.AddSetupHook(func() {
		s.Require().Nil(s.CreateTestBuckets())
	})
	s.AddTearDownHook(func() {
		s.Require().Nil(s.DeleteTestBuckets())
	})
}

// NewGSClient creates new instance of Google Storage client.
func NewGSClient(endpoint string) (*storage.Client, error) {
	client, err := storage.NewClient(
		context.TODO(), option.WithEndpoint(endpoint), option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, eris.Wrap(err, "error creating GS client")
	}

	return client, nil
}

// CreateTestBuckets creates the test buckets.
func (s *GSTestSuite) CreateTestBuckets() error {
	for _, bucket := range s.testBuckets {
		if err := s.Client.Bucket(bucket).Create(context.Background(), "", nil); err != nil {
			return eris.Wrapf(err, "failed to create bucket %q", bucket)
		}
	}
	return nil
}

// DeleteTestBuckets deletes the test buckets.
func (s *GSTestSuite) DeleteTestBuckets() error {
	for _, bucket := range s.testBuckets {
		it := s.Client.Bucket(bucket).Objects(context.Background(), &storage.Query{})
		for {
			object, err := it.Next()
			if err != nil {
				if errors.Is(err, iterator.Done) || errors.Is(err, storage.ErrBucketNotExist) {
					break
				}
				return eris.Wrapf(err, "failed to list objects in bucket %q", bucket)
			}
			if err := s.Client.Bucket(bucket).Object(object.Name).Delete(context.Background()); err != nil {
				return eris.Wrapf(err, "failed to delete objects in bucket %q", bucket)
			}
		}
		var e *googleapi.Error
		if err := s.Client.Bucket(bucket).Delete(context.Background()); errors.As(err, &e) &&
			e.Code != http.StatusNotFound {
			return eris.Wrapf(err, "failed to delete bucket %q", bucket)
		}
	}
	return nil
}
