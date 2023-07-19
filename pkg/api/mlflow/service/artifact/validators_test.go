package artifact

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestValidateListArtifactsRequest_Ok(t *testing.T) {
	err := ValidateListArtifactsRequest(&request.ListArtifactsRequest{
		RunID: "run_id",
	})
	assert.Nil(t, err)
}

func TestValidateListArtifactsRequest_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.ListArtifactsRequest{},
		},
		{
			name:  "1DotRelativePathProvided",
			error: api.NewInvalidParameterValueError("provided 'path' parameter has to be absolute"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "./",
			},
		},
		{
			name:  "2DotsRelativePathProvided",
			error: api.NewInvalidParameterValueError("provided 'path' parameter has to be absolute"),
			request: &request.ListArtifactsRequest{
				RunID: "run_id",
				Path:  "../",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListArtifactsRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
