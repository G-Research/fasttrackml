package experiment

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteExperimentTestSuite(t *testing.T) {
	suite.Run(t, &DeleteExperimentTestSuite{})
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	experiments, err := s.ExperimentFixtures.GetExperiments(context.Background())
	s.Require().Nil(err)
	length := len(experiments)

	var resp response.DeleteExperimentResponse
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodDelete,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d", *experiment.ID,
		),
	)

	remainingExperiments, err := s.ExperimentFixtures.GetExperiments(context.Background())
	s.Require().Nil(err)
	s.Equal(length-1, len(remainingExperiments))
}

func (s *DeleteExperimentTestSuite) Test_Error() {
	tests := []struct {
		name  string
		ID    string
		error *api.ErrorResponse
	}{
		{
			ID:   "123",
			name: "DeleteWithUnknownIDFails",
			error: &api.ErrorResponse{
				Message:    "experiment '123' not found",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			ID:   "incorrect_experiment_id",
			name: "DeleteIncorrectExperimentID",
			error: &api.ErrorResponse{
				Message:    `failed to decode: schema: error converting value for "id"`,
				StatusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			ID:   fmt.Sprintf("%d", *s.DefaultExperiment.ID),
			name: "DeleteDefaultExperiment",
			error: &api.ErrorResponse{
				Message:    "unable to delete default experiment",
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/experiments/%s", tt.ID,
				),
			)
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
