//go:build integration

package run

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunsActiveTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	runs               []*models.Run
}

func TestGetRunsActiveTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunsActiveTestSuite))
}

func (s *GetRunsActiveTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	s.runs, err = s.runFixtures.CreateRuns(context.Background(), exp, 3)
	assert.Nil(s.T(), err)

	// set 3rd run to status = StatusFinished
	s.runs[2].Status = models.StatusFinished
	assert.Nil(s.T(), s.runFixtures.UpdateRun(context.Background(), s.runs[2]))
}

func (s *GetRunsActiveTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name  string
		wantRunCount int
	}{
		{
			name:  "GetActiveRunsWhenPresent",
			wantRunCount: 2,
		},
		{
			name:  "GetActiveRunsWhenNotPresent",
			wantRunCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			if tt.wantRunCount == 0 {
				assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
			}
			data, err := s.client.DoStreamRequest(
				http.MethodGet,
				"/runs/active",
			)
			assert.Nil(s.T(), err)

			decodedData, err := encoding.Decode(bytes.NewBuffer(data))
			assert.Nil(s.T(), err)
			
			for idx := 0; idx < tt.wantRunCount; idx++ {
				respNameKey := fmt.Sprintf("%v.props.name", s.runs[idx].ID)
				expIdKey := fmt.Sprintf("%v.props.experiment.id", s.runs[idx].ID)
				startTimeKey := fmt.Sprintf("%v.props.creation_time", s.runs[idx].ID)
				endTimeKey := fmt.Sprintf("%v.props.end_time", s.runs[idx].ID)
				activeKey := fmt.Sprintf("%v.props.active", s.runs[idx].ID)
				archivedKey := fmt.Sprintf("%v.props.archived", s.runs[idx].ID)
				if s.runs[idx].Status == models.StatusRunning {
					assert.Equal(s.T(), s.runs[idx].Name, decodedData[respNameKey])
					assert.Equal(s.T(), fmt.Sprintf("%v", s.runs[idx].ExperimentID), decodedData[expIdKey])
					assert.Equal(s.T(), s.runs[idx].Status == models.StatusRunning, decodedData[activeKey])
					assert.Equal(s.T(), s.runs[idx].LifecycleStage == models.LifecycleStageDeleted, decodedData[archivedKey])
					assert.Equal(s.T(), s.runs[idx].StartTime.Int64, int64(decodedData[startTimeKey].(float64)))
					assert.Equal(s.T(), s.runs[idx].EndTime.Int64, int64(decodedData[endTimeKey].(float64)))
				} else {
					assert.Nil(s.T(), decodedData[respNameKey])
					assert.Nil(s.T(), decodedData[expIdKey])
					assert.Nil(s.T(), decodedData[activeKey])
					assert.Nil(s.T(), decodedData[archivedKey])
					assert.Nil(s.T(), decodedData[startTimeKey])
					assert.Nil(s.T(), decodedData[endTimeKey])
				}
			}
		})
	}
}
