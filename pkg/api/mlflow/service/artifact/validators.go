package artifact

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// ValidateListArtifactsRequest validates `GET /mlflow/artifacts/list` request.
func ValidateListArtifactsRequest(req *request.ListArtifactsRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	parsedUrl, err := url.Parse(req.Path)
	if err != nil {
		return api.NewInvalidParameterValueError("error parsing 'path' parameter")
	}
	if parsedUrl.Scheme != "" || parsedUrl.Host != "" || parsedUrl.RawQuery != "" ||
		parsedUrl.RawFragment != "" || parsedUrl.User != nil {
		return api.NewInvalidParameterValueError("provided 'path' parameter is invalid")
	}

	for _, path := range strings.Split(parsedUrl.Path, "/") {
		if path == ".." || filepath.IsAbs(path) {
			return api.NewInvalidParameterValueError("provided 'path' parameter is invalid")
		}
	}
	return nil
}
