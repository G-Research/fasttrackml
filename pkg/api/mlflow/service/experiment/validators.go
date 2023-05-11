package experiment

import (
	"net/url"
	"path/filepath"

	"github.com/G-Research/fasttrack/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrack/pkg/api/mlflow/api/request"
)

const (
	MaxResultsPerPage = 1000000
)

// AllowedViewTypeList supported list of ViewType.
var (
	AllowedViewTypeList = map[request.ViewType]struct{}{
		"":                          {},
		request.ViewTypeAll:         {},
		request.ViewTypeActiveOnly:  {},
		request.ViewTypeDeletedOnly: {},
	}
)

// ValidateCreateExperimentRequest validates `POST /mlflow/experiments/create` request.
func ValidateCreateExperimentRequest(req *request.CreateExperimentRequest) error {
	if req.Name == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'name'")
	}
	if req.ArtifactLocation != "" {
		u, err := url.Parse(req.ArtifactLocation)
		if err != nil {
			return api.NewInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
		}
		p, err := filepath.Abs(u.Path)
		if err != nil {
			return api.NewInvalidParameterValueError("Invalid value for parameter 'artifact_location': %s", err)
		}
		u.Path = p
		req.ArtifactLocation = u.String()
	}
	return nil
}

// ValidateUpdateExperimentRequest validates `POST /mlflow/experiments/update` request.
func ValidateUpdateExperimentRequest(req *request.UpdateExperimentRequest) error {
	if req.ID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	if req.Name == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'new_name'")
	}
	return nil
}

// ValidateGetExperimentByIDRequest validates `GET /mlflow/experiments/get` request.
func ValidateGetExperimentByIDRequest(req *request.GetExperimentRequest) error {
	if req.ID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}
	return nil
}

// ValidateGetExperimentByNameRequest validates `GET /mlflow/experiments/get` request.
func ValidateGetExperimentByNameRequest(req *request.GetExperimentRequest) error {
	if req.Name == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_name'")
	}
	return nil
}

// ValidateDeleteExperimentRequest validates `POST /mlflow/experiments/delete` request.
func ValidateDeleteExperimentRequest(req *request.DeleteExperimentRequest) error {
	if req.ID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}
	return nil
}

// ValidateRestoreExperimentRequest validates `POST /mlflow/experiments/restore` request.
func ValidateRestoreExperimentRequest(req *request.RestoreExperimentRequest) error {
	if req.ID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}
	return nil
}

// ValidateSearchExperimentsRequest validates `POST /mlflow/experiments/restore` request.
func ValidateSearchExperimentsRequest(req *request.SearchExperimentsRequest) error {
	if _, ok := AllowedViewTypeList[req.ViewType]; !ok {
		return api.NewInvalidParameterValueError("Invalid view_type '%s'", req.ViewType)
	}
	if req.MaxResults > MaxResultsPerPage {
		return api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied.")
	}
	return nil
}

// ValidateSetExperimentTagRequest validates `POST /mlflow/experiments/set-experiment-tag` request.
func ValidateSetExperimentTagRequest(req *request.SetExperimentTagRequest) error {
	if req.ID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'")
	}

	if req.Key == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'key'")
	}
	return nil
}
