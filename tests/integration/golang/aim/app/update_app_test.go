//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateAppTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app *database.App
}

func TestUpdateAppTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateAppTestSuite))
}

func (s *UpdateAppTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]
}

func (s *UpdateAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.AppFixtures.UnloadFixtures())
	}()
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
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.App
			err := s.AIMClient.DoPutRequest(
				fmt.Sprintf("/apps/%s", s.app.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "app-type", resp.Type)
			assert.Equal(s.T(), response.AppState{"app-state-key": "new-app-state-value"}, resp.State)
		})
	}
}

func (s *UpdateAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.AppFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody any
	}{
		{
			name: "UpdateAppWithIncorrectState",
			requestBody: map[string]any{
				"State": "this-cannot-unmarshal",
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp response.Error
			err := s.AIMClient.DoPutRequest(
				fmt.Sprintf("/apps/%s", s.app.ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp.Message, "cannot unmarshal")
		})
	}
}
