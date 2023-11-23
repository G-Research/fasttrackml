//go:build integration

package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"

	"gorm.io/datatypes"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunsActiveTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetRunsActiveTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunsActiveTestSuite))
}

func (s *GetRunsActiveTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	runToMetricContextMap := map[string]*models.Context{}

	// create test data.
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
	require.Nil(s.T(), err)

	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id1",
		Name:           "TestRun1",
		Status:         models.StatusRunning,
		StartTime:      sql.NullInt64{Int64: 123456789, Valid: true},
		EndTime:        sql.NullInt64{Int64: 123456789, Valid: true},
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext1, err := s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key1": "key1", "value1": "value1"}`),
	})
	require.Nil(s.T(), err)
	// save connection between `run` and `context` for further usage.
	runToMetricContextMap[run1.ID] = metricContext1

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key1",
		Value:     123.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		ContextID: common.GetPointer(metricContext1.ID),
	})
	require.Nil(s.T(), err)

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id2",
		Name:           "TestRun2",
		Status:         models.StatusRunning,
		StartTime:      sql.NullInt64{Int64: 123456789, Valid: true},
		EndTime:        sql.NullInt64{Int64: 123456789, Valid: true},
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext2, err := s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key2": "key2", "value2": "value2"}`),
	})
	require.Nil(s.T(), err)
	// save connection between `run` and `context` for further usage.
	runToMetricContextMap[run2.ID] = metricContext2

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key2",
		Value:     123.2,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run2.ID,
		ContextID: common.GetPointer(metricContext2.ID),
	})
	require.Nil(s.T(), err)

	run3, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id3",
		Name:           "TestRun3",
		Status:         models.StatusRunning,
		StartTime:      sql.NullInt64{Int64: 123456789, Valid: true},
		EndTime:        sql.NullInt64{Int64: 123456789, Valid: true},
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext3, err := s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key3": "key3", "value3": "value3"}`),
	})
	require.Nil(s.T(), err)
	// save connection between `run` and `context` for further usage.
	runToMetricContextMap[run3.ID] = metricContext3

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "key3",
		Value:     123.3,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run3.ID,
		ContextID: common.GetPointer(metricContext3.ID),
	})
	require.Nil(s.T(), err)

	// run tests over test data.
	tests := []struct {
		name         string
		wantRunCount int
		beforeRunFn  func()
	}{
		{
			name:         "GetActiveRuns",
			wantRunCount: 3,
		},
		{
			name:         "GetActiveRunsSkipsFinished",
			wantRunCount: 2,
			beforeRunFn: func() {
				// set 3rd run to status = StatusFinished
				run3.Status = models.StatusFinished
				require.Nil(s.T(), s.RunFixtures.UpdateRun(context.Background(), run3))
			},
		},
		{
			name:         "GetActiveRunsWithNoData",
			wantRunCount: 0,
			beforeRunFn: func() {
				// set 1t and 2d run to status = StatusFinished
				run2.Status = models.StatusFinished
				require.Nil(s.T(), s.RunFixtures.UpdateRun(context.Background(), run2))
				run1.Status = models.StatusFinished
				require.Nil(s.T(), s.RunFixtures.UpdateRun(context.Background(), run1))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.beforeRunFn != nil {
				tt.beforeRunFn()
			}
			resp := new(bytes.Buffer)
			s.Require().Nil(
				s.AIMClient().WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithResponse(
					resp,
				).DoRequest("/runs/active"),
			)
			decodedData, err := encoding.Decode(resp)
			require.Nil(s.T(), err)

			responseCount := 0
			for _, run := range []*models.Run{run1, run2, run3} {
				respNameKey := fmt.Sprintf("%v.props.name", run.ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", run.ID)
				startTimeKey := fmt.Sprintf("%v.props.creation_time", run.ID)
				endTimeKey := fmt.Sprintf("%v.props.end_time", run.ID)
				activeKey := fmt.Sprintf("%v.props.active", run.ID)
				archivedKey := fmt.Sprintf("%v.props.archived", run.ID)
				// contextKey := fmt.Sprintf("%v.traces.metric.0.context", run.ID)
				if run.Status == models.StatusRunning && run.LifecycleStage == models.LifecycleStageActive {
					assert.Equal(s.T(), run.Name, decodedData[respNameKey])
					assert.Equal(s.T(), fmt.Sprintf("%v", run.ExperimentID), decodedData[expIdKey])
					assert.Equal(s.T(), run.Status == models.StatusRunning, decodedData[activeKey])
					assert.Equal(s.T(), false, decodedData[archivedKey])
					assert.Equal(s.T(), float64(run.StartTime.Int64)/1000, decodedData[startTimeKey])
					assert.Equal(s.T(), float64(run.EndTime.Int64)/1000, decodedData[endTimeKey])
					// assert.Equal(s.T(), runToMetricContextMap[run.ID].Json.String(), decodedData[contextKey])
					responseCount++
				} else {
					s.Nil(decodedData[respNameKey])
					s.Nil(decodedData[expIdKey])
					s.Nil(decodedData[activeKey])
					s.Nil(decodedData[archivedKey])
					s.Nil(decodedData[startTimeKey])
					s.Nil(decodedData[endTimeKey])
				}
			}
			s.Equal(tt.wantRunCount, responseCount)
		})
	}
}
