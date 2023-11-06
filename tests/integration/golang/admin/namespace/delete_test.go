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
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteNamespaceTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestDeleteNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteNamespaceTestSuite))
}

func (s *DeleteNamespaceTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *DeleteNamespaceTestSuite) Test_Ok() {
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
	ns2, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  3,
		Code:                "test3",
		Description:         "test namespace 3 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name                   string
		expectedNamespaceCount int
	}{
		{
			name:                   "DeleteNamespace",
			expectedNamespaceCount: 2,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			assert.Nil(
				s.T(),
				s.AdminClient.WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/namespaces/%d", ns2.ID,
				),
			)
			namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedNamespaceCount, len(namespaces))
		})
	}
}

func (s *DeleteNamespaceTestSuite) Test_Error() {
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

	testData := []struct {
		name                    string
		ID                      string
		expectedNamespacesCount int
	}{
		{
			name:                    "DeleteNamespaceWithNotFoundID",
			ID:                      "10",
			expectedNamespacesCount: 2,
		},
		{
			name:                    "DeleteNamespaceWithEmptyID",
			ID:                      "",
			expectedNamespacesCount: 2,
		},
		{
			name:                    "DeleteNamespaceWithInvalidID",
			ID:                      "InvalidID",
			expectedNamespacesCount: 2,
		},
	}
	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			assert.Nil(
				s.T(),
				s.AdminClient.WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/namespaces/%s", tt.ID,
				),
			)
		})
		namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
		assert.Nil(s.T(), err)
		// Check that deletion failed and the namespace is still there
		assert.Equal(s.T(), tt.expectedNamespacesCount, len(namespaces))
	}
}
