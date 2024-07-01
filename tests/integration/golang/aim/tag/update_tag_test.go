package tag

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type UpdateTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestUpdateTagTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateTagTestSuite))
}

func (s *UpdateTagTestSuite) Test_Ok() {
	tag, err := s.SharedTagFixtures.CreateTag(context.Background(), "a-tag", s.DefaultNamespace.ID)
	s.Require().Nil(err)

	tests := []struct {
		name        string
		requestBody request.UpdateTagRequest
	}{
		{
			name: "UpdateTag",
			requestBody: request.UpdateTagRequest{
				ID:          tag.ID,
				Name:        "new-tag-name",
				Description: "new-tag-description",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.TagResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPut,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest(
					"/tags/%s", tag.ID,
				),
			)

			actualTag, err := s.SharedTagFixtures.GetTag(context.Background(), tag.ID)
			s.Require().Nil(err)

			s.Equal(actualTag.Name, resp.Name)
			s.Equal(actualTag.Description, resp.Description)
			s.Equal(actualTag.ID, resp.ID)
		})
	}
}

func (s *UpdateTagTestSuite) Test_Error() {
	tests := []struct {
		name        string
		ID          uuid.UUID
		requestBody request.UpdateTagRequest
	}{
		{
			name:        "UpdateTagWithUnknownID",
			ID:          uuid.New(),
			requestBody: request.UpdateTagRequest{},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithMethod(
				http.MethodPut,
			).WithRequest(
				tt.requestBody,
			).WithResponse(
				&resp,
			).DoRequest(
				"/tags/%s", tt.ID,
			))
			s.Contains(resp.Message, "Not Found")
		})
	}
}
