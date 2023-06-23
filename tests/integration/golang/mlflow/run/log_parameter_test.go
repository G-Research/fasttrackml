//go:build integration

package run

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogParamTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	run                *models.Run
}

func TestLogParamTestSuite(t *testing.T) {
	suite.Run(t, new(LogParamTestSuite))
}

func (s *LogParamTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *exp.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.runFixtures.CreateRun(context.Background(), run)
	assert.Nil(s.T(), err)
	s.run = run
}

func (s *LogParamTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	req := request.LogParamRequest{
		RunID: s.run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := map[string]any{}
	err := s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), resp)
}

func (s *LogParamTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	// missing run_id
	req := request.LogParamRequest{
		Key:   "key1",
		Value: "value1",
	}
	resp := api.ErrorResponse{}
	err := s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'").Error(), resp.Error())
}
