package response

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/artifact/storage"
)

func TestNewListArtifactsResponse_Ok(t *testing.T) {
	response := NewListArtifactsResponse("rootUri", []storage.ArtifactObject{
		{
			Path:  "path1",
			Size:  1234567890,
			IsDir: false,
		},
		{
			Path:  "path2",
			Size:  0,
			IsDir: true,
		},
	})

	assert.Equal(t, &ListArtifactsResponse{
		Files: []FilePartialResponse{
			{
				Path:     "path1",
				IsDir:    false,
				FileSize: 1234567890,
			},
			{
				Path:     "path2",
				IsDir:    true,
				FileSize: 0,
			},
		},
		RootURI: "rootUri",
	}, response)
}
