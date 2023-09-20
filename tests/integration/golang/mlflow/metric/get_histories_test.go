//go:build integration

package metric

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/metric"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetHistoriesTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetHistoriesTestSuite(t *testing.T) {
	suite.Run(t, new(GetHistoriesTestSuite))
}

func (s *GetHistoriesTestSuite) SetupTest() {
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

func (s *GetHistoriesTestSuite) Test_Ok() {
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

	tests := []struct {
		name    string
		request *request.GetMetricHistoriesRequest
	}{
		{
			name: "GetMetricHistoriesByRunIDs",
			request: &request.GetMetricHistoriesRequest{
				RunIDs: []string{run1.ID, run2.ID},
			},
		},
		{
			name: "GetMetricHistoriesByExperimentIDs",
			request: &request.GetMetricHistoriesRequest{
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			data, err := s.client.DoStreamRequest(
				http.MethodPost,
				fmt.Sprintf(
					"%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoriesRoute,
				),
				tt.request,
			)
			assert.Nil(s.T(), err)
			// TODO:DSuhinin - data is encoded so we need a bit more smart way to check the data.
			// right now we can go with this simple approach.
			assert.NotNil(s.T(), data)
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
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.MetricsRoutePrefix, mlflow.MetricsGetHistoriesRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
