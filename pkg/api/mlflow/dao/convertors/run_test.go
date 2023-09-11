package convertors

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

func TestConvertCreateRunRequestToDBModel(t *testing.T) {
	testData := []struct {
		name   string
		req    *request.CreateRunRequest
		result func(run *models.Run)
	}{
		{
			name: "WithEmptySourceType",
			req: &request.CreateRunRequest{
				ExperimentID: "experiment_id",
				UserID:       "user_id",
				Name:         "name",
				StartTime:    1234567890,
				Tags: []request.RunTagPartialRequest{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			result: func(run *models.Run) {
				assert.NotEmpty(t, run.ID)
				assert.Equal(t, "name", run.Name)
				assert.Equal(t, int32(123), run.ExperimentID)
				assert.Equal(t, "user_id", run.UserID)
				assert.Equal(t, models.StatusRunning, run.Status)
				assert.Equal(t, sql.NullInt64{Valid: true, Int64: 1234567890}, run.StartTime)
				assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
				assert.Contains(t, run.ArtifactURI, "artifacts")
				assert.Equal(t, []models.Tag{
					{
						Key:   "key",
						Value: "value",
					},
				}, run.Tags)
				assert.Equal(t, "UNKNOWN", run.SourceType)
			},
		},
		{
			name: "WithNonEmptySourceType",
			req: &request.CreateRunRequest{
				ExperimentID: "experiment_id",
				UserID:       "user_id",
				Name:         "name",
				StartTime:    1234567890,
				Tags: []request.RunTagPartialRequest{
					{
						Key:   TagKeySourceType,
						Value: "value",
					},
				},
			},
			result: func(run *models.Run) {
				assert.NotEmpty(t, run.ID)
				assert.Equal(t, "name", run.Name)
				assert.Equal(t, int32(123), run.ExperimentID)
				assert.Equal(t, "user_id", run.UserID)
				assert.Equal(t, models.StatusRunning, run.Status)
				assert.Equal(t, sql.NullInt64{Valid: true, Int64: 1234567890}, run.StartTime)
				assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
				assert.Contains(t, run.ArtifactURI, "artifacts")
				assert.Equal(t, []models.Tag{
					{
						Key:   TagKeySourceType,
						Value: "value",
					},
				}, run.Tags)
				assert.Equal(t, "value", run.SourceType)
			},
		},
		{
			name: "WithTagKeyUser",
			req: &request.CreateRunRequest{
				ExperimentID: "experiment_id",
				Name:         "name",
				StartTime:    1234567890,
				Tags: []request.RunTagPartialRequest{
					{
						Key:   TagKeyUser,
						Value: "value",
					},
				},
			},
			result: func(run *models.Run) {
				assert.NotEmpty(t, run.ID)
				assert.Equal(t, "name", run.Name)
				assert.Equal(t, int32(123), run.ExperimentID)
				assert.Equal(t, models.StatusRunning, run.Status)
				assert.Equal(t, sql.NullInt64{Valid: true, Int64: 1234567890}, run.StartTime)
				assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
				assert.Contains(t, run.ArtifactURI, "artifacts")
				assert.Equal(t, []models.Tag{
					{
						Key:   TagKeyUser,
						Value: "value",
					},
				}, run.Tags)
				assert.Equal(t, "UNKNOWN", run.SourceType)
				assert.Equal(t, "value", run.UserID)
			},
		},
		{
			name: "WithTagKeySourceName",
			req: &request.CreateRunRequest{
				ExperimentID: "experiment_id",
				UserID:       "user_id",
				Name:         "name",
				StartTime:    1234567890,
				Tags: []request.RunTagPartialRequest{
					{
						Key:   TagKeySourceName,
						Value: "value",
					},
				},
			},
			result: func(run *models.Run) {
				assert.NotEmpty(t, run.ID)
				assert.Equal(t, "name", run.Name)
				assert.Equal(t, int32(123), run.ExperimentID)
				assert.Equal(t, "user_id", run.UserID)
				assert.Equal(t, models.StatusRunning, run.Status)
				assert.Equal(t, sql.NullInt64{Valid: true, Int64: 1234567890}, run.StartTime)
				assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
				assert.Contains(t, run.ArtifactURI, "artifacts")
				assert.Equal(t, []models.Tag{
					{
						Key:   TagKeySourceName,
						Value: "value",
					},
				}, run.Tags)
				assert.Equal(t, "value", run.SourceName)
				assert.Equal(t, "UNKNOWN", run.SourceType)
			},
		},
		{
			name: "WithTagKeyRunName",
			req: &request.CreateRunRequest{
				ExperimentID: "experiment_id",
				UserID:       "user_id",
				StartTime:    1234567890,
				Tags: []request.RunTagPartialRequest{
					{
						Key:   TagKeyRunName,
						Value: "value",
					},
				},
			},
			result: func(run *models.Run) {
				assert.NotEmpty(t, run.ID)
				assert.Equal(t, "value", run.Name)
				assert.Equal(t, int32(123), run.ExperimentID)
				assert.Equal(t, "user_id", run.UserID)
				assert.Equal(t, models.StatusRunning, run.Status)
				assert.Equal(t, sql.NullInt64{Valid: true, Int64: 1234567890}, run.StartTime)
				assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
				assert.Contains(t, run.ArtifactURI, "artifacts")
				assert.Equal(t, []models.Tag{
					{
						Key:   TagKeyRunName,
						Value: "value",
					},
				}, run.Tags)
				assert.Equal(t, "UNKNOWN", run.SourceType)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			experimentID := int32(123)
			result, err := ConvertCreateRunRequestToDBModel(&models.Experiment{
				ID:               &experimentID,
				ArtifactLocation: "artifact_location",
			}, tt.req)
			assert.Nil(t, err)
			tt.result(result)
		})
	}
}

func TestConvertUpdateRunRequestToDBModel(t *testing.T) {
	req := request.UpdateRunRequest{
		Name:    "name",
		Status:  "status",
		EndTime: 1234567890,
	}
	result := ConvertUpdateRunRequestToDBModel(&models.Run{}, &req)
	assert.Equal(t, "name", result.Name)
	assert.Equal(t, models.Status("status"), result.Status)
	assert.Equal(t, int64(1234567890), result.EndTime.Int64)
}
