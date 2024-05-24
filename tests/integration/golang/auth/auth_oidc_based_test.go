package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	aimResponse "github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	mlflowResponse "github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/common/config/auth"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers/oidc"
)

type OIDCAuthTestSuite struct {
	helpers.BaseTestSuite
	namespace1     *models.Namespace
	namespace2     *models.Namespace
	namespace3     *models.Namespace
	user1Token     string
	user2Token     string
	user3Token     string
	user4Token     string
	oidcMockServer *oidc.MockServer
}

func TestOIDCAuthTestSuite(t *testing.T) {
	// create and run OIDC mock server.
	oidcMockServer, err := oidc.NewMockServer()
	assert.Nil(t, err)

	// create a service configuration with OIDC enabled option.
	testSuite := new(OIDCAuthTestSuite)
	cfg := config.Config{
		Auth: auth.Config{
			AuthOIDCAdminRole:        "admin",
			AuthOIDCClientID:         oidcMockServer.ClientID(),
			AuthOIDCClientSecret:     oidcMockServer.ClientSecret(),
			AuthOIDCClaimRoles:       "groups",
			AuthOIDCProviderEndpoint: oidcMockServer.Address(),
		},
	}
	assert.Nil(t, cfg.Validate())
	testSuite.Config = cfg
	testSuite.oidcMockServer = oidcMockServer
	suite.Run(t, testSuite)
}

func (s *OIDCAuthTestSuite) SetupTestSuite() {
	// prepare test data.
	namespace1, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "namespace1",
		Description:         "Test namespace 1",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	s.namespace1 = namespace1

	namespace2, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  3,
		Code:                "namespace2",
		Description:         "Test namespace 2",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	s.namespace2 = namespace2

	namespace3, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  4,
		Code:                "namespace3",
		Description:         "Test namespace 3",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	s.namespace3 = namespace3

	group1Role := models.Role{Name: "group1"}
	s.Nil(s.RolesFixtures.CreateRole(context.Background(), &group1Role))
	s.Nil(s.RolesFixtures.AttachNamespaceToRole(context.Background(), &group1Role, namespace1))

	group2Role := models.Role{Name: "group2"}
	s.Nil(s.RolesFixtures.CreateRole(context.Background(), &group2Role))
	s.Nil(s.RolesFixtures.AttachNamespaceToRole(context.Background(), &group2Role, namespace2))
	s.Nil(s.RolesFixtures.AttachNamespaceToRole(context.Background(), &group2Role, namespace3))

	// create test users and obtain theirs tokens.
	user1Token, err := s.oidcMockServer.Login(
		context.Background(),
		&mockoidc.MockUser{
			Email:  "test.user@example.com",
			Groups: []string{"group1"},
		}, []string{"openid", "groups"},
	)
	s.Nil(err)
	s.user1Token = user1Token

	user2Token, err := s.oidcMockServer.Login(
		context.Background(),
		&mockoidc.MockUser{
			Email:  "test.user@example.com",
			Groups: []string{"group2"},
		}, []string{"openid", "groups"},
	)
	s.Nil(err)
	s.user2Token = user2Token

	user3Token, err := s.oidcMockServer.Login(
		context.Background(),
		&mockoidc.MockUser{
			Email:  "test.user@example.com",
			Groups: []string{"group1", "group2"},
		}, []string{"openid", "groups"},
	)
	s.Nil(err)
	s.user3Token = user3Token

	user4Token, err := s.oidcMockServer.Login(
		context.Background(),
		&mockoidc.MockUser{
			Email:  "test.user@example.com",
			Groups: []string{"admin"},
		}, []string{"openid", "groups"},
	)
	s.Nil(err)
	s.user4Token = user4Token
}

func (s *OIDCAuthTestSuite) TestAIMAuth_Ok() {
	s.SetupTestSuite()

	tests := []struct {
		name  string
		check func()
	}{
		{
			name: "TestUser1NamespaceAccessLimits",
			check: func() {
				// check that user1 has access to namespace1 namespaces.
				successResponse := aimResponse.GetProjectResponse{}
				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				// check that user1 has no access to namespace2, namespace3.
				errorResponse := api.ErrorResponse{}
				client := s.AIMClient().WithResponse(
					&errorResponse,
				).WithNamespace(
					s.namespace2.Code,
				).WithHeaders(map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
				})
				s.Require().Nil(client.DoRequest("/projects"))

				s.Equal(client.GetStatusCode(), http.StatusNotFound)
				s.Equal(
					"RESOURCE_DOES_NOT_EXIST: unable to find namespace with code: namespace2", errorResponse.Error(),
				)

				client = s.AIMClient().WithResponse(
					&errorResponse,
				).WithNamespace(
					s.namespace3.Code,
				).WithHeaders(map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
				})
				s.Require().Nil(client.DoRequest("/projects"))

				s.Equal(client.GetStatusCode(), http.StatusNotFound)
				s.Equal(
					"RESOURCE_DOES_NOT_EXIST: unable to find namespace with code: namespace3", errorResponse.Error(),
				)
			},
		},
		{
			name: "TestUser2NamespaceAccessLimits",
			check: func() {
				// check that user2 has access to namespace2 and namespace3 namespaces.
				successResponse := aimResponse.GetProjectResponse{}
				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user2Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user2Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				// check that user2 has no access to namespace1
				errorResponse := api.ErrorResponse{}
				client := s.AIMClient().WithResponse(
					&errorResponse,
				).WithNamespace(
					s.namespace1.Code,
				).WithHeaders(map[string]string{
					"Authorization": fmt.Sprintf("Bearer %s", s.user2Token),
				})
				s.Require().Nil(client.DoRequest("/projects"))

				s.Equal(client.GetStatusCode(), http.StatusNotFound)
				s.Equal(
					"RESOURCE_DOES_NOT_EXIST: unable to find namespace with code: namespace1", errorResponse.Error(),
				)
			},
		},
		{
			name: "TestUser3NamespaceAccessLimits",
			check: func() {
				// check that user3 has access to namespace1, namespace2 and namespace3 namespaces.
				successResponse := aimResponse.GetProjectResponse{}
				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)
			},
		},
		{
			name: "TestAdminUserNamespaceAccessLimits",
			check: func() {
				// check that admin user has access to namespace1, namespace2 and namespace3 namespaces.
				successResponse := aimResponse.GetProjectResponse{}
				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)

				s.Require().Nil(
					s.AIMClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest("/projects"),
				)
				s.Equal("FastTrackML", successResponse.Name)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.check()
		})
	}
}

func (s *OIDCAuthTestSuite) TestMlflowAuth_Ok() {
	s.SetupTestSuite()

	tests := []struct {
		name  string
		check func()
	}{
		{
			name: "TestUser1NamespaceAccessLimits",
			check: func() {
				// check that user1 has access to namespace1 namespaces.
				successResponse := mlflowResponse.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)
			},
		},
		{
			name: "TestUser2NamespaceAccessLimits",
			check: func() {
				// check that user2 has access to namespace2 and namespace3 namespaces.
				successResponse := mlflowResponse.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user2Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user2Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)
			},
		},
		{
			name: "TestUser3NamespaceAccessLimits",
			check: func() {
				// check that user3 has access to namespace1, namespace2 and namespace3 namespaces.
				successResponse := mlflowResponse.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user3Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)
			},
		},
		{
			name: "TestAdminUserNamespaceAccessLimits",
			check: func() {
				// check that admin user has access to namespace1, namespace2 and namespace3 namespaces.
				successResponse := mlflowResponse.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace1.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace2.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						s.namespace3.Code,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.check()
		})
	}
}

func (s *OIDCAuthTestSuite) TestAdminAuth_Ok() {
	s.SetupTestSuite()

	tests := []struct {
		name  string
		check func()
	}{
		{
			name: "TestNonAdminUserHasNoAccess",
			check: func() {
				// check that user1 has no access to admin part.
				var resp goquery.Document
				s.Require().Nil(
					s.AdminClient().WithMethod(
						http.MethodGet,
					).WithResponseType(
						helpers.ResponseTypeHTML,
					).WithResponse(
						&resp,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
					}).DoRequest("/namespaces"),
				)

				s.Equal(0, resp.Find("#namespaces").Length())
			},
		},
		{
			name: "TestAdminUserHasAccess",
			check: func() {
				// check that user4(admin) has access to admin part.
				var resp goquery.Document
				s.Require().Nil(
					s.AdminClient().WithMethod(
						http.MethodGet,
					).WithResponseType(
						helpers.ResponseTypeHTML,
					).WithResponse(
						&resp,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest("/namespaces"),
				)
				s.Equal(1, resp.Find("#namespaces").Length(), "namespaces not found in HTML response")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.check()
		})
	}
}

func (s *OIDCAuthTestSuite) TestChooserAuth_Ok() {
	s.SetupTestSuite()

	tests := []struct {
		name  string
		check func()
	}{
		{
			name: "TestUserHasAccess",
			check: func() {
				// check that user1 has access to chooser part.
				var resp goquery.Document
				s.Require().Nil(
					s.ChooserClient().WithMethod(
						http.MethodGet,
					).WithResponseType(
						helpers.ResponseTypeHTML,
					).WithResponse(
						&resp,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user1Token),
					}).DoRequest("/"),
				)

				// check that HTML response has no button with id - #oidc-connect.
				s.Equal(0, resp.Find("#oidc-connect").Length())
			},
		},
		{
			name: "TestUserHasNoAccess",
			check: func() {
				// check that if there is no token, then `login` screen will be provided.
				var resp goquery.Document
				s.Require().Nil(
					s.ChooserClient().WithMethod(
						http.MethodGet,
					).WithResponseType(
						helpers.ResponseTypeHTML,
					).WithResponse(
						&resp,
					).WithHeaders(map[string]string{
						"Authorization": "Bearer incorrect-token",
					}).DoRequest("/"),
				)

				// check that HTML response has button with id - #oidc-connect.
				s.Equal(1, resp.Find("#oidc-connect").Length())
			},
		},
		{
			name: "TestAdminUserHasAccess",
			check: func() {
				// check that user4(admin) has access to chooser part.
				var resp goquery.Document
				s.Require().Nil(
					s.ChooserClient().WithMethod(
						http.MethodGet,
					).WithResponseType(
						helpers.ResponseTypeHTML,
					).WithResponse(
						&resp,
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Bearer %s", s.user4Token),
					}).DoRequest("/"),
				)
				// check that HTML response has no button with id - #oidc-connect.
				s.Equal(0, resp.Find("#oidc-connect").Length())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.check()
		})
	}
}
