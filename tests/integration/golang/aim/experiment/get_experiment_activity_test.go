package experiment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentActivityTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentActivityTestSuite(t *testing.T) {
	suite.Run(t, &GetExperimentActivityTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *GetExperimentActivityTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	s.Require().Nil(err)

	archivedRunsIds := []string{runs[0].ID, runs[1].ID}
	err = s.RunFixtures.ArchiveRuns(context.Background(), s.DefaultNamespace.ID, archivedRunsIds)
	s.Require().Nil(err)

	var resp response.GetExperimentActivity
	s.Require().Nil(
		s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%d/activity", *experiment.ID),
	)
	s.Equal(resp.NumRuns, len(runs))
	s.Equal(resp.NumArchivedRuns, len(archivedRunsIds))
	s.Equal(resp.NumActiveRuns, len(runs)-len(archivedRunsIds))
	s.Equal(resp.ActivityMap, helpers.TransformRunsToActivityMap(runs))
}

func (s *GetExperimentActivityTestSuite) Test_Error() {
	tests := []struct {
		name  string
		ID    string
		error string
	}{
		{
			name:  "GetInvalidExperimentID",
			ID:    "123",
			error: "(Not Found|not found)",
		},
		{
			name:  "DeleteIncorrectExperimentID",
			error: `(unable to parse|failed to decode)`,
			ID:    "incorrect_experiment_id",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithQuery(map[any]any{
				"limit": 4,
			}).WithResponse(&resp).DoRequest(
				"/experiments/%s/activity", tt.ID,
			))
			s.Regexp(tt.error, resp.Error())
		})
	}
}
