//go:build integration

package run

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SetRunTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestSetRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(SetRunTagTestSuite))
}

func (s *SetRunTagTestSuite) Test_Ok() {
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

	req := request.SetRunTagRequest{
		RunID: run.ID,
		Key:   "tag1",
		Value: "value1",
	}
	resp := fiber.Map{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSetTagRoute,
		),
	)
	s.Equal(fiber.Map{}, resp)

	// make sure that new tag has been created.
	tags, err := s.TagFixtures.GetByRunID(context.Background(), run.ID)
	s.Require().Nil(err)
	s.Equal(1, len(tags))
	s.Equal([]models.Tag{
		{
			RunID: run.ID,
			Key:   "tag1",
			Value: "value1",
		},
	}, tags)
}

func (s *SetRunTagTestSuite) Test_Error() {
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
		request request.SetRunTagRequest
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.SetRunTagRequest{},
			error: api.NewInvalidParameterValueError(
				"Missing value for required parameter 'run_id'",
			),
		},
		{
			name: "EmptyOrIncorrectKey",
			request: request.SetRunTagRequest{
				RunID: "id",
			},
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
		},
		{
			name: "NotFoundRun",
			request: request.SetRunTagRequest{
				Key:   "key1",
				RunID: "id",
			},
			error: api.NewResourceDoesNotExistError("Unable to find active run 'id'"),
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
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSetTagRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
