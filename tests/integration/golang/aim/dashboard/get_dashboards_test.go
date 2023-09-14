//go:build integration

package run

import (
	"context"
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

type GetDashboardsTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestGetDashboardsTestSuite(t *testing.T) {
	suite.Run(t, new(GetDashboardsTestSuite))
}

func (s *GetDashboardsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
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
		s.T().Run(tt.name, func(T *testing.T) {
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

			dashboards, err := s.DashboardFixtures.CreateDashboards(
				context.Background(), tt.expectedDashboardCount, &app.ID,
			)
			assert.Nil(s.T(), err)

			var resp []response.Dashboard
			err = s.AIMClient.DoGetRequest(
				"/dashboards",
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedDashboardCount, len(resp))
			for idx := 0; idx < tt.expectedDashboardCount; idx++ {
				assert.Equal(s.T(), dashboards[idx].ID.String(), resp[idx].ID)
				assert.Equal(s.T(), app.ID, resp[idx].AppID)
				assert.Equal(s.T(), dashboards[idx].Name, resp[idx].Name)
				assert.Equal(s.T(), dashboards[idx].Description, resp[idx].Description)
				assert.NotEmpty(s.T(), resp[idx].CreatedAt)
				assert.NotEmpty(s.T(), resp[idx].UpdatedAt)
			}
		})
	}
}
