package experiment

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentRunsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentRunsTestSuite(t *testing.T) {
	suite.Run(t, &GetExperimentRunsTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *GetExperimentRunsTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	s.Require().Nil(err)

	var resp response.ExperimentRuns
	s.Require().Nil(
		s.AIMClient().WithQuery(map[any]any{
			"limit":  4,
			"offset": runs[8].ID,
		}).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d/runs", *experiment.ID,
		),
	)

	s.Equal(4, len(resp.Runs))
	for index := 0; index < len(resp.Runs); index++ {
		r := runs[8-(index+1)]
		s.Equal(r.ID, resp.Runs[index].ID)
		s.Equal(r.Name, resp.Runs[index].Name)
		s.Equal(r.StartTime.Int64/1000, resp.Runs[index].CreationTime)
		s.Equal(r.EndTime.Int64/1000, resp.Runs[index].EndTime)
		s.Equal(r.LifecycleStage == models.LifecycleStageDeleted, resp.Runs[index].Archived)
	}
}

func (s *GetExperimentRunsTestSuite) Test_Error() {
	tests := []struct {
		ID    string
		name  string
		error *api.ErrorResponse
	}{
		{
			ID:   "incorrect_experiment_id",
			name: "IncorrectExperimentID",
			error: &api.ErrorResponse{
				Message:    `failed to decode: schema: error converting value for "id"`,
				StatusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			ID:    "123",
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError("experiment '123' not found"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%s/runs", tt.ID))
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
