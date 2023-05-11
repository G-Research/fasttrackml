package run

import (
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

// ValidateUpdateRunRequest validates `POST /mlflow/runs/update` request.
func ValidateUpdateRunRequest(req *request.UpdateRunRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateGetRunRequest validates `GET /mlflow/runs/get` request.
func ValidateGetRunRequest(req *request.GetRunRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateDeleteRunRequest validates `POST /mlflow/runs/delete` request.
func ValidateDeleteRunRequest(req *request.DeleteRunRequest) error {
	if req.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateRestoreRunRequest validates `POST /mlflow/runs/restore` request.
func ValidateRestoreRunRequest(req *request.RestoreRunRequest) error {
	if req.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateLogMetricRequest validates `POST /mlflow/runs/log-metric` request.
func ValidateLogMetricRequest(req *request.LogMetricRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	if req.Key == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'key'")
	}

	if req.Timestamp == 0 {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'timestamp'")
	}
	return nil
}

// ValidateLogParamRequest validates `POST /mlflow/runs/log-parameter` request.
func ValidateLogParamRequest(req *request.LogParamRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	if req.Key == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'key'")
	}
	return nil
}

// ValidateSetRunTagRequest validates `POST /mlflow/runs/set-tag` request.
func ValidateSetRunTagRequest(req *request.SetRunTagRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}

	if req.Key == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'key'")
	}
	return nil
}

// ValidateDeleteRunTagRequest validates `POST /mlflow/runs/delete-tag` request.
func ValidateDeleteRunTagRequest(req *request.DeleteRunTagRequest) error {
	if req.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateLogBatchRequest validates `POST /mlflow/runs/log-batch` request.
func ValidateLogBatchRequest(req *request.LogBatchRequest) error {
	if req.RunID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	return nil
}

// ValidateSearchRunsRequest validates `POST /mlflow/runs/search` request.
func ValidateSearchRunsRequest(req *request.SearchRunsRequest) error {
	if _, ok := AllowedViewTypeList[req.ViewType]; !ok {
		return api.NewInvalidParameterValueError("Invalid run_view_type '%s'", req.ViewType)
	}
	if req.MaxResults > MaxResultsPerPage {
		return api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied.")
	}
	return nil
}
