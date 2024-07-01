package run

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

func TestService_CreateRun_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:                  1,
		Code:                "code",
		DefaultExperimentID: common.GetPointer(int32(1)),
	}

	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"Create",
		context.TODO(),
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
		"GetByNamespaceIDAndExperimentID",
		context.TODO(),
		ns.ID,
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	run, err := service.CreateRun(context.TODO(), &ns, &request.CreateRunRequest{
		ExperimentID: "0", // default experiment id provided by the client is "0"
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
	require.Nil(t, err)
	assert.NotEmpty(t, run.ID)
	assert.Equal(t, "name", run.Name)
	assert.Equal(t, "1", run.UserID)
	assert.Equal(t, int32(1), run.ExperimentID) // default experiment id "0" is translated to namespace default.
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

func TestService_CreateRun_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.CreateRunRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectExperimentID",
			error:   api.NewBadRequestError(`unable to parse experiment id '': strconv.ParseInt: parsing "": invalid syntax`),
			request: &request.CreateRunRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError("unable to find experiment with id '1': database error"),
			request: &request.CreateRunRequest{
				ExperimentID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID",
					context.TODO(),
					ns.ID,
					int32(1),
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&experimentRepository,
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "CreateRunDatabaseError",
			error: api.NewInternalError("error inserting run: database error"),
			request: &request.CreateRunRequest{
				ExperimentID: "1",
				Name:         "name",
				UserID:       "1",
				Tags: []request.RunTagPartialRequest{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID",
					context.TODO(),
					ns.ID,
					int32(1),
				).Return(&models.Experiment{ID: common.GetPointer(int32(1))}, nil)
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"Create",
					context.TODO(),
					mock.MatchedBy(func(run *models.Run) bool {
						assert.NotEmpty(t, run.ID)
						assert.Equal(t, "name", run.Name)
						assert.Equal(t, int32(1), run.ExperimentID)
						assert.Equal(t, "1", run.UserID)
						assert.Equal(t, models.StatusRunning, run.Status)
						assert.NotNil(t, run.StartTime)
						assert.NotNil(t, models.LifecycleStageActive, run.LifecycleStage)
						assert.NotNil(t, []models.Tag{
							{
								Key:   "key",
								Value: "value",
							},
						}, run.Tags)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&experimentRepository,
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().CreateRun(context.TODO(), &ns, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_UpdateRun_Ok(t *testing.T) {
	// TODO:DSuhinin skip this test for now. I don't know how to mock `gorm` transaction logic.
}

func TestService_UpdateRun_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateRunRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.UpdateRunRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewResourceDoesNotExistError("unable to find run '1': database error"),
			request: &request.UpdateRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().UpdateRun(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_RestoreRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDAndRunID",
		context.TODO(),
		uint(1),
		"1",
	).Return(&models.Run{ID: "1"}, nil)
	runRepository.On(
		"Update",
		context.TODO(),
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.RestoreRun(context.TODO(), &models.Namespace{ID: 1}, &request.RestoreRunRequest{RunID: "1"})

	// compare results.
	require.Nil(t, err)
}

func TestService_RestoreRun_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreRunRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.RestoreRunRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundDatabaseError",
			error: api.NewResourceDoesNotExistError("unable to find run '1': database error"),
			request: &request.RestoreRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RestoreRunDatabaseError",
			error: api.NewInternalError("unable to restore run '1': database error"),
			request: &request.RestoreRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(&models.Run{
					ID: "1",
				}, nil)
				runRepository.On(
					"Update",
					context.TODO(),
					mock.MatchedBy(func(run *models.Run) bool {
						assert.Equal(t, "1", run.ID)
						assert.Equal(t, sql.NullInt64{}, run.DeletedTime)
						assert.Equal(t, models.LifecycleStageActive, run.LifecycleStage)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().RestoreRun(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_SetRunTag_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDRunIDAndLifecycleStage",
		context.TODO(),
		uint(1),
		"1",
		models.LifecycleStageActive,
	).Return(
		&models.Run{ID: "1", LifecycleStage: models.LifecycleStageActive}, nil,
	)
	runRepository.On(
		"SetRunTagsBatch",
		context.TODO(),
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.SetRunTag(context.TODO(), &models.Namespace{
		ID: 1,
	}, &request.SetRunTagRequest{
		RunID: "1",
		Key:   "key",
		Value: "value",
	})

	// compare results.
	require.Nil(t, err)
}
func TestService_SetRunTag_Error(t *testing.T) {}

func TestService_DeleteRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDAndRunID",
		context.TODO(),
		uint(1),
		"1",
	).Return(&models.Run{ID: "1"}, nil)
	runRepository.On(
		"Archive",
		context.TODO(),
		&models.Run{ID: "1"},
	).Return(nil)

	// call service under testing.
	service := NewService(
		&repositories.MockTagRepositoryProvider{},
		&runRepository,
		&repositories.MockParamRepositoryProvider{},
		&repositories.MockMetricRepositoryProvider{},
		&repositories.MockExperimentRepositoryProvider{},
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.DeleteRun(context.TODO(), &models.Namespace{ID: 1}, &request.DeleteRunRequest{RunID: "1"})

	// compare results.
	require.Nil(t, err)
}

func TestService_DeleteRun_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteRunRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.DeleteRunRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewResourceDoesNotExistError("unable to find run '1': database error"),
			request: &request.DeleteRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewResourceDoesNotExistError("unable to find run '1': database error"),
			request: &request.DeleteRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "DeleteRunDatabaseError",
			error: api.NewInternalError("unable to delete run '1': database error"),
			request: &request.DeleteRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(&models.Run{
					ID: "1",
				}, nil)
				runRepository.On(
					"Archive",
					context.TODO(),
					mock.MatchedBy(func(run *models.Run) bool {
						assert.Equal(t, "1", run.ID)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().DeleteRun(context.TODO(), &models.Namespace{
				ID: 1,
			}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_DeleteRunTag_Ok(t *testing.T) {}
func TestService_DeleteRunTag_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteRunTagRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.DeleteRunTagRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewInternalError("Unable to find run '1': database error"),
			request: &request.DeleteRunTagRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "ActiveRunNotFound",
			error: api.NewResourceDoesNotExistError("Run '1' not found"),
			request: &request.DeleteRunTagRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "NotFoundTag",
			error: api.NewResourceDoesNotExistError("No tag with name: key"),
			request: &request.DeleteRunTagRequest{
				RunID: "1",
				Key:   "key",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				tagRepository := repositories.MockTagRepositoryProvider{}
				tagRepository.On(
					"GetByRunIDAndKey",
					context.TODO(),
					"1",
					"key",
				).Return(nil, nil)
				return NewService(
					&tagRepository,
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "NotFoundTagDatabaseError",
			error: api.NewInternalError("Unable to find tag 'key' for run '1': database error"),
			request: &request.DeleteRunTagRequest{
				RunID: "1",
				Key:   "key",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				tagRepository := repositories.MockTagRepositoryProvider{}
				tagRepository.On(
					"GetByRunIDAndKey",
					context.TODO(),
					"1",
					"key",
				).Return(nil, errors.New("database error"))
				return NewService(
					&tagRepository,
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "DeleteRunTagDatabaseError",
			error: api.NewInternalError("unable to delete tag 'key' for run '1': database error"),
			request: &request.DeleteRunTagRequest{
				RunID: "1",
				Key:   "key",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				tagRepository := repositories.MockTagRepositoryProvider{}
				tagRepository.On(
					"GetByRunIDAndKey",
					context.TODO(),
					"1",
					"key",
				).Return(&models.Tag{
					RunID: "1",
					Key:   "key",
					Value: "value",
				}, nil)
				tagRepository.On(
					"Delete",
					context.TODO(),
					mock.MatchedBy(func(tag *models.Tag) bool {
						assert.Equal(t, "1", tag.RunID)
						assert.Equal(t, "key", tag.Key)
						assert.Equal(t, "value", tag.Value)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&tagRepository,
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().DeleteRunTag(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_GetRun_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDAndRunID",
		context.TODO(),
		uint(1),
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
				Key:      "key",
				ValueStr: common.GetPointer("value"),
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	run, err := service.GetRun(context.TODO(), &models.Namespace{
		ID: 1,
	}, &request.GetRunRequest{RunID: "1"})

	// compare results.
	require.Nil(t, err)
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
			Key:      "key",
			ValueStr: common.GetPointer("value"),
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

func TestService_GetRun_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetRunRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.GetRunRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewResourceDoesNotExistError(`unable to find run '1': database error`),
			request: &request.GetRunRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().GetRun(context.TODO(), &models.Namespace{
				ID: 1,
			}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_LogBatch_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDRunIDAndLifecycleStage",
		context.TODO(),
		uint(1),
		"1",
		models.LifecycleStageActive,
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	runRepository.On(
		"SetRunTagsBatch",
		context.TODO(),
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
		context.TODO(),
		100,
		mock.MatchedBy(func(params []models.Param) bool {
			assert.Equal(t, "1", params[0].RunID)
			assert.Equal(t, "key2", params[0].Key)
			assert.Equal(t, "value2", params[0].ValueString())
			return true
		}),
	).Return(nil)
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"CreateBatch",
		context.TODO(),
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.LogBatch(context.TODO(), &models.Namespace{
		ID: 1,
	}, &request.LogBatchRequest{
		RunID: "1",
		Tags: []request.TagPartialRequest{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		Params: []request.ParamPartialRequest{
			{
				Key:      "key2",
				ValueStr: common.GetPointer("value2"),
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
	require.Nil(t, err)
}

func TestService_LogBatch_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogBatchRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.LogBatchRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundDatabaseError",
			error: api.NewInternalError(`Unable to find run '1': database error`),
			request: &request.LogBatchRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundDatabaseNotFoundError",
			error: api.NewResourceDoesNotExistError("Run '1' not found"),
			request: &request.LogBatchRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "NoActiveRunFound",
			error: api.NewResourceDoesNotExistError("Run '1' not found"),
			request: &request.LogBatchRequest{
				RunID: "1",
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "IncorrectMetricValue",
			error: api.NewInvalidParameterValueError("invalid metric value 'incorrect_value'"),
			request: &request.LogBatchRequest{
				RunID: "1",
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key",
						Value:     "incorrect_value",
						Timestamp: 1234567890,
					},
				},
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID: "1",
				}, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "CreateBatchParamsDatabaseError",
			error: api.NewInternalError(`unable to insert params for run '1': database error`),
			request: &request.LogBatchRequest{
				RunID: "1",
				Params: []request.ParamPartialRequest{
					{
						Key:      "key",
						ValueStr: common.GetPointer("value"),
					},
				},
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID: "1",
				}, nil)
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					100,
					[]models.Param{
						{
							Key:      "key",
							ValueStr: common.GetPointer("value"),
							RunID:    "1",
						},
					},
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "CreateBatchParamsConflictError",
			error: api.NewInvalidParameterValueError(`unable to insert params for run '1': param conflict!`),
			request: &request.LogBatchRequest{
				RunID: "1",
				Params: []request.ParamPartialRequest{
					{
						Key:      "key",
						ValueStr: common.GetPointer("value"),
					},
				},
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID: "1",
				}, nil)
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					100,
					[]models.Param{
						{
							Key:      "key",
							ValueStr: common.GetPointer("value"),
							RunID:    "1",
						},
					},
				).Return(repositories.ParamConflictError{Message: "param conflict!"})
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "CreateBatchMetricsDatabaseError",
			error: api.NewInternalError(`unable to insert metrics for run '1': database error`),
			request: &request.LogBatchRequest{
				RunID: "1",
				Params: []request.ParamPartialRequest{
					{
						Key:      "key",
						ValueStr: common.GetPointer("value"),
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Step:      1,
						Key:       "key",
						Value:     1.1,
						Timestamp: 123456789,
					},
				},
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					100,
					[]models.Param{
						{
							Key:      "key",
							ValueStr: common.GetPointer("value"),
							RunID:    "1",
						},
					},
				).Return(nil)
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"CreateBatch",
					context.TODO(),
					&models.Run{
						ID:             "1",
						LifecycleStage: models.LifecycleStageActive,
					},
					100,
					[]models.Metric{
						{
							Step:      1,
							Key:       "key",
							Value:     1.1,
							RunID:     "1",
							Timestamp: 123456789,
							ContextID: models.DefaultContext.ID,
							Context:   models.DefaultContext,
						},
					},
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&metricRepository,
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "CreateBatchTagsDatabaseError",
			error: api.NewInternalError(`unable to insert tags for run '1': database error`),
			request: &request.LogBatchRequest{
				RunID: "1",
				Params: []request.ParamPartialRequest{
					{
						Key:      "key",
						ValueStr: common.GetPointer("value"),
					},
				},
				Tags: []request.TagPartialRequest{
					{
						Key:   "key",
						Value: "value",
					},
				},
				Metrics: []request.MetricPartialRequest{
					{
						Step:      1,
						Key:       "key",
						Value:     1.1,
						Timestamp: 123456789,
					},
				},
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				runRepository.On(
					"SetRunTagsBatch",
					context.TODO(),
					&models.Run{
						ID:             "1",
						LifecycleStage: models.LifecycleStageActive,
					},
					100,
					[]models.Tag{
						{
							Key:   "key",
							Value: "value",
							RunID: "1",
						},
					},
				).Return(errors.New("database error"))
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					100,
					[]models.Param{
						{
							Key:      "key",
							ValueStr: common.GetPointer("value"),
							RunID:    "1",
						},
					},
				).Return(nil)
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"CreateBatch",
					context.TODO(),
					&models.Run{
						ID:             "1",
						LifecycleStage: models.LifecycleStageActive,
					},
					100,
					[]models.Metric{
						{
							Step:      1,
							Key:       "key",
							Value:     1.1,
							RunID:     "1",
							Timestamp: 123456789,
							ContextID: models.DefaultContext.ID,
							Context:   models.DefaultContext,
						},
					},
				).Return(nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&metricRepository,
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().LogBatch(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_LogMetric_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDAndRunID",
		context.TODO(),
		uint(1),
		"1",
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	metricRepository := repositories.MockMetricRepositoryProvider{}
	metricRepository.On(
		"CreateBatch",
		context.TODO(),
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.LogMetric(context.TODO(), &models.Namespace{
		ID: 1,
	}, &request.LogMetricRequest{
		RunID:     "1",
		Key:       "key",
		Value:     1.1,
		Timestamp: 1234567890,
		Step:      1,
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_LogMetric_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogMetricRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.LogMetricRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "EmptyOrIncorrectMetricKey",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'key'`),
			request: &request.LogMetricRequest{
				RunID: "1",
			},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "EmptyOrIncorrectTimestamp",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'timestamp'`),
			request: &request.LogMetricRequest{
				RunID: "1",
				Key:   "key",
			},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundOrDatabaseError",
			error: api.NewResourceDoesNotExistError(`unable to find run '1': database error`),
			request: &request.LogMetricRequest{
				RunID:     "1",
				Key:       "key",
				Value:     "value",
				Timestamp: 1234567890,
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "IncorrectMetricValue",
			error: api.NewInvalidParameterValueError(`invalid metric value 'incorrect_value'`),
			request: &request.LogMetricRequest{
				RunID:     "1",
				Key:       "key",
				Value:     "incorrect_value",
				Timestamp: 1234567890,
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(&models.Run{
					ID: "1",
				}, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "LogMetricDatabaseError",
			error: api.NewInternalError(`unable to log metric 'key' for run '1': database error`),
			request: &request.LogMetricRequest{
				RunID:     "1",
				Key:       "key",
				Step:      1,
				Value:     common.NANValue,
				Timestamp: 1234567890,
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDAndRunID",
					context.TODO(),
					uint(1),
					"1",
				).Return(&models.Run{
					ID: "1",
				}, nil)
				metricRepository := repositories.MockMetricRepositoryProvider{}
				metricRepository.On(
					"CreateBatch",
					context.TODO(),
					mock.MatchedBy(func(run *models.Run) bool {
						assert.Equal(t, "1", run.ID)
						return true
					}),
					1,
					mock.MatchedBy(func(metrics []models.Metric) bool {
						assert.Equal(t, 1, len(metrics))
						assert.Equal(t, "key", metrics[0].Key)
						assert.Equal(t, float64(0), metrics[0].Value)
						assert.Equal(t, true, metrics[0].IsNan)
						assert.Equal(t, int64(1), metrics[0].Step)
						assert.Equal(t, int64(1234567890), metrics[0].Timestamp)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&metricRepository,
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().LogMetric(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_LogParam_Ok(t *testing.T) {
	// init repository mocks.
	runRepository := repositories.MockRunRepositoryProvider{}
	runRepository.On(
		"GetByNamespaceIDRunIDAndLifecycleStage",
		context.TODO(),
		uint(1),
		"1",
		models.LifecycleStageActive,
	).Return(&models.Run{
		ID:             "1",
		LifecycleStage: models.LifecycleStageActive,
	}, nil)
	paramRepository := repositories.MockParamRepositoryProvider{}
	paramRepository.On(
		"CreateBatch",
		context.TODO(),
		1,
		mock.MatchedBy(func(params []models.Param) bool {
			assert.Equal(t, "1", params[0].RunID)
			assert.Equal(t, "key", params[0].Key)
			assert.Equal(t, "value", params[0].ValueString())
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
		&repositories.MockLogRepositoryProvider{},
		&repositories.MockArtifactRepositoryProvider{},
	)
	err := service.LogParam(context.TODO(), &models.Namespace{
		ID: 1,
	}, &request.LogParamRequest{
		RunID:    "1",
		Key:      "key",
		ValueStr: common.GetPointer("value"),
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_LogParam_Error(t *testing.T) {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogParamRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
			request: &request.LogParamRequest{},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "EmptyOrIncorrectMetricKey",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'key'`),
			request: &request.LogParamRequest{
				RunID: "1",
			},
			service: func() *Service {
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockRunRepositoryProvider{},
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "RunNotFoundDatabaseError",
			error: api.NewInternalError(`Unable to find run '1': database error`),
			request: &request.LogParamRequest{
				RunID:    "1",
				Key:      "key",
				ValueStr: common.GetPointer("value"),
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "NoActiveRunFound",
			error: api.NewResourceDoesNotExistError("Run '1' not found"),
			request: &request.LogParamRequest{
				RunID:    "1",
				Key:      "key",
				ValueStr: common.GetPointer("value"),
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(nil, nil)
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&repositories.MockParamRepositoryProvider{},
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "LogParamDatabaseError",
			error: api.NewInternalError(`unable to insert params for run '1': database error`),
			request: &request.LogParamRequest{
				RunID:    "1",
				Key:      "key",
				ValueStr: common.GetPointer("value"),
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					1,
					mock.MatchedBy(func(params []models.Param) bool {
						assert.Equal(t, 1, len(params))
						assert.Equal(t, "key", params[0].Key)
						assert.Equal(t, "value", params[0].ValueString())
						assert.Equal(t, "1", params[0].RunID)
						return true
					}),
				).Return(errors.New("database error"))
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
		{
			name:  "LogParamConflictError",
			error: api.NewInvalidParameterValueError(`unable to insert params for run '1': conflict!`),
			request: &request.LogParamRequest{
				RunID:    "1",
				Key:      "key",
				ValueStr: common.GetPointer("value"),
			},
			service: func() *Service {
				runRepository := repositories.MockRunRepositoryProvider{}
				runRepository.On(
					"GetByNamespaceIDRunIDAndLifecycleStage",
					context.TODO(),
					uint(1),
					"1",
					models.LifecycleStageActive,
				).Return(&models.Run{
					ID:             "1",
					LifecycleStage: models.LifecycleStageActive,
				}, nil)
				paramRepository := repositories.MockParamRepositoryProvider{}
				paramRepository.On(
					"CreateBatch",
					context.TODO(),
					1,
					mock.MatchedBy(func(params []models.Param) bool {
						assert.Equal(t, 1, len(params))
						assert.Equal(t, "key", params[0].Key)
						assert.Equal(t, "value", params[0].ValueString())
						assert.Equal(t, "1", params[0].RunID)
						return true
					}),
				).Return(repositories.ParamConflictError{Message: "conflict!"})
				return NewService(
					&repositories.MockTagRepositoryProvider{},
					&runRepository,
					&paramRepository,
					&repositories.MockMetricRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
					&repositories.MockLogRepositoryProvider{},
					&repositories.MockArtifactRepositoryProvider{},
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			err := tt.service().LogParam(context.TODO(), &models.Namespace{ID: 1}, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}
