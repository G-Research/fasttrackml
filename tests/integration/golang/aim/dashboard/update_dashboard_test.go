//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:         uuid.New(),
			IsArchived: false,
			CreatedAt:  time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	require.Nil(s.T(), err)

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
	require.Nil(s.T(), err)

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
			require.Nil(
				s.T(),
				s.AIMClient.WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/dashboards/%s", dashboard.ID,
				),
			)
			require.Nil(
				s.T(),
				s.AIMClient.WithMethod(
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

			require.Nil(s.T(), err)
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
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:         uuid.New(),
			IsArchived: false,
			CreatedAt:  time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	require.Nil(s.T(), err)

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
	require.Nil(s.T(), err)

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
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			require.Nil(s.T(), s.AIMClient.WithMethod(
				http.MethodPut,
			).WithRequest(
				tt.requestBody,
			).WithResponse(
				&resp,
			).DoRequest(
				"/dashboards/%s", tt.ID,
			))
			assert.Contains(s.T(), resp.Message, tt.error)
		})
	}
}
