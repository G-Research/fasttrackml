//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteDashboardTestSuite struct {
	suite.Suite
	client            *helpers.HttpClient
	appFixtures       *fixtures.AppFixtures
	dashboardFixtures *fixtures.DashboardFixtures
	app               *database.App
	dashboard         *database.Dashboard
}

func TestDeleteDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteDashboardTestSuite))
}

func (s *DeleteDashboardTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	apps, err := s.appFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]

	dashboardFixtures, err := fixtures.NewDashboardFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.dashboardFixtures = dashboardFixtures

	dashboards, err := s.dashboardFixtures.CreateDashboards(context.Background(), 1, &s.app.ID)
	assert.Nil(s.T(), err)
	s.dashboard = dashboards[0]
}

func (s *DeleteDashboardTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.dashboardFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		expectedDashboardCount int
	}{
		{
			name:             "DeleteDashboard",
			expectedDashboardCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse response.Error
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/dashboards/%s", s.dashboard.ID),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			dashboards, err := s.dashboardFixtures.GetDashboards(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(dashboards))
		})
	}
}

func (s *DeleteDashboardTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.dashboardFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		idParam          uuid.UUID
		expectedDashboardCount int
	}{
		{
			name:             "DeleteDashboardWithNotFoundID",
			idParam:          uuid.New(),
			expectedDashboardCount: 1,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse response.Error
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/dashboards/%s", tt.idParam),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse.Message, "Not Found")

			dashboards, err := s.dashboardFixtures.GetDashboards(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(dashboards))
		})
	}
}
