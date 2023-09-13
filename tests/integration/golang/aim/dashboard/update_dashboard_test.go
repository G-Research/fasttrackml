//go:build integration

package run

import (
	"context"
	"fmt"
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

type UpdateDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestUpdateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateDashboardTestSuite))
}

func (s *UpdateDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *UpdateDashboardTestSuite) Test_Ok() {
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

	dashboard, err := s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Name:        "dashboard-exp",
		AppID:       &app.ID,
		Description: "dashboard for experiment",
	})
	assert.Nil(s.T(), err)

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
			err := s.AIMClient.DoPutRequest(fmt.Sprintf("/dashboards/%s", dashboard.ID), tt.requestBody, &resp)
			assert.Nil(s.T(), err)

			actualDashboard, err := s.DashboardFixtures.GetDashboardByID(context.Background(), dashboard.ID.String())

			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.requestBody.Name, resp.Name)
			assert.Equal(s.T(), tt.requestBody.Description, resp.Description)
			assert.Equal(s.T(), (dashboard.ID).String(), resp.ID)
			assert.Equal(s.T(), tt.requestBody.Name, actualDashboard.Name)
			assert.Equal(s.T(), tt.requestBody.Description, actualDashboard.Description)
		})
	}
}

func (s *UpdateDashboardTestSuite) Test_Error() {
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

	dashboard, err := s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Name:        "dashboard-exp",
		AppID:       &app.ID,
		Description: "dashboard for experiment",
	})
	assert.Nil(s.T(), err)

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
				fmt.Sprintf("/dashboards/%s", dashboard.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
