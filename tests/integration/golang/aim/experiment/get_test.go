//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestGetExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *GetExperimentTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	fixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *GetExperimentTestSuite) Test_Ok() {
	// 1. prepare database with test data.
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
	var resp fiber.Map
	// 2. make actual API call.
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"/experiments/%d", *experiment.ID,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	// 3. check actual API response.
	assert.Equal(s.T(), *experiment.ID, int32(resp["id"].(float64)))
	assert.Equal(s.T(), experiment.Name, resp["name"])
	assert.Equal(s.T(), experiment.LifecycleStage == models.LifecycleStageDeleted, resp["archived"])
	assert.Equal(s.T(), len(experiment.Runs), int(resp["run_count"].(float64)))
}

func (s *GetExperimentTestSuite) Test_Error() {
	testData := []struct {
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
			ID:    "1",
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			var resp api.ErrorResponse
			err := s.client.DoGetRequest(
				fmt.Sprintf(
					"/experiments/%s", tt.ID,
				),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error, resp.Error())
		})
	}
}
