//go:build integration

package metric

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoriesBulkTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetHistoriesBulkTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoriesBulkTestSuite))
}

func (s *GetHistoriesBulkTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Test Experiment",
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	require.Nil(s.T(), err)

	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run1",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run1.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	require.Nil(s.T(), err)

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run2",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     2.1,
		Timestamp: 1234567890,
		RunID:     run2.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	require.Nil(s.T(), err)

	req := request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run1.ID, run2.ID},
		MetricKey: "key1",
	}

	resp := response.GetMetricHistoryResponse{}
	require.Nil(
		s.T(),
		s.MlflowClient.WithQuery(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryBulkRoute),
		),
	)

	assert.Equal(s.T(), response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				RunID:     run1.ID,
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 1234567890,
			},
			{
				RunID:     run2.ID,
				Key:       "key1",
				Step:      1,
				Value:     2.1,
				Timestamp: 1234567890,
			},
		},
	}, resp)
}

func (s *GetHistoriesBulkTestSuite) Test_Error() {
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetMetricHistoryBulkRequest
	}{
		{
			name:    "EmptyOrIncorrectRunIDs",
			request: request.GetMetricHistoryBulkRequest{},
			error:   api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify at least one run_id."),
		},
		{
			name: "LengthOfRunIDsMoreThenAllowed",
			request: request.GetMetricHistoryBulkRequest{
				RunIDs: make([]string, metric.MaxRunIDsForMetricHistoryBulkRequest+1),
			},
			error: api.NewInvalidParameterValueError(
				"GetMetricHistoryBulk request cannot specify more than 200 run_ids. Received 201 run_ids.",
			),
		},
		{
			name: "EmptyOrIncorrectMetricKey",
			request: request.GetMetricHistoryBulkRequest{
				RunIDs: []string{"id"},
			},
			error: api.NewInvalidParameterValueError("GetMetricHistoryBulk request must specify a metric_key."),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient.WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryBulkRoute),
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
