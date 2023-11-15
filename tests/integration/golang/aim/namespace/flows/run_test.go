//go:build integration

package flows

import (
	"context"
	"net/http"
	"testing"

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
				Status:         models.StatusScheduled,
				SourceType:     "JOB",
				ExperimentID:   *experiment1.ID,
				ArtifactURI:    "artifact_uri1",
				LifecycleStage: models.LifecycleStageActive,
			})
			require.Nil(s.T(), err)

			run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             "id2",
				Name:           "TestRun2",
				UserID:         "2",
				Status:         models.StatusScheduled,
				SourceType:     "JOB",
				ExperimentID:   *experiment2.ID,
				ArtifactURI:    "artifact_uri2",
				LifecycleStage: models.LifecycleStageActive,
			})
			require.Nil(s.T(), err)

			// 2. run actual flow test over the test data.
			s.testRunFlow(tt.namespace1Code, tt.namespace2Code, run1, run2)
		})
	}
}

func (s *RunFlowTestSuite) testRunFlow(
	namespace1Code, namespace2Code string, run1, run2 *models.Run,
) {

	// test `PUT /runs/:id/` endpoint.
	s.updateRun(namespace1Code, &request.UpdateRunRequest{
		RunID:    common.GetPointer(run1.ID),
		Name:     common.GetPointer("TestRun1Updated"),
		Archived: common.GetPointer(true),
	})

	s.updateRun(namespace2Code, &request.UpdateRunRequest{
		RunID:    common.GetPointer(run2.ID),
		Name:     common.GetPointer("TestRun2Updated"),
		Archived: common.GetPointer(true),
	})

	//  test `GET /runs/:id/info` endpoint.
	// check that runs were actually updated.
	s.getRunAndCompare(namespace1Code, run1.ID, &response.GetRunInfo{
		Props: response.GetRunInfoProps{
			Name:     "TestRun1Updated",
			Archived: true,
		},
	})
	s.getRunAndCompare(namespace2Code, run2.ID, &response.GetRunInfo{
		Props: response.GetRunInfoProps{
			Name:     "TestRun2Updated",
			Archived: true,
		},
	})
}

func (s *RunFlowTestSuite) updateRun(namespace string, req *request.UpdateRunRequest) {
	require.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).DoRequest(
			"/runs/%s/", *req.RunID,
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
		).WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest("/runs/%s/info", runID),
	)
	assert.Equal(s.T(), expectedResponse.Props.Name, resp.Props.Name)
	assert.Equal(s.T(), expectedResponse.Props.Archived, resp.Props.Archived)
}
