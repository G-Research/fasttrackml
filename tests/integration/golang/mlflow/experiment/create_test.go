//go:build integration

package experiment

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateExperimentTestSuite struct {
	suite.Suite
	client   *helpers.HttpClient
	fixtures *fixtures.ExperimentFixtures
}

func TestCreateExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CreateExperimentTestSuite))
}

func (s *CreateExperimentTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(os.Getenv("SERVICE_BASE_URL"))
	fixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.fixtures = fixtures
}

func (s *CreateExperimentTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()

	req := request.CreateExperimentRequest{
		Name:             "ExperimentName",
		ArtifactLocation: "/artifact/location",
		Tags: []request.ExperimentTagPartialRequest{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2",
			},
		},
	}
	resp := response.CreateExperimentResponse{}
	err := s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsCreateRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.NotEmpty(s.T(), resp.ID)
}
func (s *CreateExperimentTestSuite) Test_Error() {
	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.CreateExperimentRequest
	}{
		{
			name:    "EmptyNameProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'name'"),
			request: &request.CreateExperimentRequest{},
		},
		{
			name: "EmptyArtifactLocationProperty",
			error: api.NewInvalidParameterValueError(
				`Invalid value for parameter 'artifact_location': parse "incorrect-protocol,:/incorrect-location": first path segment in URL cannot contain colon`,
			),
			request: &request.CreateExperimentRequest{
				Name:             "name",
				ArtifactLocation: "incorrect-protocol,:/incorrect-location",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsCreateRoute),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
