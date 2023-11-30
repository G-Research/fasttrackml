//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ExperimentFlowTestSuite struct {
	helpers.BaseTestSuite
}

// TestExperimentFlowTestSuite tests the full `experiment` flow connected to namespace functionality.
// Flow contains next endpoints:
// - `GET /experiments/:id`
// - `GET /experiments`
// - `GET /experiments/:id/runs`
// - `GET /experiments/:id/activity`
func TestExperimentFlowTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentFlowTestSuite))
}

func (s *ExperimentFlowTestSuite) TearDownTest() {
	s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
}

func (s *ExperimentFlowTestSuite) Test_Ok() {
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
			name: "TestExplicitDefaultAndCustomNamespaces",
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
		s.Run(tt.name, func() {
			defer s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())

			// 1. setup data under the test.
			namespace1, namespace2 := tt.setup()
			namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace1)
			s.Require().Nil(err)
			namespace2, err = s.NamespaceFixtures.CreateNamespace(context.Background(), namespace2)
			s.Require().Nil(err)

			experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment1",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace1.ID,
			})
			s.Require().Nil(err)

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
			s.Require().Nil(err)

			experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment2",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace2.ID,
			})
			s.Require().Nil(err)

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
			s.Require().Nil(err)

			// 2. run actual flow test over the test data.
			s.testRunFlow(tt.namespace1Code, tt.namespace2Code, experiment1, experiment2, run1, run2)
		})
	}
}

func (s *ExperimentFlowTestSuite) testRunFlow(
	namespace1Code, namespace2Code string, experiment1, experiment2 *models.Experiment, run1, run2 *models.Run,
) {
	// test `GET /experiments/:id` endpoint.
	s.getExperimentAndCompare(namespace1Code, *experiment1.ID, &response.GetExperiment{
		ID:       fmt.Sprintf("%d", *experiment1.ID),
		Name:     "Experiment1",
		RunCount: 1,
	})
	s.getExperimentAndCompare(namespace2Code, *experiment2.ID, &response.GetExperiment{
		ID:       fmt.Sprintf("%d", *experiment2.ID),
		Name:     "Experiment2",
		RunCount: 1,
	})

	// test `GET /experiments` endpoint.
	s.getExperimentsAndCompare(namespace1Code, response.Experiments{
		{
			ID:       fmt.Sprintf("%d", *experiment1.ID),
			Name:     "Experiment1",
			RunCount: 1,
		},
	})
	s.getExperimentsAndCompare(namespace2Code, response.Experiments{
		{
			ID:       fmt.Sprintf("%d", *experiment2.ID),
			Name:     "Experiment2",
			RunCount: 1,
		},
	})

	// test `GET /experiments/:id/runs` endpoint.
	s.getExperimentRunsAndCompare(namespace1Code, *experiment1.ID, response.GetExperimentRuns{
		ID: fmt.Sprintf("%d", *experiment1.ID),
		Runs: []response.ExperimentRun{
			{
				ID:   run1.ID,
				Name: run1.Name,
			},
		},
	})
	s.getExperimentRunsAndCompare(namespace2Code, *experiment2.ID, response.GetExperimentRuns{
		ID: fmt.Sprintf("%d", *experiment2.ID),
		Runs: []response.ExperimentRun{
			{
				ID:   run2.ID,
				Name: run2.Name,
			},
		},
	})

	// test `GET /experiments/:id/activity` endpoint.
	s.getExperimentActivityAndCompare(namespace1Code, *experiment1.ID, &response.GetExperimentActivity{
		NumRuns:         1,
		NumArchivedRuns: 0,
		NumActiveRuns:   1,
		ActivityMap:     nil,
	})
	s.getExperimentActivityAndCompare(namespace2Code, *experiment2.ID, &response.GetExperimentActivity{
		NumRuns:         1,
		NumArchivedRuns: 0,
		NumActiveRuns:   1,
		ActivityMap:     nil,
	})

	// test `DELETE /experiments/:id` endpoint.
	s.deleteExperiment(namespace1Code, *experiment1.ID)
	s.deleteExperiment(namespace2Code, *experiment2.ID)
}

func (s *ExperimentFlowTestSuite) getExperimentAndCompare(
	namespace string, experimentID int32, expectedResponse *response.GetExperiment,
) {
	var resp response.GetExperiment
	s.Require().Nil(
		s.AIMClient().WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d", experimentID,
		),
	)
	s.Equal(expectedResponse.ID, resp.ID)
	s.Equal(expectedResponse.Name, resp.Name)
	s.Equal(expectedResponse.Description, resp.Description)
	s.Equal(expectedResponse.Archived, resp.Archived)
	s.Equal(expectedResponse.RunCount, resp.RunCount)
}

func (s *ExperimentFlowTestSuite) getExperimentsAndCompare(
	namespace string, expectedResponse response.Experiments,
) {
	var resp response.Experiments
	s.Require().Nil(
		s.AIMClient().WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments",
		),
	)
	s.ElementsMatch(expectedResponse, resp)
}

func (s *ExperimentFlowTestSuite) getExperimentRunsAndCompare(
	namespace string, experimentID int32, expectedResponse response.GetExperimentRuns,
) {
	var resp response.GetExperimentRuns
	s.Require().Nil(
		s.AIMClient().WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d/runs", experimentID,
		),
	)
	s.Equal(expectedResponse.ID, resp.ID)
	s.ElementsMatch(expectedResponse.Runs, resp.Runs)
}

func (s *ExperimentFlowTestSuite) getExperimentActivityAndCompare(
	namespace string, experimentID int32, expectedResponse *response.GetExperimentActivity,
) {
	var resp response.GetExperimentActivity
	s.Require().Nil(
		s.AIMClient().WithResponse(
			&resp,
		).WithNamespace(
			namespace,
		).DoRequest(
			"/experiments/%d/activity", experimentID,
		),
	)
	s.Equal(expectedResponse.NumRuns, resp.NumRuns)
	s.Equal(expectedResponse.NumArchivedRuns, expectedResponse.NumArchivedRuns)
	s.Equal(expectedResponse.NumActiveRuns, expectedResponse.NumActiveRuns)
}

func (s *ExperimentFlowTestSuite) deleteExperiment(namespace string, experimentID int32) {
	var resp response.DeleteExperiment
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespace,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d", experimentID,
		),
	)
	s.Equal("OK", resp.Status)
	s.Equal(fmt.Sprintf("%d", experimentID), resp.ID)
}
