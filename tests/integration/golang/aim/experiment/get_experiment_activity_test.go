//go:build integration

package experiment

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentActivityTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetExperimentActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentActivityTestSuite))
}

func (s *GetExperimentActivityTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *GetExperimentActivityTestSuite) Test_Ok() {
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
	var resp response.GetExperimentActivity
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"/experiments/%d/activity", *experiment.ID,
		),
		&resp,
	)
	activity := map[string]int{}
	for _, r := range runs {
		key := time.UnixMilli(r.StartTime.Int64).Format("2006-01-02T15:00:00")
		activity[key] += 1
	}

	assert.Equal(s.T(), resp.NumRuns, len(runs))
	assert.Equal(s.T(), resp.NumArchivedRuns, 0)
	assert.Equal(s.T(), resp.NumActiveRuns, len(runs))
	assert.Equal(s.T(), resp.ActivityMap, activity)
}

func (s *GetExperimentActivityTestSuite) Test_Error() {
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
