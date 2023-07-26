//go:build integration

package experiment

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentRunsTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetExperimentRunsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentRunsTestSuite))
}

func (s *GetExperimentRunsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *GetExperimentRunsTestSuite) Test_Ok() {
	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	runs, err := s.runFixtures.CreateRuns(context.Background(), exp, 10)
	assert.Nil(s.T(), err)
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	offset := 8
	var resp response.GetExperimentRuns
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"/experiments/%d/runs?limit=4&offset=%s", *experiment.ID, runs[offset].ID,
		),
		&resp,
	)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), 4, len(resp.Runs))
	assert.Equal(s.T(), runs[offset-1].ID, resp.Runs[0].ID)
	assert.Equal(s.T(), runs[offset-2].ID, resp.Runs[1].ID)
	assert.Equal(s.T(), runs[offset-3].ID, resp.Runs[2].ID)
	assert.Equal(s.T(), runs[offset-4].ID, resp.Runs[3].ID)
}

func (s *GetExperimentRunsTestSuite) Test_Error() {
	tests := []struct {
		name string
		ID   string
	}{
		{
			name: "GetInvalidExperimentID",
			ID:   "123",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp api.ErrorResponse
			err := s.client.DoGetRequest(
				fmt.Sprintf(
					"/experiments/%s/runs?limit=4", tt.ID,
				),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Error(), "Not Found")
		})
	}
}
