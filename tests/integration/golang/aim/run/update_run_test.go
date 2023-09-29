//go:build integration

package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateRunTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	run *models.Run
}

func TestUpdateRunTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateRunTestSuite))
}

func (s *UpdateRunTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), experiment)
	assert.Nil(s.T(), err)
}

func (s *UpdateRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
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
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Success
			assert.Nil(
				s.T(),
				s.AIMClient.WithMethod(
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
			assert.Nil(s.T(), err)
			// TODO the PUT endpoint only updates LifecycleStage
			// assert.Equal(t, newName, run.Name)
			// assert.Equal(t, models.Status(newStatus), run.Status)
			assert.Equal(s.T(), models.LifecycleStageDeleted, run.LifecycleStage)
		})
	}
}

func (s *UpdateRunTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody any
	}{
		{
			name: "UpdateRunWithIncorrectArchived",
			requestBody: map[string]any{
				"Archived": "this-cannot-unmarshal",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			assert.Nil(
				s.T(),
				s.AIMClient.WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", s.run.ID,
				),
			)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
