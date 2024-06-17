package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	aimModels "github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RunTagTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
	tag *aimModels.SharedTag
}

func TestRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(RunTagTestSuite))
}

func (s *RunTagTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
	s.tag, err = s.SharedTagFixtures.CreateTag(context.Background(), "tag", s.DefaultExperiment.NamespaceID)
	s.Require().Nil(err)
}

func (s *RunTagTestSuite) TestAddRunTag() {
	tests := []struct {
		name    string
		runID   string
		request request.AddRunTagRequest
	}{
		{
			name:    "AddTagToExistingRun",
			runID:   s.run.ID,
			request: request.AddRunTagRequest{TagName: "tag"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.request.RunID = uuid.MustParse(tt.runID)
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).DoRequest(
					"/runs/%s/tags/new", tt.runID,
				),
			)

			// verify
			tags, err := s.SharedTagFixtures.GetByRunID(context.Background(), tt.runID)
			s.Require().Nil(err)
			s.Require().Len(tags, 1)
			s.Require().Equal(tt.request.TagName, tags[0].Name)
		})
	}
}

func (s *RunTagTestSuite) TestDeleteRunTag() {
	tests := []struct {
		name  string
		ID    string
		tagID string
		error *api.ErrorResponse
	}{
		{
			name:  "DeleteTagFromExistingRun",
			ID:    "existing-run-ID",
			tagID: "existing-tag-ID",
			error: nil,
		},
		{
			name:  "DeleteTagFromNonExistingRun",
			ID:    "non-existing-run-ID",
			tagID: "existing-tag-ID",
			error: &api.ErrorResponse{
				Message:    "run 'non-existing-run-ID' not found",
				StatusCode: http.StatusBadRequest,
			},
		},
		// Add more test cases as needed...
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
					"/runs/%s/tags/%s", tt.ID, tt.tagID,
				),
			)
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
