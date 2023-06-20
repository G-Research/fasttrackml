//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
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
	s.client = helpers.NewMlflowApiClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
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

	//check that the experiment lifecycle is active
	assert.Equal(s.T(), models.LifecycleStageActive, experiment.LifecycleStage)

	// 2. make actual API call.
	req := request.DeleteExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	}
	resp := fiber.Map{}
	err = s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsDeleteRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)

	// 3. check actual API response.
	exp, err := s.fixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), models.LifecycleStageDeleted, exp.LifecycleStage)
}

func (s *DeleteExperimentTestSuite) Test_Error() {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteExperimentRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.DeleteExperimentRequest{
				ID: "",
			},
		},
		{
			name:  "InvalidIDFormat",
			error: api.NewBadRequestError(`unable to parse experiment id 'invalid_id': strconv.ParseInt: parsing "invalid_id": invalid syntax`),
			request: &request.DeleteExperimentRequest{
				ID: "invalid_id",
			},
		},
		{
			name:  "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError("unable to find experiment '123': error getting experiment by id: 123: record not found"),
			request: &request.DeleteExperimentRequest{
				ID: "123",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsDeleteRoute),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
