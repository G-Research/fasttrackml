//go:build integration
package run

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*models.App
}

func TestGetAppTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppTestSuite))
}

func (s *GetAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))

	appFixtures, err := fixtures.NewAppFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateTestApps(context.Background(), 10)
	assert.Nil(s.T(), err)
}

func (s *GetAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		expectedAppCount int
	}{
		{
			name: "GetAppWithExistingID",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var resp models.App
			err := s.client.DoGetRequest(
				fmt.Sprintf("/apps/%v", s.apps[0].ID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), s.apps[0].ID, resp.ID)
		})
	}
}

func (s *GetAppTestSuite) Test_Error() {
	assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	tests := []struct {
		name string
	}{
		{
			name: "GetAppWithUnknownID",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var resp map[string]string
			err := s.client.DoGetRequest(
				fmt.Sprintf("/apps/%v", uuid.New().String()),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "Not Found", resp["message"]) 
		})
	}
}
