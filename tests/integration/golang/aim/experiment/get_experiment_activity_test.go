//go:build integration

package experiment

/*
import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentActivityTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetExperimentActivityTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentActivityTestSuite))
}

func (s *GetExperimentActivityTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetExperimentActivityTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.RunFixtures.UnloadFixtures())
	}()
	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), exp, 10)
	assert.Nil(s.T(), err)

	archivedRunsIds := []string{runs[0].ID, runs[1].ID}
	err = s.RunFixtures.ArchiveRuns(context.Background(), archivedRunsIds)
	assert.Nil(s.T(), err)

	var resp response.GetExperimentActivity
	err = s.AIMClient.DoGetRequest(
		fmt.Sprintf(
			"/experiments/%d/activity", *experiment.ID,
		),
		&resp,
	)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), resp.NumRuns, len(runs))
	assert.Equal(s.T(), resp.NumArchivedRuns, len(archivedRunsIds))
	assert.Equal(s.T(), resp.NumActiveRuns, len(runs)-len(archivedRunsIds))
	assert.Equal(s.T(), resp.ActivityMap, helpers.TransformRunsToActivityMap(runs))
}

func (s *GetExperimentActivityTestSuite) Test_Error() {
	tests := []struct {
		name string
		ID   string
	}{
		{
			name: "GetInvalidExperimentID",
			ID:   "123",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp api.ErrorResponse
			err := s.AIMClient.DoGetRequest(
				fmt.Sprintf(
					"/experiments/%s/runs?limit=4", tt.ID,
				),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Error(), "Not Found")
		})
	}
}
*/
