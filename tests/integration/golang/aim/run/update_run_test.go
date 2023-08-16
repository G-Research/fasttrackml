//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	run                *models.Run
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	s.run, err = s.runFixtures.CreateExampleRun(context.Background(), exp)
	assert.Nil(s.T(), err)
}

func (s *UpdateRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name    string
		request request.UpdateRunRequest
	}{
		{
			name: "UpdateOneRun",
			request: request.UpdateRunRequest{
				RunID:    &(s.run.ID),
				Name:     common.GetPointer(fmt.Sprintf("%v%v", s.run.Name, "-new")),
				Status:   common.GetPointer(string(models.StatusFinished)),
				Archived: common.GetPointer(true),
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Success
			err := s.client.DoPutRequest(
				fmt.Sprintf("/runs/%s", *tt.request.RunID),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			run, err := s.runFixtures.GetRun(context.Background(), s.run.ID)
			assert.Nil(s.T(), err)
			// TODO the PUT endpoint only updates LifecycleStage
			// assert.Equal(t, newName, run.Name)
			// assert.Equal(t, models.Status(newStatus), run.Status)
			assert.Equal(s.T(), models.LifecycleStageDeleted, run.LifecycleStage)
		})
	}
}

func (s *UpdateRunTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody any
	}{
		{
			name: "UpdateRunWithIncorrectArchived",
			requestBody: map[string]any{
				"Archived": "this-cannot-unmarshal",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.client.DoPutRequest(
				fmt.Sprintf("/runs/%s", s.run.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
