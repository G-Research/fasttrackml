//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetProjectActivityTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetProjectActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetProjectActivityTestSuite))
}

func (s *GetProjectActivityTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

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

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	s.Require().Nil(err)

	archivedRunsIds := []string{runs[0].ID, runs[1].ID}
	err = s.RunFixtures.ArchiveRuns(context.Background(), namespace.ID, archivedRunsIds)
	s.Require().Nil(err)

	var resp response.ProjectActivityResponse
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/projects/activity"))

	s.Equal(8, resp.NumActiveRuns)
	s.Equal(2, resp.NumArchivedRuns)
	s.Equal(1, resp.NumExperiments)
	s.Equal(10, resp.NumRuns)
	s.Equal(1, len(resp.ActivityMap))
	for _, v := range resp.ActivityMap {
		s.Equal(10, v)
	}
}
