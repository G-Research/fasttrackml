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

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RestoreExperimentTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestRestoreExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(RestoreExperimentTestSuite))
}

func (s *RestoreExperimentTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *RestoreExperimentTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	// 1. prepare database with test data.
	namespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
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
		LifecycleStage:   models.LifecycleStageDeleted,
		ArtifactLocation: "/artifact/location",
	})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), models.LifecycleStageDeleted, experiment.LifecycleStage)

	// 2. make actual API call.
	req := request.RestoreExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	}
	resp := fiber.Map{}
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsRestoreRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)

	// 3. check actual API response.
	exp, err := s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment.ID,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), models.LifecycleStageActive, exp.LifecycleStage)
}

func (s *RestoreExperimentTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.RestoreExperimentRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.RestoreExperimentRequest{
				ID: "",
			},
		},
		{
			name: "InvalidIDFormat",
			error: api.NewBadRequestError(
				"Unable to parse experiment id 'invalid_id': strconv.ParseInt: parsing \"invalid_id\": invalid syntax",
			),
			request: &request.RestoreExperimentRequest{
				ID: "invalid_id",
			},
		},
		{
			name: "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(
				"unable to find experiment '123': error getting experiment by id: 123: record not found",
			),
			request: &request.RestoreExperimentRequest{
				ID: "123",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.MlflowClient.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsRestoreRoute),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
