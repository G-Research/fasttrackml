//go:build integration

package flows

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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RunFlowTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

// TestExperimentFlowTestSuite tests the full `runs` flow connected to namespace functionality.
// Flow contains next endpoints:
// - `POST /runs/create`
// - `GET /runs/get`
// - `POST /runs/update`
// - `POST /runs/search`
// - `POST /runs/delete`
// - `POST /runs/restore`
// - `POST /runs/log-metric`
// - `POST /runs/log-parameter`
// - `POST /runs/set-tag`
// - `POST /runs/delete-tag`
// - `POST /runs/log-batch`
func TestRunFlowTestSuite(t *testing.T) {
	suite.Run(t, new(RunFlowTestSuite))
}

func (s *RunFlowTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *RunFlowTestSuite) TearDownTest() {
	assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *RunFlowTestSuite) Test_Ok() {
	namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "experiment-namespace-1",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)
	namespace2, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "experiment-namespace-2",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Experiment1",
		ArtifactLocation: "/artifact/location",
		LifecycleStage:   models.LifecycleStageActive,
		NamespaceID:      namespace1.ID,
	})
	assert.Nil(s.T(), err)

	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:             "Experiment2",
		ArtifactLocation: "/artifact/location",
		LifecycleStage:   models.LifecycleStageActive,
		NamespaceID:      namespace2.ID,
	})
	assert.Nil(s.T(), err)

	// 1. test `POST /runs/create` endpoint.
	// create runs in scope of different experiment namespaces.
	run1ID := s.createRun(namespace1.Code, &request.CreateRunRequest{
		Name:         "Run1",
		ExperimentID: fmt.Sprintf("%d", *experiment1.ID),
	})

	run2ID := s.createRun(namespace2.Code, &request.CreateRunRequest{
		Name:         "Run2",
		ExperimentID: fmt.Sprintf("%d", *experiment2.ID),
	})

	// 2. test `GET /runs/get` endpoint.
	// check that runs were created in scope of difference experiment namespaces.
	run1 := s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "Run1",
					Status:         string(models.StatusRunning),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)
	run2 := s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "Run2",
					Status:         string(models.StatusRunning),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)

	// 3. test `GET /runs/get` endpoint.
	// check that there is no intersection between runs, so when we request
	// run 1 in scope of namespace 2 and run 2 in scope of namespace 1 API will throw an error.
	resp := api.ErrorResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace2.Code,
		).WithQuery(
			request.GetRunRequest{
				RunID: run1.Run.Info.ID,
			},
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute),
		),
	)
	assert.Equal(s.T(), fmt.Sprintf("RESOURCE_DOES_NOT_EXIST: unable to find run '%s'", run1ID), resp.Error())
	assert.Equal(s.T(), api.ErrorCodeResourceDoesNotExist, string(resp.ErrorCode))

	resp = api.ErrorResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1.Code,
		).WithQuery(
			request.GetRunRequest{
				RunID: run2.Run.Info.ID,
			},
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute),
		),
	)
	assert.Equal(s.T(), fmt.Sprintf("RESOURCE_DOES_NOT_EXIST: unable to find run '%s'", run2ID), resp.Error())
	assert.Equal(s.T(), api.ErrorCodeResourceDoesNotExist, string(resp.ErrorCode))

	// 4. test `POST /runs/update` endpoint.
	s.updateRun(namespace1.Code, &request.UpdateRunRequest{
		RunID:  run1ID,
		Name:   "UpdatedRun1",
		Status: string(models.StatusScheduled),
	})

	s.updateRun(namespace2.Code, &request.UpdateRunRequest{
		RunID:  run2ID,
		Name:   "UpdatedRun2",
		Status: string(models.StatusScheduled),
	})

	// check that runs were updated.
	run1 = s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)
	run2 = s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)

	// 5. test `POST /runs/search` endpoint.
	s.searchRunsAndCompare(
		namespace1.Code,
		request.SearchRunsRequest{
			ExperimentIDs: []string{fmt.Sprintf("%d", *experiment1.ID)},
		},
		[]*response.RunPartialResponse{
			{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					UUID:           run1ID,
					Name:           "UpdatedRun1",
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun1",
						},
					},
				},
			},
		},
	)

	s.searchRunsAndCompare(
		namespace2.Code,
		request.SearchRunsRequest{
			ExperimentIDs: []string{fmt.Sprintf("%d", *experiment2.ID)},
		},
		[]*response.RunPartialResponse{
			{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					UUID:           run2ID,
					Name:           "UpdatedRun2",
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun2",
						},
					},
				},
			},
		},
	)

	// 6. test `POST /runs/delete` endpoint.
	s.deleteRun(namespace1.Code, &request.DeleteRunRequest{RunID: run1ID})
	s.deleteRun(namespace2.Code, &request.DeleteRunRequest{RunID: run2ID})

	// try to get deleted runs and check theirs state.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageDeleted),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageDeleted),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)

	// 7. test `POST /runs/restore` endpoint.
	s.restoreRun(namespace1.Code, &request.RestoreRunRequest{RunID: run1ID})
	s.restoreRun(namespace2.Code, &request.RestoreRunRequest{RunID: run2ID})

	// try to get restored runs and check theirs state.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{},
			},
		},
	)

	// 8. test `POST /runs/log-metric` endpoint.
	s.logRunMetric(namespace1.Code, &request.LogMetricRequest{
		RunID:     run1ID,
		Key:       "key1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
	})
	s.logRunMetric(namespace2.Code, &request.LogMetricRequest{
		RunID:     run2ID,
		Key:       "key2",
		Value:     2.2,
		Timestamp: 123456789,
		Step:      1,
	})

	// try to get runs information and compare it.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Metrics: []response.RunMetricPartialResponse{
						{
							Key:       "key1",
							Step:      1,
							Value:     1.1,
							Timestamp: 123456789,
						},
					},
				},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Metrics: []response.RunMetricPartialResponse{
						{
							Key:       "key2",
							Step:      1,
							Value:     2.2,
							Timestamp: 123456789,
						},
					},
				},
			},
		},
	)

	// 9. test `POST /runs/log-parameter` endpoint.
	s.logRunParam(namespace1.Code, &request.LogParamRequest{
		RunID: run1ID,
		Key:   "key1",
		Value: "param1",
	})
	s.logRunParam(namespace2.Code, &request.LogParamRequest{
		RunID: run2ID,
		Key:   "key2",
		Value: "param2",
	})

	// try to get runs information and compare it.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Params: []response.RunParamPartialResponse{
						{
							Key:   "key1",
							Value: "param1",
						},
					},
				},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Params: []response.RunParamPartialResponse{
						{
							Key:   "key2",
							Value: "param2",
						},
					},
				},
			},
		},
	)

	// 10. test `POST /runs/set-tag` endpoint.
	s.setRunTag(namespace1.Code, &request.SetRunTagRequest{
		RunID: run1ID,
		Key:   "mlflow.user",
		Value: "1",
	})
	s.setRunTag(namespace2.Code, &request.SetRunTagRequest{
		RunID: run2ID,
		Key:   "mlflow.user",
		Value: "2",
	})

	// try to get runs information and compare it.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					UserID:         "1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun1",
						},
						{
							Key:   "mlflow.user",
							Value: "1",
						},
					},
				},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					UserID:         "2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun2",
						},
						{
							Key:   "mlflow.user",
							Value: "2",
						},
					},
				},
			},
		},
	)

	// 11. test `POST /runs/delete-tag` endpoint.
	s.deleteRunTag(namespace1.Code, &request.DeleteRunTagRequest{
		RunID: run1ID,
		Key:   "mlflow.user",
	})
	s.deleteRunTag(namespace2.Code, &request.DeleteRunTagRequest{
		RunID: run2ID,
		Key:   "mlflow.user",
	})

	// try to get runs information and compare it.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					UserID:         "1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun1",
						},
					},
				},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					UserID:         "2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun2",
						},
					},
				},
			},
		},
	)

	// 12. test `POST /runs/log-batch` endpoint.
	s.runLogBatch(namespace1.Code, &request.LogBatchRequest{
		RunID: run1ID,
		Tags: []request.TagPartialRequest{
			{
				Key:   "mlflow.user",
				Value: "1",
			},
		},
		Params: []request.ParamPartialRequest{
			{
				Key:   "key1",
				Value: "param1",
			},
		},
		Metrics: []request.MetricPartialRequest{
			{
				Key:       "key1",
				Value:     1.1,
				Timestamp: 123456789,
				Step:      1,
			},
		},
	})
	s.runLogBatch(namespace2.Code, &request.LogBatchRequest{
		RunID: run2ID,
		Tags: []request.TagPartialRequest{
			{
				Key:   "mlflow.user",
				Value: "2",
			},
		},
		Params: []request.ParamPartialRequest{
			{
				Key:   "key2",
				Value: "param2",
			},
		},
		Metrics: []request.MetricPartialRequest{
			{
				Key:       "key2",
				Value:     2.2,
				Timestamp: 123456789,
				Step:      1,
			},
		},
	})

	// try to get runs information and compare it.
	s.getRunAndCompare(
		namespace1.Code,
		request.GetRunRequest{
			RunID: run1ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run1ID,
					Name:           "UpdatedRun1",
					UserID:         "1",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run1ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment1.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun1",
						},
						{
							Key:   "mlflow.user",
							Value: "1",
						},
					},
					Params: []response.RunParamPartialResponse{
						{
							Key:   "key1",
							Value: "param1",
						},
					},
					Metrics: []response.RunMetricPartialResponse{
						{
							Key:       "key1",
							Step:      1,
							Value:     1.1,
							Timestamp: 123456789,
						},
					},
				},
			},
		},
	)
	s.getRunAndCompare(
		namespace2.Code,
		request.GetRunRequest{
			RunID: run2ID,
		},
		&response.GetRunResponse{
			Run: &response.RunPartialResponse{
				Info: response.RunInfoPartialResponse{
					ID:             run2ID,
					Name:           "UpdatedRun2",
					UserID:         "2",
					Status:         string(models.StatusScheduled),
					ArtifactURI:    fmt.Sprintf("/artifact/location/%s/artifacts", run2ID),
					ExperimentID:   fmt.Sprintf("%d", *experiment2.ID),
					LifecycleStage: string(models.LifecycleStageActive),
				},
				Data: response.RunDataPartialResponse{
					Tags: []response.RunTagPartialResponse{
						{
							Key:   "mlflow.runName",
							Value: "UpdatedRun2",
						},
						{
							Key:   "mlflow.user",
							Value: "2",
						},
					},
					Params: []response.RunParamPartialResponse{
						{
							Key:   "key2",
							Value: "param2",
						},
					},
					Metrics: []response.RunMetricPartialResponse{
						{
							Key:       "key2",
							Step:      1,
							Value:     2.2,
							Timestamp: 123456789,
						},
					},
				},
			},
		},
	)
}

func (s *RunFlowTestSuite) createRun(
	namespace string, req *request.CreateRunRequest,
) string {
	resp := response.CreateRunResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsCreateRoute),
		),
	)
	return resp.Run.Info.ID
}

func (s *RunFlowTestSuite) getRunAndCompare(
	namespace string, req request.GetRunRequest, expectedResponse *response.GetRunResponse,
) *response.GetRunResponse {
	resp := response.GetRunResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace,
		).WithQuery(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute),
		),
	)
	assert.Equal(s.T(), expectedResponse.Run.Info.ID, resp.Run.Info.ID)
	assert.Equal(s.T(), expectedResponse.Run.Info.Name, resp.Run.Info.Name)
	assert.Equal(s.T(), expectedResponse.Run.Info.Status, resp.Run.Info.Status)
	assert.Equal(s.T(), expectedResponse.Run.Info.ArtifactURI, resp.Run.Info.ArtifactURI)
	assert.Equal(s.T(), expectedResponse.Run.Info.ExperimentID, resp.Run.Info.ExperimentID)
	assert.Equal(s.T(), expectedResponse.Run.Info.LifecycleStage, resp.Run.Info.LifecycleStage)
	if expectedResponse.Run.Data.Tags != nil {
		assert.Equal(s.T(), expectedResponse.Run.Data.Tags, resp.Run.Data.Tags)
	}
	if expectedResponse.Run.Data.Params != nil {
		assert.Equal(s.T(), expectedResponse.Run.Data.Params, resp.Run.Data.Params)
	}
	if expectedResponse.Run.Data.Metrics != nil {
		assert.Equal(s.T(), expectedResponse.Run.Data.Metrics, resp.Run.Data.Metrics)
	}
	return &resp
}

func (s *RunFlowTestSuite) updateRun(namespace string, req *request.UpdateRunRequest) string {
	resp := response.UpdateRunResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute),
		),
	)
	return resp.RunInfo.ID
}

func (s *RunFlowTestSuite) searchRunsAndCompare(
	namespace string, req request.SearchRunsRequest, expectedRuns []*response.RunPartialResponse,
) {
	searchResp := response.SearchRunsResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&searchResp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSearchRoute),
		),
	)
	assert.Equal(s.T(), len(expectedRuns), len(searchResp.Runs))
	assert.Equal(s.T(), "", searchResp.NextPageToken)
	assert.Equal(s.T(), expectedRuns, searchResp.Runs)
}

func (s *RunFlowTestSuite) deleteRun(namespace string, req *request.DeleteRunRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteRoute),
		),
	)
}

func (s *RunFlowTestSuite) restoreRun(namespace string, req *request.RestoreRunRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsRestoreRoute),
		),
	)
}

func (s *RunFlowTestSuite) logRunMetric(namespace string, req *request.LogMetricRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute),
		),
	)
}

func (s *RunFlowTestSuite) logRunParam(namespace string, req *request.LogParamRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
}

func (s *RunFlowTestSuite) setRunTag(namespace string, req *request.SetRunTagRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSetTagRoute),
		),
	)
}

func (s *RunFlowTestSuite) deleteRunTag(namespace string, req *request.DeleteRunTagRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteTagRoute),
		),
	)
}

func (s *RunFlowTestSuite) runLogBatch(namespace string, req *request.LogBatchRequest) {
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
		),
	)
}