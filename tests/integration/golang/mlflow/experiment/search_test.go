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
func executeSearchRequest(s *SearchExperimentsTestSuite, filter string) response.SearchExperimentsResponse {
	query, err := urlquery.Marshal(request.SearchExperimentsRequest{
		Filter: filter,
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

	return resp
}

func (s *SearchExperimentsTestSuite) Test_Ok() {
	// 1. prepare database with test data.
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

	// API call 1
	resp := executeSearchRequest(s, "attribute.name = 'a'")
	assert.Equal(s.T(), []string{"a"}, getExperimentNames(resp.Experiments))

	// API call 2
	resp = executeSearchRequest(s, "attribute.name != 'a'")
	assert.Equal(s.T(), []string{"Abc", "ab"}, getExperimentNames(resp.Experiments))

	// API call 3
	resp = executeSearchRequest(s, "name LIKE 'a%'")
	assert.Equal(s.T(), []string{"ab", "a"}, getExperimentNames(resp.Experiments))

	// API call 4
	resp = executeSearchRequest(s, "tag.key = 'value'")
	assert.Equal(s.T(), []string{"a"}, getExperimentNames(resp.Experiments))

	// API call 5
	resp = executeSearchRequest(s, "tag.key != 'value'")
	assert.Equal(s.T(), []string{"ab"}, getExperimentNames(resp.Experiments))

	// API call 6
	resp = executeSearchRequest(s, "tag.key ILIKE '%alu%'")
	assert.Equal(s.T(), []string{"ab", "a"}, getExperimentNames(resp.Experiments))
}

func (s *SearchExperimentsTestSuite) Test_Error() {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SearchExperimentsRequest
	}{
		{
			name:  "InvalidViewType",
			error: api.NewInvalidParameterValueError("Invalid view_type 'invalid_ViewType'"),
			request: &request.SearchExperimentsRequest{
				ViewType: "invalid_ViewType",
			},
		},
		{
			name:  "InvalidMaxResult",
			error: api.NewInvalidParameterValueError("Invalid value for parameter 'max_results' supplied."),
			request: &request.SearchExperimentsRequest{
				MaxResults: 10000000,
			},
		},
		{
			name:  "InvalidFilterValue",
			error: api.NewInvalidParameterValueError("invalid numeric value 'abc'"),
			request: &request.SearchExperimentsRequest{
				Filter: "attribute.creation_time>abc",
			},
		},
		{
			name:  "MalformedFilter",
			error: api.NewInvalidParameterValueError("malformed filter 'invalid_filter'"),
			request: &request.SearchExperimentsRequest{
				Filter: "invalid_filter",
			},
		},
		{
			name:  "InvalidAttributeComparisonOperator",
			error: api.NewInvalidParameterValueError("invalid numeric attribute comparison operator '>='"),
			request: &request.SearchExperimentsRequest{
				Filter: "attribute.creation_time>=100",
			},
		},
		{
			name:  "InvalidStringAttributeValue",
			error: api.NewInvalidParameterValueError("invalid string value '(abc'"),
			request: &request.SearchExperimentsRequest{
				Filter: "attribute.name LIKE '(abc'",
			},
		},
		{
			name:  "InvalidTagComparisonOperator",
			error: api.NewInvalidParameterValueError("invalid tag comparison operator '>='"),
			request: &request.SearchExperimentsRequest{
				Filter: "tags.name>='tag'",
			},
		},
		{
			name:  "InvalidEntity",
			error: api.NewInvalidParameterValueError("invalid entity type 'invalid_entity'. Valid values are ['tag', 'attribute']"),
			request: &request.SearchExperimentsRequest{
				Filter: "invalid_entity.name=value",
			},
		},
		{
			name:  "InvalidOrderByClause",
			error: api.NewInvalidParameterValueError("invalid order_by clause 'invalid_column'"),
			request: &request.SearchExperimentsRequest{
				OrderBy: []string{"invalid_column"},
			},
		},
		{
			name:  "InvalidOrderByAttribute",
			error: api.NewInvalidParameterValueError("invalid attribute 'invalid_attribute'. Valid values are ['name', 'experiment_id', 'creation_time', 'last_update_time']"),
			request: &request.SearchExperimentsRequest{
				OrderBy: []string{"invalid_attribute"},
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)
			resp := api.ErrorResponse{}
			err = s.client.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute, query),
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}

}
