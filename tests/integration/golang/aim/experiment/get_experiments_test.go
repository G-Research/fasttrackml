//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentsTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetExperimentsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentsTestSuite))
}

func (s *GetExperimentsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetExperimentsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiments := map[int32]*models.Experiment{}
	for i := 0; i < 5; i++ {
		experiment := &models.Experiment{
			Name: fmt.Sprintf("Test Experiment %d", i),
			Tags: []models.ExperimentTag{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
			NamespaceID: namespace.ID,
			CreationTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LastUpdateTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LifecycleStage:   models.LifecycleStageActive,
			ArtifactLocation: "/artifact/location",
		}
		experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
		assert.Nil(s.T(), err)
		experiments[*experiment.ID] = experiment
	}

	var resp response.Experiments
	assert.Nil(s.T(), s.AIMClient.WithResponse(&resp).DoRequest("/experiments/"))
	assert.Equal(s.T(), len(experiments), len(resp))
	for _, actualExperiment := range resp {
		expectedExperiment := experiments[actualExperiment.ID]
		assert.Equal(s.T(), fmt.Sprintf("%d", *expectedExperiment.ID), actualExperiment.ID)
		assert.Equal(s.T(), expectedExperiment.Name, actualExperiment.Name)
		assert.Equal(s.T(), float64(expectedExperiment.CreationTime.Int64)/1000, actualExperiment.CreationTime)
		assert.Equal(s.T(), expectedExperiment.LifecycleStage == models.LifecycleStageDeleted, actualExperiment.Archived)
		assert.Equal(s.T(), len(expectedExperiment.Runs), actualExperiment.RunCount)
	}
}
