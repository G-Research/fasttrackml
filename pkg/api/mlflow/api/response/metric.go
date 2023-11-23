package response

import (
	"encoding/json"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// MetricPartialResponse is a partial response object for GetMetricHistoryResponse.
type MetricPartialResponse struct {
	RunID     string         `json:"run_id,omitempty"`
	Key       string         `json:"key"`
	Value     any            `json:"value"`
	Timestamp int64          `json:"timestamp"`
	Step      int64          `json:"step"`
	Context   map[string]any `json:"context"`
}

// GetMetricHistoryResponse is a response object for `GET mlflow/metrics/get-history` endpoint.
type GetMetricHistoryResponse struct {
	Metrics []MetricPartialResponse `json:"metrics"`
}

// NewMetricHistoryResponse creates new GetMetricHistoryResponse object.
func NewMetricHistoryResponse(metrics []models.Metric) (*GetMetricHistoryResponse, error) {
	resp := GetMetricHistoryResponse{
		Metrics: make([]MetricPartialResponse, len(metrics)),
	}

	for n, m := range metrics {
		var context map[string]interface{}
		if err := json.Unmarshal(m.Context.Json, &context); err != nil {
			return nil, eris.Wrap(err, "error unmarshaling context")
		}
		resp.Metrics[n] = MetricPartialResponse{
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Context:   context,
		}
		if m.IsNan {
			resp.Metrics[n].Value = common.NANValue
		}
	}
	return &resp, nil
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
