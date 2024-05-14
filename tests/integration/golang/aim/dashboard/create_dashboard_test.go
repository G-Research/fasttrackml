package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateDashboardTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDashboardTestSuite))
}

func (s *CreateDashboardTestSuite) Test_Ok() {
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	tests := []struct {
		name        string
		requestBody request.CreateDashboardRequest
	}{
		{
			name: "CreateValidDashboard",
			requestBody: request.CreateDashboardRequest{
				AppID:       app.ID,
				Name:        "dashboard-name",
				Description: "dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Dashboard
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/dashboards"),
			)

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.requestBody.Name, resp.Name)
			s.Equal(tt.requestBody.Description, resp.Description)
			s.Equal(dashboards[0].Name, resp.Name)
			s.Equal(dashboards[0].Description, resp.Description)
			s.Equal(dashboards[0].ID, resp.ID)
			s.Equal(dashboards[0].AppID, &resp.AppID)
			s.NotEmpty(resp.ID)
		})
	}
}

func (s *CreateDashboardTestSuite) Test_Error() {
	tests := []struct {
		name        string
		requestBody request.CreateDashboardRequest
	}{
		{
			name: "CreateDashboardWithNonExistentAppID",
			requestBody: request.CreateDashboardRequest{
				AppID:       uuid.New(),
				Name:        "dashboard-name",
				Description: "dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/dashboards"),
			)
			s.Contains(resp.Message, "Not Found")
		})
	}
}
