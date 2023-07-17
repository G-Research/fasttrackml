package artifact

import (
	"regexp"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// RelativePathRegExp checks for the sequence `./` or `../` in provided path.
var RelativePathRegExp = regexp.MustCompile(`\.{1,2}\/`)

// ValidateListArtifactsRequest validates `GET /mlflow/artifacts/list` request.
func ValidateListArtifactsRequest(req *request.ListArtifactsRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	if RelativePathRegExp.MatchString(req.Path) {
		return api.NewInvalidParameterValueError("incorrect path has been provided. path has to be absolute")
	}
	return nil
}
