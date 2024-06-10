package run

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogOutputTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogOutputTestSuite(t *testing.T) {
	suite.Run(t, new(LogOutputTestSuite))
}

func (s *LogOutputTestSuite) Test_Ok() {
	tests := []struct {
		name string
		data []string
	}{
		{
			name: "TestWithValidParameters",
			data: []string{"a log row"},
		},
		{
			name: "TestTruncation",
			data: []string{
				"log row 1",
				"log row 2",
				"log row 3",
				"log row 4",
				"log row 5",
				"log row 6",
				"log row 7",
				"log row 8",
				"log row 9",
				"log row 10",
				"log row 11",
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
				ExperimentID:   *s.DefaultExperiment.ID,
				SourceType:     "JOB",
				LifecycleStage: models.LifecycleStageActive,
				Status:         models.StatusRunning,
			})
			s.Require().Nil(err)

			for i := range tt.data {
				req := request.LogOutputRequest{
					RunID: run.ID,
					Data:  tt.data[i],
				}
				resp := map[string]any{}
				s.Require().Nil(
					s.MlflowClient().WithMethod(
						http.MethodPost,
					).WithRequest(
						req,
					).WithResponse(
						&resp,
					).DoRequest(
						"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogOutputRoute,
					),
				)
				s.Empty(resp)
			}

			// verify truncation to helpers.MaxLogRows
			max := helpers.MaxLogRows
			if len(tt.data) < max {
				max = len(tt.data)
			}
			logs, err := s.LogFixtures.GetByRunID(context.Background(), run.ID)
			s.Require().Nil(err)
			s.Assert().Equal(max, len(logs))
			for i := 1; i <= max; i++ {
				intputIndex := len(tt.data) - i
				outputIndex := max - i
				s.Assert().Equal(tt.data[intputIndex], logs[outputIndex].Value)
				s.Assert().WithinRange(time.Unix(logs[outputIndex].Timestamp, 0),
					time.Unix(time.Now().Unix()-1, 0),
					time.Unix(time.Now().Unix(), 0),
				)
			}
		})
	}
}

func (s *LogOutputTestSuite) Test_Error() {
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *s.DefaultExperiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)
	tests := []struct {
		name  string
		runID string
		data  string
		error *api.ErrorResponse
	}{
		{
			name:  "MissingData",
			runID: run.ID,
			error: &api.ErrorResponse{
				Message:    "Missing value for required parameter 'data'",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "MissingRunID",
			data: "some log message",
			error: &api.ErrorResponse{
				Message:    "Missing value for required parameter 'run_id'",
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			req := request.LogOutputRequest{
				RunID: tt.runID,
				Data:  tt.data,
			}
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					req,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogOutputRoute,
				),
			)
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
