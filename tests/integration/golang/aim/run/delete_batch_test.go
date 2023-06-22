//go:build integration

package run

import (
	"context"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteBatchTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	runs               []*models.Run
}

func TestDeleteBatchTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteBatchTestSuite))
}

func (s *DeleteBatchTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	s.runs, err = s.runFixtures.CreateTestRuns(context.Background(), exp, 10)
	assert.Nil(s.T(), err)
}

func (s *DeleteBatchTestSuite) Test_Ok() {
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
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			resp := fiber.Map{}
			err = s.client.DoPostRequest(
				"/runs/delete-batch",
				tt.runIDs,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), fiber.Map{"status": "OK"}, resp)

			runs, err := s.runFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedRunCount, len(runs))

			newMinRowNum, newMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), originalMinRowNum, newMinRowNum)
			assert.Greater(s.T(), originalMaxRowNum, newMaxRowNum)

		})
	}
}

func (s *DeleteBatchTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
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
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			var resp api.ErrorResponse
			err = s.client.DoPostRequest(
				"/runs/delete-batch",
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Error(), "count of deleted runs does not match length of ids input")

			runs, err := s.runFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedRunCount, len(runs))

			newMinRowNum, newMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), originalMinRowNum, newMinRowNum)
			assert.Equal(s.T(), originalMaxRowNum, newMaxRowNum)
		})
	}
}
