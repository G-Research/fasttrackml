package run

import (
	"context"
	"testing"

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
	dashboard, err := s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Name: "dashboard-exp",
		App: database.App{
			Type:        "mpi",
			State:       database.AppState{},
			NamespaceID: s.DefaultNamespace.ID,
		},
		Description: "dashboard for experiment",
	})
	s.Require().Nil(err)

	var resp response.Dashboard
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/dashboards/%s", dashboard.ID))
	s.Equal(dashboard.ID, resp.ID)
	s.Equal(dashboard.App.ID, resp.AppID)
	s.Equal(dashboard.App.Type, resp.AppType)
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
