//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteBatchTestSuite struct {
	runs []*models.Run
	helpers.BaseTestSuite
}

func TestDeleteBatchTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteBatchTestSuite))
}

func (s *DeleteBatchTestSuite) SetupTest() {
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

	s.runs, err = s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	s.Require().Nil(err)
}

func (s *DeleteBatchTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		runIDs           []string
		expectedRunCount int
	}{
		{
			name:             "DeleteBatchOfOne",
			runIDs:           []string{s.runs[4].ID},
			expectedRunCount: 9,
		},
		{
			name:             "DeleteBatchOfTwo",
			runIDs:           []string{s.runs[3].ID, s.runs[5].ID},
			expectedRunCount: 7,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			resp := fiber.Map{}
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodPost).WithRequest(
					tt.runIDs,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/delete-batch",
				),
			)
			s.Equal(fiber.Map{"status": "OK"}, resp)

			runs, err := s.RunFixtures.GetRuns(context.Background(), s.runs[0].ExperimentID)
			s.Require().Nil(err)
			s.Equal(tt.expectedRunCount, len(runs))

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)
			s.Equal(originalMinRowNum, newMinRowNum)
			s.Greater(originalMaxRowNum, newMaxRowNum)
		})
	}
}

func (s *DeleteBatchTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		request          []string
		expectedRunCount int
	}{
		{
			name:             "DeleteWithUnknownID",
			request:          []string{s.runs[1].ID, "some-other-id"},
			expectedRunCount: 10,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodPost).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/delete-batch",
				),
			)
			s.Contains(resp.Error(), "count of deleted runs does not match length of ids input")

			runs, err := s.RunFixtures.GetRuns(context.Background(), s.runs[0].ExperimentID)
			s.Require().Nil(err)
			s.Equal(tt.expectedRunCount, len(runs))

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)
			s.Equal(originalMinRowNum, newMinRowNum)
			s.Equal(originalMaxRowNum, newMaxRowNum)
		})
	}
}
