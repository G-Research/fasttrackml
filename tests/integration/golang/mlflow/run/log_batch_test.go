//go:build integration

package run

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogBatchTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestLogBatchTestSuite(t *testing.T) {
	suite.Run(t, new(LogBatchTestSuite))
}

func (s *LogBatchTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *LogBatchTestSuite) TestTags_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name    string
		request *request.LogBatchRequest
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: run.ID,
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
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name    string
		request *request.LogBatchRequest
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: run.ID,
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
				RunID: run.ID,
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
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name                  string
		request               *request.LogBatchRequest
		latestMetricIteration map[string]int64
	}{
		{
			name: "LogOne",
			request: &request.LogBatchRequest{
				RunID: run.ID,
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
				RunID: run.ID,
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
				RunID: run.ID,
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
		{
			name: "LogMany",
			request: &request.LogBatchRequest{
				RunID: run.ID,
				Metrics: func() []request.MetricPartialRequest {
					metrics := make([]request.MetricPartialRequest, 100*1000)
					for k := 0; k < 100; k++ {
						key := fmt.Sprintf("many%d", k)
						for i := 0; i < 1000; i++ {
							metrics[k*1000+i] = request.MetricPartialRequest{
								Key:       key,
								Value:     float64(i) + 0.1,
								Timestamp: 1687325991,
								Step:      1,
							}
						}
					}
					return metrics
				}(),
			},
			latestMetricIteration: func() map[string]int64 {
				metrics := make(map[string]int64, 100)
				for k := 0; k < 100; k++ {
					key := fmt.Sprintf("many%d", k)
					metrics[key] = 1000
				}
				return metrics
			}(),
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

	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	assert.Nil(s.T(), err)

	testData := []struct {
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
			error: api.NewInternalError("unable to insert params for run"),
			request: &request.LogBatchRequest{
				RunID: run.ID,
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
