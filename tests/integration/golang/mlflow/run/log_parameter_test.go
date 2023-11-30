//go:build integration

package run

import (
	"context"
	"fmt"
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

type LogParamTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogParamTestSuite(t *testing.T) {
	suite.Run(t, new(LogParamTestSuite))
}

func (s *LogParamTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment := &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
	s.Require().Nil(err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.RunFixtures.CreateRun(context.Background(), run)
	s.Require().Nil(err)

	req := request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := map[string]any{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
		),
	)
	s.Empty(resp)

	// log duplicate, which is OK
	req = request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
		),
	)
	s.Empty(resp)
}

func (s *LogParamTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment := &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
	s.Require().Nil(err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.RunFixtures.CreateRun(context.Background(), run)
	s.Require().Nil(err)

	// setup param OK
	req := request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := api.ErrorResponse{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
		),
	)
	s.Empty(resp)

	// error conditions

	// missing run_id
	req = request.LogParamRequest{
		Key:   "key1",
		Value: "value1",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
		),
	)
	s.Equal(
		api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'").Error(),
		resp.Error(),
	)

	// conflicting param
	req = request.LogParamRequest{
		RunID: run.ID,
		Key:   "key1",
		Value: "value2",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute,
		),
	)
	s.Equal(
		api.NewInvalidParameterValueError(
			fmt.Sprintf(`unable to insert params for run '%s': conflicting params found: `+
				`[{run_id: %s, key: %s, old_value: %s, new_value: %s}]`,
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
