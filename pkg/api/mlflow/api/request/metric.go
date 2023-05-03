package request

// GetMetricHistoriesRequest is a request object for `POST mlflow/get-histories` endpoint.
type GetMetricHistoriesRequest struct {
	ExperimentIDs []string `json:"experiment_ids"`
	RunIDs        []string `json:"run_ids"`
	MetricKeys    []string `json:"metric_keys"`
	ViewType      string   `json:"run_view_type"`
	MaxResults    int32    `json:"max_results"`
}
