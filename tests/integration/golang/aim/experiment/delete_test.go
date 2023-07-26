//go:build integration

package experiment

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
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
	defer func() {
		assert.Nil(s.T(), s.fixtures.UnloadFixtures())
	}()
	test_experiments, err := s.fixtures.CreateExperiments(context.Background(), 1)
	assert.Nil(s.T(), err)
	experiment := test_experiments[0]

	experiments, err := s.fixtures.GetTestExperiments(context.Background())
	length := len(experiments)

	var resp response.DeleteExperiment
	err = s.client.DoDeleteRequest(
		fmt.Sprintf("/experiments/%d", *experiment.ID),
		&resp,
	)
	assert.Nil(s.T(), err)

	remainingExperiments, err := s.fixtures.GetTestExperiments(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), length-1, len(remainingExperiments))
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
