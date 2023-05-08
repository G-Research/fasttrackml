package artifact

import (
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
)

// ValidateListArtifactsRequestRequest validates `GET /mlflow/artifacts/list` request.
func ValidateListArtifactsRequestRequest(req *request.ListArtifactsRequest) error {
	if req.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}
