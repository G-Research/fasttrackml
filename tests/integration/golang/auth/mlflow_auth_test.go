package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zeebo/assert"
	"gopkg.in/yaml.v3"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MlflowAuthTestSuite struct {
	helpers.BaseTestSuite
}

func TestMlflowAuthTestSuite(t *testing.T) {
	// create users configuration firstly.
	data, err := yaml.Marshal(auth.YamlConfig{
		Users: []auth.YamlUserConfig{
			{
				Name: "user1",
				Roles: []string{
					"ns:namespace1",
					"ns:namespace2",
				},
				Password: "user1password",
			},
			{
				Name: "user2",
				Roles: []string{
					"ns:namespace2",
					"ns:namespace3",
				},
				Password: "user2password",
			},
			{
				Name: "user3",
				Roles: []string{
					"admin",
				},
				Password: "user3password",
			},
		},
	})
	assert.Nil(t, err)

	configPath := fmt.Sprintf("%s/users-config.yaml", t.TempDir())
	// #nosec G304
	f, err := os.Create(configPath)
	_, err = f.Write(data)
	assert.Nil(t, err)
	assert.Nil(t, f.Close())

	// run test suite with newly created configuration.
	testSuite := new(MlflowAuthTestSuite)
	testSuite.Config = config.ServiceConfig{
		Auth: auth.Config{
			AuthType:        auth.TypeUser,
			AuthUsersConfig: configPath,
		},
	}
	suite.Run(t, testSuite)
}

func (s *MlflowAuthTestSuite) Test_Ok() {
	// create test namespaces.
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "namespace1",
		Description:         "Test namespace 1",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  3,
		Code:                "namespace2",
		Description:         "Test namespace 2",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  4,
		Code:                "namespace3",
		Description:         "Test namespace 3",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)

	tests := []struct {
		name  string
		check func()
	}{
		{
			name: "TestUser1Access",
			check: func() {
				// check that user1 has access to namespace1 and namespace2 namespaces.
				basicAuthToken := base64.StdEncoding.EncodeToString(
					[]byte(fmt.Sprintf("%s:%s", "user1", "user1password")),
				)
				successResponse := response.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace1",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace2",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest("/projects"),
				)
				s.Empty(0, successResponse.Experiments)

				// check that user1 has no access to namespace3 namespace.
				errorResponse := api.ErrorResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&errorResponse,
					).WithNamespace(
						"namespace3",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest("/projects"),
				)
				s.Equal(
					"RESOURCE_DOES_NOT_EXIST: unable to find namespace with code: namespace3", errorResponse.Error(),
				)
			},
		},
		{
			name: "TestUser2Access",
			check: func() {
				// check that user2 has access to namespace2 and namespace3 namespaces.
				basicAuthToken := base64.StdEncoding.EncodeToString(
					[]byte(fmt.Sprintf("%s:%s", "user2", "user2password")),
				)
				successResponse := response.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace2",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace3",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				// check that user2 has no access to namespace1 namespace.
				errorResponse := api.ErrorResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&errorResponse,
					).WithNamespace(
						"namespace1",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Equal(
					"RESOURCE_DOES_NOT_EXIST: unable to find namespace with code: namespace1", errorResponse.Error(),
				)
			},
		},
		{
			name: "TestUser3Access",
			check: func() {
				// check that user3 has access to namespace1, namespace2, namespace3 namespaces because of admin role.
				basicAuthToken := base64.StdEncoding.EncodeToString(
					[]byte(fmt.Sprintf("%s:%s", "user3", "user3password")),
				)
				successResponse := response.SearchExperimentsResponse{}
				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace1",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace2",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
					}).DoRequest(
						"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSearchRoute,
					),
				)
				s.Empty(0, successResponse.Experiments)

				s.Require().Nil(
					s.MlflowClient().WithResponse(
						&successResponse,
					).WithNamespace(
						"namespace3",
					).WithHeaders(map[string]string{
						"Authorization": fmt.Sprintf("Basic %s", basicAuthToken),
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
