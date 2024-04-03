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
	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchMetricsTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchMetricsTestSuite))
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
		Key:      "param1",
		ValueStr: common.GetPointer[string]("value1"),
		RunID:    run1.ID,
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
		Key:      "param2",
		ValueStr: common.GetPointer[string]("value2"),
		RunID:    run2.ID,
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
		Key:      "param3",
		ValueStr: common.GetPointer[string]("value3"),
		RunID:    run3.ID,
	})
	s.Require().Nil(err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "mlflow.runName",
		Value: "TestRunTag3",
		RunID: run3.ID,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1, run2, run3}
	contextValue := "{\"testkey\":\"testvalue\"}"
	tests := []struct {
		name    string
		request request.SearchMetricsRequest
		metrics []*models.LatestMetric
	}{
		{
			name: "SearchMetric",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.archived == True`,
			},
		},
		{
			name: "SearchRunNotArchived",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.active == True`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.duration == 0`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.created_at > 123456789`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunStartTimeOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.created_at >= 123456789`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.created_at == 123456789`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.finalized_at == 123456789`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.finalized_at < 444444444`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: fmt.Sprintf(`run.hash == "%s"`, run1.ID),
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name == "TestRun1"`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `"Run3" in run.name`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunNameOperationNotIn",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name.endswith('3')`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
				metric2Run3,
			},
		},
		{
			name: "SearchRunExperimentOperationEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.tags['mlflow.runName'] == "TestRunTag1"`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
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
			name: "SearchMetricNameAndRunNameOperationRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `re.match("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `re.search("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `run.name == "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.name != "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationStartsWith",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.name.startswith("Test")`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.name.endswith("Run2")`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpMatchFunction",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `re.match("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndRegexpSearchFunction",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `re.search("TestRun1", run.name)`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name == "TestRun1"`,
			},
			metrics: []*models.LatestMetric{
				metric2Run1,
				metric3Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunNameOperationNotEqualsAndNotEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name != "TestRun1"`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name.startswith("Test")`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `run.name.endswith("Run2")`,
			},
			metrics: []*models.LatestMetric{
				metric2Run2,
				metric3Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `run.duration == 222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationNotEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.duration != 222222`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGreater",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.duration > 0`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationGreaterOrEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.duration >= 0`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.duration < 333333)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunDurationOperationLessOrEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `run.duration <= 333333`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: fmt.Sprintf(`(run.hash == "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunHashOperationNotEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: fmt.Sprintf(`(run.hash != "%s")`, run1.ID),
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.finalized_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.finalized_at >= 123456789)`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.finalized_at < 444444444)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.finalized_at <= 444444444)`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.finalized_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.finalized_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameAndRunFinalizedAtOperationNotEqualsAndGreater",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `(run.finalized_at > 123456789)`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric3",
						Context: "{}",
					},
				},
				Query: `(run.finalized_at >= 123456789)`,
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
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreater",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.created_at > 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationGreaterOrEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.created_at >= 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLess",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.created_at < 222222222)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
				metric1Run2,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationLessOrEqual",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.created_at <= 222222222)`,
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
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
				},
				Query: `(run.created_at == 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
		{
			name: "SearchMetricNameEqualsAndRunCreatedAtOperationNotEquals",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.created_at != 123456789)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric1Run3,
			},
		},
		{
			name: "SearchMetricComplexQuery",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: "{}",
					},
					{
						Key:     "TestMetric2",
						Context: "{}",
					},
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				Query: `(run.name.endswith("2") or re.match("TestRun1", run.name) and run.duration > 0)`,
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
				metric2Run2,
			},
		},
		{
			name: "SearchMetricContext",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
			},
			metrics: []*models.LatestMetric{
				metric1Run2,
			},
		},
		{
			name: "SearchMetricContextWithXAxis",
			request: request.SearchMetricsRequest{
				Metrics: []request.MetricTuple{
					{
						Key:     "TestMetric1",
						Context: contextValue,
					},
				},
				XAxis: `TestMetric2`,
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
