//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetDashboardsTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app *database.App
}

func TestGetDashboardsTestSuite(t *testing.T) {
	suite.Run(t, new(GetDashboardsTestSuite))
}

func (s *GetDashboardsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]
}

func (s *GetDashboardsTestSuite) Test_Ok() {
	tests := []struct {
		name                   string
		expectedDashboardCount int
	}{
		{
			name:                   "GetDashboardsWithExistingRows",
			expectedDashboardCount: 2,
		},
		{
			name:                   "GetDashboardsWithNoRows",
			expectedDashboardCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			defer func() {
				assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures())
			}()

			dashboards, err := s.DashboardFixtures.CreateDashboards(context.Background(), tt.expectedDashboardCount, &s.app.ID)
			assert.Nil(s.T(), err)

			var resp []response.Dashboard
			err = s.AIMClient.DoGetRequest(
				"/dashboards",
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(resp))
			for idx := 0; idx < tt.expectedDashboardCount; idx++ {
				assert.Equal(s.T(), dashboards[idx].ID.String(), resp[idx].ID)
				assert.Equal(s.T(), s.app.ID, resp[idx].AppID)
				assert.Equal(s.T(), dashboards[idx].Name, resp[idx].Name)
				assert.Equal(s.T(), dashboards[idx].Description, resp[idx].Description)
				assert.NotEmpty(s.T(), resp[idx].CreatedAt)
				assert.NotEmpty(s.T(), resp[idx].UpdatedAt)
			}
		})
	}
}
