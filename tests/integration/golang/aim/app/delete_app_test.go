//go:build integration

package run

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteAppTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteAppTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteAppTestSuite))
}

func (s *DeleteAppTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/apps/%s", app.ID,
				),
			)
			apps, err := s.AppFixtures.GetApps(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedAppCount, len(apps))
		})
	}
}

func (s *DeleteAppTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	_, err = s.AppFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: namespace.ID,
	})
	s.Require().Nil(err)

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
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", tt.idParam,
				),
			)
			s.Contains(resp.Message, "Not Found")

			apps, err := s.AppFixtures.GetApps(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedAppCount, len(apps))
		})
	}
}
