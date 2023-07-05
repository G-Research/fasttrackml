//go:build integration

package run

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppsTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*models.App
}

func TestGetAppsTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppsTestSuite))
}

func (s *GetAppsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))

	appFixtures, err := fixtures.NewAppFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateTestApps(context.Background(), 10)
	assert.Nil(s.T(), err)
}

func (s *GetAppsTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		expectedAppCount int
	}{
		{
			name:             "GetAppsWithExistingRows",
			expectedAppCount: 10,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var resp []models.App
			err := s.client.DoGetRequest(
				"/apps",
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), s.apps[0].ID, resp[0].ID)
			assert.Equal(s.T(), tt.expectedAppCount, len(resp))
		})
	}
}

func (s *GetAppsTestSuite) Test_Error() {
	assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	tests := []struct {
		name             string
		expectedAppCount int
	}{
		{
			name:             "GetAppsWithNoRows",
			expectedAppCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var resp []any
			err := s.client.DoGetRequest(
				"/apps",
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(resp))
		})
	}
}
