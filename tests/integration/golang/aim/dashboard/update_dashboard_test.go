//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app       *database.App
	dashboard *database.Dashboard
}

func TestUpdateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateDashboardTestSuite))
}

func (s *UpdateDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]

	dashboards, err := s.DashboardFixtures.CreateDashboards(context.Background(), 1, &s.app.ID)
	assert.Nil(s.T(), err)
	s.dashboard = dashboards[0]
}

func (s *UpdateDashboardTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody request.UpdateDashboard
	}{
		{
			name: "UpdateDashboard",
			requestBody: request.UpdateDashboard{
				Name:        "new-dashboard-name",
				Description: "new-dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Dashboard
			err := s.AIMClient.DoPutRequest(
				fmt.Sprintf("/dashboards/%s", s.dashboard.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			s.dashboard = &dashboards[0]

			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.requestBody.Name, resp.Name)
			assert.Equal(s.T(), tt.requestBody.Description, resp.Description)
			assert.Equal(s.T(), (s.dashboard.ID).String(), resp.ID)
			assert.Equal(s.T(), tt.requestBody.Name, s.dashboard.Name)
			assert.Equal(s.T(), tt.requestBody.Description, s.dashboard.Description)
			assert.Equal(s.T(), s.dashboard.Name, resp.Name)
			assert.Equal(s.T(), s.dashboard.Description, resp.Description)
		})
	}
}

func (s *UpdateDashboardTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "UpdateDashboardWithIncorrectDescriptionType",
			requestBody: map[string]interface{}{
				"Description": map[string]interface{}{"Description": "latest-description"},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.AIMClient.DoPutRequest(
				fmt.Sprintf("/dashboards/%s", s.dashboard.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
