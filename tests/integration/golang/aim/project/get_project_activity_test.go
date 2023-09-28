//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectActivityTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	projectFixtures    *fixtures.ProjectFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	runFixtures        *fixtures.RunFixtures
}

func TestGetProjectActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectActivityTestSuite))
}

func (s *GetProjectActivityTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	projectFixtures, err := fixtures.NewProjectFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.projectFixtures = projectFixtures

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

	_, err = s.runFixtures.CreateExampleRuns(context.Background(), exp, 5)
	assert.Nil(s.T(), err)
}

func (s *GetProjectActivityTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.projectFixtures.UnloadFixtures())
	}()
	var resp response.ProjectActivityResponse
	err := s.client.DoGetRequest(
		"/projects/activity",
		&resp,
	)
	assert.Nil(s.T(), err)

	activity, err := s.projectFixtures.GetProjectActivity(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), activity.NumActiveRuns, resp.NumActiveRuns)
	assert.Equal(s.T(), activity.NumArchivedRuns, resp.NumArchivedRuns)
	assert.Equal(s.T(), activity.NumExperiments, resp.NumExperiments)
	assert.Equal(s.T(), activity.NumRuns, resp.NumRuns)
}
