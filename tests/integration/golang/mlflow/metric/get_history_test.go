package metric

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoryTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoryTestSuite))
}

func (s *GetHistoryTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
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

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
		Context: models.Context{
			Json: types.JSONB(`{"key": "key", "value": "value"}`),
		},
	})
	s.Require().Nil(err)

	req := request.GetMetricHistoryRequest{
		RunID:     run.ID,
		MetricKey: "key1",
	}

	resp := response.GetMetricHistoryResponse{}
	s.Require().Nil(
		s.MlflowClient().WithQuery(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute,
		),
	)
	s.Equal(response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 1234567890,
				Context: map[string]any{
					"key":   "key",
					"value": "value",
				},
			},
		},
	}, resp)
}

func (s *GetHistoryTestSuite) Test_Error() {
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetMetricHistoryRequest
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.GetMetricHistoryRequest{},
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
		},
		{
			name: "EmptyOrIncorrectMetricKey",
			request: request.GetMetricHistoryRequest{
				RunID: "id",
			},
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'metric_key'"),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
