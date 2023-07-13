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

type DeleteAppTestSuite struct {
	suite.Suite
	client      *helpers.HttpClient
	appFixtures *fixtures.AppFixtures
	apps        []*database.App
}

func TestDeleteAppTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteAppTestSuite))
}

func (s *DeleteAppTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.appFixtures = appFixtures

	s.apps, err = s.appFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
}

func (s *DeleteAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		expectedAppCount int
	}{
		{
			name:             "DeleteApp",
			expectedAppCount: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse map[string]any
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/apps/%s", s.apps[0].ID),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)

			apps, err := s.appFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}

func (s *DeleteAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.appFixtures.UnloadFixtures())
	}()
	tests := []struct {
		name             string
		idParam          uuid.UUID
		expectedAppCount int
	}{
		{
			name:             "DeleteAppWithNotFoundID",
			idParam:          uuid.New(),
			expectedAppCount: 1,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var deleteResponse map[string]any
			err := s.client.DoDeleteRequest(
				fmt.Sprintf("/apps/%s", tt.idParam),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse["message"], "Not Found")

			apps, err := s.appFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}
