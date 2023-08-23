//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectParamsTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetProjectParamsTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectParamsTestSuite))
}

func (s *GetProjectParamsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetProjectParamsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.ExperimentFixtures.UnloadFixtures())
	}()

	// 1. create test `experiment` and connect test `run`.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	assert.Nil(s.T(), err)

	// 2. create latest metric.
	metric, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key",
		Value:     123.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
	})
	assert.Nil(s.T(), err)

	// 3. create test param and tag.
	tag, err := s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	param, err := s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	// 3. check that response contains metric from previous step.
	resp := response.ProjectParamsResponse{}
	err = s.AIMClient.DoGetRequest(
		"/projects/params?sequence=metric",
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(resp.Metric))
	_, ok := resp.Metric[metric.Key]
	assert.True(s.T(), ok)
	assert.Equal(s.T(), map[string]interface{}{
		param.Key: map[string]interface{}{
			"__example_type__": "<class 'str'>",
		},
		"tags": map[string]interface{}{
			tag.Key: map[string]interface{}{
				"__example_type__": "<class 'str'>",
			},
		},
	}, resp.Params)

	// 4. mark run as `deleted`.
	run.LifecycleStage = models.LifecycleStageDeleted
	assert.Nil(s.T(), s.RunFixtures.UpdateRun(context.Background(), run))

	// 5. check that endpoint returns an empty response.
	resp = response.ProjectParamsResponse{}
	err = s.AIMClient.DoGetRequest(
		"/projects/params?sequence=metric",
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(resp.Metric))
	_, ok = resp.Metric[metric.Key]
	assert.False(s.T(), ok)
	assert.Equal(s.T(), map[string]interface{}{"tags": map[string]interface{}{}}, resp.Params)
}

func (s *GetProjectParamsTestSuite) Test_Error() {}
