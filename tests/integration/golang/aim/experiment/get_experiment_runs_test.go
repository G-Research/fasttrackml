//go:build integration

package experiment

import (
	"context"
	"fmt"
	"testing"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentRunsTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetExperimentRunsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentRunsTestSuite))
}

func (s *GetExperimentRunsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetExperimentRunsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err := s.RunFixtures.CreateExampleRuns(context.Background(), experiment, 10)
	assert.Nil(s.T(), err)

	var resp response.GetExperimentRuns
	err = s.AIMClient.DoGetRequest(
		fmt.Sprintf("/experiments/%d/runs?limit=4&offset=%s", *experiment.ID, runs[8].ID),
		&resp,
	)
	assert.Nil(s.T(), err)

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
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	tests := []struct {
		name  string
		error string
		ID    string
	}{
		{
			name:  "IncorrectExperimentID",
			error: `: unable to parse experiment id "incorrect_experiment_id": strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`,
			ID:    "incorrect_experiment_id",
		},
		{
			name:  "NotFoundExperiment",
			error: `: Not Found`,
			ID:    "123",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			var resp api.ErrorResponse
			err := s.AIMClient.DoGetRequest(
				fmt.Sprintf(
					"/experiments/%s/runs", tt.ID,
				),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error, resp.Error())
		})
	}
}
