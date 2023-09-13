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
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetDashboardTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetDashboardTestSuite(t *testing.T) {
	suite.Run(t, new(GetDashboardTestSuite))
}

func (s *GetDashboardTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *GetDashboardTestSuite) Test_Ok() {
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	assert.Nil(s.T(), err)

	dashboard, err := s.DashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Name:        "dashboard-exp",
		AppID:       &app.ID,
		Description: "dashboard for experiment",
	})
	assert.Nil(s.T(), err)

	var resp database.Dashboard
	err = s.AIMClient.DoGetRequest(fmt.Sprintf("/dashboards/%v", dashboard.ID), &resp)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), dashboard.ID, resp.ID)
	assert.Equal(s.T(), &app.ID, resp.AppID)
	assert.Equal(s.T(), dashboard.Name, resp.Name)
	assert.Equal(s.T(), dashboard.Description, resp.Description)
	assert.NotEmpty(s.T(), resp.CreatedAt)
	assert.NotEmpty(s.T(), resp.UpdatedAt)
}

func (s *GetDashboardTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
			err := s.AIMClient.DoGetRequest(fmt.Sprintf("/dashboards/%v", tt.idParam), &resp)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "Not Found", resp.Message)
		})
	}
}
