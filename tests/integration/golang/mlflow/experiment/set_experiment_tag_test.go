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

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SetExperimentTagTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestSetExperimentTagTestSuite(t *testing.T) {
	suite.Run(t, new(SetExperimentTagTestSuite))
}

func (s *SetExperimentTagTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func SetExperimentTag(s *SetExperimentTagTestSuite, experiment *models.Experiment, key, value string) {
	req := request.SetExperimentTagRequest{
		ID:    fmt.Sprintf("%d", *experiment.ID),
		Key:   key,
		Value: value,
	}
	err := s.client.DoPostRequest(
		fmt.Sprintf(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
		req,
		&struct{}{},
	)
	assert.Nil(s.T(), err)
}

func (s *SetExperimentTagTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()
	// 1. prepare database with test data.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
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
	experiment1, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
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

	// Set tag on experiment
	SetExperimentTag(s, experiment, "KeyTag1", "ValueTag1")
	exp, err := s.experimentFixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)

	assert.True(s.T(), helpers.CheckTagExists(exp.Tags, "KeyTag1", "ValueTag1"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'")

	// Update tag on experiment
	SetExperimentTag(s, experiment, "KeyTag1", "ValueTag2")
	exp, err = s.experimentFixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)

	assert.True(s.T(), helpers.CheckTagExists(exp.Tags, "KeyTag1", "ValueTag2"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'")

	// test that setting a tag on 1 experiment does not impact another experiment.
	exp, err = s.experimentFixtures.GetExperimentByID(context.Background(), *experiment1.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(exp.Tags), 0)

	// test that setting a tag on different experiments maintain different values across experiments
	SetExperimentTag(s, experiment1, "KeyTag1", "ValueTag3")
	exp, err = s.experimentFixtures.GetExperimentByID(context.Background(), *experiment.ID)
	assert.Nil(s.T(), err)
	exp1, err := s.experimentFixtures.GetExperimentByID(context.Background(), *experiment1.ID)
	assert.Nil(s.T(), err)
	assert.True(s.T(), helpers.CheckTagExists(exp.Tags, "KeyTag1", "ValueTag2"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag2'")
	assert.True(s.T(), helpers.CheckTagExists(exp1.Tags, "KeyTag1", "ValueTag3"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag3'")
}

func (s *SetExperimentTagTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetExperimentTagRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.SetExperimentTagRequest{
				ID: "",
			},
		},
		{
			name:  "EmptyKeyProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "",
			},
		},
		{
			name:  "IncorrectExperimentID",
			error: api.NewBadRequestError(`Unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`),
			request: &request.SetExperimentTagRequest{
				ID:  "incorrect_experiment_id",
				Key: "test_key",
			},
		},
		{
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': error getting experiment by id: 1: record not found`),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "test_key",
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
