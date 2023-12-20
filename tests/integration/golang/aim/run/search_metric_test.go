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
	// create test experiments.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    s.DefaultNamespace.ID,
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
	run1.Experiment = *experiment
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
	run2.Experiment = *experiment
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
	run3.Experiment = *experiment2
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
	run4.Experiment = *experiment2
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
				s.AIMClient().WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithQuery(
					tt.request,
				).WithResponse(
					resp,
				).DoRequest("/runs/search/metric"),
			)

			decodedData, err := encoding.NewDecoder(resp).Decode()
			s.Require().Nil(err)

			for _, run := range runs {
				respNameKey := fmt.Sprintf("%v.props.name", run.ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", run.ID)
				expNameKey := fmt.Sprintf("%v.props.experiment.name", run.ID)
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
					s.Equal(run.Experiment.Name, decodedData[expNameKey])
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
