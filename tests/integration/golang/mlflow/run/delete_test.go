//go:build integration

package run

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	run                *models.Run
}

func TestDeleteRunTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTestSuite))
}

func (s *DeleteRunTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	// create experiment
	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	// create run for the experiment
	runs, err := s.runFixtures.CreateTestRuns(context.Background(), exp, 1)
	assert.Nil(s.T(), err)
	s.run = runs[0]
}

func (s *DeleteRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name    string
		request request.DeleteRunRequest
	}{
		{
			name:    "DeleteRunSucceedsWithExistingRunID",
			request: request.DeleteRunRequest{RunID: s.run.ID},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := map[string]any{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)

			archivedRuns, err := s.runFixtures.GetTestRuns(context.Background(), s.run.ExperimentID)

			assert.Nil(s.T(), err)
			assert.Equal(s.T(), s.run.ID, archivedRuns[0].ID)
			assert.Equal(s.T(), models.LifecycleStageDeleted, archivedRuns[0].LifecycleStage)
		})
	}
}

func (s *DeleteRunTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name    string
		request request.DeleteRunRequest
	}{
		{
			name:    "DeleteRunFailsWithNonExistingRunID",
			request: request.DeleteRunRequest{RunID: "not-an-id"},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "RESOURCE_DOES_NOT_EXIST: unable to find run 'not-an-id': error getting `run` entity by id: not-an-id: record not found", resp.Error())

		})
	}
}
