//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DashboardFlowTestSuite struct {
	helpers.BaseTestSuite
}

func TestDashboardFlowTestSuite(t *testing.T) {
	suite.Run(t, &DashboardFlowTestSuite{})
}

func (s *DashboardFlowTestSuite) TearDownTest() {
	s.Nil(s.NamespaceFixtures.UnloadFixtures())
}

func (s *DashboardFlowTestSuite) Test_Ok() {
	tests := []struct {
		name           string
		namespace1Code string
		namespace2Code string
	}{
		{
			name:           "TestCustomNamespaces",
			namespace1Code: "namespace-1",
			namespace2Code: "namespace-2",
		},
		{
			name:           "TestExplicitDefaultAndCustomNamespaces",
			namespace1Code: "default",
			namespace2Code: "namespace-1",
		},
		{
			name:           "TestImplicitDefaultAndCustomNamespaces",
			namespace1Code: "",
			namespace2Code: "namespace-1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			defer func() {
				s.Nil(s.NamespaceFixtures.UnloadFixtures())
			}()

			// setup namespaces
			for _, nsCode := range []string{"default", tt.namespace1Code, tt.namespace2Code} {
				_, err := s.NamespaceFixtures.UpsertNamespace(context.Background(), &models.Namespace{
					Code:                nsCode,
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				s.Require().Nil(err)
			}

			// run actual flow test over the test data.
			s.testDashboardFlow(tt.namespace1Code, tt.namespace2Code)
		})
	}
}

func (s *DashboardFlowTestSuite) testDashboardFlow(
	namespace1Code, namespace2Code string,
) {
	// create apps
	app1ID := s.createApp(namespace1Code, &request.CreateApp{
		Type: "tf",
		State: request.AppState{
			"app-state-key": "app-state-value1",
		},
	})

	app2ID := s.createApp(namespace2Code, &request.CreateApp{
		Type: "mpi",
		State: request.AppState{
			"app-state-key": "app-state-value2",
		},
	})

	// create dashboards
	dashboard1ID := s.createDashboard(namespace1Code, &request.CreateDashboard{
		Name:        "dashboard1-name",
		Description: "dashboard1-description",
		AppID:       uuid.MustParse(app1ID),
	})

	dashboard2ID := s.createDashboard(namespace2Code, &request.CreateDashboard{
		Name:        "dashboard2-name",
		Description: "dashboard2-description",
		AppID:       uuid.MustParse(app2ID),
	})

	// test `GET /dashboards` endpoint with namespace 1
	resp := s.getDashboards(namespace1Code)
	// only dashboard 1 should be present
	s.Equal(1, len(resp))
	s.Equal(dashboard1ID, resp[0].ID)

	// test `GET /dashboards` endpoint with namespace 2
	resp = s.getDashboards(namespace2Code)
	// only dashboard 2 should be present
	s.Equal(1, len(resp))
	s.Equal(dashboard2ID, resp[0].ID)

	// IDs from active namespace can be fetched, updated, and deleted
	s.getDashboardAndCompare(namespace1Code, dashboard1ID)
	s.updateDashboardAndCompare(namespace1Code, dashboard1ID)
	s.deleteDashboardAndCompare(namespace2Code, dashboard2ID)

	// IDs from other namespace cannot be fetched, updated, or deleted
	errResp := response.Error{}
	client := s.AIMClient()
	s.Require().Nil(
		client.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1Code,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/dashboards/%s", dashboard2ID),
		),
	)
	s.Equal(fiber.ErrNotFound.Code, client.GetStatusCode())

	client = s.AIMClient()
	s.Require().Nil(
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespace2Code,
		).WithRequest(
			request.UpdateDashboard{
				Name:        "new-dashboard-name",
				Description: "new-dashboard-description",
			},
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/dashboards/%s", dashboard1ID),
		),
	)
	s.Equal(fiber.ErrNotFound.Code, client.GetStatusCode())

	client = s.AIMClient()
	s.Require().Nil(
		client.WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespace2Code,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/dashboards/%s", dashboard1ID),
		),
	)
	s.Equal(fiber.ErrNotFound.Code, client.GetStatusCode())
}

func (s *DashboardFlowTestSuite) deleteDashboardAndCompare(namespaceCode string, dashboardID string) {
	client := s.AIMClient()
	dashboardResp := response.Dashboard{}
	s.Require().Nil(
		client.WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&dashboardResp,
		).DoRequest(
			"/dashboards/%s", dashboardID,
		),
	)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
}

func (s *DashboardFlowTestSuite) updateDashboardAndCompare(namespaceCode string, dashboardID string) {
	client := s.AIMClient()
	dashboardResp := response.Dashboard{}
	s.Require().Nil(
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespaceCode,
		).WithRequest(
			request.UpdateDashboard{
				Name:        "new-dashboard-name",
				Description: "new-dashboard-description",
			},
		).WithResponse(
			&dashboardResp,
		).DoRequest(
			fmt.Sprintf("/dashboards/%s", dashboardID),
		),
	)
	s.Equal(dashboardID, dashboardResp.ID)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
}

func (s *DashboardFlowTestSuite) getDashboardAndCompare(namespaceCode string, dashboardID string) response.Dashboard {
	dashboardResp := response.Dashboard{}
	client := s.AIMClient()
	s.Require().Nil(
		client.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&dashboardResp,
		).DoRequest(
			fmt.Sprintf("/dashboards/%s", dashboardID),
		),
	)
	s.Equal(dashboardID, dashboardResp.ID)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
	return dashboardResp
}

func (s *DashboardFlowTestSuite) getDashboards(namespaceCode string) []response.Dashboard {
	resp := []response.Dashboard{}
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&resp,
		).DoRequest(
			"/dashboards",
		),
	)
	return resp
}

func (s *DashboardFlowTestSuite) createApp(namespace string, req *request.CreateApp) string {
	var resp response.App
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"/apps",
		),
	)
	return resp.ID
}

func (s *DashboardFlowTestSuite) createDashboard(namespace string, req *request.CreateDashboard) string {
	var resp response.Dashboard
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"/dashboards",
		),
	)
	s.Equal(req.Name, resp.Name)
	s.NotEmpty(resp.ID)
	return resp.ID
}
