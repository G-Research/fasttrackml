//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DashboardFlowTestSuite struct {
	helpers.BaseTestSuite
	testBuckets []string
	s3Client    *s3.Client
}

func TestDashboardFlowTestSuite(t *testing.T) {
	suite.Run(t, &DashboardFlowTestSuite{})
}

func (s *DashboardFlowTestSuite) TearDownTest() {
	assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *DashboardFlowTestSuite) Test_Ok() {
	s3Client, err := helpers.NewS3Client(helpers.GetS3EndpointUri())
	require.Nil(s.T(), err)
	s.s3Client = s3Client

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
		s.T().Run(tt.name, func(T *testing.T) {
			defer func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			}()

			// setup namespaces
			for _, nsCode := range []string{"default", tt.namespace1Code, tt.namespace2Code} {
				// ignore errors here since default exists on first run
				//nolint:errcheck
				s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					Code:                nsCode,
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
			}

			// run actual flow test over the test data.
			s.testDashboardFlow(tt.namespace1Code, tt.namespace2Code)
		})
	}
}

func (s *DashboardFlowTestSuite) testDashboardFlow(
	namespace1Code, namespace2Code string,
) {
	// create Dashboards
	Dashboard1ID := s.createDashboard(namespace1Code, &request.CreateDashboard{
		Type: "tf",
		State: request.DashboardState{
			"Dashboard-state-key": "Dashboard-state-value1",
		},
	})

	Dashboard2ID := s.createDashboard(namespace2Code, &request.CreateDashboard{
		Type: "mpi",
		State: request.DashboardState{
			"Dashboard-state-key": "Dashboard-state-value2",
		},
	})

	// test `GET /Dashboards` endpoint with namespace 1
	resp := []response.Dashboard{}
	require.Nil(
		s.T(),
		s.AIMClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&resp,
		).DoRequest(
			"/Dashboards",
		),
	)
	// only Dashboard 1 should be present
	assert.Equal(s.T(), 1, len(resp))
	assert.Equal(s.T(), Dashboard1ID, resp[0].ID)

	// test `GET /Dashboards` endpoint with namespace 2
	require.Nil(
		s.T(),
		s.AIMClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace2Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&resp,
		).DoRequest(
			"/Dashboards",
		),
	)
	// only Dashboard 2 should be present
	assert.Equal(s.T(), 1, len(resp))
	assert.Equal(s.T(), Dashboard2ID, resp[0].ID)

	// IDs from other namespace cannot be fetched, updated, or deleted
	errResp := response.Error{}
	client := s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard2ID),
		),
	)
	assert.Equal(s.T(), fiber.ErrNotFound.Code, client.GetStatusCode())

	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespace2Code,
		).WithRequest(
			request.UpdateDashboard{
				Type: "Dashboard-type",
				State: request.DashboardState{
					"Dashboard-state-key": "new-Dashboard-state-value",
				},
			},
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard1ID),
		),
	)
	assert.Equal(s.T(), fiber.ErrNotFound.Code, client.GetStatusCode())

	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespace2Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard1ID),
		),
	)
	assert.Equal(s.T(), fiber.ErrNotFound.Code, client.GetStatusCode())

	// IDs from active namespace can be fetched, updated, and deleted
	DashboardResp := response.Dashboard{}
	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&DashboardResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard1ID),
		),
	)
	assert.Equal(s.T(), Dashboard1ID, DashboardResp.ID)
	assert.Equal(s.T(), fiber.StatusOK, client.GetStatusCode())

	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespace1Code,
		).WithRequest(
			request.UpdateDashboard{
				Type: "Dashboard-type",
				State: request.DashboardState{
					"Dashboard-state-key": "new-Dashboard-state-value",
				},
			},
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&DashboardResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard1ID),
		),
	)
	assert.Equal(s.T(), Dashboard1ID, DashboardResp.ID)
	assert.Equal(s.T(), fiber.StatusOK, client.GetStatusCode())

	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodDelete,
		).WithNamespace(
			namespace2Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&DashboardResp,
		).DoRequest(
			fmt.Sprintf("/Dashboards/%s", Dashboard2ID),
		),
	)
	assert.Equal(s.T(), fiber.StatusOK, client.GetStatusCode())
}

func (s *DashboardFlowTestSuite) createDashboard(namespace string, req *request.CreateDashboard) string {
	var resp response.Dashboard
	require.Nil(
		s.T(),
		s.AIMClient().WithMethod(
			http.MethodPost,
		).WithNamespace(
			namespace,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"/Dashboards",
		),
	)
	assert.Equal(s.T(), req.Name, resp.Name)
	assert.NotEmpty(s.T(), resp.ID)
	return resp.ID
}
