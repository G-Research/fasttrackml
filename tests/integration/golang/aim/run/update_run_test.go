package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
}

func (s *UpdateRunTestSuite) Test_Ok() {
	tests := []struct {
		name    string
		request request.UpdateRunRequest
	}{
		{
			name: "UpdateOneRun",
			request: request.UpdateRunRequest{
				ID:       s.run.ID,
				RunID:    &(s.run.ID),
				Name:     common.GetPointer(fmt.Sprintf("%v%v", s.run.Name, "-new")),
				Status:   common.GetPointer(string(models.StatusFinished)),
				Archived: common.GetPointer(true),
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Success
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", *tt.request.RunID,
				),
			)
			run, err := s.RunFixtures.GetRun(context.Background(), s.run.ID)
			s.Require().Nil(err)
			// TODO the PUT endpoint only updates LifecycleStage
			// s.Equal(newName, run.Name)
			// s.Equal(models.Status(newStatus), run.Status)
			s.Equal(models.LifecycleStageDeleted, run.LifecycleStage)
		})
	}
}

func (s *UpdateRunTestSuite) Test_Error() {
	tests := []struct {
		ID          string
		name        string
		requestBody any
		error       *api.ErrorResponse
	}{
		{
			name: "UpdateRunWithIncorrectArchived",
			ID:   s.run.ID,
			requestBody: map[string]any{
				"Archived": "this-cannot-unmarshal",
			},
			error: &api.ErrorResponse{
				Message:    "json: cannot unmarshal string into Go struct field UpdateRunRequest.archived of type bool",
				StatusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name:        "UpdateRunWithUnknownID",
			ID:          "incorrect-ID",
			requestBody: map[string]any{},
			error: &api.ErrorResponse{
				Message:    "run 'incorrect-ID' not found",
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", tt.ID,
				),
			)
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
