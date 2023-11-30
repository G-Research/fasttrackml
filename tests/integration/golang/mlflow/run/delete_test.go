//go:build integration

package run

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteRunTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTestSuite))
}

func (s *DeleteRunTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create experiment
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

	// create run for the experiment
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		Name:           "TestRun",
		Status:         models.StatusRunning,
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	tests := []struct {
		name    string
		request request.DeleteRunRequest
	}{
		{
			name:    "DeleteRunSucceedsWithExistingRunID",
			request: request.DeleteRunRequest{RunID: run.ID},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := map[string]any{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteRoute,
				),
			)
			s.Empty(resp)

			archivedRuns, err := s.RunFixtures.GetRuns(context.Background(), run.ExperimentID)

			s.Require().Nil(err)
			s.Equal(1, len(archivedRuns))
			s.Equal(run.ID, archivedRuns[0].ID)
			s.Equal(models.LifecycleStageDeleted, archivedRuns[0].LifecycleStage)
		})
	}
}

func (s *DeleteRunTestSuite) Test_Error() {
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
		request request.DeleteRunRequest
	}{
		{
			name:    "DeleteRunFailsWithNonExistingRunID",
			request: request.DeleteRunRequest{RunID: "not-an-id"},
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
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteRoute,
				),
			)
			s.Require().Nil(err)
			s.Equal(api.NewResourceDoesNotExistError("unable to find run 'not-an-id'").Error(), resp.Error())
		})
	}
}
