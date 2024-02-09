package experiment

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RestoreExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestRestoreExperimentTestSuite(t *testing.T) {
	suite.Run(t, &RestoreExperimentTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *RestoreExperimentTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageDeleted,
	})
	s.Require().Nil(err)
	s.Equal(models.LifecycleStageDeleted, experiment.LifecycleStage)

	// 2. make actual API call.
	req := request.RestoreExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
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
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsRestoreRoute,
		),
	)

	// 3. check actual API response.
	exp, err := s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment.ID,
	)
	s.Require().Nil(err)
	s.Equal(models.LifecycleStageActive, exp.LifecycleStage)
}

func (s *RestoreExperimentTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreExperimentRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.RestoreExperimentRequest{
				ID: "",
			},
		},
		{
			name: "InvalidIDFormat",
			error: api.NewBadRequestError(
				"Unable to parse experiment id 'invalid_id': strconv.ParseInt: parsing \"invalid_id\": invalid syntax",
			),
			request: &request.RestoreExperimentRequest{
				ID: "invalid_id",
			},
		},
		{
			name: "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(
				"unable to find experiment '123': error getting experiment by id: 123: record not found",
			),
			request: &request.RestoreExperimentRequest{
				ID: "123",
			},
		},
	}

	for _, tt := range testData {
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
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsRestoreRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
