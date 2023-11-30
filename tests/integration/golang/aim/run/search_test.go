//go:build integration

package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
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

func (s *SearchTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	// create test experiments.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    namespace.ID,
	})
	s.Require().Nil(err)

	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    namespace.ID,
	})
	s.Require().Nil(err)

	// create 3 different test runs and attach tags, metrics, params, etc.
	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id1",
		Name:       "TestRun1",
		UserID:     "1",
		Status:     models.StatusRunning,
		RowNum:     1,
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
		Key:       "TestMetric",
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
		RowNum:     2,
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
		Key:       "TestMetric",
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
		RowNum:     3,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 333444444,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 444555555,
			Valid: true,
		},
		ExperimentID:   *experiment2.ID,
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
		Key:       "TestMetric",
		Value:     3.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  3,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param3",
		Value: "value3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)

	run4, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id4",
		Name:       "TestRun4",
		UserID:     "4",
		Status:     models.StatusScheduled,
		RowNum:     4,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 111111111,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 150000000,
			Valid: true,
		},
		ExperimentID:   *experiment2.ID,
		ArtifactURI:    "artifact_uri4",
		LifecycleStage: models.LifecycleStageDeleted,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag4",
		RunID: run4.ID,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric",
		Value:     4.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run4.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param4",
		Value: "value4",
		RunID: run4.ID,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1, run2, run3, run4}

	tests := []struct {
		name    string
		request request.SearchRunsRequest
		runs    []*models.Run
	}{
		{
			name: "SearchFirstPage",
			request: request.SearchRunsRequest{
				Query: `run.archived == True or run.archived == False`,
				Limit: 2,
			},
			runs: []*models.Run{
				run4,
				run3,
			},
		},
		{
			name: "SearchSecondPage",
			request: request.SearchRunsRequest{
				Query:  `run.archived == True or run.archived == False`,
				Offset: run3.ID,
			},
			runs: []*models.Run{
				run2,
				run1,
			},
		},
		{
			name: "SearchThirdPage",
			request: request.SearchRunsRequest{
				Query:  `run.archived == True or run.archived == False`,
				Offset: run1.ID,
			},
		},
		{
			name: "SearchArchived",
			request: request.SearchRunsRequest{
				Query: `run.archived == True`,
			},

			runs: []*models.Run{
				run2,
				run4,
			},
		},
		{
			name: "SearchNotArchived",
			request: request.SearchRunsRequest{
				Query: `run.archived == False`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchActive",
			request: request.SearchRunsRequest{
				Query: `run.active == True`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchNotActive",
			request: request.SearchRunsRequest{
				Query: `run.active == False`,
			},

			runs: []*models.Run{},
		},
		{
			name: "SearchDurationOperationGrater",
			request: request.SearchRunsRequest{
				Query: `run.duration > 0`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchDurationOperationGraterOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.duration >= 0`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchDurationOperationLess",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf("run.duration < %d", (run3.EndTime.Int64-run3.StartTime.Int64)/1000),
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchDurationOperationLessOrEqual",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf("run.duration <= %d", (run3.EndTime.Int64-run3.StartTime.Int64)/1000),
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchDurationOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.duration == 0`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchDurationOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.duration != 0`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchStartTimeOperationGrater",
			request: request.SearchRunsRequest{
				Query: `run.created_at > 123456789`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchStartTimeOperationGraterOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.created_at >= 123456789`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchStartTimeOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.created_at != 123456789`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchStartTimeOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.created_at == 123456789`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchStartTimeOperationLess",
			request: request.SearchRunsRequest{
				Query: `run.created_at < 333444444`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchStartTimeOperationLessOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.created_at <= 333444444`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchFinalizedAtOperationGrater",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at > 123456789`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchFinalizedAtOperationGraterOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at >= 123456789`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchFinalizedAtOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at != 123456789`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchFinalizedAtOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at == 123456789`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchFinalizedAtOperationLess",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at < 333444444`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchFinalizedAtOperationLessOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.finalized_at <= 444555555`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchRunHashOperationEqual",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf(`run.hash == "%s"`, run1.ID),
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchRunHashOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf(`run.hash != "%s"`, run1.ID),
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchRunNameOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.name != "TestRun1"`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchRunNameOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.name == "TestRun1"`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchRunNameOperationIn",
			request: request.SearchRunsRequest{
				Query: `"Run3" in run.name`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchRunNameOperationNotIn",
			request: request.SearchRunsRequest{
				Query: `"Run3" not in run.name`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchRunNameOperationStartsWith",
			request: request.SearchRunsRequest{
				Query: `run.name.startswith("Test")`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchRunNameOperationStartsWith",
			request: request.SearchRunsRequest{
				Query: `run.name.endswith('3')`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchRunExperimentOperationEqual",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf(`run.experiment == "%s"`, experiment.Name),
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchRunExperimentOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: fmt.Sprintf(`run.experiment != "%s"`, experiment.Name),
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchMetricLastOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last == 3.1`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchMetricLastOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last != 3.1`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchMetricLastOperationGrater",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last > 1.1`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchMetricLastOperationGraterOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last >= 1.1`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchMetricLastOperationLess",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last < 3.1`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchMetricLastOperationLessOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last <= 3.1`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step == 1`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchMetricLastStepOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step != 1`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationGrater",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step > 1`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationGraterOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step >= 1`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationLess",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step < 3`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchMetricLastStepOperationLessOrEqual",
			request: request.SearchRunsRequest{
				Query: `run.metrics['TestMetric'].last_step <= 3`,
			},

			runs: []*models.Run{
				run1,
				run3,
			},
		},
		{
			name: "SearchTagOperationEqual",
			request: request.SearchRunsRequest{
				Query: `run.tags['mlflow.runName'] == "TestRunTag1"`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchTagOperationNotEqual",
			request: request.SearchRunsRequest{
				Query: `run.tags['mlflow.runName'] != "TestRunTag1"`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		// node: re
		{
			name: "SearchRunNameOperationRegexpMatchFunction",
			request: request.SearchRunsRequest{
				Query: `re.match('TestRun1', run.name)`,
			},

			runs: []*models.Run{
				run1,
			},
		},
		{
			name: "SearchRunNameOperationRegexpSearchFunction",
			request: request.SearchRunsRequest{
				Query: `re.search('TestRun3', run.name)`,
			},

			runs: []*models.Run{
				run3,
			},
		},
		{
			name: "SearchComplexQuery",
			request: request.SearchRunsRequest{
				Query: `(run.archived == True or run.archived == False) and run.duration > 0` +
					`and run.metrics['TestMetric'].last > 2.5 and not run.name.endswith('4')`,
			},

			runs: []*models.Run{
				run3,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := new(bytes.Buffer)
			s.Require().Nil(
				s.AIMClient().WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithQuery(
					tt.request,
				).WithResponse(
					resp,
				).DoRequest("/runs/search/run"),
			)

			decodedData, err := encoding.Decode(resp)
			s.Require().Nil(err)

			for _, run := range runs {
				respNameKey := fmt.Sprintf("%v.props.name", run.ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", run.ID)
				startTimeKey := fmt.Sprintf("%v.props.creation_time", run.ID)
				endTimeKey := fmt.Sprintf("%v.props.end_time", run.ID)
				activeKey := fmt.Sprintf("%v.props.active", run.ID)
				archivedKey := fmt.Sprintf("%v.props.archived", run.ID)
				if !slices.Contains(tt.runs, run) {
					s.Nil(decodedData[respNameKey])
				} else {
					s.Equal(run.Name, decodedData[respNameKey])
					s.Equal(
						fmt.Sprintf("%v", run.ExperimentID),
						decodedData[expIdKey])
					s.Equal(
						run.Status == models.StatusRunning,
						decodedData[activeKey])
					s.Equal(run.LifecycleStage == models.LifecycleStageDeleted, decodedData[archivedKey])
					s.Equal(
						run.StartTime.Int64,
						int64(decodedData[startTimeKey].(float64)*1000))
					s.Equal(
						run.EndTime.Int64,
						int64(decodedData[endTimeKey].(float64)*1000))
					metricCount := 0
					for _, metric := range run.LatestMetrics {
						metricNameKey := fmt.Sprintf("%v.traces.metric.%d.name", run.ID, metricCount)
						metricValueKey := fmt.Sprintf("%v.traces.metric.%d.last_value.last", run.ID, metricCount)
						metricStepKey := fmt.Sprintf("%v.traces.metric.%d.last_value.last_step", run.ID, metricCount)
						s.Equal(metric.Value, decodedData[metricValueKey])
						s.Equal(metric.LastIter, decodedData[metricStepKey])
						s.Equal(metric.Key, decodedData[metricNameKey])
						metricCount++
					}
					for _, tag := range run.Tags {
						tagKey := fmt.Sprintf("%v.params.tags.mlflow.runName", run.ID)
						s.Equal(tag.Value, decodedData[tagKey])
					}
				}
			}
		})
	}
}
