//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteAppTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
	app *database.App
}

func TestDeleteAppTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteAppTestSuite))
}

func (s *DeleteAppTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())

	apps, err := s.AppFixtures.CreateApps(context.Background(), 1)
	assert.Nil(s.T(), err)
	s.app = apps[0]
}

func (s *DeleteAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.AppFixtures.UnloadFixtures())
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
			var deleteResponse response.Error
			err := s.AIMClient.DoDeleteRequest(
				fmt.Sprintf("/apps/%s", s.app.ID),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			apps, err := s.AppFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}

func (s *DeleteAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.AppFixtures.UnloadFixtures())
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
			var deleteResponse response.Error
			err := s.AIMClient.DoDeleteRequest(
				fmt.Sprintf("/apps/%s", tt.idParam),
				&deleteResponse,
			)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse.Message, "Not Found")

			apps, err := s.AppFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}
