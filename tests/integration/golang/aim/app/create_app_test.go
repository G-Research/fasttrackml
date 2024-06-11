package run

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateAppTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateAppTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAppTestSuite))
}

func (s *CreateAppTestSuite) Test_Ok() {
	tests := []struct {
		name        string
		requestBody request.CreateAppRequest
	}{
		{
			name: "CreateValidApp",
			requestBody: request.CreateAppRequest{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.App
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps",
				),
			)
			s.Equal(tt.requestBody.Type, resp.Type)
			s.Equal(tt.requestBody.State["app-state-key"], resp.State["app-state-key"])
			s.NotEmpty(resp.ID)
			s.NotEmpty(resp.CreatedAt)
			s.NotEmpty(resp.UpdatedAt)
		})
	}
}

func (s *CreateAppTestSuite) Test_Error() {
	tests := []struct {
		name        string
		requestBody any
	}{
		{
			name: "CreateAppWithIncorrectJson",
			requestBody: map[string]any{
				"State": "bad json",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/apps"),
			)
			s.Contains(resp.Message, "cannot unmarshal")
		})
	}
}
