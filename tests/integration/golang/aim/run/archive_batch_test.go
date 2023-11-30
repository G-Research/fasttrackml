//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ArchiveBatchTestSuite struct {
	helpers.BaseTestSuite
	runs []*models.Run
}

func TestArchiveBatchTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveBatchTestSuite))
}

func (s *ArchiveBatchTestSuite) SetupTest() {
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

func (s *ArchiveBatchTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
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
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			resp := map[string]any{}
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodPost).WithQuery(map[any]any{
					"archive": tt.archiveParam,
				}).WithRequest(
					tt.runIDs,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/archive-batch",
				),
			)
			s.Equal(map[string]interface{}{"status": "OK"}, resp)

			runs, err := s.RunFixtures.GetRuns(context.Background(), s.runs[0].ExperimentID)
			s.Require().Nil(err)
			s.Equal(10, len(runs))
			archiveCount := 0
			for _, run := range runs {
				if run.LifecycleStage == models.LifecycleStageDeleted {
					archiveCount++
				}
			}
			s.Equal(tt.expectedArchiveCount, archiveCount)

			newMinRowNum, newMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)
			s.Equal(originalMinRowNum, newMinRowNum)
			s.Equal(originalMaxRowNum, newMaxRowNum)
		})
	}
}

func (s *ArchiveBatchTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
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
		s.Run(tt.name, func() {
			originalMinRowNum, originalMaxRowNum, err := s.RunFixtures.FindMinMaxRowNums(
				context.Background(), s.runs[0].ExperimentID,
			)
			s.Require().Nil(err)

			var resp fiber.Map
			s.Require().Nil(
				s.AIMClient().WithMethod(http.MethodPost).WithQuery(map[any]any{
					"archive": true,
				}).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/archive-batch",
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
			s.Equal(originalMaxRowNum, newMaxRowNum)
		})
	}
}
