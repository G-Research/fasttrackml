//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentsTestSuite))
}

func (s *GetExperimentsTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	experiments := map[string]*models.Experiment{}
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
		require.Nil(s.T(), err)
		experiments[fmt.Sprintf("%d", *experiment.ID)] = experiment
	}

	var resp response.Experiments
	require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/experiments/"))
	assert.Equal(s.T(), len(experiments), len(resp))
	for _, actualExperiment := range resp {
		id, err := strconv.ParseInt(actualExperiment.ID, 10, 32)
		require.Nil(s.T(), err)
		expectedExperiment := experiments[int32(id)]
		assert.Equal(s.T(), fmt.Sprintf("%d", *expectedExperiment.ID), actualExperiment.ID)
		assert.Equal(s.T(), expectedExperiment.Name, actualExperiment.Name)
		assert.Equal(s.T(), helpers.GetDescriptionFromExperiment(*expectedExperiment), actualExperiment.Description)
		assert.Equal(s.T(), float64(expectedExperiment.CreationTime.Int64)/1000, actualExperiment.CreationTime)
		assert.Equal(s.T(), expectedExperiment.LifecycleStage == models.LifecycleStageDeleted, actualExperiment.Archived)
		assert.Equal(s.T(), len(expectedExperiment.Runs), actualExperiment.RunCount)
	}
}
