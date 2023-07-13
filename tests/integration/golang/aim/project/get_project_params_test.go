//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectParamsTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetProjectParamsTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectParamsTestSuite))
}

func (s *GetProjectParamsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
}

func (s *GetProjectParamsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// 1. create test `experiment` and connect test `run`.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	assert.Nil(s.T(), err)

	// 2. create latest metric.
	metric, err := s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key",
		Value:     123.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
	})
	assert.Nil(s.T(), err)

	// 3. check that response contains metric from previous step.
	resp := response.ProjectParamsResponse{}
	err = s.client.DoGetRequest(
		"/projects/params?sequence=metric",
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(resp.Metric))
	_, ok := resp.Metric[metric.Key]
	assert.True(s.T(), ok)

	// 4. mark run as `deleted`.
	run.LifecycleStage = models.LifecycleStageDeleted
	assert.Nil(s.T(), s.runFixtures.UpdateRun(context.Background(), run))

	// 5. check that endpoint returns an empty response.
	resp = response.ProjectParamsResponse{}
	err = s.client.DoGetRequest(
		"/projects/params?sequence=metric",
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(resp.Metric))
	_, ok = resp.Metric[metric.Key]
	assert.False(s.T(), ok)
}

func (s *GetProjectParamsTestSuite) Test_Error() {}
