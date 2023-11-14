//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RestoreRunTestSuite struct {
	helpers.BaseTestSuite
}

func TestRestoreRunTestSuite(t *testing.T) {
	suite.Run(t, new(RestoreRunTestSuite))
}

func (s *RestoreRunTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

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
		LifecycleStage: models.LifecycleStageDeleted,
	})
	require.Nil(s.T(), err)

	// create tags, metrics, params.
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "metric1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
	})
	require.Nil(s.T(), err)

	req := request.RestoreRunRequest{
		RunID: run.ID,
	}
	resp := fiber.Map{}
	require.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsRestoreRoute),
		),
	)
	assert.Equal(s.T(), fiber.Map{}, resp)

	// check that run has been updated in database.
	run, err = s.RunFixtures.GetRun(context.Background(), run.ID)
	require.Nil(s.T(), err)
	assert.Equal(s.T(), models.LifecycleStageActive, run.LifecycleStage)
}

func (s *RestoreRunTestSuite) Test_Error() {
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.RestoreRunRequest
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.RestoreRunRequest{},
			error: api.NewInvalidParameterValueError(
				"Missing value for required parameter 'run_id'",
			),
		},
		{
			name: "NotFoundRun",
			request: request.RestoreRunRequest{
				RunID: "id",
			},
			error: api.NewResourceDoesNotExistError("unable to find run 'id'"),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient.WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsRestoreRoute),
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
