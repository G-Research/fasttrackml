//go:build integration

package namespace

import (
	"context"
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateNamespaceTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(CreateNamespaceTestSuite))
}

func (s *CreateNamespaceTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
		s.Require().Nil(
			s.AdminClient().WithMethod(
				http.MethodPost,
			).WithRequest(
				request,
			).DoRequest("/namespaces"),
		)
	}

	namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
	s.Require().Nil(err)
	s.True(helpers.CheckNamespaces(namespaces, requests))

	// Check the length of the namespaces considering the default namespace
	s.Equal(len(requests)+1, len(namespaces))
}

func (s *CreateNamespaceTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	testData := []struct {
		name    string
		request *request.Namespace
		error   string
	}{
		{
			name: "EmptyCode",
			request: &request.Namespace{
				Code:        "",
				Description: "description",
			},
			error: "The namespace code is invalid.",
		},
		{
			name: "CodeLenghtLessThan2",
			request: &request.Namespace{
				Code:        "a",
				Description: "description",
			},
			error: "The namespace code is invalid.",
		},
		{
			name: "CodeLenghtGreaterThan12",
			request: &request.Namespace{
				Code:        "TooLongNamespaceCode",
				Description: "description",
			},
			error: "The namespace code is invalid.",
		},
		{
			name: "InvalidCode",
			request: &request.Namespace{
				Code:        "test#",
				Description: "description",
			},
			error: "The namespace code is invalid.",
		},
		{
			name: "CodeAlreadyExists",
			request: &request.Namespace{
				Code:        "default",
				Description: "description",
			},
			error: "The namespace code is already in use.",
		},
	}
	for _, tt := range testData {
		s.Run(tt.name, func() {
			var resp goquery.Document
			s.Require().Nil(
				s.AdminClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponseType(
					helpers.ResponseTypeHTML,
				).WithResponse(
					&resp,
				).DoRequest("/namespaces"),
			)

			msg := resp.Find(".error-message").Text()
			s.Equal(tt.error, msg)

			namespaces, err := s.NamespaceFixtures.GetNamespaces(context.Background())
			s.Require().Nil(err)

			// Check that creation failed, only the default namespace is present
			s.Equal(1, len(namespaces))
		})
	}
}
