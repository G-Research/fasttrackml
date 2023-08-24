//go:build integration

package run

/*
import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
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
		requestBody request.CreateApp
	}{
		{
			name: "CreateValidApp",
			requestBody: request.CreateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.App
			err := s.client.DoPostRequest(
				"/apps",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.requestBody.Type, resp.Type)
			assert.Equal(s.T(), tt.requestBody.State["app-state-key"], resp.State["app-state-key"])
			assert.NotEmpty(s.T(), resp.ID)
			// TODO these timestamps are not set by the create endpoint
			// assert.NotEmpty(s.T(), resp.CreatedAt)
			// assert.NotEmpty(s.T(), resp.UpdatedAt)
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
				"State": "bad json",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.client.DoPostRequest(
				"/apps",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
*/
