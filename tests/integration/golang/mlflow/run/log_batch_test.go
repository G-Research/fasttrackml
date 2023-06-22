//go:build integration

package run

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogBatchTestSuite struct {
	suite.Suite
	run                *models.Run
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestLogBatchTestSuite(t *testing.T) {
	suite.Run(t, new(LogBatchTestSuite))
}

func (s *LogBatchTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *exp.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.runFixtures.CreateTestRun(context.Background(), run)
	assert.Nil(s.T(), err)
	s.run = run
}

func (s *LogBatchTestSuite) TestTags_Ok() {
	tests := []struct {
		name    string
		request *request.LogBatchRequest
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Tags: []request.TagPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := map[string]any{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)
		})
	}
}

func (s *LogBatchTestSuite) TestParams_Ok() {
	tests := []struct {
		name    string
		request *request.LogBatchRequest
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
			},
		},
		{
			name: "LogDuplicate",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key2",
						Value: "value2",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := map[string]any{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)
		})
	}
}

func (s *LogBatchTestSuite) TestMetrics_Ok() {
	tests := []struct {
		name                  string
		request               *request.LogBatchRequest
		latestMetricIteration map[string]int64
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key1",
						Value:     1.0,
						Timestamp: 1687325991,
						Step:      1,
					},
				},
			},
			latestMetricIteration: map[string]int64{
				"key1": 1,
			},
		},
		{
			name: "LogSeveral",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key1",
						Value:     1.1,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key1",
						Value:     1.1,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key2",
						Value:     1.1,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key2",
						Value:     1.2,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key2",
						Value:     1.3,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key2",
						Value:     1.4,
						Timestamp: 1687325991,
						Step:      1,
					},
				},
			},
			latestMetricIteration: map[string]int64{
				"key1": 3,
				"key2": 4,
			},
		},
		{
			name: "LogDuplicate",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Metrics: []request.MetricPartialRequest{
					{
						Key:       "key3",
						Value:     1.0,
						Timestamp: 1687325991,
						Step:      1,
					},
					{
						Key:       "key3",
						Value:     1.0,
						Timestamp: 1687325991,
						Step:      1,
					},
				},
			},
			latestMetricIteration: map[string]int64{
				"key3": 2,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			// do actual call to API.
			resp := map[string]any{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)

			// make sure that `iter` and `last_iter` for each metric has been updated correctly.
			for key, iteration := range tt.latestMetricIteration {
				lastMetric, err := s.metricFixtures.GetLatestMetricByKey(context.Background(), key)
				assert.Nil(s.T(), err)
				assert.Equal(s.T(), iteration, lastMetric.LastIter)
			}
		})
	}
}

func (s *LogBatchTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.LogBatchRequest
	}{
		{
			name:    "MissingRunIDFails",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.LogBatchRequest{},
		},
		{
			name:  "DuplicateKeyDifferentValueFails",
			error: api.NewInternalError("duplicate key"),
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key1",
						Value: "value2",
					},
				},
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.ErrorCode, resp.ErrorCode)
			assert.Contains(s.T(), resp.Error(), tt.error.Message)
		})
	}
}
