//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*database.App
}

func TestGetAppTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppTestSuite))
}

func (s *GetAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
}

func (s *GetAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name string
	}{
		{
			name: "GetAppWithExistingID",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp database.App
			err := s.client.DoGetRequest(
				fmt.Sprintf("/apps/%v", s.apps[0].ID),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), s.apps[0].ID, resp.ID)
			assert.Equal(s.T(), s.apps[0].Type, resp.Type)
			assert.Equal(s.T(), s.apps[0].State, resp.State)
			assert.NotEmpty(s.T(), resp.CreatedAt)
			assert.NotEmpty(s.T(), resp.UpdatedAt)
		})
	}
}

func (s *GetAppTestSuite) Test_Error() {
	assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	tests := []struct {
		name    string
		idParam uuid.UUID
	}{
		{
			name:    "GetAppWithUnknownID",
			idParam: uuid.New(),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp map[string]string
			err := s.client.DoGetRequest(
				fmt.Sprintf("/apps/%v", tt.idParam),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), "Not Found", resp["message"])
		})
	}
}
