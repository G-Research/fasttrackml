package artifact

import (
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

// ValidateListArtifactsRequest validates `GET /mlflow/artifacts/list` request.
func ValidateListArtifactsRequest(req *request.ListArtifactsRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	return validatePath(req.Path)
}

// ValidateGetArtifactRequest validates `GET /artifacts/get` request.
func ValidateGetArtifactRequest(req *request.GetArtifactRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	return validatePath(req.Path)
}

// validatePath validates path parameter.
func validatePath(path string) error {
	parsedUrl, err := url.Parse(path)
	if err != nil ||
		parsedUrl.Scheme != "" ||
		parsedUrl.Host != "" ||
		parsedUrl.RawQuery != "" ||
		parsedUrl.RawFragment != "" ||
		parsedUrl.User != nil ||
		filepath.IsAbs(parsedUrl.Path) ||
		slices.Contains(strings.Split(parsedUrl.Path, "/"), "..") {
		return api.NewInvalidParameterValueError("Invalid path")
	}
	return nil
}
