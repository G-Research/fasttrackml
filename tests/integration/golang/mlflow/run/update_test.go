//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func (s *UpdateRunTestSuite) Test_Ok() {
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

	resp := response.UpdateRunResponse{}
	err = s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute),
		request.UpdateRunRequest{
			RunID:   run.ID,
			Name:    "UpdatedName",
			Status:  string(models.StatusScheduled),
			EndTime: 1111111111,
		},
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.NotEmpty(s.T(), resp.RunInfo.ID)
	assert.NotEmpty(s.T(), resp.RunInfo.UUID)
	assert.Equal(s.T(), "UpdatedName", resp.RunInfo.Name)
	assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.RunInfo.ExperimentID)
	assert.Equal(s.T(), int64(1234567890), resp.RunInfo.StartTime)
	assert.Equal(s.T(), int64(1111111111), resp.RunInfo.EndTime)
	assert.Equal(s.T(), string(models.StatusScheduled), resp.RunInfo.Status)
	assert.NotEmpty(s.T(), resp.RunInfo.ArtifactURI)
	assert.Equal(s.T(), string(models.LifecycleStageActive), resp.RunInfo.LifecycleStage)

	// check that run has been updated in database.
	run, err = s.runFixtures.GetRun(context.Background(), run.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "UpdatedName", run.Name)
	assert.Equal(s.T(), models.StatusScheduled, run.Status)
	assert.Equal(s.T(), int64(1111111111), run.EndTime.Int64)
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
			error: api.NewResourceDoesNotExistError(
				"unable to find run '1': error getting 'run' entity by id: 1: record not found",
			),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsUpdateRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
