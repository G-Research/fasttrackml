//go:build integration

package run

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteAppTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestDeleteAppTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteAppTestSuite))
}

func (s *DeleteAppTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *DeleteAppTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  0,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	assert.Nil(s.T(), err)

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
			err := s.AIMClient.DoDeleteRequest(fmt.Sprintf("/apps/%s", app.ID), &deleteResponse)
			assert.Nil(s.T(), err)
			apps, err := s.AppFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}

func (s *DeleteAppTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  0,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	_, err = s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	assert.Nil(s.T(), err)

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
			err := s.AIMClient.DoDeleteRequest(fmt.Sprintf("/apps/%s", tt.idParam), &deleteResponse)
			assert.Nil(s.T(), err)
			assert.Contains(s.T(), deleteResponse.Message, "Not Found")

			apps, err := s.AppFixtures.GetApps(context.Background())
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.expectedAppCount, len(apps))
		})
	}
}
