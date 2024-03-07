//go:build pipeline

package namespace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/admin/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetCurrentNamespacesTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetCurrentNamespacesTestSuite(t *testing.T) {
	suite.Run(t, new(GetCurrentNamespacesTestSuite))
}

func (s *GetCurrentNamespacesTestSuite) Test_Ok() {
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  2,
		Code:                "ns1",
		Description:         "Test namespace 1 description",
		DefaultExperimentID: common.GetPointer(models.DefaultExperimentID),
	})
	s.Require().Nil(err)

	var resp response.Namespace
	s.Require().Nil(s.AdminClient().WithNamespace(namespace.Code).WithResponse(&resp).DoRequest("/namespaces/current"))

	s.Equal(namespace.ID, resp.ID)
	s.Equal(namespace.Code, resp.Code)
	s.Equal(namespace.Description, resp.Description)
}
