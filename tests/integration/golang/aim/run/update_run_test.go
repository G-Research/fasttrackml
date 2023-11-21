//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), experiment)
	s.Require().Nil(err)
}

func (s *UpdateRunTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name    string
		request request.UpdateRunRequest
	}{
		{
			name: "UpdateOneRun",
			request: request.UpdateRunRequest{
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
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		ID          string
		requestBody any
		error       string
	}{
		{
			name: "UpdateRunWithIncorrectArchived",
			ID:   s.run.ID,
			requestBody: map[string]any{
				"Archived": "this-cannot-unmarshal",
			},
			error: "cannot unmarshal",
		},
		{
			name:        "UpdateRunWithUnknownID",
			ID:          "incorrect-ID",
			requestBody: map[string]any{},
			error:       "unable to find run 'incorrect-ID'",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
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
			s.Contains(resp.Message, tt.error)
		})
	}
}
