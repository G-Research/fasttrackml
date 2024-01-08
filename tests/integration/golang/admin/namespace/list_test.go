package namespace

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/admin/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListNamespacesTestSuite struct {
	helpers.BaseTestSuite
}

func TestListNamespacesTestSuite(t *testing.T) {
	suite.Run(t, new(ListNamespacesTestSuite))
}

func (s *ListNamespacesTestSuite) Test_Ok() {
	namespaces := map[string]*models.Namespace{}
	defaultNamespace, err := s.NamespaceFixtures.GetDefaultNamespace(context.Background())
	s.Require().Nil(err)
	namespaces[fmt.Sprintf("%d", defaultNamespace.ID)] = defaultNamespace

	for i := 0; i < 5; i++ {
		namespace := &models.Namespace{
			ID:                  uint(i + 2),
			Code:                fmt.Sprintf("Test Namespace %d", i),
			Description:         fmt.Sprintf("Test namespace %d description", i),
			DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
		}
		namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace)
		s.Require().Nil(err)
		namespaces[fmt.Sprintf("%d", namespace.ID)] = namespace
	}

	var resp response.ListNamespaces
	s.Require().Nil(s.AdminClient().WithResponse(&resp).DoRequest("/namespaces/list"))
	// +1 for default namespace
	s.Require().Equal(len(namespaces), len(resp))
	for _, actualNamespace := range resp {
		expectedNamespace := namespaces[fmt.Sprintf("%d", actualNamespace.ID)]
		s.Equal(expectedNamespace.ID, actualNamespace.ID)
		s.Equal(expectedNamespace.Code, actualNamespace.Code)
		s.Equal(expectedNamespace.Description, actualNamespace.Description)
	}
}
