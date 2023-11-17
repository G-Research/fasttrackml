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

type AppFlowTestSuite struct {
	helpers.BaseTestSuite
	testBuckets []string
	s3Client    *s3.Client
}

func TestAppFlowTestSuite(t *testing.T) {
	suite.Run(t, &AppFlowTestSuite{})
}

func (s *AppFlowTestSuite) TearDownTest() {
	assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *AppFlowTestSuite) Test_Ok() {
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
			s.testAppFlow(tt.namespace1Code, tt.namespace2Code)
		})
	}
}

func (s *AppFlowTestSuite) testAppFlow(
	namespace1Code, namespace2Code string,
) {
	// create Apps
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

	// test `GET /apps` endpoint with namespace 1
	resp := []response.App{}
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
			"/apps",
		),
	)
	// only app 1 should be present
	assert.Equal(s.T(), 1, len(resp))
	assert.Equal(s.T(), app1ID, resp[0].ID)

	// test `GET /apps` endpoint with namespace 2
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
			"/apps",
		),
	)
	// only app 2 should be present
	assert.Equal(s.T(), 1, len(resp))
	assert.Equal(s.T(), app2ID, resp[0].ID)

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
			fmt.Sprintf("/apps/%s", app2ID),
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
			request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&errResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app1ID),
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
			fmt.Sprintf("/apps/%s", app1ID),
		),
	)
	assert.Equal(s.T(), fiber.ErrNotFound.Code, client.GetStatusCode())

	// IDs from active namespace can be fetched, updated, and deleted
	appResp := response.App{}
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
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app1ID),
		),
	)
	assert.Equal(s.T(), app1ID, appResp.ID)
	assert.Equal(s.T(), fiber.StatusOK, client.GetStatusCode())

	client = s.AIMClient()
	require.Nil(
		s.T(),
		client.WithMethod(
			http.MethodPut,
		).WithNamespace(
			namespace1Code,
		).WithRequest(
			request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app1ID),
		),
	)
	assert.Equal(s.T(), app1ID, appResp.ID)
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
			&appResp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app2ID),
		),
	)
	assert.Equal(s.T(), fiber.StatusOK, client.GetStatusCode())
}

func (s *AppFlowTestSuite) createApp(namespace string, req *request.CreateApp) string {
	var resp response.App
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
			"/apps",
		),
	)
	assert.Equal(s.T(), req.Type, resp.Type)
	assert.Equal(s.T(), req.State["app-state-key"], resp.State["app-state-key"])
	assert.NotEmpty(s.T(), resp.ID)
	return resp.ID
}
