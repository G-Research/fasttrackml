package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTestSuite struct {
	helpers.BaseTestSuite
	runs []*models.Run
}

func TestDeleteRunTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTestSuite))
}

func (s *DeleteRunTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	var err error
	s.runs, err = s.RunFixtures.CreateExampleRuns(context.Background(), s.DefaultExperiment, 10)
	s.Require().Nil(err)
}

func (s *DeleteRunTestSuite) Test_Ok() {
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
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(),
				s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			var resp fiber.Map
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodDelete).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", tt.request.RunID,
				),
			)

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

func (s *DeleteRunTestSuite) Test_Error() {
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
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodDelete).WithRequest(
					tt.request.RunID,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s", tt.request.RunID,
				),
			)
			s.Regexp("unable to find|not found", resp.Error())

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)
			s.Equal(originalMinRowNum, newMinRowNum)
			s.Equal(originalMaxRowNum, newMaxRowNum)
		})
	}
}
