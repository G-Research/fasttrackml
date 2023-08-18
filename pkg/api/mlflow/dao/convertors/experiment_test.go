package convertors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

func TestConvertCreateExperimentToDBModel_Ok(t *testing.T) {
	req := request.CreateExperimentRequest{
		Name: "name",
		Tags: []request.ExperimentTagPartialRequest{
			{
				Key:   "key",
				Value: "value",
			},
		},
		ArtifactLocation: "s3://location",
	}
	result, err := ConvertCreateExperimentToDBModel(&req)
	assert.Nil(t, err)
	assert.Equal(t, "name", result.Name)
	assert.Equal(t, models.LifecycleStageActive, result.LifecycleStage)
	assert.Equal(t, "s3://location", result.ArtifactLocation)
	assert.Equal(t, []models.ExperimentTag{
		{
			Key:   "key",
			Value: "value",
		},
	}, result.Tags)
	assert.NotNil(t, result.CreationTime)
	assert.NotNil(t, result.LastUpdateTime)
}

func TestConvertUpdateExperimentToDBModel_Ok(t *testing.T) {
	req := request.UpdateExperimentRequest{
		Name: "name",
	}
	result := ConvertUpdateExperimentToDBModel(&models.Experiment{}, &req)
	assert.Equal(t, "name", result.Name)
	assert.NotNil(t, result.LastUpdateTime)
}
