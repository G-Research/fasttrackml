package metric

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

const (
	MaxResultsPerPage = 1000000000
)

// ValidateGetMetricHistoryRequest validates `GET /mlflow/metrics/get-history` request.
func ValidateGetMetricHistoryRequest(req *request.GetMetricHistoryRequest) error {
	if req.RunID == "" && req.RunUUID == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'")
	}
	if req.MetricKey == "" {
		return api.NewInvalidParameterValueError("Missing value for required parameter 'metric_key'")
	}
	return nil
}

// ValidateGetMetricHistoryBulkRequest validates `GET /mlflow/metrics/get-history-bulk` request.
func ValidateGetMetricHistoryBulkRequest(req *request.GetMetricHistoryBulkRequest) error {
	if len(req.RunIDs) == 0 {
		return api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify at least one run_id.")
	}

	if len(req.RunIDs) > 200 {
		return api.NewInvalidParameterValueError(
			"GetMetricHistoryBulk request cannot specify more than 200 run_ids. Received %d run_ids.", len(req.RunIDs),
		)
	}

	if req.MetricKey == "" {
		return api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify a metric_key.")
	}

	// TODO:DSuhinin - I don't like the idea of that but let's keep it for now.
	if req.MaxResults == 0 {
		req.MaxResults = 25000
	}
	return nil
}

// ValidateGetMetricHistoriesRequest validates `GET /mlflow/metrics/get-histories` request.
func ValidateGetMetricHistoriesRequest(req *request.GetMetricHistoriesRequest) error {
	if len(req.ExperimentIDs) > 0 && len(req.RunIDs) > 0 {
		return api.NewInvalidParameterValueError(
			"experiment_ids and run_ids cannot both be specified at the same time",
		)
	}
	if req.MaxResults > MaxResultsPerPage {
		return api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied.")
	}
	return nil
}
