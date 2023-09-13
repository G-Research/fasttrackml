//go:build integration

package metric

import (
	"context"
	"fmt"
	"testing"

	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoryTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoryTestSuite))
}

func (s *GetHistoryTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetHistoryTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Test Experiment",
		NamespaceID:      namespace.ID,
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
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

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	assert.Nil(s.T(), err)

	query, err := urlquery.Marshal(&request.GetMetricHistoryRequest{
		RunID:     run.ID,
		MetricKey: "key1",
	})
	assert.Nil(s.T(), err)

	resp := response.GetMetricHistoryResponse{}
	err = s.MlflowClient.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), response.GetMetricHistoryResponse{
		Metrics: []response.MetricPartialResponse{
			{
				Key:       "key1",
				Step:      1,
				Value:     1.1,
				Timestamp: 1234567890,
			},
		},
	}, resp)
}

func (s *GetHistoryTestSuite) Test_Error() {
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
		s.T().Run(tt.name, func(T *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)

			resp := api.ErrorResponse{}
			err = s.MlflowClient.DoGetRequest(
				fmt.Sprintf(
					"%s%s?%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryRoute, query,
				),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
