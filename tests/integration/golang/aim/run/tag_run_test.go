package run

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type RunTagTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
	tag *models.SharedTag
}

func (s *RunTagTestSuite) SetupSuite() {
	// Setup code...
}

func (s *RunTagTestSuite) TearDownSuite() {
	// Teardown code...
}

func (s *RunTagTestSuite) TestAddRunTag() {
	tests := []struct {
		name    string
		runID   string
		request request.AddRunTagRequest
	}{
		{
			name:    "AddTagToExistingRun",
			runID:   "existing-run-ID",
			request: request.AddRunTagRequest{TagName: "tag"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp 
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
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
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

func TestRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(RunTagTestSuite))
}
