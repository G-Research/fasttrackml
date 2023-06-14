package run

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/repositories"
)

func TestService_CreateRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"Create",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.MatchedBy(func(run *models.Run) bool {
			assert.NotEmpty(t, run.ID)
			assert.Equal(t, "name", run.Name)
			assert.Equal(t, int32(1), run.ExperimentID)
			assert.Equal(t, "1", run.UserID)
			assert.Equal(t, models.StatusRunning, run.Status)
			assert.NotEmpty(t, run.StartTime.Int64)
			assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
			assert.Contains(t, run.ArtifactURI, "/artifact/location")
			assert.Equal(t, []models.Tag{
				{
					Key:   "key",
					Value: "value",
				},
			}, run.Tags)
			return true
		}),
	).Return(nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		int32(1),
	).Return(&models.Experiment{
		ID:               common.GetPointer(int32(1)),
		ArtifactLocation: "/artifact/location",
	}, nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&experimentRepository,
	)
	run, err := service.CreateRun(context.TODO(), &request.CreateRunRequest{
		ExperimentID: "1",
		UserID:       "1",
		Name:         "name",
		StartTime:    12345,
		Tags: []request.RunTagPartialRequest{
			{
				Key:   "key",
				Value: "value",
			},
		},
	})

	// compare results.
	assert.Nil(t, err)
	assert.NotEmpty(t, run.ID)
	assert.Equal(t, "name", run.Name)
	assert.Equal(t, "1", run.UserID)
	assert.Equal(t, int32(1), run.ExperimentID)
	assert.Equal(t, models.StatusRunning, run.Status)
	assert.Equal(t, int64(12345), run.StartTime.Int64)
	assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
	assert.Equal(t, []models.Tag{
		{
			Key:   "key",
			Value: "value",
		},
	}, run.Tags)
}
func TestService_CreateRun_Error(t *testing.T) {}

func TestService_UpdateRun_Ok(t *testing.T) {
	// TODO:DSuhinin skip this test for now. I don't know how to mock `gorm` transaction logic.
}
func TestService_UpdateRun_Error(t *testing.T) {}

func TestService_RestoreRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{ID: "1"}, nil)
	runRepository.On(
		"Update",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{
			ID:             "1",
			DeletedTime:    sql.NullInt64{Valid: false},
			LifecycleStage: models.LifecycleStageActive,
		},
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.RestoreRun(context.TODO(), &request.RestoreRunRequest{RunID: "1"})

	// compare results.
	assert.Nil(t, err)
}
func TestService_RestoreRun_Error(t *testing.T) {}

func TestService_SetRunTag_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive}, nil,
	)
	runRepository.On(
		"SetRunTagsBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive},
		1,
		[]models.Tag{{RunID: "1", Key: "key", Value: "value"}},
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.SetRunTag(context.TODO(), &request.SetRunTagRequest{
		RunID: "1",
		Key:   "key",
		Value: "value",
	})

	// compare results.
	assert.Nil(t, err)
}
func TestService_SetRunTag_Error(t *testing.T) {}

func TestService_DeleteRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{ID: "1"}, nil)
	runRepository.On(
		"Archive",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{ID: "1"},
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.DeleteRun(context.TODO(), &request.DeleteRunRequest{RunID: "1"})

	// compare results.
	assert.Nil(t, err)
}
func TestService_DeleteRun_Error(t *testing.T) {}

func TestService_DeleteRunTag_Ok(t *testing.T)    {}
func TestService_DeleteRunTag_Error(t *testing.T) {}

func TestService_GetRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{
		ID:             "1",
		Name:           "name",
		SourceType:     "source_type",
		SourceName:     "source_name",
		EntryPointName: "entry_point_name",
		UserID:         "user_id",
		Status:         models.StatusRunning,
		StartTime:      sql.NullInt64{Int64: 111111111, Valid: true},
		EndTime:        sql.NullInt64{Int64: 222222222, Valid: true},
		SourceVersion:  "source_version",
		LifecycleStage: models.LifecycleStageActive,
		ArtifactURI:    "artifact_uri",
		ExperimentID:   1,
		RowNum:         1,
		Params: []models.Param{
			{
				Key:   "key",
				Value: "value",
			},
		},
		Tags: []models.Tag{
			{
				Key:   "key",
				Value: "value",
			},
		},
		Metrics: []models.Metric{
			{
				Key:       "key",
				Value:     1.1,
				Timestamp: 1234567890,
				Step:      2,
			},
		},
	}, nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
	)
	run, err := service.GetRun(context.TODO(), &request.GetRunRequest{RunID: "1"})

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, "1", run.ID)
	assert.Equal(t, "name", run.Name)
	assert.Equal(t, "source_type", run.SourceType)
	assert.Equal(t, "source_name", run.SourceName)
	assert.Equal(t, "entry_point_name", run.EntryPointName)
	assert.Equal(t, "user_id", run.UserID)
	assert.Equal(t, models.StatusRunning, run.Status)
	assert.Equal(t, sql.NullInt64{Int64: 111111111, Valid: true}, run.StartTime)
	assert.Equal(t, sql.NullInt64{Int64: 222222222, Valid: true}, run.EndTime)
	assert.Equal(t, "source_version", run.SourceVersion)
	assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
	assert.Equal(t, "artifact_uri", run.ArtifactURI)
	assert.Equal(t, int32(1), run.ExperimentID)
	assert.Equal(t, models.RowNum(1), run.RowNum)
	assert.Equal(t, []models.Param{
		{
			Key:   "key",
			Value: "value",
		},
	}, run.Params)
	assert.Equal(t, []models.Tag{
		{
			Key:   "key",
			Value: "value",
		},
	}, run.Tags)
	assert.Equal(t, []models.Metric{
		{
			Key:       "key",
			Value:     1.1,
			Timestamp: 1234567890,
			Step:      2,
		},
	}, run.Metrics)
}
func TestService_GetRun_Error(t *testing.T) {}

func TestService_LogBatch_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	runRepository.On(
		"SetRunTagsBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive},
		100,
		mock.MatchedBy(func(tags []models.Tag) bool {
			assert.Equal(t, "1", tags[0].RunID)
			assert.Equal(t, "key1", tags[0].Key)
			assert.Equal(t, "value1", tags[0].Value)
			return true
		}),
	).Return(nil)
	paramRepository := repositories.MockParamRepositoryProvider{}
	paramRepository.On(
		"CreateBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		100,
		mock.MatchedBy(func(params []models.Param) bool {
			assert.Equal(t, "1", params[0].RunID)
			assert.Equal(t, "key2", params[0].Key)
			assert.Equal(t, "value2", params[0].Value)
			return true
		}),
	).Return(nil)
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"CreateBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive},
		100,
		mock.MatchedBy(func(metrics []models.Metric) bool {
			assert.Equal(t, "1", metrics[0].RunID)
			assert.Equal(t, "key3", metrics[0].Key)
			assert.Equal(t, 1.1, metrics[0].Value)
			assert.Equal(t, int64(1), metrics[0].Step)
			assert.Equal(t, int64(1234567890), metrics[0].Timestamp)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&paramRepository,
		&metricRepository,
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.LogBatch(context.TODO(), &request.LogBatchRequest{
		RunID: "1",
		Tags: []request.TagPartialRequest{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		Params: []request.ParamPartialRequest{
			{
				Key:   "key2",
				Value: "value2",
			},
		},
		Metrics: []request.MetricPartialRequest{
			{
				Key:       "key3",
				Value:     1.1,
				Timestamp: 1234567890,
				Step:      1,
			},
		},
	})

	// compare results.
	assert.Nil(t, err)
}
func TestService_LogBatch_Error(t *testing.T) {}

func TestService_LogMetric_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"CreateBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive},
		1,
		mock.MatchedBy(func(metrics []models.Metric) bool {
			assert.Equal(t, "1", metrics[0].RunID)
			assert.Equal(t, "key", metrics[0].Key)
			assert.Equal(t, 1.1, metrics[0].Value)
			assert.Equal(t, int64(1), metrics[0].Step)
			assert.Equal(t, int64(1234567890), metrics[0].Timestamp)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&metricRepository,
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.LogMetric(context.TODO(), &request.LogMetricRequest{
		RunID:     "1",
		Key:       "key",
		Value:     1.1,
		Timestamp: 1234567890,
		Step:      1,
	})

	// compare results.
	assert.Nil(t, err)
}
func TestService_LogMetric_Error(t *testing.T) {}

func TestService_LogParam_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByID",
		mock.AnythingOfType("*context.emptyCtx"),
		"1",
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	paramRepository := repositories.MockParamRepositoryProvider{}
	paramRepository.On(
		"CreateBatch",
		mock.AnythingOfType("*context.emptyCtx"),
		1,
		mock.MatchedBy(func(params []models.Param) bool {
			assert.Equal(t, "1", params[0].RunID)
			assert.Equal(t, "key", params[0].Key)
			assert.Equal(t, "value", params[0].Value)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&paramRepository,
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
	)
	err := service.LogParam(context.TODO(), &request.LogParamRequest{
		RunID: "1",
		Key:   "key",
		Value: "value",
	})

	// compare results.
	assert.Nil(t, err)
}
func TestService_LogParam_Error(t *testing.T) {}
