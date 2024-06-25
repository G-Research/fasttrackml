package tag

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type CreateTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestCreateTagTestSuite(t *testing.T) {
	suite.Run(t, new(CreateTagTestSuite))
}

func (s *CreateTagTestSuite) Test_Ok() {
	tests := []struct {
		name        string
		requestBody request.CreateTagRequest
	}{
		{
			name: "CreateValidTag",
			requestBody: request.CreateTagRequest{
				Name:        "tag-name",
				Description: "tag-description",
				Color:       "#cccccc",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.TagResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/tags"),
			)

			tags, err := s.SharedTagFixtures.GetTags(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.requestBody.Name, resp.Name)
			s.Equal(tt.requestBody.Description, resp.Description)
			s.Equal(tags[0].Name, resp.Name)
			s.Equal(tags[0].Description, resp.Description)
			s.Equal(tags[0].ID, resp.ID)
			s.NotEmpty(resp.ID)
		})
	}
}

func (s *CreateTagTestSuite) Test_Error() {
	tests := []struct {
		name        string
		requestBody request.CreateTagRequest
	}{
		{
			name: "CreateTagWithMissingField",
			requestBody: request.CreateTagRequest{
				Name:        "",
				Description: "tag-description",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.requestBody,
				).WithResponse(
					&resp,
				).DoRequest("/tags"),
			)
			s.Contains(resp.Message, "not a valid tag name")
		})
	}
}
