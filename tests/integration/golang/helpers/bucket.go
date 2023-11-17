package helpers

import (
	"github.com/stretchr/testify/require"
)

type BucketStorageClient interface {
	CreateBuckets([]string) error
	DeleteBuckets([]string) error
}

type BucketStorageTestSuite struct {
	BaseTestSuite
	client      BucketStorageClient
	testBuckets []string
}

func NewBucketStorageTestSuite(client BucketStorageClient, testBuckets []string) *BucketStorageTestSuite {
	return &BucketStorageTestSuite{
		client:      client,
		testBuckets: testBuckets,
	}
}

func (s *BucketStorageTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	require.Nil(s.T(), s.client.CreateBuckets(s.testBuckets))
}

func (s *BucketStorageTestSuite) TearDownTest() {
	s.BaseTestSuite.TearDownTest()
	require.Nil(s.T(), s.client.DeleteBuckets(s.testBuckets))
}
