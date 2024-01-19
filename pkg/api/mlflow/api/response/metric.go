package response

import (
	"encoding/json"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// MetricPartialResponseBulk is a partial response object for GetMetricHistoryBulkResponse.
type MetricPartialResponseBulk struct {
	RunID     string `json:"run_id,omitempty"`
	Key       string `json:"key"`
	Value     any    `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Step      int64  `json:"step"`
}

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

	var mappedContext map[string]map[string]interface{}
	for n, m := range metrics {
		resp.Metrics[n] = MetricPartialResponse{
			Key:       m.Key,
			Step:      m.Step,
			Value:     m.Value,
			Timestamp: m.Timestamp,
		}

		// avoid serialization of the same context.
		// if context has been already serialized just use it.
		if context, ok := mappedContext[m.Context.GetJsonHash()]; ok {
			resp.Metrics[n].Context = context
		} else {
			var deserializedContext map[string]interface{}
			if err := json.Unmarshal(m.Context.Json, &deserializedContext); err != nil {
				return nil, eris.Wrap(err, "error unmarshaling context")
			}
			resp.Metrics[n].Context = deserializedContext
			mappedContext[m.Context.GetJsonHash()] = deserializedContext
		}
		if m.IsNan {
			resp.Metrics[n].Value = common.NANValue
		}
	}
	return &resp, nil
}

// GetMetricHistoryBulkResponse is a response object for `GET mlflow/metrics/get-history-bulk` endpoint.
type GetMetricHistoryBulkResponse struct {
	Metrics []MetricPartialResponseBulk `json:"metrics"`
}

// NewMetricHistoryBulkResponse creates new GetMetricHistoryBulkResponse object.
func NewMetricHistoryBulkResponse(metrics []models.Metric) *GetMetricHistoryBulkResponse {
	resp := GetMetricHistoryBulkResponse{
		Metrics: make([]MetricPartialResponseBulk, len(metrics)),
	}

	for n, m := range metrics {
		resp.Metrics[n] = MetricPartialResponseBulk{
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
