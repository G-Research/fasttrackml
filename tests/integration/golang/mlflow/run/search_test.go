//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	tagFixtures        *fixtures.TagFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestSearchTestSuite(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}

func (s *SearchTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	tagFixtures, err := fixtures.NewTagFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.tagFixtures = tagFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *SearchTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// create test experiment.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// create 3 different test runs and attach tags, metrics, params, etc.
	run1, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id1",
		Name:       "TestRun1",
		Status:     models.StatusRunning,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri",
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)
	_, err = s.tagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag1",
		RunID: run1.ID,
	})
	assert.Nil(s.T(), err)

	run2, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id2",
		Name:       "TestRun2",
		Status:     models.StatusScheduled,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 111111111,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 222222222,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri",
		LifecycleStage: models.LifecycleStageDeleted,
	})
	assert.Nil(s.T(), err)
	_, err = s.tagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag2",
		RunID: run2.ID,
	})
	assert.Nil(s.T(), err)

	run3, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id3",
		Name:       "TestRun3",
		Status:     models.StatusRunning,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 333444444,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 444555555,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri",
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)
	_, err = s.tagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag3",
		RunID: run3.ID,
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name     string
		error    *api.ErrorResponse
		request  *request.SearchRunsRequest
		response *response.SearchRunsResponse
	}{
		{
			name: "SearchWithViewTypeAllParameter3RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				ViewType:      request.ViewTypeAll,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run2.ID,
							Name:           "TestRunTag2",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusScheduled),
							StartTime:      111111111,
							EndTime:        222222222,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageDeleted),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithViewTypeActiveOnlyParameter2RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				ViewType:      request.ViewTypeActiveOnly,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithViewTypeDeletedOnlyParameter1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				ViewType:      request.ViewTypeDeletedOnly,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run2.ID,
							Name:           "TestRunTag2",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusScheduled),
							StartTime:      111111111,
							EndTime:        222222222,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageDeleted),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationGrater1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time > 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationGraterOrEqual2RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time >= 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationNotEqual1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time != 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationEqual1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time = 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationLess1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time < 333444444`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationLessOrEqual2RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.start_time <= 333444444`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationGrater1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time > 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationGraterOrEqual2RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time >= 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationNotEqual1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time != 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationEqual1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time = 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationLess1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time < 444555555`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationLessOrEqual2RunsShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.end_time <= 444555555`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunNameOperationEqual1RunShouldBeReturned",
			request: &request.SearchRunsRequest{
				Filter:        `attributes.run_name = "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", *experiment.ID)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri",
							LifecycleStage: string(models.LifecycleStageActive),
						},
					},
				},
				NextPageToken: "",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := &response.SearchRunsResponse{}
			err = s.client.DoPostRequest(
				fmt.Sprintf("%s%s?%s", mlflow.RunsRoutePrefix, mlflow.RunsSearchRoute, query),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			helpers.CompareExpectedSearchRunsResponseWithActualSearchRunsResponse(s.T(), tt.response, resp)
		})
	}
}
