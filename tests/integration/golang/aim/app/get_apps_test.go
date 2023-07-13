//go:build integration

package run

import (
	"context"
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
}

func TestGetAppsTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppsTestSuite))
}

func (s *GetAppsTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures
}

func (s *GetAppsTestSuite) Test_Ok() {
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
			defer func() {
				assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
			}()

			apps, err := s.appFixtures.CreateApps(context.Background(), tt.expectedAppCount)
			assert.Nil(s.T(), err)

			var resp []database.App
			err = s.client.DoGetRequest(
				"/apps",
				&resp,
			)
			assert.Nil(s.T(), err)
			for idx := 0; idx < tt.expectedAppCount; idx++ {
				assert.Equal(s.T(), tt.expectedAppCount, len(resp))
				assert.Equal(s.T(), apps[0].ID, resp[0].ID)
				assert.Equal(s.T(), apps[0].Type, resp[0].Type)
				assert.Equal(s.T(), apps[0].State, resp[0].State)
				assert.NotEmpty(s.T(), resp[0].CreatedAt)
				assert.NotEmpty(s.T(), resp[0].UpdatedAt)
			}
		})
	}
}
