package response

import (
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

// MetricPartialResponse is a partial response object for GetMetricHistoryResponse.
type MetricPartialResponse struct {
	RunID     string `json:"run_id,omitempty"`
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
func NewMetricHistoryResponse(metrics []models.Metric) *GetMetricHistoryResponse {
	response := GetMetricHistoryResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		response.Metrics[n] = MetricPartialResponse{
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}
		if m.IsNan {
			response.Metrics[n].Value = "NaN"
		}
	}
	return &response
}

// GetMetricHistoryBulkResponse is a response object for `GET mlflow/metrics/get-history-bulk` endpoint.
type GetMetricHistoryBulkResponse struct {
	Metrics []MetricPartialResponse `json:"metrics"`
}

// NewMetricHistoryBulkResponse creates new GetMetricHistoryBulkResponse object.
func NewMetricHistoryBulkResponse(metrics []models.Metric) *GetMetricHistoryBulkResponse {
	response := GetMetricHistoryBulkResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		response.Metrics[n] = MetricPartialResponse{
			RunID:     m.RunID,
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}
		if m.IsNan {
			response.Metrics[n].Value = "NaN"
		}
	}
	return &response
}
