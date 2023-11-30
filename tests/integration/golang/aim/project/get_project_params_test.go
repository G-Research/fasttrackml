//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectParamsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectParamsTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectParamsTestSuite))
}

func (s *GetProjectParamsTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// 1. create test `namespace` and connect test `run`.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	// 2. create test `experiment` and connect test `run`.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	s.Require().Nil(err)

	// 3. create latest metric.
	metric, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key",
		Value:     123.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)

	// 4. create test param and tag.
	tag, err := s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	s.Require().Nil(err)

	param, err := s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run.ID,
	})
	s.Require().Nil(err)

	// 5. check that response contains metric from previous step.
	resp := response.ProjectParamsResponse{}
	s.Require().Nil(
		s.AIMClient().WithQuery(
			map[any]any{"sequence": "metric"},
		).WithResponse(
			&resp,
		).DoRequest("/projects/params"),
	)

	s.Equal(1, len(resp.Metric))
	_, ok := resp.Metric[metric.Key]
	s.True(ok)
	s.Equal(map[string]interface{}{
		param.Key: map[string]interface{}{
			"__example_type__": "<class 'str'>",
		},
		"tags": map[string]interface{}{
			tag.Key: map[string]interface{}{
				"__example_type__": "<class 'str'>",
			},
		},
	}, resp.Params)

	// 6. mark run as `deleted`.
	run.LifecycleStage = models.LifecycleStageDeleted
	s.Require().Nil(s.RunFixtures.UpdateRun(context.Background(), run))

	// 7. check that endpoint returns an empty response.
	resp = response.ProjectParamsResponse{}
	s.Require().Nil(
		s.AIMClient().WithQuery(
			map[any]any{"sequence": "metric"},
		).WithResponse(
			&resp,
		).DoRequest("/projects/params"),
	)
	s.Equal(0, len(resp.Metric))
	_, ok = resp.Metric[metric.Key]
	s.False(ok)
	s.Equal(map[string]interface{}{"tags": map[string]interface{}{}}, resp.Params)
}

func (s *GetProjectParamsTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
}
