//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	runs []*models.Run
}

func TestDeleteRunTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTestSuite))
}

func (s *DeleteRunTestSuite) SetupTest() {
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

	s.runs, err = s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	assert.Nil(s.T(), err)
}

func (s *DeleteRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		request          request.DeleteRunRequest
		expectedRunCount int
	}{
		{
			name:             "DeleteOneRun",
			request:          request.DeleteRunRequest{RunID: s.runs[4].ID},
			expectedRunCount: 9,
		},
		{
			name:             "RowNumbersAreRecalculated",
			request:          request.DeleteRunRequest{RunID: s.runs[1].ID},
			expectedRunCount: 8,
		},
		{
			name:             "DeleteFirstRun",
			request:          request.DeleteRunRequest{RunID: s.runs[0].ID},
			expectedRunCount: 7,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(),
				s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			var resp fiber.Map
			assert.Nil(
				s.T(),
				s.AIMClient.WithMethod(http.MethodDelete).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", tt.request.RunID,
				),
			)

			runs, err := s.RunFixtures.GetRuns(context.Background(), s.runs[0].ExperimentID)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedRunCount, len(runs))

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), originalMinRowNum, newMinRowNum)
			assert.Greater(s.T(), originalMaxRowNum, newMaxRowNum)
		})
	}
}

func (s *DeleteRunTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name    string
		request request.DeleteRunRequest
	}{
		{
			name:    "DeleteWithUnknownID",
			request: request.DeleteRunRequest{RunID: "some-other-id"},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			var resp api.ErrorResponse
			assert.Nil(
				s.T(),
				s.AIMClient.WithMethod(http.MethodDelete).WithRequest(
					tt.request.RunID,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", tt.request.RunID,
				),
			)
			assert.Contains(s.T(), resp.Error(), "unable to find run 'some-other-id'")

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), originalMinRowNum, newMinRowNum)
			assert.Equal(s.T(), originalMaxRowNum, newMaxRowNum)
		})
	}
}
