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

type AddRunTagTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
	tag *aimModels.SharedTag
}

func TestAddRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(AddRunTagTestSuite))
}

func (s *AddRunTagTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
	s.tag, err = s.SharedTagFixtures.CreateTag(context.Background(), "tag", s.DefaultExperiment.NamespaceID)
	s.Require().Nil(err)
}

func (s *AddRunTagTestSuite) Test_Ok() {
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
			tt.request.RunID = tt.runID
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

func (s *AddRunTagTestSuite) Test_Error() {
	tests := []struct {
		name    string
		runID   string
		request request.AddRunTagRequest
	}{
		{
			name:    "AddNonExistingTag",
			runID:   s.run.ID,
			request: request.AddRunTagRequest{TagName: "tag-not"},
		},
		{
			name:    "AddNonExistingRun",
			runID:   uuid.NewString(),
			request: request.AddRunTagRequest{TagName: "tag"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s/tags/new", tt.runID,
				),
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
	}
}
