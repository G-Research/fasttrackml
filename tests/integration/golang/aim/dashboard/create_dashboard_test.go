//go:build integration

package run

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app *database.App
}

func TestCreateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDashboardTestSuite))
}

func (s *CreateDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]
}

func (s *CreateDashboardTestSuite) Test_Ok() {
	defer func() { assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures()) }()
	tests := []struct {
		name        string
		requestBody request.CreateDashboard
	}{
		{
			name: "CreateValidDashboard",
			requestBody: request.CreateDashboard{
				AppID:       s.app.ID,
				Name:        "dashboard-name",
				Description: "dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Dashboard
			err := s.AIMClient.DoPostRequest(
				"/dashboards",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.requestBody.Name, resp.Name)
			assert.Equal(s.T(), tt.requestBody.Description, resp.Description)
			assert.Equal(s.T(), dashboards[0].Name, resp.Name)
			assert.Equal(s.T(), dashboards[0].Description, resp.Description)
			assert.Equal(s.T(), dashboards[0].ID.String(), resp.ID)
			assert.Equal(s.T(), dashboards[0].AppID, &resp.AppID)
			assert.NotEmpty(s.T(), resp.ID)
		})
	}
}

func (s *CreateDashboardTestSuite) Test_Error() {
	tests := []struct {
		name        string
		requestBody request.CreateDashboard
	}{
		{
			name: "CreateDashboardWithNon-ExistentAppID",
			requestBody: request.CreateDashboard{
				AppID:       uuid.New(),
				Name:        "dashboard-name",
				Description: "dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.AIMClient.DoPostRequest(
				"/dashboards",
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "Not Found")
		})
	}
}
