//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ArchiveBatchTestSuite struct {
	suite.Suite
	runs               []*models.Run
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestArchiveBatchTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveBatchTestSuite))
}

func (s *ArchiveBatchTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	s.runs, err = s.runFixtures.CreateExampleRuns(context.Background(), exp, 10)
	assert.Nil(s.T(), err)
}

func (s *ArchiveBatchTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name                 string
		runIDs               []string
		expectedArchiveCount int
		archiveParam         string
	}{
		{
			name:                 "ArchiveBatchOfOne",
			runIDs:               []string{s.runs[4].ID},
			expectedArchiveCount: 1,
			archiveParam:         "true",
		},
		{
			name:                 "ArchiveBatchOfTwo",
			runIDs:               []string{s.runs[3].ID, s.runs[5].ID},
			expectedArchiveCount: 3,
			archiveParam:         "true",
		},
		{
			name:                 "RestoreBatchOfTwo",
			runIDs:               []string{s.runs[3].ID, s.runs[5].ID},
			expectedArchiveCount: 1,
			archiveParam:         "false",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			resp := map[string]any{}
			err = s.client.DoPostRequest(
				fmt.Sprintf("/runs/archive-batch?archive=%s", tt.archiveParam),
				tt.runIDs,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), map[string]interface{}{"status": "OK"}, resp)

			runs, err := s.runFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), 10, len(runs))
			archiveCount := 0
			for _, run := range runs {
				if run.LifecycleStage == models.LifecycleStageDeleted {
					archiveCount++
				}
			}
			assert.Equal(s.T(), tt.expectedArchiveCount, archiveCount)

			newMinRowNum, newMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), originalMinRowNum, newMinRowNum)
			assert.Equal(s.T(), originalMaxRowNum, newMaxRowNum)
		})
	}
}

func (s *ArchiveBatchTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		request          []string
		expectedRunCount int
	}{
		{
			name:             "ArchiveWithUnknownID",
			request:          []string{"some-other-id"},
			expectedRunCount: 10,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			originalMinRowNum, originalMaxRowNum, err := s.runFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			assert.Nil(s.T(), err)

			var resp fiber.Map
			err = s.client.DoPostRequest(
				"/runs/archive-batch?archive=true",
				tt.request,
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
			assert.Equal(s.T(), originalMaxRowNum, newMaxRowNum)
		})
	}
}
