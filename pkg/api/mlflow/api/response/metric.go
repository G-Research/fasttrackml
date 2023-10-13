package response

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
	resp := GetMetricHistoryResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		resp.Metrics[n] = MetricPartialResponse{
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}
		if m.IsNan {
			resp.Metrics[n].Value = common.NANValue
		}
	}
	return &resp
}

// GetMetricHistoryBulkResponse is a response object for `GET mlflow/metrics/get-history-bulk` endpoint.
type GetMetricHistoryBulkResponse struct {
	Metrics []MetricPartialResponse `json:"metrics"`
}

// NewMetricHistoryBulkResponse creates new GetMetricHistoryBulkResponse object.
func NewMetricHistoryBulkResponse(metrics []models.Metric) *GetMetricHistoryBulkResponse {
	resp := GetMetricHistoryBulkResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		resp.Metrics[n] = MetricPartialResponse{
			RunID:     m.RunID,
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}
		if m.IsNan {
			resp.Metrics[n].Value = common.NANValue
		}
	}
	return &resp
}
