package helpers

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/api/googleapi"

	"cloud.google.com/go/storage"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type BaseArtifactGSTestSuite struct {
	BaseTestSuite
	GsClient    *storage.Client
	TestBuckets []string
}

func NewBaseArtifactGSTestSuite(testBuckets ...string) BaseArtifactGSTestSuite {
	return BaseArtifactGSTestSuite{
		TestBuckets: testBuckets,
	}
}

func (s *BaseArtifactGSTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	gsClient, err := NewGSClient(GetGSEndpointUri())
	require.Nil(s.T(), err)
	s.GsClient = gsClient
}

func (s *BaseArtifactGSTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	require.Nil(s.T(), CreateGSBuckets(s.GsClient, s.TestBuckets))
}

func (s *BaseArtifactGSTestSuite) TearDownTest() {
	require.Nil(s.T(), DeleteGSBuckets(s.GsClient, s.TestBuckets))
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

// CreateGSBuckets creates tests buckets in the Google Storage.
func CreateGSBuckets(client *storage.Client, buckets []string) error {
	for _, bucket := range buckets {
		if err := client.Bucket(bucket).Create(context.Background(), "", nil); err != nil {
			return err
		}
	}
	return nil
}

// DeleteGSBuckets deletes tests buckets from the Google Storage.
func DeleteGSBuckets(client *storage.Client, buckets []string) error {
	for _, bucket := range buckets {
		it := client.Bucket(bucket).Objects(context.Background(), &storage.Query{})
		for {
			object, err := it.Next()
			if errors.Is(err, iterator.Done) || errors.Is(err, storage.ErrBucketNotExist) {
				break
			}
			if err := client.Bucket(bucket).Object(object.Name).Delete(context.Background()); err != nil {
				return err
			}
		}
		var e *googleapi.Error
		if err := client.Bucket(bucket).Delete(context.Background()); errors.As(err, &e) &&
			e.Code != http.StatusNotFound {
			return err
		}
	}
	return nil
}
