//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunMetricsTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	metricFixtures     *fixtures.MetricFixtures
	run                *models.Run
}

func TestGetRunMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunMetricsTestSuite))
}

func (s *GetRunMetricsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	s.run, err = s.runFixtures.CreateExampleRun(context.Background(), exp)
	assert.Nil(s.T(), err)
}

func (s *GetRunMetricsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		runID            string
		request          request.GetRunMetrics
		expectedResponse response.GetRunMetrics
	}{
		{
			name:  "GetOneRun",
			runID: s.run.ID,
			request: request.GetRunMetrics{
				{
					Context: map[string]string{},
					Name:    "key1",
				},
				{
					Context: map[string]string{},
					Name:    "key2",
				},
			},
			expectedResponse: response.GetRunMetrics{
				response.RunMetrics{
					Name:    "key1",
					Context: map[string]interface{}{},
					Values:  []float64{124.1, 125.1},
					Iters:   []int64{1, 2},
				},
				response.RunMetrics{
					Name:    "key2",
					Context: map[string]interface{}{},
					Values:  []float64{124.1, 125.1},
					Iters:   []int64{1, 2},
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.GetRunMetrics
			err := s.client.DoPostRequest(
				fmt.Sprintf("/runs/%s/metric/get-batch", tt.runID),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.ElementsMatch(s.T(), tt.expectedResponse, resp)
		})
	}
}

func (s *GetRunMetricsTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name  string
		runID string
		error string
	}{
		{
			name:  "GetNonexistentRun",
			runID: uuid.NewString(),
			error: "Not Found",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.client.DoGetRequest(
				fmt.Sprintf("/runs/%s/metric/get-batch", tt.runID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error, resp.Message)
		})
	}
}
