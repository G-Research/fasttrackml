//go:build integration

package run

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogParamTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestLogParamTestSuite(t *testing.T) {
	suite.Run(t, new(LogParamTestSuite))
}

func (s *LogParamTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *LogParamTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	experiment := &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
	assert.Nil(s.T(), err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.RunFixtures.CreateRun(context.Background(), run)
	assert.Nil(s.T(), err)

	req := request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := map[string]any{}
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), resp)
}

func (s *LogParamTestSuite) Test_Error() {
	_, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	// missing run_id
	req := request.LogParamRequest{
		Key:   "key1",
		Value: "value1",
	}
	resp := api.ErrorResponse{}
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(
		s.T(),
		api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'").Error(),
		resp.Error(),
	)
}
