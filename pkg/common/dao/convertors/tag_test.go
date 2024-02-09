package convertors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestConvertSetRunTagRequestToDBModel_Ok(t *testing.T) {
	req := request.SetRunTagRequest{
		Key:   "key",
		Value: "value",
	}
	result := ConvertSetRunTagRequestToDBModel("run_id", &req)
	assert.Equal(t, "key", result.Key)
	assert.Equal(t, "value", result.Value)
}

func TestConvertSetExperimentTagRequestToDBModel_Ok(t *testing.T) {
	req := request.SetExperimentTagRequest{
		Key:   "key",
		Value: "value",
	}
	result := ConvertSetExperimentTagRequestToDBModel(1, &req)
	assert.Equal(t, "key", result.Key)
	assert.Equal(t, "value", result.Value)
	assert.Equal(t, int32(1), result.ExperimentID)
}
