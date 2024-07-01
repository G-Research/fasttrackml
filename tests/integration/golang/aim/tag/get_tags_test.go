package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetTagsTestSuite struct {
	helpers.BaseTestSuite
	tags []*models.SharedTag
}

func TestGetTagsTestSuite(t *testing.T) {
	suite.Run(t, new(GetTagsTestSuite))
}

func (s *GetTagsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	tag1, err := s.SharedTagFixtures.CreateTag(context.Background(), "tag-exp1", s.DefaultNamespace.ID)
	s.Require().Nil(err)
	tag2, err := s.SharedTagFixtures.CreateTag(context.Background(), "tag-exp2", s.DefaultNamespace.ID)
	s.Require().Nil(err)
	s.tags = []*models.SharedTag{tag1, tag2}
}

func (s *GetTagsTestSuite) Test_Ok() {
	tests := []struct {
		name             string
		expectedTagCount int
	}{
		{
			name:             "GetTagsWithExistingRows",
			expectedTagCount: 2,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.GetTagsResponse
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/tags"))
			s.Equal(tt.expectedTagCount, len(resp))
			for idx := 0; idx < tt.expectedTagCount; idx++ {
				s.Equal(s.tags[idx].ID, resp[idx].ID)
				s.Equal(s.tags[idx].Name, resp[idx].Name)
				s.Equal(s.tags[idx].Description, resp[idx].Description)
			}
		})
	}
}
