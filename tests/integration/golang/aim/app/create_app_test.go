//go:build integration

package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
}

func TestCreateAppTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAppTestSuite))
}

func (s *CreateAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures
}

func (s *CreateAppTestSuite) Test_Ok() {
	defer func() { assert.Nil(s.T(), s.appFixtures.UnloadFixtures()) }()
	tests := []struct {
		name        string
		requestBody map[string]any
	}{
		{
			name: "CreateValidApp",
			requestBody: map[string]any{
				"type": "app-type",
				"state": map[string]any{
					"app-state-key": "app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp map[string]any
			err := s.client.DoPostRequest(
				"/apps",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.requestBody["type"], resp["type"])
			assert.Equal(s.T(), tt.requestBody["state"], resp["state"])
			assert.NotEmpty(s.T(), resp["id"])
			assert.NotEmpty(s.T(), resp["created_at"])
			assert.NotEmpty(s.T(), resp["updated_at"])
		})
	}
}

func (s *CreateAppTestSuite) Test_Error() {
	tests := []struct {
		name        string
		requestBody any
	}{
		{
			name: "CreateAppWithIncorrectJson",
			requestBody: map[string]any{
				"State": "this-will-not-unmarshal",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp map[string]any
			err := s.client.DoPostRequest(
				"/apps",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp["message"], "cannot unmarshal")
		})
	}
}
