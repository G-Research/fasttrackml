//go:build integration

package run

import (
	"context"
	"fmt"
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
	runs               []*models.Run
	activity           response.ProjectActivity
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

	s.runs, err = s.runFixtures.CreateRuns(context.Background(), exp, 5)
	assert.Nil(s.T(), err)

}

func (s *GetProjectActivityTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.projectFixtures.UnloadFixtures())
	}()
	var resp response.ProjectActivity
	err := s.client.DoGetRequest(
		fmt.Sprintf("/projects/activity"),
		&resp,
	)
	assert.Nil(s.T(), err)

	activity, err := s.projectFixtures.GetProjectActivity(context.Background())
	s.activity = *activity

	assert.Equal(s.T(), s.activity.NumActiveRuns, resp.NumActiveRuns)
	assert.Equal(s.T(), s.activity.NumArchivedRuns, resp.NumArchivedRuns)
	assert.Equal(s.T(), s.activity.NumExperiments, resp.NumExperiments)
	assert.Equal(s.T(), s.activity.NumRuns, resp.NumRuns)
}
