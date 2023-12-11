package run

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
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
			dashboards, err := s.DashboardFixtures.CreateDashboards(
				context.Background(), s.DefaultNamespace, tt.expectedDashboardCount,
			)
			s.Require().Nil(err)

			var resp []response.Dashboard
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/dashboards"))
			s.Equal(tt.expectedDashboardCount, len(resp))
			for idx := 0; idx < tt.expectedDashboardCount; idx++ {
				s.Equal(dashboards[idx].ID, resp[idx].ID)
				s.Equal(dashboards[idx].App.ID, resp[idx].AppID)
				s.Equal(dashboards[idx].App.Type, resp[idx].AppType)
				s.Equal(dashboards[idx].Name, resp[idx].Name)
				s.Equal(dashboards[idx].Description, resp[idx].Description)
				s.NotEmpty(resp[idx].CreatedAt)
				s.NotEmpty(resp[idx].UpdatedAt)
			}
		})
	}
}
