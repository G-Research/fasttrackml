package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractS3BucketAndPrefix_Ok(t *testing.T) {
	uri := "s3://fasttrackml/2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts"
	bucket, prefix, err := ExtractBucketAndPrefix(uri)
	require.Nil(t, err)
	assert.Equal(t, "fasttrackml", bucket)
	assert.Equal(t, "2/30357ed2eaac4f2cacdbcd0e06e9e48a/artifacts", prefix)
}
