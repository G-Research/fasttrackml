//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchExperimentsTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestSearchExperimentsTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *SearchExperimentsTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

type Experiment struct {
	Name string
	Tags []models.ExperimentTag
}

func getExperimentNames(experiments []*response.ExperimentPartialResponse) []string {
	names := make([]string, len(experiments))
	for i, e := range experiments {
		names[i] = e.Name
	}
	return names
}

func (s *SearchExperimentsTestSuite) Test_Ok() {
	experiments := []Experiment{
		{
			Name: "a",
			Tags: []models.ExperimentTag{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
		},
		{
			Name: "ab",
			Tags: []models.ExperimentTag{
				{
					Key:   "key2",
					Value: "value2",
				},
			},
		},
		{
			Name: "Abc",
			Tags: nil,
		},
	}
	for _, ex := range experiments {
		_, err := s.fixtures.CreateTestExperiment(context.Background(), &models.Experiment{
			Name: ex.Name,
			Tags: ex.Tags,
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
	}
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()

	query, err := urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "attribute.name = 'a'",
	})
	assert.Nil(s.T(), err)

	resp := response.SearchExperimentsResponse{}
	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"a"}, getExperimentNames(resp.Experiments))

	query, err = urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "attribute.name != 'a'",
	})
	assert.Nil(s.T(), err)

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"Abc", "ab"}, getExperimentNames(resp.Experiments))

	query, err = urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "name LIKE 'a%'",
	})
	assert.Nil(s.T(), err)

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"ab", "a"}, getExperimentNames(resp.Experiments))

	query, err = urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "tag.key = 'value'",
	})
	assert.Nil(s.T(), err)

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"a"}, getExperimentNames(resp.Experiments))

	query, err = urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "tag.key != 'value'",
	})
	assert.Nil(s.T(), err)

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"ab"}, getExperimentNames(resp.Experiments))

	query, err = urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: "tag.key ILIKE '%alu%'",
	})
	assert.Nil(s.T(), err)

	err = s.client.DoGetRequest(
		fmt.Sprintf(
			"%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query,
		),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"ab", "a"}, getExperimentNames(resp.Experiments))
}

func (s *SearchExperimentsTestSuite) Test_Error() {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.GetExperimentRequest
	}{
		{
			name:  "IncorrectExperimentID",
			error: api.NewBadRequestError(`unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`),
			request: &request.GetExperimentRequest{
				ID: "incorrect_experiment_id",
			},
		},
		{
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': error getting experiment by id: 1: record not found`),
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
