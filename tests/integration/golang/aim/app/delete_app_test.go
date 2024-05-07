package run

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/common/api"
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
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
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
	_, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
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
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/apps/%s", tt.idParam,
				),
			)
			s.Contains(strings.ToLower(resp.Message), "not found")

			apps, err := s.AppFixtures.GetApps(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedAppCount, len(apps))
		})
	}
}
