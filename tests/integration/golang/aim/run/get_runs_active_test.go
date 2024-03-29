package run

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
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
	tests := []struct {
		name         string
		wantRunCount int
		beforeRunFn  func()
	}{
		{
			name:         "GetActiveRuns",
			wantRunCount: 3,
			beforeRunFn: func() {
				experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
					Name:           uuid.New().String(),
					NamespaceID:    s.DefaultNamespace.ID,
					LifecycleStage: models.LifecycleStageActive,
				})
				s.Require().Nil(err)

				s.runs, err = s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 3)
				s.Require().Nil(err)
			},
		},
		{
			name:         "GetActiveRunsSkipsFinished",
			wantRunCount: 2,
			beforeRunFn: func() {
				// set 3rd run to status = StatusFinished
				s.runs[2].Status = models.StatusFinished
				s.Require().Nil(s.RunFixtures.UpdateRun(context.Background(), s.runs[2]))
			},
		},
		{
			name:         "GetActiveRunsWithNoData",
			wantRunCount: 0,
			beforeRunFn: func() {
				// set 1t and 2d run to status = StatusFinished
				s.runs[1].Status = models.StatusFinished
				s.Require().Nil(s.RunFixtures.UpdateRun(context.Background(), s.runs[1]))
				s.runs[0].Status = models.StatusFinished
				s.Require().Nil(s.RunFixtures.UpdateRun(context.Background(), s.runs[0]))
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
			decodedData, err := encoding.NewDecoder(resp).Decode()
			s.Require().Nil(err)

			responseCount := 0
			for _, run := range s.runs {
				respNameKey := fmt.Sprintf("%v.props.name", run.ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", run.ID)
				expNameKey := fmt.Sprintf("%v.props.experiment.name", run.ID)
				startTimeKey := fmt.Sprintf("%v.props.creation_time", run.ID)
				endTimeKey := fmt.Sprintf("%v.props.end_time", run.ID)
				activeKey := fmt.Sprintf("%v.props.active", run.ID)
				archivedKey := fmt.Sprintf("%v.props.archived", run.ID)
				if run.Status == models.StatusRunning && run.LifecycleStage ==
					models.LifecycleStageActive {
					s.Equal(run.Name, decodedData[respNameKey])
					s.Equal(fmt.Sprintf("%v", run.ExperimentID), decodedData[expIdKey])
					s.Equal(run.Experiment.Name, decodedData[expNameKey])
					s.Equal(run.Status == models.StatusRunning, decodedData[activeKey])
					s.Equal(false, decodedData[archivedKey])
					s.Equal(float64(run.StartTime.Int64)/1000, decodedData[startTimeKey])
					s.Equal(float64(run.EndTime.Int64)/1000, decodedData[endTimeKey])
					responseCount++
				} else {
					s.Nil(decodedData[respNameKey])
					s.Nil(decodedData[expIdKey])
					s.Nil(decodedData[expNameKey])
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
