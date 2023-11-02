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

// TestExperimentFlowTestSuite tests the full `run` flow connected with namespace functionality.
// Flow contains next endpoints:
// - `POST /runs/create`
// - `GET /runs/get`
// - `POST /runs/update`
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
	run1 := s.getRun(
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
	run2 := s.getRun(
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
	run1 = s.getRun(
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
	run2 = s.getRun(
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
	s.searchRuns(
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

	s.searchRuns(
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

func (s *RunFlowTestSuite) getRun(
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

func (s *RunFlowTestSuite) searchRuns(
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
