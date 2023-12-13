package experiment

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateExperimentTestSuite(t *testing.T) {
	suite.Run(t, &CreateExperimentTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *CreateExperimentTestSuite) Test_Ok() {
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
			name: "IncorrectArtifactLocationProperty",
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
