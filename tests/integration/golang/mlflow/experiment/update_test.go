package experiment

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateExperimentTestSuite(t *testing.T) {
	suite.Run(t, &UpdateExperimentTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *UpdateExperimentTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	req := request.UpdateExperimentRequest{
		ID:   fmt.Sprintf("%d", *experiment.ID),
		Name: "Test Updated Experiment",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&struct{}{},
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsUpdateRoute,
		),
	)

	exp, err := s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment.ID,
	)
	s.Require().Nil(err)
	s.Equal("Test Updated Experiment", exp.Name)
}

func (s *UpdateExperimentTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.UpdateExperimentRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.UpdateExperimentRequest{
				ID: "",
			},
		},
		{
			name:  "EmptyNameProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'new_name'"),
			request: &request.UpdateExperimentRequest{
				ID:   "1",
				Name: "",
			},
		},
		{
			name: "InvalidIDFormat",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'invalid_id': strconv.ParseInt: parsing "invalid_id": invalid syntax`,
			),
			request: &request.UpdateExperimentRequest{
				ID:   "invalid_id",
				Name: "New Name",
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
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsUpdateRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
