//go:build integration

package flows

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RunFlowTestSuite struct {
	helpers.BaseTestSuite
}

func TestRunFlowTestSuite(t *testing.T) {
	suite.Run(t, new(RunFlowTestSuite))
}

func (s *RunFlowTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *RunFlowTestSuite) Test_Ok() {
	tests := []struct {
		name           string
		setup          func() (*models.Namespace, *models.Namespace)
		namespace1Code string
		namespace2Code string
	}{
		{
			name: "TestCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-2",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "namespace-1",
			namespace2Code: "namespace-2",
		},
		{
			name: "TestObviousDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "default",
			namespace2Code: "namespace-1",
		},
		{
			name: "TestImplicitDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "",
			namespace2Code: "namespace-1",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			defer require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())

			// 1. setup data under the test.
			namespace1, namespace2 := tt.setup()
			namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace1)
			require.Nil(s.T(), err)
			namespace2, err = s.NamespaceFixtures.CreateNamespace(context.Background(), namespace2)
			require.Nil(s.T(), err)

			experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment1",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace1.ID,
			})
			require.Nil(s.T(), err)

			experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment2",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace2.ID,
			})
			require.Nil(s.T(), err)

			run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             "id1",
				Name:           "TestRun1",
				UserID:         "2",
				Status:         models.StatusRunning,
				SourceType:     "JOB",
				ExperimentID:   *experiment1.ID,
				ArtifactURI:    "artifact_uri1",
				LifecycleStage: models.LifecycleStageActive,
			})
			require.Nil(s.T(), err)
			require.Nil(
				s.T(), s.RunFixtures.CreateMetric(
					context.Background(),
					&models.Metric{
						Key:       "key1",
						Value:     1111.1,
						Timestamp: 1234567890,
						RunID:     run1.ID,
						Step:      1,
						IsNan:     false,
						Iter:      1,
					},
				),
			)

			run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             "id2",
				Name:           "TestRun2",
				UserID:         "2",
				Status:         models.StatusRunning,
				SourceType:     "JOB",
				ExperimentID:   *experiment2.ID,
				ArtifactURI:    "artifact_uri2",
				LifecycleStage: models.LifecycleStageActive,
			})
			require.Nil(s.T(), err)
			require.Nil(
				s.T(), s.RunFixtures.CreateMetric(
					context.Background(),
					&models.Metric{
						Key:       "key2",
						Value:     2222.2,
						Timestamp: 1234567890,
						RunID:     run2.ID,
						Step:      2,
						IsNan:     false,
						Iter:      2,
					},
				),
			)

			// 2. run actual flow test over the test data.
			s.testRunFlow(tt.namespace1Code, tt.namespace2Code, experiment1, experiment2, run1, run2)
		})
	}
}

func (s *RunFlowTestSuite) testRunFlow(
	namespace1Code, namespace2Code string, experiment1, experiment2 *models.Experiment, run1, run2 *models.Run,
) {

	// test `PUT /runs/:id` endpoint.
	s.updateRun(namespace1Code, &request.UpdateRunRequest{
		RunID: common.GetPointer(run1.ID),
		Name:  common.GetPointer("TestRun1Updated"),
	})

	s.updateRun(namespace2Code, &request.UpdateRunRequest{
		RunID: common.GetPointer(run2.ID),
		Name:  common.GetPointer("TestRun2Updated"),
	})

	// test `GET /runs/:id/info` endpoint.
	// check that runs were actually updated.
	s.getRunAndCompare(namespace1Code, run1.ID, &response.GetRunInfo{
		Props: response.GetRunInfoProps{
			Name: "TestRun1Updated",
		},
	})
	s.getRunAndCompare(namespace2Code, run2.ID, &response.GetRunInfo{
		Props: response.GetRunInfoProps{
			Name: "TestRun2Updated",
		},
	})

	// test `GET /runs/search/run` endpoint.
	s.searchRunsAndCompare(namespace1Code, request.SearchRunsRequest{}, []models.Run{
		{
			ID:             run1.ID,
			Name:           "TestRun1Updated",
			Status:         models.StatusRunning,
			LifecycleStage: models.LifecycleStageActive,
			ExperimentID:   *experiment1.ID,
		},
	})
	s.searchRunsAndCompare(namespace2Code, request.SearchRunsRequest{}, []models.Run{
		{
			ID:             run2.ID,
			Name:           "TestRun2Updated",
			Status:         models.StatusRunning,
			LifecycleStage: models.LifecycleStageActive,
			ExperimentID:   *experiment2.ID,
		},
	})

	// test `GET /runs/active` endpoint.
	s.getActiveRunsAndCompare(namespace1Code, []models.Run{
		{
			ID:             run1.ID,
			Name:           "TestRun1Updated",
			Status:         models.StatusRunning,
			LifecycleStage: models.LifecycleStageActive,
			ExperimentID:   *experiment1.ID,
		},
	})
	s.getActiveRunsAndCompare(namespace2Code, []models.Run{
		{
			ID:             run2.ID,
			Name:           "TestRun2Updated",
			Status:         models.StatusRunning,
			LifecycleStage: models.LifecycleStageActive,
			ExperimentID:   *experiment2.ID,
		},
	})

	// test `GET /runs/:id/metric/get-batch` endpoint.
	s.getRunMetricsAndCompare(
		namespace1Code,
		run1.ID,
		&request.GetRunMetrics{
			{
				Name: "key1",
			},
		},
		response.GetRunMetrics{
			response.RunMetrics{
				Name:    "key1",
				Context: map[string]interface{}{},
				Values:  []float64{1111.1},
				Iters:   []int64{1},
			},
		},
	)
	s.getRunMetricsAndCompare(
		namespace2Code,
		run2.ID,
		&request.GetRunMetrics{
			{
				Name: "key2",
			},
		},
		response.GetRunMetrics{
			response.RunMetrics{
				Name:    "key2",
				Context: map[string]interface{}{},
				Values:  []float64{2222.2},
				Iters:   []int64{2},
			},
		},
	)
}

func (s *RunFlowTestSuite) updateRun(namespace string, req *request.UpdateRunRequest) {
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodPut,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			"/runs/%s", *req.RunID,
		),
	)
}

func (s *RunFlowTestSuite) getRunAndCompare(
	namespace string, runID string, expectedResponse *response.GetRunInfo,
) {
	var resp response.GetRunInfo
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodGet,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest("/runs/%s/info", runID),
	)
	assert.Equal(s.T(), expectedResponse.Props.Name, resp.Props.Name)
	assert.Equal(s.T(), expectedResponse.Props.Archived, resp.Props.Archived)
}

func (s *RunFlowTestSuite) searchRunsAndCompare(
	namespace string, request request.SearchRunsRequest, expectedRunList []models.Run,
) {
	resp := new(bytes.Buffer)
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodGet,
		).WithResponseType(
			helpers.ResponseTypeBuffer,
		).WithQuery(
			request,
		).WithNamespace(
			namespace,
		).WithResponse(
			resp,
		).DoRequest("/runs/search/run"),
	)

	decodedData, err := encoding.Decode(resp)
	require.Nil(s.T(), err)
	for _, expectedRun := range expectedRunList {
		assert.Equal(
			s.T(),
			expectedRun.Name,
			decodedData[fmt.Sprintf("%v.props.name", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			fmt.Sprintf("%d", expectedRun.ExperimentID),
			decodedData[fmt.Sprintf("%v.props.experiment.id", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			expectedRun.Status == models.StatusRunning,
			decodedData[fmt.Sprintf("%v.props.active", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			expectedRun.LifecycleStage == models.LifecycleStageDeleted,
			decodedData[fmt.Sprintf("%v.props.archived", expectedRun.ID)],
		)
	}
}

func (s *RunFlowTestSuite) getActiveRunsAndCompare(namespace string, expectedRunList []models.Run) {
	resp := new(bytes.Buffer)
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodGet,
		).WithResponseType(
			helpers.ResponseTypeBuffer,
		).WithNamespace(
			namespace,
		).WithResponse(
			resp,
		).DoRequest("/runs/active"),
	)

	decodedData, err := encoding.Decode(resp)
	require.Nil(s.T(), err)
	for _, expectedRun := range expectedRunList {
		assert.Equal(
			s.T(),
			expectedRun.Name,
			decodedData[fmt.Sprintf("%v.props.name", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			fmt.Sprintf("%d", expectedRun.ExperimentID),
			decodedData[fmt.Sprintf("%v.props.experiment.id", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			expectedRun.Status == models.StatusRunning,
			decodedData[fmt.Sprintf("%v.props.active", expectedRun.ID)],
		)
		assert.Equal(
			s.T(),
			expectedRun.LifecycleStage == models.LifecycleStageDeleted,
			decodedData[fmt.Sprintf("%v.props.archived", expectedRun.ID)],
		)
	}
}

func (s *RunFlowTestSuite) getRunMetricsAndCompare(
	namespace, runID string, request *request.GetRunMetrics, expectedMetrics response.GetRunMetrics,
) {
	var resp response.GetRunMetrics
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodPost,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithRequest(
			request,
		).WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest(
			"/runs/%s/metric/get-batch", runID,
		),
	)
	assert.ElementsMatch(s.T(), expectedMetrics, resp)
}
