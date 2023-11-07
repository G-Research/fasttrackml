//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
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

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
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
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
	assert.Empty(s.T(), resp)

	// log duplicate, which is OK
	req = request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
	assert.Empty(s.T(), resp)
}

func (s *LogParamTestSuite) Test_Error() {
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
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

	// setup param OK
	req := request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := api.ErrorResponse{}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
	assert.Empty(s.T(), resp)

	// error conditions

	// missing run_id
	req = request.LogParamRequest{
		Key:   "key1",
		Value: "value1",
	}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
	assert.Equal(
		s.T(),
		api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'").Error(),
		resp.Error(),
	)

	// conflicting param
	req = request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value2",
	}
	assert.Nil(
		s.T(),
		s.MlflowClient.WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		),
	)
	assert.Equal(
		s.T(),
		api.NewInvalidParameterValueError(
			fmt.Sprintf("unable to insert params for run '%s': conflicting params found: [{run_id: %s, key: %s, old_value: %s, new_value: %s}]",
				req.RunID,
				req.RunID,
				req.Key,
				"value1",
				req.Value,
			),
		).Error(),
		resp.Error(),
	)
}
