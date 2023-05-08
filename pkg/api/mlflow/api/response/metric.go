package response

import "github.com/G-Research/fasttrack/pkg/database"

// MetricPartialResponse is a partial response object for GetMetricHistoryResponse.
type MetricPartialResponse struct {
	RunID     string `json:"run_id"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

// GetMetricHistoryResponse is a response object for `GET mlflow/metrics/get-history` endpoint.
type GetMetricHistoryResponse struct {
	Metrics []MetricPartialResponse `json:"metrics"`
}

// NewMetricHistoryResponse creates new GetMetricHistoryResponse object.
func NewMetricHistoryResponse(metrics []database.Metric) *GetMetricHistoryResponse {
	response := GetMetricHistoryResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		response.Metrics[n] = MetricPartialResponse{
			RunID:     m.RunID,
			Key:       m.Key,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Step:      m.Step,
		}
		if m.IsNan {
			response.Metrics[n].Value = "NaN"
		}
	}
	return &response
}
