package artifact

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

// RelativePathRegExp checks for the `../` in provided path.
var RelativePathRegExp = regexp.MustCompile(`\^\.{2}$`)

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
		return api.NewInvalidParameterValueError("incorrect 'path' parameter has been provided")
	}

	for _, path := range strings.Split(parsedUrl.Path, "/") {
		if RelativePathRegExp.MatchString(path) {
			return api.NewInvalidParameterValueError("provided 'path' parameter has to be absolute")
		}
	}
	return nil
}
