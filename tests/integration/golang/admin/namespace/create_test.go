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

type CreateNamespaceTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestCreateNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(CreateNamespaceTestSuite))
}

func (s *CreateNamespaceTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *CreateNamespaceTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	testNamespaces := []models.Namespace{
		{
			ID:                  2,
			Code:                "test2",
			Description:         "test namespace 2 description",
			DefaultExperimentID: common.GetPointer(int32(0)),
		},
		{
			ID:                  3,
			Code:                "test3",
			Description:         "test namespace 3 description",
			DefaultExperimentID: common.GetPointer(int32(0)),
		},
	}
	for _, namespace := range testNamespaces {
		request := request.Namespace{
			Code:        namespace.Code,
			Description: namespace.Description,
		}
		assert.Nil(
			s.T(),
			s.AdminClient.WithMethod(
				http.MethodPost,
			).WithRequest(
				request,
			).DoRequest("/namespaces"),
		)
	}

	namespaces, err := s.NamespaceFixtures.GetTestNamespaces(context.Background())
	assert.Nil(s.T(), err)

	// Check that the namespace has been created
	for _, namespace := range namespaces {
		for _, testNamespace := range testNamespaces {
			if namespace.Code == testNamespace.Code {
				assert.Equal(s.T(), testNamespace.Description, namespace.Description)
			}
		}
	}

	// Check the length of the namespaces considering the default namespace
	assert.Equal(s.T(), len(testNamespaces)+1, len(namespaces))
}

func (s *CreateNamespaceTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	testData := []struct {
		name    string
		request *request.Namespace
	}{
		{
			name: "EmptyCode",
			request: &request.Namespace{
				Code:        "",
				Description: "description",
			},
		},
		{
			name: "CodeLenghtLessThan2",
			request: &request.Namespace{
				Code:        "a",
				Description: "description",
			},
		},
		{
			name: "CodeLenghtGreaterThan12",
			request: &request.Namespace{
				Code:        "TooLongNamespaceCode",
				Description: "description",
			},
		},
		{
			name: "InvalidCode",
			request: &request.Namespace{
				Code:        "test#",
				Description: "description",
			},
		},
		{
			name: "CodeAlreadyExists",
			request: &request.Namespace{
				Code:        "default",
				Description: "description",
			},
		},
	}
	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			assert.Nil(
				s.T(),
				s.AdminClient.WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).DoRequest("/namespaces"),
			)
			namespaces, err := s.NamespaceFixtures.GetTestNamespaces(context.Background())
			assert.Nil(s.T(), err)
			// Check that creation failed, only the default namespace is present
			assert.Equal(s.T(), 1, len(namespaces))
		})
	}
}
