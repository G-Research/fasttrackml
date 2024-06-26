package tag

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetTagTestSuite(t *testing.T) {
	suite.Run(t, new(GetTagTestSuite))
}

func (s *GetTagTestSuite) Test_Ok() {
	tag, err := s.SharedTagFixtures.CreateTag(context.Background(), "a-tag", s.DefaultNamespace.ID)
	s.Require().Nil(err)

	var resp response.TagResponse
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/tags/%s", tag.ID))
	s.Equal(tag.ID, resp.ID)
	s.Equal(tag.Name, resp.Name)
	s.Equal(tag.Description, resp.Description)
}

func (s *GetTagTestSuite) Test_Error() {
	tests := []struct {
		name    string
		idParam string
	}{
		{
			name:    "GetTagWithNotFoundID",
			idParam: uuid.New().String(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/tags/%s", tt.idParam),
			)
			s.Equal("Not Found", resp.Message)
		})
	}
}
