//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogBatchTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogBatchTestSuite(t *testing.T) {
	suite.Run(t, new(LogBatchTestSuite))
}

func (s *LogBatchTestSuite) TestTags_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			resp := map[string]any{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute,
				),
			)
			s.Empty(resp)
		})
	}
}

func (s *LogBatchTestSuite) TestParams_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)

	// create preexisting param (other batch) for conflict testing
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	})
	s.Require().Nil(err)

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
			name: "LogDuplicateSeparateBatch",
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
			name: "LogDuplicateSameBatch",
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
		s.Run(tt.name, func() {
			resp := map[string]any{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute,
				),
			)
			s.Empty(resp)

			// verify params are inserted
			params, err := s.ParamFixtures.GetParamsByRunID(context.Background(), run.ID)
			s.Require().Nil(err)
			for _, param := range tt.request.Params {
				s.Contains(params, models.Param{Key: param.Key, Value: param.Value, RunID: run.ID})
			}
		})
	}
}

func (s *LogBatchTestSuite) TestMetrics_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			// do actual call to API.
			resp := map[string]any{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute,
				),
			)
			s.Empty(resp)

			// make sure that `iter` and `last_iter` for each metric has been updated correctly.
			for key, iteration := range tt.latestMetricIteration {
				lastMetric, err := s.MetricFixtures.GetLatestMetricByKey(context.Background(), key)
				s.Require().Nil(err)
				s.Equal(iteration, lastMetric.LastIter)
			}
		})
	}
}

func (s *LogBatchTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)

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
			error: api.NewInvalidParameterValueError("unable to insert params for run"),
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
					{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
		},
	}

	for _, tt := range testData {
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
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute,
				),
			)
			s.Equal(tt.error.ErrorCode, resp.ErrorCode)
			s.Contains(resp.Error(), tt.error.Message)

			// there should be no params inserted when error occurs.
			params, err := s.ParamFixtures.GetParamsByRunID(context.Background(), run.ID)
			s.Require().Nil(err)
			s.Empty(params)
		})
	}
}
