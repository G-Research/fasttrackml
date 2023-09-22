//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
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

			var resp []response.App
			err = s.client.DoGetRequest(
				"/apps",
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(resp))
			for idx := 0; idx < tt.expectedAppCount; idx++ {
				assert.Equal(s.T(), apps[idx].ID.String(), resp[idx].ID)
				assert.Equal(s.T(), apps[idx].Type, resp[idx].Type)
				assert.Equal(s.T(), apps[idx].State, database.AppState(resp[idx].State))
				// TODO these timestamps are not populated by the endpoint -- should they be?
				// assert.NotEmpty(s.T(), resp[idx].CreatedAt)
				// assert.NotEmpty(s.T(), resp[idx].UpdatedAt)
			}
		})
	}
}
