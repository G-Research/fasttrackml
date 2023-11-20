//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *GetExperimentTestSuite) Test_Ok() {
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
		Name: "Test Experiment",
		Tags: []models.ExperimentTag{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		CreationTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		NamespaceID: namespace.ID,
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	require.Nil(s.T(), err)

	var resp response.GetExperiment
	require.Nil(s.T(), s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%d", *experiment.ID))

	assert.Equal(s.T(), *experiment.ID, resp.ID)
	assert.Equal(s.T(), experiment.Name, resp.Name)
	assert.Equal(s.T(), "", resp.Description)
	assert.Equal(s.T(), float64(experiment.CreationTime.Int64)/1000, resp.CreationTime)
	assert.Equal(s.T(), false, resp.Archived)
	assert.Equal(s.T(), len(experiment.Runs), resp.RunCount)
}

func (s *GetExperimentTestSuite) Test_Error() {
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
			ID:    "1",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			var resp api.ErrorResponse
			require.Nil(t, s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%s", tt.ID))
			assert.Equal(s.T(), tt.error, resp.Error())
		})
	}
}
