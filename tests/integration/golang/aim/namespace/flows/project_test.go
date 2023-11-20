//go:build integration

package flows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ProjectFlowTestSuite struct {
	helpers.BaseTestSuite
}

// TestProjectFlowTestSuite tests the full `projects` flow connected to namespace functionality.
// Flow contains next endpoints:
// - `GET /projects`
// - `GET /projects/status`
// - `GET /projects/params`
// - `GET /projects/activity`
func TestProjectFlowTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectFlowTestSuite))
}

func (s *ProjectFlowTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *ProjectFlowTestSuite) Test_Ok() {
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

			metric1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
				Key:       "metric1",
				Value:     123.1,
				Timestamp: 1234567890,
				Step:      1,
				IsNan:     false,
				RunID:     run1.ID,
				LastIter:  1,
			})
			require.Nil(s.T(), err)

			tag1, err := s.TagFixtures.CreateTag(context.Background(), &models.Tag{
				Key:   "tag1",
				Value: "value1",
				RunID: run1.ID,
			})
			require.Nil(s.T(), err)

			param1, err := s.ParamFixtures.CreateParam(context.Background(), &models.Param{
				Key:   "param1",
				Value: "value1",
				RunID: run1.ID,
			})
			require.Nil(s.T(), err)

			experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:             "Experiment2",
				ArtifactLocation: "/artifact/location",
				LifecycleStage:   models.LifecycleStageActive,
				NamespaceID:      namespace2.ID,
			})
			require.Nil(s.T(), err)

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

			metric2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
				Key:       "metric2",
				Value:     123.1,
				Timestamp: 1234567890,
				Step:      1,
				IsNan:     false,
				RunID:     run2.ID,
				LastIter:  1,
			})
			require.Nil(s.T(), err)

			tag2, err := s.TagFixtures.CreateTag(context.Background(), &models.Tag{
				Key:   "tag2",
				Value: "value2",
				RunID: run2.ID,
			})
			require.Nil(s.T(), err)

			param2, err := s.ParamFixtures.CreateParam(context.Background(), &models.Param{
				Key:   "param2",
				Value: "value2",
				RunID: run2.ID,
			})
			require.Nil(s.T(), err)

			// 2. run actual flow test over the test data.
			s.testRunFlow(
				tt.namespace1Code, tt.namespace2Code, metric1, metric2, param1, param2, tag1, tag2,
			)
		})
	}
}

func (s *ProjectFlowTestSuite) testRunFlow(
	namespace1Code, namespace2Code string,
	metric1, metric2 *models.LatestMetric,
	param1, param2 *models.Param,
	tag1, tag2 *models.Tag,
) {
	// test `GET /projects` endpoint.
	s.getProjectAndCompare(namespace1Code, &response.GetProjectResponse{
		Name: "FastTrackML",
	})
	s.getProjectAndCompare(namespace2Code, &response.GetProjectResponse{
		Name: "FastTrackML",
	})

	// test `GET /projects/status` endpoint.
	s.getProjectStatusAndCompare(namespace1Code)
	s.getProjectStatusAndCompare(namespace2Code)

	// test `GET /projects/params` endpoint.
	s.getProjectParamsAndCompare(namespace1Code, metric1, param1, tag1)
	s.getProjectParamsAndCompare(namespace2Code, metric2, param2, tag2)

	// test `GET /projects/activity` endpoint.
	s.getProjectActivityAndCompare(namespace1Code, &response.ProjectActivityResponse{
		NumExperiments:  1,
		NumRuns:         1,
		NumActiveRuns:   1,
		NumArchivedRuns: 0,
		ActivityMap:     map[string]int{},
	})
	s.getProjectActivityAndCompare(namespace2Code, &response.ProjectActivityResponse{
		NumExperiments:  1,
		NumRuns:         1,
		NumActiveRuns:   1,
		NumArchivedRuns: 0,
		ActivityMap:     map[string]int{},
	})
}

func (s *ProjectFlowTestSuite) getProjectAndCompare(
	namespace string, expectedResponse *response.GetProjectResponse,
) {
	var resp response.GetProjectResponse
	require.Nil(
		s.T(), s.AIMClient().WithNamespace(namespace).WithResponse(&resp).DoRequest("/projects"),
	)
	assert.Equal(s.T(), expectedResponse.Name, resp.Name)
	assert.Equal(s.T(), expectedResponse.Description, resp.Description)
	assert.Equal(s.T(), expectedResponse.TelemetryEnabled, resp.TelemetryEnabled)
}

func (s *ProjectFlowTestSuite) getProjectStatusAndCompare(namespace string) {
	var resp string
	require.Nil(
		s.T(), s.AIMClient().WithNamespace(namespace).WithResponse(&resp).DoRequest("/projects/status"),
	)
	assert.Equal(s.T(), "up-to-date", resp)
}

func (s *ProjectFlowTestSuite) getProjectParamsAndCompare(
	namespace string, metric *models.LatestMetric, param *models.Param, tag *models.Tag,
) {
	resp := response.ProjectParamsResponse{}
	require.Nil(
		s.T(),
		s.AIMClient().WithNamespace(
			namespace,
		).WithQuery(
			map[any]any{"sequence": "metric"},
		).WithResponse(
			&resp,
		).DoRequest("/projects/params"),
	)

	assert.Equal(s.T(), 1, len(resp.Metric))
	_, ok := resp.Metric[metric.Key]
	assert.True(s.T(), ok)
	assert.Equal(s.T(), map[string]interface{}{
		param.Key: map[string]interface{}{
			"__example_type__": "<class 'str'>",
		},
		"tags": map[string]interface{}{
			tag.Key: map[string]interface{}{
				"__example_type__": "<class 'str'>",
			},
		},
	}, resp.Params)
}

func (s *ProjectFlowTestSuite) getProjectActivityAndCompare(
	namespace string, expectedResponse *response.ProjectActivityResponse,
) {
	var resp response.ProjectActivityResponse
	require.Nil(
		s.T(), s.AIMClient().WithNamespace(namespace).WithResponse(&resp).DoRequest("/projects/activity"),
	)

	assert.Equal(s.T(), expectedResponse.NumActiveRuns, resp.NumActiveRuns)
	assert.Equal(s.T(), expectedResponse.NumArchivedRuns, resp.NumArchivedRuns)
	assert.Equal(s.T(), expectedResponse.NumExperiments, resp.NumExperiments)
	assert.Equal(s.T(), expectedResponse.NumRuns, resp.NumRuns)
}
