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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoriesBulkTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetHistoriesBulkTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoriesBulkTestSuite))
}

func (s *GetHistoriesBulkTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func (s *GetHistoriesBulkTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Test Experiment",
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	assert.Nil(s.T(), err)

	run1, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run1",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	assert.Nil(s.T(), err)

	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run1.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	assert.Nil(s.T(), err)

	run2, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "run2",
		Name:           "chill-run",
		Status:         models.StatusScheduled,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		ExperimentID:   *experiment.ID,
	})
	assert.Nil(s.T(), err)

	_, err = s.metricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     2.1,
		Timestamp: 1234567890,
		RunID:     run2.ID,
		Step:      1,
		IsNan:     false,
		Iter:      1,
	})
	assert.Nil(s.T(), err)

	query, err := urlquery.Marshal(&request.GetMetricHistoryBulkRequest{
		RunIDs:    []string{run1.ID, run2.ID},
		MetricKey: "key1",
	})
	assert.Nil(s.T(), err)

	resp := response.GetMetricHistoryResponse{}
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryBulkRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
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
				RunIDs: make(
					[]string, metric.MaxRunIDsForMetricHistoryBulkRequest+1, metric.MaxRunIDsForMetricHistoryBulkRequest+1,
				),
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
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)

			resp := api.ErrorResponse{}
			err = s.client.DoGetRequest(
				fmt.Sprintf(
					"%s%s?%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoryBulkRoute, query,
				),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
