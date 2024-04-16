package run

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) Test_Ok() {
	// create test runs.
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:     strings.ReplaceAll(uuid.New().String(), "-", ""),
		Name:   "TestRun",
		Status: models.StatusRunning,
		StartTime: sql.NullInt64{
			Int64: 1234567890,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 1234567899,
			Valid: true,
		},
		SourceType:     "JOB",
		ArtifactURI:    "artifact_uri",
		ExperimentID:   *s.DefaultExperiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	finishedRun, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:     strings.ReplaceAll(uuid.New().String(), "-", ""),
		Name:   "TestFinishedRun",
		Status: models.StatusFinished,
		StartTime: sql.NullInt64{
			Int64: 1234567890,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 1234567899,
			Valid: true,
		},
		SourceType:     "JOB",
		ArtifactURI:    "artifact_uri",
		ExperimentID:   *s.DefaultExperiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	tests := []struct {
		name            string
		request         request.UpdateRunRequest
		expectedRunInfo response.UpdateRunResponse
	}{
		{
			name: "UpdateRun",
			request: request.UpdateRunRequest{
				RunID:   run.ID,
				Name:    "UpdatedName",
				Status:  string(models.StatusScheduled),
				EndTime: 1111111111,
			},
			expectedRunInfo: response.UpdateRunResponse{
				RunInfo: response.RunInfoPartialResponse{
					ID:             run.ID,
					UUID:           run.ID,
					Name:           "UpdatedName",
					ExperimentID:   fmt.Sprintf("%d", *s.DefaultExperiment.ID),
					ArtifactURI:    run.ArtifactURI,
					Status:         string(models.StatusScheduled),
					StartTime:      1234567890,
					EndTime:        1111111111,
					LifecycleStage: string(models.LifecycleStageActive),
				},
			},
		},
		{
			name: "RestartRun",
			request: request.UpdateRunRequest{
				RunID:  finishedRun.ID,
				Name:   "RestartedRun",
				Status: string(models.StatusRunning),
			},
			expectedRunInfo: response.UpdateRunResponse{
				RunInfo: response.RunInfoPartialResponse{
					ID:             finishedRun.ID,
					UUID:           finishedRun.ID,
					Name:           "RestartedRun",
					ExperimentID:   fmt.Sprintf("%d", *s.DefaultExperiment.ID),
					ArtifactURI:    finishedRun.ArtifactURI,
					Status:         string(models.StatusRunning),
					StartTime:      1234567890,
					EndTime:        0,
					LifecycleStage: string(models.LifecycleStageActive),
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := response.UpdateRunResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute,
				),
			)
			s.Equal(tt.expectedRunInfo, resp)

			// check that run has been updated in database.
			run, err = s.RunFixtures.GetRun(context.Background(), tt.request.RunID)
			s.Require().Nil(err)
			s.Equal(tt.expectedRunInfo.RunInfo.Name, run.Name)
			s.Equal(tt.expectedRunInfo.RunInfo.Status, string(run.Status))
			s.Equal(tt.expectedRunInfo.RunInfo.EndTime, run.EndTime.Int64)
		})
	}
}

func (s *UpdateRunTestSuite) Test_Error() {
	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.UpdateRunRequest
	}{
		{
			name:    "UpdateWithInvalidExperimentID",
			request: request.UpdateRunRequest{},
			error:   api.NewInvalidParameterValueError(`Missing value for required parameter 'run_id'`),
		},
		{
			name: "UpdateWithNotExistingRun",
			request: request.UpdateRunRequest{
				RunID: "1",
			},
			error: api.NewResourceDoesNotExistError("unable to find run '1'"),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
