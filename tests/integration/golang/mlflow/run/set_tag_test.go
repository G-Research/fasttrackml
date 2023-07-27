//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"

	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SetRunTagTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	tagFixtures        *fixtures.TagFixtures
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestSetRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(SetRunTagTestSuite))
}

func (s *SetRunTagTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	tagFixtures, err := fixtures.NewTagFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.tagFixtures = tagFixtures
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func (s *SetRunTagTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// create test experiment.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// create test run for the experiment
	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
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
	assert.Nil(s.T(), err)

	resp := fiber.Map{}
	err = s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSetTagRoute),
		request.SetRunTagRequest{
			RunID: run.ID,
			Key:   "tag1",
			Value: "value1",
		},
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), fiber.Map{}, resp)

	// make sure that new tag has been created.
	tags, err := s.tagFixtures.GetByRunID(context.Background(), run.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(tags))
	assert.Equal(s.T(), []models.Tag{
		{
			RunID: run.ID,
			Key:   "tag1",
			Value: "value1",
		},
	}, tags)
}

func (s *SetRunTagTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

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
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsSetTagRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
