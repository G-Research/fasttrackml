//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *GetExperimentTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
}

func (s *GetExperimentTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// 1. prepare database with test data.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
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

	// 2. make actual API call.
	query, err := urlquery.Marshal(request.GetExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	})
	assert.Nil(s.T(), err)

	resp := response.GetExperimentResponse{}
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	// 3. check actual API response.
	assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.Experiment.ID)
	assert.Equal(s.T(), experiment.Name, resp.Experiment.Name)
	assert.Equal(s.T(), string(experiment.LifecycleStage), resp.Experiment.LifecycleStage)
	assert.Equal(s.T(), experiment.ArtifactLocation, resp.Experiment.ArtifactLocation)
	assert.Equal(s.T(), []models.ExperimentTag{
		{
			Key:          "key1",
			Value:        "value1",
			ExperimentID: *experiment.ID,
		},
	}, experiment.Tags)
}

func (s *GetExperimentTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetExperimentRequest
	}{
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`,
			),
			request: &request.GetExperimentRequest{
				ID: "incorrect_experiment_id",
			},
		},
		{
			name: "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(
				`unable to find experiment '1': error getting experiment by id: 1: record not found`,
			),
			request: &request.GetExperimentRequest{
				ID: "1",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := api.ErrorResponse{}
			err = s.client.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute, query),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
