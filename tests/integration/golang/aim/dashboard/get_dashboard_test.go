//go:build integration

package run

/*
import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app       *database.App
	dashboard *database.Dashboard
}

func TestGetDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(GetDashboardTestSuite))
}

func (s *GetDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]

	dashboards, err := s.DashboardFixtures.CreateDashboards(context.Background(), 1, &s.app.ID)
	assert.Nil(s.T(), err)
	s.dashboard = dashboards[0]
}

func (s *GetDashboardTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures())
	}()
	var resp database.Dashboard
	err := s.AIMClient.DoGetRequest(
		fmt.Sprintf("/dashboards/%v", s.dashboard.ID),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.dashboard.ID, resp.ID)
	assert.Equal(s.T(), &s.app.ID, resp.AppID)
	assert.Equal(s.T(), s.dashboard.Name, resp.Name)
	assert.Equal(s.T(), s.dashboard.Description, resp.Description)
	assert.NotEmpty(s.T(), resp.CreatedAt)
	assert.NotEmpty(s.T(), resp.UpdatedAt)
}

func (s *GetDashboardTestSuite) Test_Error() {
	assert.Nil(s.T(), s.DashboardFixtures.UnloadFixtures())
	tests := []struct {
		name    string
		idParam uuid.UUID
	}{
		{
			name:    "GetDashboardWithNotFoundID",
			idParam: uuid.New(),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.AIMClient.DoGetRequest(
				fmt.Sprintf("/dashboards/%v", tt.idParam),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "Not Found", resp.Message)
		})
	}
}
*/
