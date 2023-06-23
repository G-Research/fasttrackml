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
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
	}{
		{
			name:    "EmptyRunIDAndRunUUID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.ListArtifactsRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListArtifactsRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
