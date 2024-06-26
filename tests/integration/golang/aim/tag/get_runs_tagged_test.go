package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunsTaggedTestSuite struct {
	helpers.BaseTestSuite
	tag1 *models.SharedTag
	tag2 *models.SharedTag
}

func TestGetRunsTaggedTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunsTaggedTestSuite))
}

func (s *GetRunsTaggedTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	run1, err := s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)

	run2, err := s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)

	s.tag1, err = s.SharedTagFixtures.CreateTag(context.Background(), "tag-exp1", s.DefaultNamespace.ID)
	s.Require().Nil(err)
	s.tag2, err = s.SharedTagFixtures.CreateTag(context.Background(), "tag-exp2", s.DefaultNamespace.ID)
	s.Require().Nil(err)

	s.Require().Nil(s.SharedTagFixtures.Associate(context.Background(), s.tag1.ID.String(), run1.ID))
	s.Require().Nil(s.SharedTagFixtures.Associate(context.Background(), s.tag1.ID.String(), run2.ID))
}

func (s *GetRunsTaggedTestSuite) Test_Ok() {
	tests := []struct {
		name             string
		tagID            string
		expectedRunCount int
	}{
		{
			name:             "GetTagsWithRows",
			tagID:            s.tag1.ID.String(),
			expectedRunCount: 2,
		},
		{
			name:             "GetTagsWithNoRows",
			tagID:            s.tag2.ID.String(),
			expectedRunCount: 0,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.GetRunsTaggedResponse
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/tags/%s/runs", tt.tagID))
			s.Equal(tt.expectedRunCount, len(resp.Runs))
		})
	}
}
