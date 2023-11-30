//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteDashboardTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteDashboardTestSuite))
}

func (s *DeleteDashboardTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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

	tests := []struct {
		name                   string
		expectedDashboardCount int
	}{
		{
			name:                   "DeleteDashboard",
			expectedDashboardCount: 0,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/dashboards/%s", dashboard.ID,
				),
			)
			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedDashboardCount, len(dashboards))
		})
	}
}

func (s *DeleteDashboardTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
	s.Require().Nil(err)

	_, err = s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
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

	tests := []struct {
		name                   string
		idParam                uuid.UUID
		expectedDashboardCount int
	}{
		{
			name:                   "DeleteDashboardWithNotFoundID",
			idParam:                uuid.New(),
			expectedDashboardCount: 1,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/dashboards/%s", tt.idParam,
				),
			)
			s.Contains(resp.Message, "Not Found")

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedDashboardCount, len(dashboards))
		})
	}
}
