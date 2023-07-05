//go:build integration
package run

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*models.App
}

func TestDeleteAppTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteAppTestSuite))
}

func (s *DeleteAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))

	appFixtures, err := fixtures.NewAppFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateTestApps(context.Background(), 10)
	assert.Nil(s.T(), err)
}

func (s *DeleteAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		wantAppCount int
	}{
		{
			name: "DeleteApp",
			wantAppCount: 9,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var deleteResponse map[string]any
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/apps/%s", s.apps[0].ID),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)

			var getResponse []models.App
			err = s.client.DoGetRequest(
				"/apps",
				&getResponse,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.wantAppCount, len(getResponse))

		})
	}
}

func (s *DeleteAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody any
		wantAppCount int
	}{
		{
			name: "DeleteAppWithIncorrectBody",
			requestBody: map[string]any{
				"State": "this-cannot-unmarshal",
			},
			wantAppCount: 10,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var deleteResponse map[string]any
			err := s.client.DoPutRequest(
				fmt.Sprintf("/apps/%s", s.apps[0].ID),
				tt.requestBody,
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse["message"], "cannot unmarshal")

			var getResponse []models.App
			err = s.client.DoGetRequest(
				"/apps",
				&getResponse,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.wantAppCount, len(getResponse))

		})
	}
}
