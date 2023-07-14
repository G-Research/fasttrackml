//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunInfoTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	run                *models.Run
}

func TestGetRunInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunInfoTestSuite))
}

func (s *GetRunInfoTestSuite) SetupTest() {
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

	runs, err := s.runFixtures.CreateRuns(context.Background(), exp, 1)
	assert.Nil(s.T(), err)
	s.run = runs[0]
}

func (s *GetRunInfoTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		runID            string          
	}{
		{
			name:             "GetOneRun",
			runID:          s.run.ID,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.GetRunInfoResponse
			err := s.client.DoGetRequest(
				fmt.Sprintf("/runs/%s/info", tt.runID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), s.run.Name, resp.Props.Name)
		})
	}
}
