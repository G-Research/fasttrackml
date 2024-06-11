package experiment

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
	"github.com/G-Research/fasttrackml/pkg/common/config"
)

func TestService_CreateExperiment_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
	).Return(nil, nil)
	experimentRepository.On(
		"Create", context.TODO(), mock.Anything,
	).Return(nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	experiment, err := service.CreateExperiment(context.TODO(), &ns, &request.CreateExperimentRequest{
		Name: "name",
		Tags: []request.ExperimentTagPartialRequest{
			{
				Key:   "key",
				Value: "value",
			},
		},
		ArtifactLocation: "/artifact/location",
	})

	// compare results.
	require.Nil(t, err)
	assert.Equal(t, "name", experiment.Name)
	assert.Equal(t, []models.ExperimentTag{
		{
			Key:   "key",
			Value: "value",
		},
	}, experiment.Tags)
	assert.Equal(t, "/artifact/location", experiment.ArtifactLocation)
	assert.Equal(t, models.LifecycleStageActive, experiment.LifecycleStage)
	assert.NotEmpty(t, experiment.CreationTime.Int64)
	assert.NotEmpty(t, experiment.LastUpdateTime.Int64)
}

func TestService_CreateExperiment_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.CreateExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyOrIncorrectExperimentName",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'name'"),
			request: &request.CreateExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "EmptyOrIncorrectArtifactLocation",
			error: api.NewInvalidParameterValueError(
				`Invalid value for parameter 'artifact_location': error parsing artifact location: parse ` +
					`"incorrect-protocol,:/incorrect-location": first path segment in URL cannot contain colon`,
			),
			request: &request.CreateExperimentRequest{
				Name:             "name",
				ArtifactLocation: "incorrect-protocol,:/incorrect-location",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, nil)
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "GettingExperimentByNameDatabaseError",
			error: api.NewInternalError(`error getting experiment with name: 'name', error: database error`),
			request: &request.CreateExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "ExperimentWithProvidedNameExists",
			error: api.NewResourceAlreadyExistsError(`experiment(name=name) already exists`),
			request: &request.CreateExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(&models.Experiment{}, nil)
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "CreateExperimentDatabaseError",
			error: api.NewInternalError(`error inserting experiment 'name': database error`),
			request: &request.CreateExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, nil)
				experimentRepository.On(
					"Create", context.TODO(), mock.Anything,
				).Return(errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "UpdateExperimentArtifactLocationDatabaseError",
			error: api.NewInternalError(`error updating artifact_location for experiment 'name': database error`),
			request: &request.CreateExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, nil)
				experimentRepository.On(
					"Create",
					context.TODO(),
					mock.MatchedBy(func(experiment *models.Experiment) bool {
						experiment.ID = common.GetPointer(int32(1))
						return true
					}),
				).Return(nil)
				experimentRepository.On(
					"Update", context.TODO(), mock.AnythingOfType("*models.Experiment"),
				).Return(errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().CreateExperiment(context.TODO(), &ns, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_DeleteExperiment_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
	).Return(&models.Experiment{
		ID: common.GetPointer(int32(1)),
	}, nil)
	experimentRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(experiment *models.Experiment) bool {
			assert.Equal(t, experiment.LifecycleStage, models.LifecycleStageDeleted)
			assert.NotNil(t, experiment.LastUpdateTime)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	err := service.DeleteExperiment(context.TODO(), &ns, &request.DeleteExperimentRequest{
		ID: "1",
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_DeleteExperiment_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:                  1,
		Code:                "code",
		DefaultExperimentID: common.GetPointer(int32(0)),
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_id'`),
			request: &request.DeleteExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'incorrect_id': strconv.ParseInt: parsing "incorrect_id": invalid syntax`,
			),
			request: &request.DeleteExperimentRequest{
				ID: "incorrect_id",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': experiment not found`),
			request: &request.DeleteExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(nil, errors.New("experiment not found"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "DefaultExperiment",
			error: api.NewBadRequestError("unable to delete default experiment"),
			request: &request.DeleteExperimentRequest{
				ID: "0",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(0),
				).Return(&models.Experiment{
					ID:   common.GetPointer(models.DefaultExperimentID),
					Name: models.DefaultExperimentName,
				}, nil)
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "UpdateExperimentDatabaseError",
			error: api.NewInternalError(`unable to delete experiment '1': database error`),
			request: &request.DeleteExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(&models.Experiment{
					ID: common.GetPointer(int32(1)),
				}, nil)
				experimentRepository.On(
					"Update", context.TODO(), mock.AnythingOfType("*models.Experiment"),
				).Return(errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			assert.Equal(t, tt.error, tt.service().DeleteExperiment(context.TODO(), &ns, tt.request))
		})
	}
}

func TestService_GetExperiment_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
	).Return(&models.Experiment{
		ID:   common.GetPointer(int32(1)),
		Name: "name",
		Tags: []models.ExperimentTag{
			{
				Key:   "key",
				Value: "value",
			},
		},
		Runs: []models.Run{
			{
				ID: "1",
			},
		},
		LifecycleStage:   models.LifecycleStageActive,
		CreationTime:     sql.NullInt64{Int64: 1234567890, Valid: true},
		LastUpdateTime:   sql.NullInt64{Int64: 1234567891, Valid: true},
		ArtifactLocation: "/artifact/location",
	}, nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	experiment, err := service.GetExperiment(context.TODO(), &ns, &request.GetExperimentRequest{
		ID: "1",
	})

	// compare results.
	require.Nil(t, err)
	assert.Equal(t, int32(1), *experiment.ID)
	assert.Equal(t, "name", experiment.Name)
	assert.Equal(t, []models.ExperimentTag{
		{
			Key:   "key",
			Value: "value",
		},
	}, experiment.Tags)
	assert.Equal(t, []models.Run{
		{
			ID: "1",
		},
	}, experiment.Runs)
	assert.Equal(t, models.LifecycleStageActive, experiment.LifecycleStage)
	assert.Equal(t, sql.NullInt64{Int64: 1234567890, Valid: true}, experiment.CreationTime)
	assert.Equal(t, sql.NullInt64{Int64: 1234567891, Valid: true}, experiment.LastUpdateTime)
	assert.Equal(t, "/artifact/location", experiment.ArtifactLocation)
}

func TestService_GetExperiment_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_id'`),
			request: &request.GetExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'incorrect_id': strconv.ParseInt: parsing "incorrect_id": invalid syntax`,
			),
			request: &request.GetExperimentRequest{
				ID: "incorrect_id",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': experiment not found`),
			request: &request.GetExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(nil, errors.New("experiment not found"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().GetExperiment(context.TODO(), &ns, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_GetExperimentByName_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
	).Return(&models.Experiment{
		ID:   common.GetPointer(int32(1)),
		Name: "name",
		Tags: []models.ExperimentTag{
			{
				Key:   "key",
				Value: "value",
			},
		},
		Runs: []models.Run{
			{
				ID: "1",
			},
		},
		LifecycleStage:   models.LifecycleStageActive,
		CreationTime:     sql.NullInt64{Int64: 1234567890, Valid: true},
		LastUpdateTime:   sql.NullInt64{Int64: 1234567891, Valid: true},
		ArtifactLocation: "/artifact/location",
	}, nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	experiment, err := service.GetExperimentByName(
		context.TODO(),
		&ns,
		&request.GetExperimentRequest{
			Name: "name",
		},
	)

	// compare results.
	require.Nil(t, err)
	assert.Equal(t, int32(1), *experiment.ID)
	assert.Equal(t, "name", experiment.Name)
	assert.Equal(t, []models.ExperimentTag{
		{
			Key:   "key",
			Value: "value",
		},
	}, experiment.Tags)
	assert.Equal(t, []models.Run{
		{
			ID: "1",
		},
	}, experiment.Runs)
	assert.Equal(t, models.LifecycleStageActive, experiment.LifecycleStage)
	assert.Equal(t, sql.NullInt64{Int64: 1234567890, Valid: true}, experiment.CreationTime)
	assert.Equal(t, sql.NullInt64{Int64: 1234567891, Valid: true}, experiment.LastUpdateTime)
	assert.Equal(t, "/artifact/location", experiment.ArtifactLocation)
}

func TestService_GetExperimentByName_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentName",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_name'`),
			request: &request.GetExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "GetExperimentByNameDatabaseError",
			error: api.NewInternalError(`unable to get experiment by name 'name': database error`),
			request: &request.GetExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment 'name'`),
			request: &request.GetExperimentRequest{
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndName", context.TODO(), ns.ID, "name",
				).Return(nil, nil)
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			_, err := tt.service().GetExperimentByName(context.TODO(), &ns, tt.request)
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestService_RestoreExperiment_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
	).Return(&models.Experiment{
		ID: common.GetPointer(int32(1)),
	}, nil)
	experimentRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(experiment *models.Experiment) bool {
			assert.Equal(t, models.LifecycleStageActive, experiment.LifecycleStage)
			assert.NotNil(t, experiment.LastUpdateTime)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	err := service.RestoreExperiment(context.TODO(), &ns, &request.RestoreExperimentRequest{
		ID: "1",
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_RestoreExperiment_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_id'`),
			request: &request.RestoreExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`Unable to parse experiment id 'incorrect_id': strconv.ParseInt: parsing "incorrect_id": invalid syntax`,
			),
			request: &request.RestoreExperimentRequest{
				ID: "incorrect_id",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': experiment not found`),
			request: &request.RestoreExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(nil, errors.New("experiment not found"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "UpdateExperimentDatabaseError",
			error: api.NewInternalError(`Unable to restore experiment '1': database error`),
			request: &request.RestoreExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(&models.Experiment{
					ID: common.GetPointer(int32(1)),
				}, nil)
				experimentRepository.On(
					"Update", context.TODO(), mock.AnythingOfType("*models.Experiment"),
				).Return(errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			assert.Equal(t, tt.error, tt.service().RestoreExperiment(context.TODO(), &ns, tt.request))
		})
	}
}

func TestService_SetExperimentTag_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
	).Return(&models.Experiment{
		ID: common.GetPointer(int32(1)),
	}, nil)

	tagsRepository := repositories.MockTagRepositoryProvider{}
	tagsRepository.On(
		"CreateExperimentTag",
		context.TODO(),
		mock.MatchedBy(func(tag *models.ExperimentTag) bool {
			assert.Equal(t, "key", tag.Key)
			assert.Equal(t, "value", tag.Value)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&tagsRepository,
		&experimentRepository,
	)
	err := service.SetExperimentTag(context.TODO(), &ns, &request.SetExperimentTagRequest{
		ID:    "1",
		Key:   "key",
		Value: "value",
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_SetExperimentTag_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetExperimentTagRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_id'`),
			request: &request.SetExperimentTagRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "EmptyOrIncorrectTagKey",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'key'`),
			request: &request.SetExperimentTagRequest{
				ID: "1",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`Unable to parse experiment id 'incorrect_id': strconv.ParseInt: parsing "incorrect_id": invalid syntax`,
			),
			request: &request.SetExperimentTagRequest{
				ID:  "incorrect_id",
				Key: "key",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': experiment not found`),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "key",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(nil, errors.New("experiment not found"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "SetExperimentTagDatabaseError",
			error: api.NewInternalError(`Unable to set tag for experiment '1': database error`),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "key",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(&models.Experiment{
					ID: common.GetPointer(int32(1)),
				}, nil)
				tagRepository := repositories.MockTagRepositoryProvider{}
				tagRepository.On(
					"CreateExperimentTag",
					context.TODO(),
					mock.AnythingOfType("*models.ExperimentTag"),
				).Return(errors.New("database error"))

				return NewService(
					&config.Config{},
					&tagRepository,
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			assert.Equal(t, tt.error, tt.service().SetExperimentTag(context.TODO(), &ns, tt.request))
		})
	}
}

func TestService_UpdateExperiment_Ok(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
	).Return(&models.Experiment{
		ID: common.GetPointer(int32(1)),
	}, nil)
	experimentRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(experiment *models.Experiment) bool {
			assert.Equal(t, "name", experiment.Name)
			assert.NotNil(t, experiment.LastUpdateTime)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&config.Config{},
		&repositories.MockTagRepositoryProvider{},
		&experimentRepository,
	)
	err := service.UpdateExperiment(context.TODO(), &ns, &request.UpdateExperimentRequest{
		ID:   "1",
		Name: "name",
	})

	// compare results.
	require.Nil(t, err)
}

func TestService_UpdateExperiment_Error(t *testing.T) {
	// initialise namespace to which experiment under the test belongs to.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateExperimentRequest
		service func() *Service
	}{
		{
			name:    "EmptyExperimentID",
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_id'`),
			request: &request.UpdateExperimentRequest{},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "EmptyExperimentName",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'new_name'`),
			request: &request.UpdateExperimentRequest{
				ID: "1",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'incorrect_id': strconv.ParseInt: parsing "incorrect_id": invalid syntax`,
			),
			request: &request.UpdateExperimentRequest{
				ID:   "incorrect_id",
				Name: "name",
			},
			service: func() *Service {
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&repositories.MockExperimentRepositoryProvider{},
				)
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': experiment not found`),
			request: &request.UpdateExperimentRequest{
				ID:   "1",
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(nil, errors.New("experiment not found"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
		{
			name:  "UpdateExperimentDatabaseError",
			error: api.NewInternalError(`unable to update experiment '1': database error`),
			request: &request.UpdateExperimentRequest{
				ID:   "1",
				Name: "name",
			},
			service: func() *Service {
				experimentRepository := repositories.MockExperimentRepositoryProvider{}
				experimentRepository.On(
					"GetByNamespaceIDAndExperimentID", context.TODO(), ns.ID, int32(1),
				).Return(&models.Experiment{
					ID: common.GetPointer(int32(1)),
				}, nil)
				experimentRepository.On(
					"Update", context.TODO(), mock.AnythingOfType("*models.Experiment"),
				).Return(errors.New("database error"))
				return NewService(
					&config.Config{},
					&repositories.MockTagRepositoryProvider{},
					&experimentRepository,
				)
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// call service under testing.
			assert.Equal(t, tt.error, tt.service().UpdateExperiment(context.TODO(), &ns, tt.request))
		})
	}
}
