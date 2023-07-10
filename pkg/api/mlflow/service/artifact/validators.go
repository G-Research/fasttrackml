package artifact

import (
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// ValidateListArtifactsRequest validates `GET /mlflow/artifacts/list` request.
func ValidateListArtifactsRequest(req *request.ListArtifactsRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	if strings.Contains(req.Path, ".") || strings.Contains(req.Path, "..") {
		return api.NewInvalidParameterValueError("incorrect path has been provided. path has to be absolute")
	}
	return nil
}
