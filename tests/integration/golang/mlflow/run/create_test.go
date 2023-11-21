//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
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
	helpers.BaseTestSuite
}

func TestCreateRunTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRunTestSuite))
}

func (s *CreateRunTestSuite) Test_DefaultNamespace_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	s.successCases(namespace, experiment, false, *experiment.ID)
}

func (s *CreateRunTestSuite) Test_DefaultNamespaceExperimentZero_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// update default experiment id for namespace.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	s.Require().Nil(err)

	s.successCases(namespace, experiment, false, int32(0))
}

func (s *CreateRunTestSuite) Test_CustomNamespace_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	s.successCases(namespace, experiment, true, *experiment.ID)
}

func (s *CreateRunTestSuite) Test_CustomNamespaceExperimentZero_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// update default experiment id.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	s.Require().Nil(err)

	s.successCases(namespace, experiment, true, int32(0))
}

func (s *CreateRunTestSuite) successCases(
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
	client := s.MlflowClient().WithMethod(
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
	s.Require().Nil(
		client.DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute,
		),
	)
	s.NotEmpty(resp.Run.Info.ID)
	s.NotEmpty(resp.Run.Info.UUID)
	s.Equal("TestRun", resp.Run.Info.Name)
	s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.Run.Info.ExperimentID)
	s.Equal(int64(1234567890), resp.Run.Info.StartTime)
	s.Equal(int64(0), resp.Run.Info.EndTime)
	s.Equal(string(models.StatusRunning), resp.Run.Info.Status)
	s.NotEmpty(resp.Run.Info.ArtifactURI)
	s.Equal(string(models.LifecycleStageActive), resp.Run.Info.LifecycleStage)
	s.Equal([]response.RunTagPartialResponse{
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
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// set namespace default experiment.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	s.Require().Nil(err)

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
			name: "CreateRunWithNotExistingExperiment",
			request: request.CreateRunRequest{
				ExperimentID: fmt.Sprintf("%d", -1),
			},
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					`unable to find experiment with id '%d': error getting experiment by id: %d: record not found`,
					-1,
					-1,
				),
			),
		},
		{
			name: "CreateRunWithExistingNamespaceAndNotExistingExperiment",
			request: request.CreateRunRequest{
				ExperimentID: fmt.Sprintf("%d", -1),
			},
			error: api.NewResourceDoesNotExistError(
				fmt.Sprintf(
					`unable to find experiment with id '%d': error getting experiment by id: %d: record not found`,
					-1,
					-1,
				),
			),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			client := s.MlflowClient().WithMethod(
				http.MethodPost,
			).WithRequest(
				tt.request,
			).WithNamespace(
				tt.namespace,
			).WithResponse(
				&resp,
			)
			s.Require().Nil(
				client.DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
