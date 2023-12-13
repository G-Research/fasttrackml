package convertors

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestConvertMetricParamRequestToDBModel_Ok(t *testing.T) {
	testData := []struct {
		name           string
		request        *request.LogMetricRequest
		expectedMetric *models.Metric
	}{
		{
			name: "WithMetricNormalValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     1.1,
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     1.1,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context:   models.Context{
					Json: []byte(`{}`),
				},
			},
		},
		{
			name: "WithMetricNaNValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     "NaN",
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     0,
				IsNan:     true,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context:   models.Context{
					Json: []byte(`{}`),
				},
			},
		},
		{
			name: "WithMetricPositiveInfinityValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     "Infinity",
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     math.MaxFloat64,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context:   models.Context{
					Json: []byte(`{}`),
				},
			},
		},
		{
			name: "WithMetricNegativeInfinityValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     "-Infinity",
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     -math.MaxFloat64,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context:   models.Context{
					Json: []byte(`{}`),
				},
			},
		},
		{
			name: "WithMetricContext",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     1.1,
				RunID:     "run_id",
				Timestamp: 1234567890,
				Context: map[string]interface{}{
					"key1": "value1",
					"key2": 2,
				},
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     1.1,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context: models.Context{
					Json: []byte(`{"key1":"value1","key2":2}`),
				},
			},
		},
		{
			name: "WithMetricContext",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     1.1,
				RunID:     "run_id",
				Timestamp: 1234567890,
				Context: map[string]interface{}{
					"key1": "value1",
					"key2": 2,
				},
			},
			expectedMetric: &models.Metric{
				Key:       "key",
				Value:     1.1,
				Timestamp: 1234567890,
				RunID:     "run_id",
				Step:      1,
				Context: &models.Context{
					Json: []byte(`{"key1":"value1","key2":2}`),
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := ConvertLogMetricRequestToDBModel("run_id", tt.request)
			require.Nil(t, err)
			assert.Equal(t, tt.expectedMetric, metric)
		})
	}
}

func TestConvertMetricParamRequestToDBModel_Error(t *testing.T) {
	testData := []struct {
		name    string
		request *request.LogMetricRequest
		error   error
	}{
		{
			name: "WithUnsupportedMetricValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     struct{}{},
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			error: errors.New("invalid metric value '{}'"),
		},
		{
			name: "WithUnsupportedNaNMetricValue",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     "UnsupportedNaNValue",
				RunID:     "run_id",
				Timestamp: 1234567890,
			},
			error: errors.New("invalid metric value 'UnsupportedNaNValue'"),
		},
		{
			name: "WithUnsupportedContext",
			request: &request.LogMetricRequest{
				Key:       "key",
				Step:      1,
				Value:     1.1,
				RunID:     "run_id",
				Timestamp: 1234567890,
				Context: map[string]interface{}{
					"unsupported": func() {},
				},
			},
			error: errors.New("error marshalling context: json: unsupported type: func()"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConvertLogMetricRequestToDBModel("run_id", tt.request)
			assert.Equal(t, tt.error.Error(), err.Error())
		})
	}
}
