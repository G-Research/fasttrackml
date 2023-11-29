//go:build integration

package metric

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoriesTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetHistoriesTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoriesTestSuite))
}

func (s *GetHistoriesTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment2",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run1",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	s.Require().Nil(err)

	metric, err := s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run1.ID,
		Step:      1,
		Iter:      1,
	})
	s.Require().Nil(err)
	s.Require().Nil(metric.ContextID)
	s.Require().Nil(metric.Context)

	metric, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key2",
		Value:     1.1,
		Timestamp: 2234567890,
		RunID:     run1.ID,
		Step:      1,
		Iter:      1,
		Context: &models.Context{
			Json: datatypes.JSON([]byte(`
				{
					"metrickey1": "metricvalue1",
					"metrickey2": "metricvalue2",
					"metricnested": { "metricnestedkey": "metricnestedvalue" }
				}`,
			)),
		},
	})
	s.Require().Nil(err)
	s.Require().NotNil(metric.ContextID)
	s.Require().NotNil(metric.Context)

	// verify metric contexts are persisting
	metrics, err := s.MetricFixtures.GetMetricsByRunID(context.Background(), run1.ID)
	s.Require().Nil(err)
	s.Require().Nil(metrics[0].ContextID)
	s.Require().NotNil(metrics[1].ContextID)

	// verify metric contexts can be used for selection (toplevel key)
	metrics, err = s.MetricFixtures.GetMetricsByContext(context.Background(), map[string]string{
		"metrickey1": "metricvalue1",
	})
	s.Require().Nil(err)
	s.Require().Len(metrics, 1)
	s.Require().NotNil(metrics[0].ContextID)

	// nested key
	metrics, err = s.MetricFixtures.GetMetricsByContext(context.Background(), map[string]string{
		"metricnested.metricnestedkey": "metricnestedvalue",
	})
	s.Require().Nil(err)
	s.Require().Len(metrics, 1)
	s.Require().NotNil(metrics[0].ContextID)

	metrics, err = s.MetricFixtures.GetMetricsByContext(
		context.Background(),
		map[string]string{"metrickey2": "metricvalue1"},
	)
	s.Require().Nil(err)
	s.Require().Len(metrics, 0)

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run2",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment2.ID,
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     2.1,
		Timestamp: 1234567890,
		RunID:     run2.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	s.Require().Nil(err)

	tests := []struct {
		name           string
		request        *request.GetMetricHistoriesRequest
		expectResponse func(response string)
	}{
		{
			name: "GetMetricHistoriesByRunIDs",
			request: &request.GetMetricHistoriesRequest{
				RunIDs: []string{run1.ID, run2.ID},
			},
			expectResponse: func(resp string) {
				s.Contains(resp, "run1")
				s.Contains(resp, "run2")
			},
		},
		{
			name: "GetMetricHistoriesByExperimentIDs",
			request: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			expectResponse: func(resp string) {
				s.Contains(resp, "run1")
				s.NotContains(resp, "run2")
			},
		},
		{
			name: "GetMetricHistoriesByContextMatch",
			request: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
				Context:       map[string]string{"metrickey1": "metricvalue1"},
			},
			expectResponse: func(resp string) {
				s.Contains(resp, "run1")
				s.NotContains(resp, "run2")
			},
		},
		{
			name: "GetMetricHistoriesByContextNoMatch",
			request: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
				Context:       map[string]string{"metrickey1": "metricvalue2"},
			},
			expectResponse: func(resp string) {
				s.NotContains(resp, "run1")
				s.NotContains(resp, "run2")
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := new(bytes.Buffer)
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithResponse(
					resp,
				).DoRequest(
					"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoriesRoute,
				),
			)

			// TODO:DSuhinin - data is encoded so we need a bit more smart way to check the data.
			// right now we can go with this simple approach.
			tt.expectResponse(resp.String())
		})
	}
}

func (s *GetHistoriesTestSuite) Test_Error() {
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetMetricHistoriesRequest
	}{
		{
			name: "RunIDsAndExperimentIDsPopulatedAtTheSameTime",
			request: request.GetMetricHistoriesRequest{
				RunIDs:        []string{"id"},
				ExperimentIDs: []string{"id"},
			},
			error: api.NewInvalidParameterValueError(
				"experiment_ids and run_ids cannot both be specified at the same time",
			),
		},
		{
			name: "IncorrectOrUnsupportedViewType",
			request: request.GetMetricHistoriesRequest{
				RunIDs:   []string{"id"},
				ViewType: "unsupported_view_type",
			},
			error: api.NewInvalidParameterValueError("Invalid run_view_type 'unsupported_view_type'"),
		},
		{
			name: "LengthOfRunIDsMoreThenAllowed",
			request: request.GetMetricHistoriesRequest{
				RunIDs:     []string{"id"},
				ViewType:   request.ViewTypeAll,
				MaxResults: metric.MaxResultsForMetricHistoriesRequest + 1,
			},
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied."),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoriesRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
