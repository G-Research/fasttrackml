//go:build integration

package experiment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentActivityTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentActivityTestSuite))
}

func (s *GetExperimentActivityTestSuite) Test_Ok() {
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

	var resp response.GetExperimentActivity
	require.Nil(
		s.T(),
		s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%d/activity", *experiment.ID),
	)
	assert.Equal(s.T(), resp.NumRuns, len(runs))
	assert.Equal(s.T(), resp.NumArchivedRuns, len(archivedRunsIds))
	assert.Equal(s.T(), resp.NumActiveRuns, len(runs)-len(archivedRunsIds))
	assert.Equal(s.T(), resp.ActivityMap, helpers.TransformRunsToActivityMap(runs))
}

func (s *GetExperimentActivityTestSuite) Test_Error() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	tests := []struct {
		name  string
		ID    string
		error string
	}{
		{
			name:  "GetInvalidExperimentID",
			ID:    "123",
			error: "Not Found",
		},
		{
			name: "DeleteIncorrectExperimentID",
			error: `: unable to parse experiment id "incorrect_experiment_id": strconv.ParseInt:` +
				` parsing "incorrect_experiment_id": invalid syntax`,
			ID: "incorrect_experiment_id",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp api.ErrorResponse
			require.Nil(s.T(), s.AIMClient().WithQuery(map[any]any{
				"limit": 4,
			}).WithResponse(&resp).DoRequest(
				"/experiments/%s/activity", tt.ID,
			))
			assert.Contains(s.T(), resp.Error(), tt.error)
		})
	}
}
