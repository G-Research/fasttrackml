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

type SetExperimentTagTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestSetExperimentTagTestSuite(t *testing.T) {
	suite.Run(t, new(SetExperimentTagTestSuite))
}

func (s *SetExperimentTagTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func setExperimentTag(s *SetExperimentTagTestSuite, experiment *models.Experiment, key, value string) {
	req := request.SetExperimentTagRequest{
		ID:    fmt.Sprintf("%d", *experiment.ID),
		Key:   key,
		Value: value,
	}
	resp := fiber.Map{}
	err := s.client.DoPostRequest(
		fmt.Sprintf(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
}

func (s *SetExperimentTagTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
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
	experiment1, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment2",
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

	// Set tag on experiment
	setExperimentTag(s, experiment, "dataset", "imagenet1K")
	exp, err := s.fixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)

	found := false
	for _, tag := range exp.Tags {
		if tag.Key == "dataset" && tag.Value == "imagenet1K" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected 'experiment.tags' to contain 'dataset' with value 'imagenet1K'")

	// Update tag on experiment
	setExperimentTag(s, experiment, "dataset", "birdbike")
	exp, err = s.fixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)

	found = false
	for _, tag := range exp.Tags {
		if tag.Key == "dataset" && tag.Value == "birdbike" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected 'experiment.tags' to contain 'dataset' with value 'imagenet1K'")

	//test that setting a tag on 1 experiment does not impact another experiment.
	exp, err = s.fixtures.GetExperimentByID(context.Background(), *experiment1.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(exp.Tags), 0)

	//test that setting a tag on different experiments maintain different values across experiments
	setExperimentTag(s, experiment1, "dataset", "birds200")
	exp, err = s.fixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)
	exp1, err := s.fixtures.GetExperimentByID(context.Background(), *experiment1.ID)
	assert.Nil(s.T(), err)
	found = false
	for _, tag := range exp.Tags {
		if tag.Key == "dataset" && tag.Value == "birdbike" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected 'experiment.tags' to contain 'dataset' with value 'birdbike'")

	found = false
	for _, tag := range exp1.Tags {
		if tag.Key == "dataset" && tag.Value == "birds200" {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected 'experiment.tags' to contain 'dataset' with value 'birds200'")

}
func (s *SetExperimentTagTestSuite) Test_Error() {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetExperimentTagRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.SetExperimentTagRequest{
				ID:    "",
				Key:   "test_key",
				Value: "test_value",
			},
		},
		{
			name:  "EmptyKeyProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.SetExperimentTagRequest{
				ID:    "1",
				Key:   "",
				Value: "test_value",
			},
		},
		{
			name:  "IncorrectExperimentID",
			error: api.NewBadRequestError(`Unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`),
			request: &request.SetExperimentTagRequest{
				ID:    "incorrect_experiment_id",
				Key:   "test_key",
				Value: "test_value",
			},
		},
		{
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': error getting experiment by id: 1: record not found`),
			request: &request.SetExperimentTagRequest{
				ID:    "1",
				Key:   "test_key",
				Value: "test_value",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
