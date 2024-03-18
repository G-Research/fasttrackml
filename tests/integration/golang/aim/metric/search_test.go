package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchMetricsTestSuite(t *testing.T) {
	flag, ok := os.LookupEnv("FML_RUN_ORIGINAL_AIM_SERVICE")
	if ok && flag == "true" {
		suite.Run(t, new(SearchMetricsTestSuite))
	}
}

func (s *SearchMetricsTestSuite) Test_Ok() {
	// create test experiments.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    s.DefaultNamespace.ID,
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
		Context: models.Context{
			Json: types.JSONB(`{"testkey":"testvalue"}`),
		},
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
		Context: models.Context{
			Json: types.JSONB(`{"testkey":"testvalue"}`),
		},
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
			name: "SearchMetric",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchRunArchived",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.archived == True`,
			},
		},
		{
			name: "SearchRunNotArchived",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.archived == False`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.active == True`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.active == False`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.duration > 0`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.duration >= 0`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query: fmt.Sprintf(
					`run.duration < %d`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query: fmt.Sprintf(
					`run.duration <= %d`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.duration == 0`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.duration != 0`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.created_at > 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.created_at >= 123456789`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.created_at != 123456789`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.created_at == 123456789`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.created_at < 222222222`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.created_at <= 222222222`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.finalized_at > 123456789`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.finalized_at >= 123456789`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.finalized_at != 123456789`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.finalized_at == 123456789`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.finalized_at < 444444444`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.finalized_at <= 444444444`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              fmt.Sprintf(`run.hash == "%s"`, run1.ID),
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: fmt.Sprintf(`run.hash != "%s"`, run1.ID),
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.name != "TestRun1"`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name == "TestRun1"`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `"Run3" in run.name`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationNotIn",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `"Run3" not in run.name`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.name.startswith("Test")`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name.endswith('3')`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunExperimentOperationEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: fmt.Sprintf(
					`run.experiment == "%s"`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query: fmt.Sprintf(
					`run.experiment != "%s"`,
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
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.tags['mlflow.runName'] == "TestRunTag1"`,
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
				MetricsWithContext: []string{
					"TestMetric1-{}",
					"TestMetric2-{}",
					"TestMetric3-{}",
					`TestMetric1-{"testkey":"testvalue"}`,
				},
				Query: `run.tags['mlflow.runName'] != "TestRunTag1"`,
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
			name: "SearchMetricAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `re.match("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `re.search("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricAndRunNameOperationEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `run.name == "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.name != "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.name.startswith("Test")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.name.endswith("Run2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricsAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `re.match("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricsAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `re.search("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricsAndRunNameOperationdEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name == "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricsAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name != "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricsAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name.startswith("Test")`,
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
			name: "SearchMetricsxsAndRunNameOperationEndsWith",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `run.name.endswith("Run2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `run.duration == 222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationNotEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.duration != 222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationGreater",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.duration > 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationGreaterOrEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.duration >= 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationLess",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunDurationOperationLessOrEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `run.duration <= 333333`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunHashOperationEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              fmt.Sprintf(`(run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricsAndRunHashOperationNotEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              fmt.Sprintf(`(run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationGreater",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.finalized_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationLess",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.finalized_at <= 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedOperationAtEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `(run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
				metric2Run3,
			},
		},
		{
			name: "SearchMetricsAndRunFinalizedAtOperationNotEqualsAndGreaterOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric2-{}", "TestMetric3-{}"},
				Query:              `(run.finalized_at >= 123456789)`,
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
			name: "SearchMetricEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.created_at <= 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricEqualsAndRunCreatedAtOperationEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}"},
				Query:              `(run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricComplexQuery",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{"TestMetric1-{}", "TestMetric2-{}", `TestMetric1-{"testkey":"testvalue"}`},
				Query:              `(run.name.endswith("2") or re.match("TestRun1", run.name) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricContext",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{`TestMetric1-{"testkey":"testvalue"}`},
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricContextWithXAxis",
			request: request.SearchMetricsRequest{
				MetricsWithContext: []string{`TestMetric1-{"testkey":"testvalue"}`},
				XAxis:              `TestMetric2`,
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

			decodedData, err := encoding.NewDecoder(resp).Decode()
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

					contextPrefix := prefix + ".context"
					contx, err := helpers.ExtractContextBytes(contextPrefix, decodedData)
					s.Require().Nil(err)

					decodedContext, err := s.ContextFixtures.GetContextByJSON(
						context.Background(),
						string(contx),
					)
					s.Require().Nil(err)

					m := models.LatestMetric{
						Key:       decodedData[nameKey].(string),
						Value:     decodedData[valuesKey].([]float64)[0],
						Timestamp: int64(decodedData[timestampsKey].([]float64)[0] * 1000),
						Step:      int64(decodedData[epochsKey].([]float64)[0]),
						IsNan:     false,
						RunID:     run.ID,
						LastIter:  int64(decodedData[itersKey].([]float64)[0]),
						ContextID: decodedContext.ID,
						Context:   *decodedContext,
					}
					decodedMetrics = append(decodedMetrics, &m)
					metricCount++
				}
			}
			// Check if the received metrics match the expected ones
			s.Equal(len(tt.metrics), len(decodedMetrics))
			for i, metric := range tt.metrics {
				s.Equal(metric, decodedMetrics[i])
			}
		})
	}
}
