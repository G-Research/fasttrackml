package convertors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

func TestConvertLogParamRequestToDBModel_Ok(t *testing.T) {
	req := request.LogParamRequest{
		Key:   "key",
		Value: "value",
	}
	result := ConvertLogParamRequestToDBModel("run_id", &req)
	assert.Equal(t, "key", result.Key)
	assert.Equal(t, "value", result.Value)
	assert.Equal(t, "run_id", result.RunID)
}

func TestConvertLogBatchRequestToDBModel_Ok(t *testing.T) {
	req := request.LogBatchRequest{
		Tags: []request.TagPartialRequest{{
			Key:   "key",
			Value: "value",
		}},
		Params: []request.ParamPartialRequest{
			{
				Key:   "key",
				Value: "value",
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
	}
	contexts := []*models.Context{nil}
	metrics, params, tags, err := ConvertLogBatchRequestToDBModel("run_id", &req, contexts)
	require.Nil(t, err)
	assert.Equal(t, []models.Tag{
		{
			RunID: "run_id",
			Key:   "key",
			Value: "value",
		},
	}, tags)
	assert.Equal(t, []models.Param{
		{
			RunID: "run_id",
			Key:   "key",
			Value: "value",
		},
	}, params)
	assert.Equal(t, []models.Metric{
		{
			Key:       "key",
			Value:     1.1,
			Timestamp: 1234567890,
			RunID:     "run_id",
			Step:      1,
		},
	}, metrics)
}
