package tag

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteTagTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteTagTestSuite))
}

func (s *DeleteTagTestSuite) Test_Ok() {
	tag, err := s.SharedTagFixtures.CreateTag(context.Background(), "a-tag", s.DefaultNamespace.ID)
	s.Require().Nil(err)

	tests := []struct {
		name string
	}{
		{
			name: "DeleteTag",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/tags/%s", tag.ID,
				),
			)
			fetchedTag, err := s.SharedTagFixtures.GetTag(context.Background(), tag.ID)
			s.Require().Nil(err)
			s.Require().True(fetchedTag.IsArchived)
		})
	}
}

func (s *DeleteTagTestSuite) Test_Error() {
	_, err := s.SharedTagFixtures.CreateTag(context.Background(), "tag-exp", s.DefaultNamespace.ID)
	s.Require().Nil(err)

	tests := []struct {
		name             string
		idParam          uuid.UUID
		expectedTagCount int
	}{
		{
			name:             "DeleteTagWithNotFoundID",
			idParam:          uuid.New(),
			expectedTagCount: 1,
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
					"/tags/%s", tt.idParam,
				),
			)
			s.Contains(resp.Message, "Not Found")

			tags, err := s.SharedTagFixtures.GetTags(context.Background())
			s.Require().Nil(err)
			s.Equal(tt.expectedTagCount, len(tags))
		})
	}
}
