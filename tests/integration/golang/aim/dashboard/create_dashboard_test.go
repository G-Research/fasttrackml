//go:build integration

package run

import (
	"context"
	"testing"
	"time"

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
}

func TestCreateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDashboardTestSuite))
}

func (s *CreateDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *CreateDashboardTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name        string
		requestBody request.CreateDashboard
	}{
		{
			name: "CreateValidDashboard",
			requestBody: request.CreateDashboard{
				AppID:       app.ID,
				Name:        "dashboard-name",
				Description: "dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Dashboard
			err := s.AIMClient.DoPostRequest("/dashboards", tt.requestBody, &resp)
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
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	assert.Nil(s.T(), err)

	tests := []struct {
		name        string
		requestBody request.CreateDashboard
	}{
		{
			name: "CreateDashboardWithNonExistentAppID",
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
