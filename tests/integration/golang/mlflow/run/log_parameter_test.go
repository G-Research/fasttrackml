package run

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogParamTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogParamTestSuite(t *testing.T) {
	suite.Run(t, new(LogParamTestSuite))
}

func (s *LogParamTestSuite) Test_Ok() {
	tests := []struct {
		name       string
		runID      string
		key        string
		valueStr   *string
		valueFloat *float64
		valueInt   *int64
	}{
		{
			name:     "TestWithValidParameters",
			runID:    strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:      "key1",
			valueStr: common.GetPointer("value1"),
		},
		{
			name:     "TestWithDuplicate",
			runID:    strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:      "key1",
			valueStr: common.GetPointer("value1"),
		},
		{
			name:       "TestWithFloat",
			runID:      strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:        "key2",
			valueFloat: common.GetPointer(float64(123.45)),
		},
		{
			name:     "TestWithInt",
			runID:    strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:      "key2",
			valueInt: common.GetPointer(int64(123)),
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             tt.runID,
				ExperimentID:   *s.DefaultExperiment.ID,
				SourceType:     "JOB",
				LifecycleStage: models.LifecycleStageActive,
				Status:         models.StatusRunning,
			})
			s.Require().Nil(err)

			req := request.LogParamRequest{
				RunID:      run.ID,
				Key:        tt.key,
				ValueStr:   tt.valueStr,
				ValueFloat: tt.valueFloat,
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
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
				),
			)
			s.Empty(resp)
		})
	}
}

func (s *LogParamTestSuite) Test_Error() {
	tests := []struct {
		key      string
		name     string
		runID    string
		valueStr *string
		error    *api.ErrorResponse
	}{
		{
			name:     "TestMissingParamKey",
			runID:    strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:      "",
			valueStr: common.GetPointer("value1"),
			error: &api.ErrorResponse{
				Message:    "Missing value for required parameter 'key'",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "TestConflictingParam",
			runID:    strings.ReplaceAll(uuid.NewString(), "-", ""),
			key:      "key1",
			valueStr: common.GetPointer("value2"),
			error: &api.ErrorResponse{
				Message:    "unable to insert params for run",
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
				ID:             tt.runID,
				ExperimentID:   *s.DefaultExperiment.ID,
				SourceType:     "JOB",
				LifecycleStage: models.LifecycleStageActive,
				Status:         models.StatusRunning,
			})
			s.Require().Nil(err)

			param := models.Param{
				Key:      tt.key,
				ValueStr: common.GetPointer("value1"),
				RunID:    tt.runID,
			}
			_, err = s.ParamFixtures.CreateParam(context.Background(), &param)
			s.Require().Nil(err)

			req := request.LogParamRequest{
				RunID:    run.ID,
				Key:      tt.key,
				ValueStr: tt.valueStr,
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
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
				),
			)
			s.Contains(resp.Message, tt.error.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
