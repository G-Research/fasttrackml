package response

// MetricPartialResponse is a partial response object for GetMetricHistoryResponse.
type MetricPartialResponse struct {
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

// GetMetricHistoryResponse is a response object for `GET mlflow/metrics/get-history` endpoint.
type GetMetricHistoryResponse struct {
	Metrics []MetricPartialResponse `json:"metrics"`
}
