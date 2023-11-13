//go:build integration

package run

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunsActiveTestSuite struct {
	helpers.BaseTestSuite
	runs []*models.Run
}

func TestGetRunsActiveTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunsActiveTestSuite))
}

func (s *GetRunsActiveTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name         string
		wantRunCount int
		beforeRunFn  func()
	}{
		{
			name:         "GetActiveRunsWithNoData",
			wantRunCount: 0,
		},
		{
			name:         "GetActiveRuns",
			wantRunCount: 3,
			beforeRunFn: func() {
				experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
					Name:           uuid.New().String(),
					NamespaceID:    namespace.ID,
					LifecycleStage: models.LifecycleStageActive,
				})
				require.Nil(s.T(), err)

				s.runs, err = s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 3)
				require.Nil(s.T(), err)
			},
		},
		{
			name:         "GetActiveRunsSkipsFinished",
			wantRunCount: 2,
			beforeRunFn: func() {
				// set 3rd run to status = StatusFinished
				s.runs[2].Status = models.StatusFinished
				require.Nil(s.T(), s.RunFixtures.UpdateRun(context.Background(), s.runs[2]))
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			if tt.beforeRunFn != nil {
				tt.beforeRunFn()
			}
			resp := new(bytes.Buffer)
			require.Nil(
				s.T(),
				s.AIMClient.WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithResponse(
					resp,
				).DoRequest("/runs/active"),
			)
			decodedData, err := encoding.Decode(resp)
			require.Nil(s.T(), err)

			responseCount := 0
			for _, run := range s.runs {
				respNameKey := fmt.Sprintf("%v.props.name", run.ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", run.ID)
				startTimeKey := fmt.Sprintf("%v.props.creation_time", run.ID)
				endTimeKey := fmt.Sprintf("%v.props.end_time", run.ID)
				activeKey := fmt.Sprintf("%v.props.active", run.ID)
				archivedKey := fmt.Sprintf("%v.props.archived", run.ID)
				if run.Status == models.StatusRunning && run.LifecycleStage ==
					models.LifecycleStageActive {
					assert.Equal(s.T(), run.Name, decodedData[respNameKey])
					assert.Equal(s.T(), fmt.Sprintf("%v", run.ExperimentID), decodedData[expIdKey])
					assert.Equal(s.T(), run.Status == models.StatusRunning, decodedData[activeKey])
					assert.Equal(s.T(), false, decodedData[archivedKey])
					assert.Equal(s.T(), run.StartTime.Int64, int64(decodedData[startTimeKey].(float64)))
					assert.Equal(s.T(), run.EndTime.Int64, int64(decodedData[endTimeKey].(float64)))
					responseCount++
				} else {
					assert.Nil(s.T(), decodedData[respNameKey])
					assert.Nil(s.T(), decodedData[expIdKey])
					assert.Nil(s.T(), decodedData[activeKey])
					assert.Nil(s.T(), decodedData[archivedKey])
					assert.Nil(s.T(), decodedData[startTimeKey])
					assert.Nil(s.T(), decodedData[endTimeKey])
				}
			}
			assert.Equal(s.T(), tt.wantRunCount, responseCount)
		})
	}
}
