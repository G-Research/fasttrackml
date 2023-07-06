//go:build integration

package run

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppsTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*database.App
}

func TestGetAppsTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppsTestSuite))
}

func (s *GetAppsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(os.Getenv("SERVICE_BASE_URL"))

	appFixtures, err := fixtures.NewAppFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateApps(context.Background(), 2)
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
			expectedAppCount: 2,
		},
		{
			name:             "GetAppsWithNoRows",
			expectedAppCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {

			var resp []database.App
			err := s.client.DoGetRequest(
				"/apps",
				&resp,
			)
			assert.Nil(s.T(), err)
			for idx := 0; idx < tt.expectedAppCount; idx++ {
				assert.Equal(s.T(), tt.expectedAppCount, len(resp))
				assert.Equal(s.T(), s.apps[0].ID, resp[0].ID)
				assert.Equal(s.T(), s.apps[0].Type, resp[0].Type)
				assert.Equal(s.T(), s.apps[0].State, resp[0].State)
				assert.NotEmpty(s.T(), resp[0].CreatedAt)
				assert.NotEmpty(s.T(), resp[0].UpdatedAt)
			}
		})
	}
}

func (s *GetAppsTestSuite) Test_Empty() {
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
