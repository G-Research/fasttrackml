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

type GetExperimentRunsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentRunsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentRunsTestSuite))
}

func (s *GetExperimentRunsTestSuite) Test_Ok() {
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

	var resp response.GetExperimentRuns
	require.Nil(
		s.T(),
		s.AIMClient().WithQuery(map[any]any{
			"limit":  4,
			"offset": runs[8].ID,
		}).WithResponse(&resp).DoRequest(
			"/experiments/%d/runs", *experiment.ID,
		),
	)

	assert.Equal(s.T(), 4, len(resp.Runs))
	for index := 0; index < len(resp.Runs); index++ {
		r := runs[8-(index+1)]
		assert.Equal(s.T(), r.ID, resp.Runs[index].ID)
		assert.Equal(s.T(), r.Name, resp.Runs[index].Name)
		assert.Equal(s.T(), float64(r.StartTime.Int64)/1000, resp.Runs[index].CreationTime)
		assert.Equal(s.T(), float64(r.EndTime.Int64)/1000, resp.Runs[index].EndTime)
		assert.Equal(s.T(), r.LifecycleStage == models.LifecycleStageDeleted, resp.Runs[index].Archived)
	}
}

func (s *GetExperimentRunsTestSuite) Test_Error() {
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
		error string
		ID    string
	}{
		{
			name: "IncorrectExperimentID",
			error: `: unable to parse experiment id "incorrect_experiment_id": strconv.ParseInt: ` +
				`parsing "incorrect_experiment_id": invalid syntax`,
			ID: "incorrect_experiment_id",
		},
		{
			name:  "NotFoundExperiment",
			error: `: Not Found`,
			ID:    "123",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%s/runs", tt.ID))
			assert.Equal(s.T(), tt.error, resp.Error())
		})
	}
}
