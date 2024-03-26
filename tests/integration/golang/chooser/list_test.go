package chooser

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/api/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ListNamespacesTestSuite struct {
	helpers.BaseTestSuite
}

func TestListNamespacesTestSuite(t *testing.T) {
	suite.Run(t, new(ListNamespacesTestSuite))
}

func (s *ListNamespacesTestSuite) Test_Ok() {
	namespaces := map[uint]*models.Namespace{}
	defaultNamespace, err := s.NamespaceFixtures.GetNamespaceByCode(context.Background(), s.DefaultNamespace.Code)
	s.Require().Nil(err)
	namespaces[defaultNamespace.ID] = defaultNamespace

	for i := 0; i < 5; i++ {
		namespace := &models.Namespace{
			ID:                  uint(i + 2),
			Code:                fmt.Sprintf("ns%d", i),
			Description:         fmt.Sprintf("Test namespace %d description", i),
			DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
		}
		namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), namespace)
		s.Require().Nil(err)
		namespaces[namespace.ID] = namespace
	}

	var resp response.ListNamespaces
	s.Require().Nil(s.AdminClient().WithResponse(&resp).DoRequest("/namespaces/list"))

	s.Require().Equal(len(namespaces), len(resp))
	for _, actualNamespace := range resp {
		expectedNamespace := namespaces[actualNamespace.ID]
		s.Equal(expectedNamespace.ID, actualNamespace.ID)
		s.Equal(expectedNamespace.Code, actualNamespace.Code)
		s.Equal(expectedNamespace.Description, actualNamespace.Description)
	}
}
