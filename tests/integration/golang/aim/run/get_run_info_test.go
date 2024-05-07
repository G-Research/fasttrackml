package run

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunInfoTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
}

func TestGetRunInfoTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunInfoTestSuite))
}

func (s *GetRunInfoTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
}

func (s *GetRunInfoTestSuite) Test_Ok() {
	tests := []struct {
		name  string
		runID string
	}{
		{
			name:  "GetOneRun",
			runID: s.run.ID,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.GetRunInfoResponse
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/info", tt.runID),
			)
			s.Equal(s.run.Name, resp.Props.Name)
			s.Equal(fmt.Sprintf("%v", s.run.ExperimentID), resp.Props.Experiment.ID)
			s.Equal(s.run.Experiment.Name, resp.Props.Experiment.Name)
			s.Equal(float64(s.run.StartTime.Int64)/1000, resp.Props.CreationTime)
			s.Equal(float64(s.run.EndTime.Int64)/1000, resp.Props.EndTime)
			expectedTags := make(map[string]interface{}, len(s.run.Tags))
			for _, tag := range s.run.Tags {
				expectedTags[tag.Key] = tag.Value
			}
			s.Equal(expectedTags, resp.Params["tags"])
		})
	}
}

func (s *GetRunInfoTestSuite) Test_Error() {
	tests := []struct {
		name  string
		runID string
		error *api.ErrorResponse
	}{
		{
			name:  "GetNonexistentRun",
			runID: "not-found-id",
			error: &api.ErrorResponse{
				Message:    "run 'not-found-id' not found",
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/info", tt.runID))
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
