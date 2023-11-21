//go:build integration

package namespace

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteNamespaceTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteNamespaceTestSuite))
}

func (s *DeleteNamespaceTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "test2",
		Description:         "test namespace 2 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)
	ns2, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  3,
		Code:                "test3",
		Description:         "test namespace 3 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			s.Require().Nil(
				s.AdminClient().WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/namespaces/%d", ns2.ID,
				),
			)
			namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedNamespaceCount, len(namespaces))
		})
	}
}

func (s *DeleteNamespaceTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)
	_, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "test2",
		Description:         "test namespace 2 description",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	testData := []struct {
		name                    string
		ID                      string
		expectedNamespacesCount int
		response                map[string]any
	}{
		{
			name:                    "DeleteNamespaceWithNotFoundID",
			ID:                      "10",
			expectedNamespacesCount: 2,
			response: map[string]any{
				"message": "An unexepected error was encountered: namespace not found by id: 10",
				"status":  "error",
			},
		},
	}
	for _, tt := range testData {
		s.Run(tt.name, func() {
			var resp any
			s.Require().Nil(
				s.AdminClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/namespaces/%s", tt.ID,
				),
			)
			s.Equal(resp, tt.response)
		})
		namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
		s.Require().Nil(err)
		// Check that deletion failed and the namespace is still there
		s.Equal(tt.expectedNamespacesCount, len(namespaces))
	}
}
