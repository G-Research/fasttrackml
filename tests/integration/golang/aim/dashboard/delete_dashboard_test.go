//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestDeleteDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteDashboardTestSuite))
}

func (s *DeleteDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *DeleteDashboardTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

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
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse response.Error
			err := s.AIMClient.DoDeleteRequest(fmt.Sprintf("/dashboards/%s", dashboard.ID), &deleteResponse)
			assert.Nil(s.T(), err)
			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(dashboards))
		})
	}
}

func (s *DeleteDashboardTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

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
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse response.Error
			err := s.AIMClient.DoDeleteRequest(fmt.Sprintf("/dashboards/%s", tt.idParam), &deleteResponse)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse.Message, "Not Found")

			dashboards, err := s.DashboardFixtures.GetDashboards(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(dashboards))
		})
	}
}
