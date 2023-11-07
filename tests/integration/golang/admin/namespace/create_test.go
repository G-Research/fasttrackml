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

	requests := []request.Namespace{
		{
			Code:        "test2",
			Description: "test namespace 2 description",
		},
		{
			Code:        "test3",
			Description: "test namespace 3 description",
		},
	}
	for _, request := range requests {
		assert.Nil(
			s.T(),
			s.AdminClient.WithMethod(
				http.MethodPost,
			).WithRequest(
				request,
			).DoRequest("/namespaces"),
		)
	}

	namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
	assert.Nil(s.T(), err)
	assert.True(s.T(), helpers.CheckNamespaces(namespaces, requests))

	// Check the length of the namespaces considering the default namespace
	assert.Equal(s.T(), len(requests)+1, len(namespaces))
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
			namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
			assert.Nil(s.T(), err)
			// Check that creation failed, only the default namespace is present
			assert.Equal(s.T(), 1, len(namespaces))
		})
	}
}
