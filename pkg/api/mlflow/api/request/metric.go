package request

// GetMetricHistoryRequest is a request object for `GET /mlflow/metrics/get-history` endpoint.
type GetMetricHistoryRequest struct {
	RunID     string            `query:"run_id"`
	RunUUID   string            `query:"run_uuid"`
	MetricKey string            `query:"metric_key"`
	Context   map[string]string `json:"context"`
}

// GetRunID returns Run RunID.
func (r GetMetricHistoryRequest) GetRunID() string {
	if r.RunID != "" {
		return r.RunID
	}
	return r.RunUUID
}

// GetMetricHistoryBulkRequest is a request object for `GET /mlflow/metrics/get-history-bulk` endpoint.
type GetMetricHistoryBulkRequest struct {
	RunIDs     []string `query:"run_id"`
	MetricKey  string   `query:"metric_key"`
	MaxResults int      `query:"max_results"`
}

// GetMetricHistoriesRequest is a request object for `POST /mlflow/metrics/get-histories` endpoint.
type GetMetricHistoriesRequest struct {
	ExperimentIDs []string          `json:"experiment_ids"`
	RunIDs        []string          `json:"run_ids"`
	MetricKeys    []string          `json:"metric_keys"`
	ViewType      ViewType          `json:"run_view_type"`
	MaxResults    int32             `json:"max_results"`
	Context       map[string]string `json:"context"`
}
