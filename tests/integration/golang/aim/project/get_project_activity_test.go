//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	require.Nil(s.T(), err)

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	require.Nil(s.T(), err)

	archivedRunsIds := []string{runs[0].ID, runs[1].ID}
	err = s.RunFixtures.ArchiveRuns(context.Background(), archivedRunsIds)
	require.Nil(s.T(), err)

	var resp response.ProjectActivityResponse
	require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/projects/activity"))

	assert.Equal(s.T(), 8, resp.NumActiveRuns)
	assert.Equal(s.T(), 2, resp.NumArchivedRuns)
	assert.Equal(s.T(), 1, resp.NumExperiments)
	assert.Equal(s.T(), 10, resp.NumRuns)
	assert.Equal(s.T(), 1, len(resp.ActivityMap))
	for _, v := range resp.ActivityMap {
		assert.Equal(s.T(), 10, v)
	}
}
