//go:build integration

package flows

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim"
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
	suite.Run(t, &AppFlowTestSuite{
		testBuckets: []string{"bucket1", "bucket2"},
	})
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
		setup          func() (*models.Namespace, *models.Namespace)
		namespace1Code string
		namespace2Code string
	}{
		{
			name: "TestCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-2",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "namespace-1",
			namespace2Code: "namespace-2",
		},
		{
			name: "TestExplicitDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "default",
			namespace2Code: "namespace-1",
		},
		{
			name: "TestImplicitDefaultAndCustomNamespaces",
			setup: func() (*models.Namespace, *models.Namespace) {
				return &models.Namespace{
						Code:                "default",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}, &models.Namespace{
						Code:                "namespace-1",
						DefaultExperimentID: common.GetPointer(int32(0)),
					}
			},
			namespace1Code: "",
			namespace2Code: "namespace-1",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			defer func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			}()

			// setup data under the test.
			namespace1, namespace2 := tt.setup()
			namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace1)
			require.Nil(s.T(), err)
			namespace2, err = s.NamespaceFixtures.CreateNamespace(context.Background(), namespace2)
			require.Nil(s.T(), err)

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
		s.AimClient().WithMethod(
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
		s.AimClient().WithMethod(
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

	// IDs from other namespace cannot be fetched
	resp = response.Error{}
	require.Nil(
		s.T(),
		s.AimClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace1Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app2ID),
		),
	)
	assert.Equal(s.T(), "Not Found", resp.Message)

	require.Nil(
		s.T(),
		s.AimClient().WithMethod(
			http.MethodGet,
		).WithNamespace(
			namespace2Code,
		).WithResponseType(
			helpers.ResponseTypeJSON,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("/apps/%s", app1ID),
		),
	)
	assert.Equal(s.T(), "Not Found", resp.Message)

	// IDs from active namespace can be fetched
	appResp := response.App{}
	require.Nil(
		s.T(),
		s.AimClient().WithMethod(
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
	assert.Equal(s.T(), app1ID, appResp[0].ID)

	require.Nil(
		s.T(),
		s.AimClient().WithMethod(
			http.MethodGet,
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
	assert.Equal(s.T(), app1ID, appResp[0].ID)
}

func (s *AppFlowTestSuite) createApp(namespace string, req *request.CreateApp) string {
	var resp response.App
	require.Nil(
		s.T(),
		s.AIMClient().WithMethod(
			http.MethodPost,
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

