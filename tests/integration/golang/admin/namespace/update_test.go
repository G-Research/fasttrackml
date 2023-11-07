//go:build integration

package namespace

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateNamespaceTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestUpdateNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateNamespaceTestSuite))
}

func (s *UpdateNamespaceTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *UpdateNamespaceTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)
	ns, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "test2",
		Description:         "test namespace 2 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	request := request.Namespace{
		Code:        "test2Updated",
		Description: "test namespace 2 description updated",
	}
	assert.Nil(
		s.T(),
		s.AdminClient.WithMethod(
			http.MethodPut,
		).WithRequest(
			request,
		).DoRequest("/namespaces/%d", ns.ID),
	)

	namespace, err := s.NamespaceFixtures.GetNamespaceByID(context.Background(), ns.ID)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), namespace.Code, request.Code)
	assert.Equal(s.T(), namespace.Description, request.Description)
}

func (s *UpdateNamespaceTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "test2",
		Description:         "test namespace 2 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)
	expectedNamespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
	assert.Nil(s.T(), err)

	testData := []struct {
		name    string
		ID      string
		request *request.Namespace
	}{
		{
			name: "UpdateNamespaceWithNotFoundID",
			ID:   "10",
			request: &request.Namespace{
				Code:        "testUpdated",
				Description: "test namespace updated",
			},
		},
		{
			name: "UpdateNamespaceWithEmptyID",
			ID:   "",
			request: &request.Namespace{
				Code:        "testUpdated",
				Description: "test namespace updated",
			},
		},
		{
			name: "UpdateNamespaceWithInvalidID",
			ID:   "InvalidID",
			request: &request.Namespace{
				Code:        "testUpdated",
				Description: "test namespace updated",
			},
		},
		{
			name: "UpdateNamespaceWithDuplicatedCode",
			ID:   "2",
			request: &request.Namespace{
				Code:        "default",
				Description: "test namespace updated",
			},
		},
	}
	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			assert.Nil(
				s.T(),
				s.AdminClient.WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.request,
				).DoRequest(
					"/namespaces/%s", tt.ID,
				),
			)
		})
		actualNamespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
		assert.Nil(s.T(), err)
		assert.Equal(s.T(), expectedNamespaces, actualNamespaces)
	}
}
