//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type AppFlowTestSuite struct {
	helpers.BaseTestSuite
}

func TestAppFlowTestSuite(t *testing.T) {
	suite.Run(t, &AppFlowTestSuite{})
}

func (s *AppFlowTestSuite) TearDownTest() {
	s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
}

func (s *AppFlowTestSuite) Test_Ok() {
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
				s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
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
			s.testAppFlow(tt.namespace1Code, tt.namespace2Code)
		})
	}
}

func (s *AppFlowTestSuite) testAppFlow(
	namespace1Code, namespace2Code string,
) {
	// create Apps
	app1ID := s.createAppAndCompare(namespace1Code, &request.CreateApp{
		Type: "tf",
		State: request.AppState{
			"app-state-key": "app-state-value1",
		},
	})

	app2ID := s.createAppAndCompare(namespace2Code, &request.CreateApp{
		Type: "mpi",
		State: request.AppState{
			"app-state-key": "app-state-value2",
		},
	})

	// test `GET /apps` endpoint with namespace 1
	resp := s.getApps(namespace1Code)
	// only app 1 should be present
	s.Equal(1, len(resp))
	s.Equal(app1ID, resp[0].ID)

	// test `GET /apps` endpoint with namespace 2
	resp = s.getApps(namespace2Code)
	// only app 2 should be present
	s.Equal(1, len(resp))
	s.Equal(app2ID, resp[0].ID)

	// IDs from active namespace can be fetched, updated, and deleted
	s.getAppAndCompare(namespace1Code, app1ID)
	s.updateAppAndCompare(namespace1Code, app1ID)
	s.deleteAppAndCompare(namespace2Code, app2ID)

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
			fmt.Sprintf("/apps/%s", app2ID),
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
			request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app1ID),
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
			fmt.Sprintf("/apps/%s", app1ID),
		),
	)
	s.Equal(fiber.ErrNotFound.Code, client.GetStatusCode())
}

func (s *AppFlowTestSuite) deleteAppAndCompare(namespaceCode string, appID string) {
	client := s.AIMClient()
	appResp := response.App{}
	s.Require().Nil(
		client.WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", appID),
		),
	)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
}

func (s *AppFlowTestSuite) updateAppAndCompare(namespaceCode string, appID string) {
	client := s.AIMClient()
	appResp := response.App{}
	s.Require().Nil(
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespaceCode,
		).WithRequest(
			request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		).WithResponse(
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", appID),
		),
	)
	s.Equal(appID, appResp.ID)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
}

func (s *AppFlowTestSuite) getAppAndCompare(namespaceCode string, appID string) response.App {
	appResp := response.App{}
	client := s.AIMClient()
	s.Require().Nil(
		client.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", appID),
		),
	)
	s.Equal(appID, appResp.ID)
	s.Equal(fiber.StatusOK, client.GetStatusCode())
	return appResp
}

func (s *AppFlowTestSuite) getApps(namespaceCode string) []response.App {
	resp := []response.App{}
	s.Require().Nil(
		s.AIMClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespaceCode,
		).WithResponse(
			&resp,
		).DoRequest(
			"/apps",
		),
	)
	return resp
}

func (s *AppFlowTestSuite) createAppAndCompare(namespace string, req *request.CreateApp) string {
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
	s.Equal(req.Type, resp.Type)
	s.Equal(req.State["app-state-key"], resp.State["app-state-key"])
	s.NotEmpty(resp.ID)
	return resp.ID
}
