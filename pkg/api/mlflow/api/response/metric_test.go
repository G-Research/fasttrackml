package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestNewMetricHistoryResponse_Ok(t *testing.T) {
	testData := []struct {
		name             string
		metrics          []models.Metric
		expectedResponse *GetMetricHistoryResponse
	}{
		{
			name: "WithNaNValue",
			metrics: []models.Metric{
				{
					Key:       "key",
					Value:     123.4,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					IsNan:     true,
					Iter:      1,
				},
			},
			expectedResponse: &GetMetricHistoryResponse{
				Metrics: []MetricPartialResponse{
					{
						Key:       "key",
						Timestamp: 1234567890,
						Step:      1,
						Value:     common.NANValue,
					},
				},
			},
		},
		{
			name: "WithNotNaNValue",
			metrics: []models.Metric{
				{
					Key:       "key",
					Value:     123.4,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					IsNan:     false,
					Iter:      1,
				},
			},
			expectedResponse: &GetMetricHistoryResponse{
				Metrics: []MetricPartialResponse{
					{
						Key:       "key",
						Timestamp: 1234567890,
						Step:      1,
						Value:     123.4,
					},
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			actualResponse, err := NewMetricHistoryResponse(tt.metrics)
			require.Nil(t, err)
			assert.Equal(t, tt.expectedResponse, actualResponse)
		})
	}
}

func TestNewMetricHistoryBulkResponse_Ok(t *testing.T) {
	testData := []struct {
		name             string
		metrics          []models.Metric
		expectedResponse *GetMetricHistoryBulkResponse
	}{
		{
			name: "WithNaNValue",
			metrics: []models.Metric{
				{
					Key:       "key",
					Value:     123.4,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					IsNan:     true,
					Iter:      1,
				},
			},
			expectedResponse: &GetMetricHistoryBulkResponse{
				Metrics: []MetricPartialResponse{
					{
						RunID:     "run_id",
						Key:       "key",
						Timestamp: 1234567890,
						Step:      1,
						Value:     common.NANValue,
					},
				},
			},
		},
		{
			name: "WithNotNaNValue",
			metrics: []models.Metric{
				{
					Key:       "key",
					Value:     123.4,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					IsNan:     false,
					Iter:      1,
				},
			},
			expectedResponse: &GetMetricHistoryBulkResponse{
				Metrics: []MetricPartialResponse{
					{
						RunID:     "run_id",
						Key:       "key",
						Timestamp: 1234567890,
						Step:      1,
						Value:     123.4,
					},
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			actualResponse := NewMetricHistoryBulkResponse(tt.metrics)
			assert.Equal(t, tt.expectedResponse, actualResponse)
		})
	}
}
