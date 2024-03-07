//go:build pipeline

package run

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateAppTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateAppTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateAppTestSuite))
}

func (s *UpdateAppTestSuite) Test_Ok() {
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	tests := []struct {
		name        string
		requestBody request.UpdateApp
	}{
		{
			name: "UpdateApplication",
			requestBody: request.UpdateApp{
				Type: "app-type",
				State: request.AppState{
					"app-state-key": "new-app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.App
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", app.ID,
				),
			)
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", app.ID,
				),
			)
			s.Equal("app-type", resp.Type)
			s.Equal(response.AppState{"app-state-key": "new-app-state-value"}, resp.State)
		})
	}
}

func (s *UpdateAppTestSuite) Test_Error() {
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	tests := []struct {
		name        string
		ID          uuid.UUID
		requestBody any
		error       string
	}{
		{
			name: "UpdateAppWithIncorrectState",
			ID:   app.ID,
			requestBody: map[string]any{
				"State": "this-cannot-unmarshal",
			},
			error: "cannot unmarshal",
		},
		{
			name:        "UpdateAppWithUnknownID",
			ID:          uuid.New(),
			requestBody: map[string]any{},
			error:       "not found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", tt.ID,
				),
			)
			s.Contains(strings.ToLower(resp.Message), tt.error)
		})
	}
}
