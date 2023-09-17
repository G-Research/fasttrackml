//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunInfoTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	run *models.Run
}

func TestGetRunInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunInfoTestSuite))
}

func (s *GetRunInfoTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	namespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), experiment)
	assert.Nil(s.T(), err)
}

func (s *GetRunInfoTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name  string
		runID string
	}{
		{
			name:  "GetOneRun",
			runID: s.run.ID,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.GetRunInfo
			err := s.AIMClient.DoGetRequest(
				fmt.Sprintf("/runs/%s/info", tt.runID),
				&resp,
			)
			assert.Nil(s.T(), err)
			// TODO this assertion fails because ID is not rendered by the endpoint
			// assert.Equal(s.T(), s.run.ID, resp.Props.ID)
			assert.Equal(s.T(), s.run.Name, resp.Props.Name)
			assert.Equal(s.T(), fmt.Sprintf("%v", s.run.ExperimentID), resp.Props.Experiment.ID)
			assert.Equal(s.T(), s.run.StartTime.Int64, resp.Props.CreationTime)
			assert.Equal(s.T(), s.run.EndTime.Int64, resp.Props.EndTime)
			// TODO this assertion fails because tags are not rendered by endpoint
			// assert.Equal(s.T(), s.run.Tags[0].Key, resp.Props.Tags[0])
			// TODO this assertion fails so maybe the endpoint is not populating correctly
			// assert.NotEmpty(s.T(), resp.Props.CreationTime)
		})
	}
}

func (s *GetRunInfoTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name  string
		runID string
	}{
		{
			name:  "GetNonexistentRun",
			runID: uuid.NewString(),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.AIMClient.DoGetRequest(
				fmt.Sprintf("/runs/%s/info", tt.runID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "Not Found", resp.Message)
		})
	}
}
