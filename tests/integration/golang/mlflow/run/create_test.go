//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateRunTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestCreateRunTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRunTestSuite))
}

func (s *CreateRunTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func (s *CreateRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// create test experiment.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	resp := response.CreateRunResponse{}
	err = s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute),
		request.CreateRunRequest{
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
			ExperimentID: fmt.Sprintf("%d", *experiment.ID),
		},
		&resp,
	)
	assert.Nil(s.T(), err)
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
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.CreateRunRequest
	}{
		{
			name: "CreateRunWithInvalidExperimentID",
			request: request.CreateRunRequest{
				ExperimentID: "invalid_experiment_id",
			},
			error: api.NewBadRequestError(
				`unable to parse experiment id 'invalid_experiment_id': strconv.ParseInt: parsing "invalid_experiment_id": invalid syntax`,
			),
		},
		{
			name: "CreateRunWithNotExistingExperiment",
			request: request.CreateRunRequest{
				ExperimentID: "1",
			},
			error: api.NewResourceDoesNotExistError(
				`unable to find experiment with id '1': error getting experiment by id: 1: record not found`,
			),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
