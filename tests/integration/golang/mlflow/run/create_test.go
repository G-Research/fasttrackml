//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateRunTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestCreateRunTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRunTestSuite))
}

func (s *CreateRunTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *CreateRunTestSuite) Test_DefaultNamespace_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	s.testCases(namespace, experiment, false, *experiment.ID)
}

func (s *CreateRunTestSuite) Test_DefaultNamespaceExperimentZero_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// update default experiment id for namespace.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	assert.Nil(s.T(), err)

	s.testCases(namespace, experiment, false, int32(0))
}

func (s *CreateRunTestSuite) Test_CustomNamespace_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	s.testCases(namespace, experiment, true, *experiment.ID)
}

func (s *CreateRunTestSuite) Test_CustomNamespaceExperimentZero_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// update default experiment id.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	assert.Nil(s.T(), err)

	s.testCases(namespace, experiment, true, int32(0))
}

func (s *CreateRunTestSuite) testCases(
	namespace *models.Namespace,
	experiment *models.Experiment,
	useNamespaceInRequest bool,
	experimentIDInRequest int32,
) {
	req := request.CreateRunRequest{
		Name: "TestRun",
		Tags: []request.RunTagPartialRequest{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2",
			},
		},
		StartTime:    1234567890,
		ExperimentID: fmt.Sprintf("%d", experimentIDInRequest),
	}

	resp := response.CreateRunResponse{}
	client := s.MlflowClient.WithMethod(
		http.MethodPost,
	).WithRequest(
		req,
	).WithResponse(
		&resp,
	)
	if useNamespaceInRequest {
		client = client.WithNamespace(
			namespace.Code,
		)
	}
	assert.Nil(
		s.T(),
		client.DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute),
		),
	)
	assert.NotEmpty(s.T(), resp.Run.Info.ID)
	assert.NotEmpty(s.T(), resp.Run.Info.UUID)
	assert.Equal(s.T(), "TestRun", resp.Run.Info.Name)
	assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.Run.Info.ExperimentID)
	assert.Equal(s.T(), int64(1234567890), resp.Run.Info.StartTime)
	assert.Equal(s.T(), int64(0), resp.Run.Info.EndTime)
	assert.Equal(s.T(), string(models.StatusRunning), resp.Run.Info.Status)
	assert.NotEmpty(s.T(), resp.Run.Info.ArtifactURI)
	assert.Equal(s.T(), string(models.LifecycleStageActive), resp.Run.Info.LifecycleStage)
	assert.Equal(s.T(), []response.RunTagPartialResponse{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}, resp.Run.Data.Tags)
}

func (s *CreateRunTestSuite) Test_Error() {
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// set namespace default experiment.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	assert.Nil(s.T(), err)
	nonExistingExperimentID := *experiment.ID + 1

	tests := []struct {
		name      string
		error     *api.ErrorResponse
		namespace string
		request   request.CreateRunRequest
	}{
		{
			name: "CreateRunWithInvalidExperimentID",
			request: request.CreateRunRequest{
				ExperimentID: "invalid_experiment_id",
			},
			error: api.NewBadRequestError(
				`unable to parse experiment id 'invalid_experiment_id': strconv.ParseInt: ` +
					`parsing "invalid_experiment_id": invalid syntax`,
			),
		},
		{
			name:      "CreateRunWithNotExistingNamespaceAndExistingExperimentID",
			namespace: "not_existing_namespace",
			request: request.CreateRunRequest{
				ExperimentID: fmt.Sprintf("%d", *experiment.ID),
			},
			error: api.NewResourceDoesNotExistError(
				`unable to find namespace with code: not_existing_namespace`,
			),
		},
		{
			name:      "CreateRunWithNotExistingExperiment",
			namespace: "default",
			request: request.CreateRunRequest{
				ExperimentID: fmt.Sprintf("%d", nonExistingExperimentID),
			},
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					`unable to find experiment with id '%d': error getting experiment by id: %d: record not found`,
					nonExistingExperimentID,
					nonExistingExperimentID,
				),
			),
		},
		{
			name:      "CreateRunWithExistingNamespaceAndNotExistingExperiment",
			namespace: "default",
			request: request.CreateRunRequest{
				ExperimentID: fmt.Sprintf("%d", nonExistingExperimentID),
			},
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					`unable to find experiment with id '%d': error getting experiment by id: %d: record not found`,
					nonExistingExperimentID,
					nonExistingExperimentID,
				),
			),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			client := s.MlflowClient.WithMethod(
				http.MethodPost,
			).WithRequest(
				tt.request,
			).WithResponse(
				&resp,
			)
			if tt.namespace != "" {
				client = client.WithNamespace(
					tt.namespace,
				)
			}
			assert.Nil(
				s.T(),
				client.DoRequest(
					fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute),
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
