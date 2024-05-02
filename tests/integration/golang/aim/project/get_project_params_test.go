package run

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
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
	// create test run.
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *s.DefaultExperiment.ID,
	})
	s.Require().Nil(err)

	// create latest metric.
	metric, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key",
		Value:     123.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
		Context: models.Context{
			ID:   2,
			Json: []byte(`{"key":"value"}`),
		},
	})
	s.Require().Nil(err)

	// create test param and tag.
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

	tests := []struct {
		name     string
		request  map[any]any
		response response.ProjectParamsResponse
	}{
		{
			name:    "RequestProjectParamsWithoutExperimentFilter",
			request: map[any]any{"sequence": "metric"},
			response: response.ProjectParamsResponse{
				Metric: map[string][]fiber.Map{
					"key": {
						{
							"key": "value",
						},
					},
				},
				Params: map[string]interface{}{
					param.Key: map[string]interface{}{
						"__example_type__": "<class 'str'>",
					},
					"tags": map[string]interface{}{
						tag.Key: map[string]interface{}{
							"__example_type__": "<class 'str'>",
						},
					},
				},
			},
		},
		{
			name: "RequestProjectParamsFilteredByExistingExperiment",
			request: map[any]any{
				"experiments": *s.DefaultExperiment.ID,
			},
			response: response.ProjectParamsResponse{
				Metric: map[string][]fiber.Map{
					"key": {
						{
							"key": "value",
						},
					},
				},
				Params: map[string]interface{}{
					param.Key: map[string]interface{}{
						"__example_type__": "<class 'str'>",
					},
					"tags": map[string]interface{}{
						tag.Key: map[string]interface{}{
							"__example_type__": "<class 'str'>",
						},
					},
				},
			},
		},
		{
			name: "RequestProjectParamsFilteredByNotExistingExperiment",
			request: map[any]any{
				"experiments": 999,
			},
			response: response.ProjectParamsResponse{
				Params: map[string]interface{}{
					"tags": map[string]interface{}{},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := response.ProjectParamsResponse{}
			s.Require().Nil(
				s.AIMClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest("/projects/params"),
			)
			s.Equal(tt.response.Metric, resp.Metric)
			s.Equal(tt.response.Params, resp.Params)
		})
	}

	// mark run as `deleted`.
	run.LifecycleStage = models.LifecycleStageDeleted
	s.Require().Nil(s.RunFixtures.UpdateRun(context.Background(), run))

	// check that endpoint returns an empty response.
	resp := response.ProjectParamsResponse{}
	s.Require().Nil(
		s.AIMClient().WithQuery(
			map[any]any{"sequence": "metric"},
		).WithResponse(
			&resp,
		).DoRequest("/projects/params"),
	)
	s.Equal(0, len(resp.Metric))
	_, ok := resp.Metric[metric.Key]
	s.False(ok)
	s.Equal(map[string]interface{}{"tags": map[string]interface{}{}}, resp.Params)
}

func (s *GetProjectParamsTestSuite) Test_Error() {}
