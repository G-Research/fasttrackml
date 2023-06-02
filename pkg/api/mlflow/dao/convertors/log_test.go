package convertors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/models"
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
	}
	params, tags := ConvertLogBatchRequestToDBModel("run_id", &req)
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
}
