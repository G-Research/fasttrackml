//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchTestSuite(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}

func (s *SearchTestSuite) Test_DefaultNamespace_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create default namespace and experiment.
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

	s.testCases(namespace, experiment, false, *experiment.ID)
}

func (s *SearchTestSuite) Test_DefaultNamespaceExerimentZero_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create default namespace and experiment.
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

	// update default experiment id.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	s.Require().Nil(err)

	s.testCases(namespace, experiment, false, int32(0))
}

func (s *SearchTestSuite) Test_CustomNamespace_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create custom namespace and experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	s.testCases(namespace, experiment, true, *experiment.ID)
}

func (s *SearchTestSuite) Test_CustomNamespaceExperimentZero_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create custom namespace and experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		Code:                "custom",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// update default experiment id.
	namespace.DefaultExperimentID = experiment.ID
	_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), namespace)
	s.Require().Nil(err)

	s.testCases(namespace, experiment, true, int32(0))
}

func (s *SearchTestSuite) testCases(
	namespace *models.Namespace,
	experiment *models.Experiment,
	useNamespaceInRequest bool,
	experimentIDInRequest int32,
) {
	// create 3 different test runs and attach tags, metrics, params, etc.
	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id1",
		Name:       "TestRun1",
		UserID:     "1",
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
		ArtifactURI:    "artifact_uri1",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag1",
		RunID: run1.ID,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "run1",
		Value:     1.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run1.ID,
	})
	s.Require().Nil(err)

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id2",
		Name:       "TestRun2",
		UserID:     "2",
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
		ArtifactURI:    "artifact_uri2",
		LifecycleStage: models.LifecycleStageDeleted,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag2",
		RunID: run2.ID,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "run2",
		Value:     2.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param2",
		Value: "value2",
		RunID: run2.ID,
	})
	s.Require().Nil(err)

	run3, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id3",
		Name:       "TestRun3",
		UserID:     "3",
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
		ArtifactURI:    "artifact_uri3",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "run3",
		Value:     3.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param3",
		Value: "value3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)

	tests := []struct {
		name     string
		error    *api.ErrorResponse
		request  request.SearchRunsRequest
		response *response.SearchRunsResponse
	}{
		{
			name: "SearchWithViewTypeAllParameter3RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				ViewType:      request.ViewTypeAll,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run2.ID,
							Name:           "TestRunTag2",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "2",
							Status:         string(models.StatusScheduled),
							StartTime:      111111111,
							EndTime:        222222222,
							ArtifactURI:    "artifact_uri2",
							LifecycleStage: string(models.LifecycleStageDeleted),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag2",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param2",
									Value: "value2",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run2",
									Value:     2.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithViewTypeActiveOnlyParameter2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				ViewType:      request.ViewTypeActiveOnly,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithViewTypeDeletedOnlyParameter1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				ViewType:      request.ViewTypeDeletedOnly,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run2.ID,
							Name:           "TestRunTag2",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "2",
							Status:         string(models.StatusScheduled),
							StartTime:      111111111,
							EndTime:        222222222,
							ArtifactURI:    "artifact_uri2",
							LifecycleStage: string(models.LifecycleStageDeleted),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag2",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param2",
									Value: "value2",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run2",
									Value:     2.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationGrater1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time > 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationGraterOrEqual2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time >= 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time != 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time = 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationLess1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time < 333444444`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStartTimeOperationLessOrEqual2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.start_time <= 333444444`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationGrater1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time > 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationGraterOrEqual2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time >= 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time != 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time = 123456789`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationLess1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time < 444555555`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeEndTimeOperationLessOrEqual2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.end_time <= 444555555`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunNameOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.run_name != "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunNameOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.run_name = "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunNameOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.run_name LIKE "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunNameOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.run_name ILIKE "testruntag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStatusOperationNotEqualNoRunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.status != "RUNNING"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{},
		},
		{
			name: "SearchWithAttributeStatusOperationEqual2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.status = "RUNNING"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStatusOperationLike2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.status LIKE "RUNNING"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeStatusOperationILike2RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.status ILIKE "running"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeUserIDOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.user_id != 1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeUserIDOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.user_id = 3`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeUserIDOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.user_id LIKE "3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeUserIDOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.user_id ILIKE "3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeArtifactURIOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.artifact_uri != "artifact_uri1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeArtifactURIOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.artifact_uri = "artifact_uri3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeArtifactURIOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.artifact_uri LIKE "artifact_uri3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeArtifactURIOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `attributes.artifact_uri ILIKE "ArTiFaCt_UrI3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id != "%s"`, run1.ID),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id = "%s"`, run3.ID),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id LIKE "%s"`, run3.ID),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id ILIKE "%s"`, strings.ToUpper(run3.ID)),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationIN1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id IN ('%s')`, run3.ID),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeRunIDOperationIN1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        fmt.Sprintf(`attributes.run_id NOT IN ('%s')`, run1.ID),
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationGrater1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 > 1.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationGraterOrEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 >= 1.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 != 1.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 = 3.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationLess0RunsShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 < 3.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeMetricsOperationLessOrEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `metrics.run3 <= 3.1`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeParamsOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `params.param3 != "value1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeParamsOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `params.param3 = "value3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeParamsOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `params.param3 LIKE "value3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeParamsOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `params.param3 ILIKE "VaLuE3"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeTagsOperationNotEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `tags.mlflow.runName != "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run3.ID,
							Name:           "TestRunTag3",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "3",
							Status:         string(models.StatusRunning),
							StartTime:      333444444,
							EndTime:        444555555,
							ArtifactURI:    "artifact_uri3",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag3",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param3",
									Value: "value3",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run3",
									Value:     3.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeTagsOperationEqual1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `tags.mlflow.runName = "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeTagsOperationLike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `tags.mlflow.runName LIKE "TestRunTag1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "SearchWithAttributeTagsOperationILike1RunShouldBeReturned",
			request: request.SearchRunsRequest{
				Filter:        `tags.mlflow.runName ILIKE "TeStRuNTaG1"`,
				ExperimentIDs: []string{fmt.Sprintf("%d", experimentIDInRequest)},
			},
			response: &response.SearchRunsResponse{
				Runs: []*response.RunPartialResponse{
					{
						Info: response.RunInfoPartialResponse{
							ID:             run1.ID,
							Name:           "TestRunTag1",
							ExperimentID:   fmt.Sprintf("%d", *experiment.ID),
							UserID:         "1",
							Status:         string(models.StatusRunning),
							StartTime:      123456789,
							EndTime:        123456789,
							ArtifactURI:    "artifact_uri1",
							LifecycleStage: string(models.LifecycleStageActive),
						},
						Data: response.RunDataPartialResponse{
							Tags: []response.RunTagPartialResponse{
								{
									Key:   "mlflow.runName",
									Value: "TestRunTag1",
								},
							},
							Params: []response.RunParamPartialResponse{
								{
									Key:   "param1",
									Value: "value1",
								},
							},
							Metrics: []response.RunMetricPartialResponse{
								{
									Key:       "run1",
									Value:     1.1,
									Timestamp: 1234567890,
									Step:      1,
								},
							},
						},
					},
				},
				NextPageToken: "",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := &response.SearchRunsResponse{}
			client := s.MlflowClient().WithMethod(
				http.MethodPost,
			).WithRequest(
				tt.request,
			).WithResponse(
				&resp,
			)
			if useNamespaceInRequest {
				client = client.WithNamespace(
					namespace.Code,
				)
			}
			s.Require().Nil(
				client.DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSearchRoute,
				),
			)
			s.Equal(len(tt.response.Runs), len(resp.Runs))
			s.Equal(len(tt.response.NextPageToken), len(resp.NextPageToken))

			mappedExpectedResult := make(map[string]*response.RunPartialResponse, len(tt.response.Runs))
			for _, run := range tt.response.Runs {
				mappedExpectedResult[run.Info.ID] = run
			}

			if tt.response.Runs != nil && resp.Runs != nil {
				for _, actualRun := range resp.Runs {
					expectedRun, ok := mappedExpectedResult[actualRun.Info.ID]
					s.True(ok)
					s.NotEmpty(actualRun.Info.ID)
					s.Equal(expectedRun.Info.Name, actualRun.Info.Name)
					s.Equal(expectedRun.Info.Name, actualRun.Info.Name)
					s.Equal(expectedRun.Info.UserID, actualRun.Info.UserID)
					s.Equal(expectedRun.Info.Status, actualRun.Info.Status)
					s.Equal(expectedRun.Info.EndTime, actualRun.Info.EndTime)
					s.Equal(expectedRun.Info.StartTime, actualRun.Info.StartTime)
					s.Equal(expectedRun.Info.ArtifactURI, actualRun.Info.ArtifactURI)
					s.Equal(expectedRun.Info.ExperimentID, actualRun.Info.ExperimentID)
					s.Equal(expectedRun.Info.LifecycleStage, actualRun.Info.LifecycleStage)
					s.Equal(expectedRun.Data.Tags, actualRun.Data.Tags)
					s.Equal(expectedRun.Data.Params, actualRun.Data.Params)
					s.Equal(expectedRun.Data.Metrics, actualRun.Data.Metrics)
				}
			}
		})
	}
}
