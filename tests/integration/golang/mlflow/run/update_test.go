//go:build integration

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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
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
	s.Require().Nil(err)

	// create test run for the experiment
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
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	req := request.UpdateRunRequest{
		RunID:   run.ID,
		Name:    "UpdatedName",
		Status:  string(models.StatusScheduled),
		EndTime: 1111111111,
	}
	resp := response.UpdateRunResponse{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute,
		),
	)
	s.NotEmpty(resp.RunInfo.ID)
	s.NotEmpty(resp.RunInfo.UUID)
	s.Equal("UpdatedName", resp.RunInfo.Name)
	s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.RunInfo.ExperimentID)
	s.Equal(int64(1234567890), resp.RunInfo.StartTime)
	s.Equal(int64(1111111111), resp.RunInfo.EndTime)
	s.Equal(string(models.StatusScheduled), resp.RunInfo.Status)
	s.NotEmpty(resp.RunInfo.ArtifactURI)
	s.Equal(string(models.LifecycleStageActive), resp.RunInfo.LifecycleStage)

	// check that run has been updated in database.
	run, err = s.RunFixtures.GetRun(context.Background(), run.ID)
	s.Require().Nil(err)
	s.Equal("UpdatedName", run.Name)
	s.Equal(models.StatusScheduled, run.Status)
	s.Equal(int64(1111111111), run.EndTime.Int64)
}

func (s *UpdateRunTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
