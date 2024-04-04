package convertors

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestConvertLogParamRequestToDBModel_Ok(t *testing.T) {
	req := request.LogParamRequest{
		Key:   "key",
		ValueStr: common.GetPointer[string]("value"),
	}
	result := ConvertLogParamRequestToDBModel("run_id", &req)
	assert.Equal(t, "key", result.Key)
	assert.Equal(t, "value", *result.ValueStr)
	assert.Equal(t, "run_id", result.RunID)
}

func TestConvertLogBatchRequestToDBModel_Ok(t *testing.T) {
	testData := []struct {
		name            string
		request         *request.LogBatchRequest
		expectedTags    []models.Tag
		expectedParams  []models.Param
		expectedMetrics []models.Metric
	}{
		{
			name: "WithMetricNormalValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     1.1,
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			expectedTags: []models.Tag{
				{
					RunID: "run_id",
					Key:   "key",
					Value: "value",
				},
			},
			expectedParams: []models.Param{
				{
					RunID:    "run_id",
					Key:      "key",
					ValueStr: common.GetPointer[string]("value"),
				},
			},
			expectedMetrics: []models.Metric{
				{
					Key:       "key",
					Value:     1.1,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					Context:   models.DefaultContext,
				},
			},
		},
		{
			name: "WithMetricNaNValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     "NaN",
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			expectedTags: []models.Tag{
				{
					RunID: "run_id",
					Key:   "key",
					Value: "value",
				},
			},
			expectedParams: []models.Param{
				{
					RunID:    "run_id",
					Key:      "key",
					ValueStr: common.GetPointer[string]("value"),
				},
			},
			expectedMetrics: []models.Metric{
				{
					Key:       "key",
					Value:     0,
					IsNan:     true,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					Context:   models.DefaultContext,
				},
			},
		},
		{
			name: "WithMetricPositiveInfinityValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     "Infinity",
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			expectedTags: []models.Tag{
				{
					RunID: "run_id",
					Key:   "key",
					Value: "value",
				},
			},
			expectedParams: []models.Param{
				{
					RunID:    "run_id",
					Key:      "key",
					ValueStr: common.GetPointer[string]("value"),
				},
			},
			expectedMetrics: []models.Metric{
				{
					Key:       "key",
					Value:     math.MaxFloat64,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					Context:   models.DefaultContext,
				},
			},
		},
		{
			name: "WithMetricNegativeInfinityValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     "-Infinity",
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			expectedTags: []models.Tag{
				{
					RunID: "run_id",
					Key:   "key",
					Value: "value",
				},
			},
			expectedParams: []models.Param{
				{
					RunID:    "run_id",
					Key:      "key",
					ValueStr: common.GetPointer[string]("value"),
				},
			},
			expectedMetrics: []models.Metric{
				{
					Key:       "key",
					Value:     -math.MaxFloat64,
					Timestamp: 1234567890,
					RunID:     "run_id",
					Step:      1,
					Context:   models.DefaultContext,
				},
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			metrics, params, tags, err := ConvertLogBatchRequestToDBModel("run_id", tt.request)
			require.Nil(t, err)
			assert.Equal(t, tt.expectedTags, tags)
			assert.Equal(t, tt.expectedParams, params)
			assert.Equal(t, tt.expectedMetrics, metrics)
		})
	}
}

func TestConvertLogBatchRequestToDBModel_Error(t *testing.T) {
	testData := []struct {
		name    string
		request *request.LogBatchRequest
		error   error
	}{
		{
			name: "WithUnsupportedMetricValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     struct{}{},
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			error: errors.New("invalid metric value '{}'"),
		},
		{
			name: "WithUnsupportedNaNMetricValue",
			request: &request.LogBatchRequest{
				Tags: []request.TagPartialRequest{{
					Key:   "key",
					Value: "value",
				}},
				Params: []request.ParamPartialRequest{
					{
						Key:   "key",
						ValueStr: common.GetPointer[string]("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     "UnsupportedNaNValue",
						Timestamp: 1234567890,
						Step:      1,
					},
				},
			},
			error: errors.New("invalid metric value 'UnsupportedNaNValue'"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := ConvertLogBatchRequestToDBModel("run_id", tt.request)
			assert.Equal(t, tt.error.Error(), err.Error())
		})
	}
}
