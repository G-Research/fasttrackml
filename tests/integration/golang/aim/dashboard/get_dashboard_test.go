//go:build integration

package run

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetDashboardTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(GetDashboardTestSuite))
}

func (s *GetDashboardTestSuite) Test_Ok() {
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:         uuid.New(),
			IsArchived: false,
			CreatedAt:  time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	dashboard, err := s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:         uuid.New(),
			IsArchived: false,
			CreatedAt:  time.Now(),
		},
		Name:        "dashboard-exp",
		AppID:       &app.ID,
		Description: "dashboard for experiment",
	})
	s.Require().Nil(err)

	var resp response.Dashboard
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/dashboards/%s", dashboard.ID))
	s.Equal(dashboard.ID, resp.ID)
	s.Equal(app.ID, resp.AppID)
	s.Equal(app.Type, resp.AppType)
	s.Equal(dashboard.Name, resp.Name)
	s.Equal(dashboard.Description, resp.Description)
	s.NotEmpty(resp.CreatedAt)
	s.NotEmpty(resp.UpdatedAt)
}

func (s *GetDashboardTestSuite) Test_Error() {
	tests := []struct {
		name    string
		idParam string
	}{
		{
			name:    "GetDashboardWithNotFoundID",
			idParam: uuid.New().String(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/dashboards/%s", tt.idParam),
			)
			s.Equal("Not Found", resp.Message)
		})
	}
}
