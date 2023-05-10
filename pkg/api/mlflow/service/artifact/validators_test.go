package artifact

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
)

func TestValidateListArtifactsRequestRequest_Ok(t *testing.T) {
	err := ValidateListArtifactsRequestRequest(&request.ListArtifactsRequest{
		RunID: "run_id",
	})
	assert.Nil(t, err)
}
func TestValidateListArtifactsRequestRequest_Error(t *testing.T) {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.ListArtifactsRequest
	}{
		{
			name:    "EmptyRunIDProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.ListArtifactsRequest{},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListArtifactsRequestRequest(tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
