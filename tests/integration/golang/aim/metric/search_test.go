//go:build integration

package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
}

func (s *SearchMetricsTestSuite) Test_Ok() {
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

	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    namespace.ID,
	})
	s.Require().Nil(err)

	// create different test runs and attach tags, metrics, params, etc.
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
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric1Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      2,
	})
	s.Require().Nil(err)
	metric2Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 123456789,
		Step:      5,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  2,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      3,
	})
	s.Require().Nil(err)
	metric3Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 123456789,
		Step:      10,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  3,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run1.ID,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag1",
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
			Int64: 444444444,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri2",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      3,
	})
	s.Require().Nil(err)
	metric1Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     0.5,
		Timestamp: 111111111,
		Step:      4,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  3,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      2,
	})
	s.Require().Nil(err)
	metric2Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     2.1,
		Timestamp: 222222222,
		Step:      5,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  2,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		Iter:      3,
	})
	s.Require().Nil(err)
	metric3Run2, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric3",
		Value:     3.1,
		Timestamp: 333333333,
		Step:      10,
		IsNan:     false,
		RunID:     run2.ID,
		LastIter:  3,
	})
	s.Require().Nil(err)
	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param2",
		Value: "value2",
		RunID: run2.ID,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag2",
		RunID: run2.ID,
	})
	s.Require().Nil(err)

	run3, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id3",
		Name:       "TestRun3",
		UserID:     "3",
		Status:     models.StatusScheduled,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 222222222,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 444444444,
			Valid: true,
		},
		ExperimentID:   *experiment1.ID,
		ArtifactURI:    "artifact_uri3",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      3,
	})
	s.Require().Nil(err)
	metric1Run3, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.2,
		Timestamp: 1511111111,
		Step:      6,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  3,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		Iter:      4,
	})
	s.Require().Nil(err)
	metric2Run3, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric2",
		Value:     1.6,
		Timestamp: 2522222222,
		Step:      2,
		IsNan:     false,
		RunID:     run3.ID,
		LastIter:  4,
	})
	s.Require().Nil(err)

	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param3",
		Value: "value3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1, run2, run3}
	tests := []struct {
		name    string
		request request.SearchMetricsRequest
		metrics []*models.LatestMetric
	}{
		{
			name: "SearchMetricNameOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric3")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric1Run2,
				metric2Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("3"))`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastOperationLess",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 1)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastStepOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 1)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 1)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationLess",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 1)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchRunArchived",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.archived == True`,
			},
		},
		{
			name: "SearchRunNotArchived",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.archived == False`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunActive",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.active == True`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunNotActive",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.active == False`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchDurationOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.duration > 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunDurationOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.duration >= 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunDurationOperationLess",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and run.duration < %d`,
					(run3.EndTime.Int64-run3.StartTime.Int64)/1000,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunDurationOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and run.duration <= %d`,
					(run3.EndTime.Int64-run3.StartTime.Int64)/1000,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunDurationOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.duration == 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunDurationOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.duration != 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at > 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at >= 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at != 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at == 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunStartTimeOperationLess",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at < 222222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchRunStartTimeOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.created_at <= 222222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at > 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at >= 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at != 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at == 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationLess",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at < 444444444`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunFinalizedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.finalized_at <= 444444444`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunHashOperationEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
					`or (metric.name == "TestMetric3")) and run.hash == "%s"`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunHashOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
					`or (metric.name == "TestMetric3")) and run.hash != "%s"`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.name != "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.name == "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunNameOperationIn",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and "Run3" in run.name`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationNotIn",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and "Run3" not in run.name`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.name.startswith("Test")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.name.endswith('3')`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunExperimentOperationEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and run.experiment == "%s"`,
					experiment.Name,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchRunExperimentOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and run.experiment != "%s"`,
					experiment.Name,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunTagOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.tags['mlflow.runName'] == "TestRunTag1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchRunTagOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and run.tags['mlflow.runName'] != "TestRunTag1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWithAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEndsWithAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationStartsWithAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEndsWithAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name == "TestMetric1" and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name == "TestMetric1" and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name != "TestMetric1" and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name != "TestMetric1" and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationStartsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name.startswith("Test") and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationStartsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name.startswith("Test") and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationEndsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name.endswith("Metric2") and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationEndsWithAndNotEqual",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(metric.name.endswith("Metric2") and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedOperationAtEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNamehAndRunFinalizedAtOperationStartsWithAndLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationStartsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationEndsWithAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name == "TestMetric1" and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameNotEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name != "TestMetric1" and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameStartsWithAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.startswith("Test") and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndLessOrEqual",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricNameEndsWithAndRunCreatedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(metric.name.endswith("Metric2") and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and re.match("TestRun3", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and re.search("TestRun3", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name == "TestRun3")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name != "TestRun3")`,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.name.endswith("Run2"))`,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and re.match("TestRun3", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and re.search("TestRun3", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name == "TestRun3")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name != "TestRun3")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.name.endswith("Run3"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationGreaterOrEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and re.match("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and re.search("TestRun1", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name == "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.name != "TestRun1")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunNameOperationLessOrEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration != 222222)`,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration < 222222)`,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.duration <= 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration < 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.duration <= 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration == 222222)`,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration == 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration != 222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunDurationOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.hash == "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2")`+
						`or (metric.name == "TestMetric3")) and (metric.last != 1.1) and run.hash == "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last != 1.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last > 1.1) and run.hash == "%s")`,
					run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last > 1.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last >= 1.1) and run.hash == "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last >= 1.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last < 3.1) and run.hash == "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last < 3.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last <= 3.1) and run.hash == "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunHashOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(
					`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
						`or (metric.name == "TestMetric3")) and (metric.last <= 3.1) and run.hash != "%s")`,
					run1.ID,
				),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at > 123456789)`,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at < 444444444)`,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunFinalizedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at > 123456789)`,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.1) and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at < 222222222)`,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last == 1.6) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at == 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at != 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last != 1.6) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at == 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at != 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at < 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last > 1.6) and run.created_at <= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at == 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at != 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last >= 1.6) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at == 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at != 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last < 1.6) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at == 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at != 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricLastAndRunCreatedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last <= 1.6) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") or ` +
					`(metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationNotEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationGreaterOrEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and re.match("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and re.search("TestRun2", run.name))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name == "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name != "TestRun2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessrOrEqualsAndStartsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name.startswith("Test"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunNameOperationLessOrEqualsAndEndsWith",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.name.endswith("Run2"))`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  == 2) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  != 2) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  > 2) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  >= 2) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  < 3) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration == 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration != 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration >= 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunDurationOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step  <= 3) and run.duration <= 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunHashOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: fmt.Sprintf(`(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") `+
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunFinalizedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at > 123456789)`,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at < 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step == 2) and run.created_at <= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationNotEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step != 2) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step > 2) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric3Run1,
				metric1Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationGreaterOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step >= 2) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at < 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step < 3) and run.created_at <= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at > 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at >= 111111111)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndLess",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricLastStepAndRunCreatedAtOperationLessOrEqualsAndLessOrEquals",
			request: request.SearchMetricsRequest{
				Query: `(((metric.name == "TestMetric1") or (metric.name == "TestMetric2") ` +
					`or (metric.name == "TestMetric3")) and (metric.last_step <= 3) and run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric2Run1,
				metric3Run1,
				metric1Run2,
				metric2Run2,
				metric3Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricComplexQuery",
			request: request.SearchMetricsRequest{
				Query: `((metric.name == "TestMetric1") or (metric.name == "TestMetric2")) ` +
					`and metric.last_step >= 1 and (run.name.endswith("2") or re.match("TestRun1", run.name)) ` +
					`and (metric.last < 1.6) and run.duration > 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := new(bytes.Buffer)
			s.Require().Nil(
				s.AIMClient().WithQuery(
					tt.request,
				).WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithResponse(
					resp,
				).DoRequest("/runs/search/metric"),
			)
			decodedData, err := encoding.Decode(resp)
			s.Require().Nil(err)

			var decodedMetrics []*models.LatestMetric
			for _, run := range runs {
				metricCount := 0
				for decodedData[fmt.Sprintf("%v.traces.%d.name", run.ID, metricCount)] != nil {
					prefix := fmt.Sprintf("%v.traces.%d", run.ID, metricCount)
					epochsKey := prefix + ".epochs.blob"
					itersKey := prefix + ".iters.blob"
					nameKey := prefix + ".name"
					timestampsKey := prefix + ".timestamps.blob"
					valuesKey := prefix + ".values.blob"

					m := models.LatestMetric{
						Key:       decodedData[nameKey].(string),
						Value:     decodedData[valuesKey].([]float64)[0],
						Timestamp: int64(decodedData[timestampsKey].([]float64)[0] * 1000),
						Step:      int64(decodedData[epochsKey].([]float64)[0]),
						IsNan:     false,
						RunID:     run.ID,
						LastIter:  int64(decodedData[itersKey].([]float64)[0]),
					}
					decodedMetrics = append(decodedMetrics, &m)
					metricCount++
				}
			}

			// Check if the received metrics match the expected ones
			s.Equal(tt.metrics, decodedMetrics)
		})
	}
}
