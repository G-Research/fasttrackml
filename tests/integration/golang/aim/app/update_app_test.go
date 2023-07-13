//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*database.App
}

func TestUpdateAppTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateAppTestSuite))
}

func (s *UpdateAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
}

func (s *UpdateAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name        string
		requestBody map[string]any
	}{
		{
			name: "UpdateValidApp",
			requestBody: map[string]any{
				"type": "app-type",
				"state": map[string]string{
					"app-state-key": "new-app-state-value",
				},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp database.App
			err := s.client.DoPutRequest(
				fmt.Sprintf("/apps/%s", s.apps[0].ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "app-type", resp.Type)
			assert.Equal(s.T(), database.AppState{"app-state-key": "new-app-state-value"}, resp.State)
		})
	}
}

func (s *UpdateAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
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
			var resp map[string]any
			err := s.client.DoPutRequest(
				fmt.Sprintf("/apps/%s", s.apps[0].ID),
				tt.requestBody,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), resp["message"], "cannot unmarshal")
		})
	}
}
