package run

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetAppTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetAppTestSuite(t *testing.T) {
	suite.Run(t, new(GetAppTestSuite))
}

func (s *GetAppTestSuite) Test_Ok() {
	app, err := s.AppFixtures.CreateApp(context.Background(), &database.App{
		Type:        "mpi",
		State:       database.AppState{},
		NamespaceID: s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	var resp database.App
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/apps/%s", app.ID.String()))
	s.Equal(app.ID, resp.ID)
	s.Equal(app.Type, resp.Type)
	s.Equal(app.State, resp.State)
	s.NotEmpty(resp.CreatedAt)
	s.NotEmpty(resp.UpdatedAt)
}

func (s *GetAppTestSuite) Test_Error() {
	tests := []struct {
		name    string
		idParam uuid.UUID
	}{
		{
			name:    "GetAppWithNotFoundID",
			idParam: uuid.New(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/apps/%v", tt.idParam))
			s.Contains(strings.ToLower(resp.Message), "not found")
		})
	}
}
