package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTagTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
	tag *aimModels.SharedTag
}

func TestDeleteRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTagTestSuite))
}

func (s *DeleteRunTagTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
	s.tag, err = s.SharedTagFixtures.CreateTag(context.Background(), "tag", s.DefaultExperiment.NamespaceID)
	s.Require().Nil(err)
}

func (s *DeleteRunTagTestSuite) Test_Ok() {
	tests := []struct {
		name  string
		runID string
		tagID string
	}{
		{
			name:  "DeleteTagFromExistingRun",
			runID: s.run.ID,
			tagID: s.tag.ID.String(),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().Nil(s.SharedTagFixtures.Associate(context.Background(), tt.tagID, tt.runID))

			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).DoRequest(
					"/runs/%s/tags/%s", tt.runID, tt.tagID,
				),
			)

			// verify
			tags, err := s.SharedTagFixtures.GetByRunID(context.Background(), tt.runID)
			s.Require().Nil(err)
			s.Require().Len(tags, 0)
		})
	}
}

func (s *DeleteRunTagTestSuite) Test_Error() {
	tests := []struct {
		name  string
		runID string
		tagID string
	}{
		{
			name:  "DeleteNonExistingTag",
			runID: s.run.ID,
			tagID: uuid.NewString(),
		},
		{
			name:  "DeleteNonExistingRun",
			runID: uuid.NewString(),
			tagID: s.tag.ID.String(),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s/tags/%s", tt.runID, tt.tagID,
				),
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
	}
}
