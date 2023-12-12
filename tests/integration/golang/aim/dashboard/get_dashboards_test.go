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

type GetDashboardsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetDashboardsTestSuite(t *testing.T) {
	suite.Run(t, &GetDashboardsTestSuite{
		helpers.BaseTestSuite{
			ResetOnSubTest: true,
		},
	})
}

func (s *GetDashboardsTestSuite) Test_Ok() {
	tests := []struct {
		name                   string
		expectedDashboardCount int
	}{
		{
			name:                   "GetDashboardsWithExistingRows",
			expectedDashboardCount: 2,
		},
		{
			name:                   "GetDashboardsWithNoRows",
			expectedDashboardCount: 0,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
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

			dashboards, err := s.DashboardFixtures.CreateDashboards(
				context.Background(), tt.expectedDashboardCount, &app.ID,
			)
			s.Require().Nil(err)

			var resp []response.Dashboard
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/dashboards"))
			s.Equal(tt.expectedDashboardCount, len(resp))
			for idx := 0; idx < tt.expectedDashboardCount; idx++ {
				s.Equal(dashboards[idx].ID, resp[idx].ID)
				s.Equal(app.ID, resp[idx].AppID)
				s.Equal(app.Type, resp[idx].AppType)
				s.Equal(dashboards[idx].Name, resp[idx].Name)
				s.Equal(dashboards[idx].Description, resp[idx].Description)
				s.NotEmpty(resp[idx].CreatedAt)
				s.NotEmpty(resp[idx].UpdatedAt)
			}
		})
	}
}
