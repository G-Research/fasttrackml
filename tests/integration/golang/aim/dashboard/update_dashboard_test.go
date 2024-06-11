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

type UpdateDashboardTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateDashboardTestSuite))
}

func (s *UpdateDashboardTestSuite) Test_Ok() {
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

	tests := []struct {
		name        string
		requestBody request.UpdateDashboardRequest
	}{
		{
			name: "UpdateDashboard",
			requestBody: request.UpdateDashboardRequest{
				ID:          dashboard.ID,
				Name:        "new-dashboard-name",
				Description: "new-dashboard-description",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Dashboard
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/dashboards/%s", dashboard.ID,
				),
			)
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/dashboards/%s", dashboard.ID,
				),
			)

			actualDashboard, err := s.DashboardFixtures.GetDashboardByID(context.Background(), dashboard.ID.String())

			s.Require().Nil(err)
			s.Equal(tt.requestBody.Name, resp.Name)
			s.Equal(tt.requestBody.Description, resp.Description)
			s.Equal(dashboard.ID, resp.ID)
			s.Equal(tt.requestBody.Name, actualDashboard.Name)
			s.Equal(tt.requestBody.Description, actualDashboard.Description)
		})
	}
}

func (s *UpdateDashboardTestSuite) Test_Error() {
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

	tests := []struct {
		name        string
		ID          uuid.UUID
		requestBody map[string]interface{}
		error       string
	}{
		{
			name: "UpdateDashboardWithIncorrectDescriptionType",
			ID:   dashboard.ID,
			requestBody: map[string]interface{}{
				"Description": map[string]interface{}{"Description": "latest-description"},
			},
			error: "cannot unmarshal",
		},
		{
			name:  "UpdateDashboardWithUnknownID",
			ID:    uuid.New(),
			error: "Not Found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithMethod(
				http.MethodPut,
			).WithRequest(
				tt.requestBody,
			).WithResponse(
				&resp,
			).DoRequest(
				"/dashboards/%s", tt.ID,
			))
			s.Contains(resp.Message, tt.error)
		})
	}
}
