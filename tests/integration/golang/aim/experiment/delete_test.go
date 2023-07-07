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

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteExperimentTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestDeleteExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteExperimentTestSuite))
}

func (s *DeleteExperimentTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	fixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
	experiment, err := s.fixtures.CreateExperiment(context.Background(), &models.Experiment{
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
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	assert.Nil(s.T(), err)
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()

	experiments, err := s.fixtures.GetTestExperiments(context.Background())
	assert.Nil(s.T(), err)
	length := len(experiments)

	var resp any
	err = s.client.DoDeleteRequest(
		fmt.Sprintf("/experiments/%d", *experiment.ID),
		&resp,
	)
	assert.Nil(s.T(), err)

	experiments, err = s.fixtures.GetTestExperiments(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(experiments), length-1)
}

func (s *DeleteExperimentTestSuite) Test_Error() {
	tests := []struct {
		name string
		ID   string
	}{
		{
			name: "DeleteWithUnknownIDFails",
			ID:   "123",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp api.ErrorResponse
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/experiments/%s", tt.ID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Error(), "count of deleted experiments does not match length of ids input")

			assert.NoError(s.T(), err)
		})
	}
}
