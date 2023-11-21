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

type CreateDashboardTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDashboardTestSuite))
}

func (s *CreateDashboardTestSuite) Test_Ok() {
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
		s.Run(tt.name, func() {
			var resp response.Dashboard
			require.Nil(
				s.T(),
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/dashboards"),
			)

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			require.Nil(s.T(), err)
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
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

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
		s.Run(tt.name, func() {
			var resp response.Error
			require.Nil(
				s.T(),
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/dashboards"),
			)
			assert.Contains(s.T(), resp.Message, "Not Found")
		})
	}
}
