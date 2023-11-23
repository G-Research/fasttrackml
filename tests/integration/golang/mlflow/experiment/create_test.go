//go:build integration

package experiment

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CreateExperimentTestSuite))
}

func (s *CreateExperimentTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	req := request.CreateExperimentRequest{
		Name:             "ExperimentName",
		ArtifactLocation: "/artifact/location",
		Tags: []request.ExperimentTagPartialRequest{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2",
			},
		},
	}
	resp := response.CreateExperimentResponse{}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsCreateRoute,
		),
	)
	s.NotEmpty(resp.ID)
}

func (s *CreateExperimentTestSuite) Test_Error() {
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
		error   *api.ErrorResponse
		request *request.CreateExperimentRequest
	}{
		{
			name:    "EmptyNameProperty",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'name'"),
			request: &request.CreateExperimentRequest{},
		},
		{
			name: "EmptyArtifactLocationProperty",
			error: api.NewInvalidParameterValueError(
				`Invalid value for parameter 'artifact_location': error parsing artifact location: parse ` +
					`"incorrect-protocol,:/incorrect-location": first path segment in URL cannot contain colon`,
			),
			request: &request.CreateExperimentRequest{
				Name:             "name",
				ArtifactLocation: "incorrect-protocol,:/incorrect-location",
			},
		},
	}

	for _, tt := range testData {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsCreateRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
