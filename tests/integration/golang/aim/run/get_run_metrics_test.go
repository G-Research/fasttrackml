package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunMetricsTestSuite struct {
	helpers.BaseTestSuite
	run *models.Run
}

func TestGetRunMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunMetricsTestSuite))
}

func (s *GetRunMetricsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	var err error
	s.run, err = s.RunFixtures.CreateExampleRun(context.Background(), s.DefaultExperiment)
	s.Require().Nil(err)
}

func (s *GetRunMetricsTestSuite) Test_Ok() {
	tests := []struct {
		name             string
		runID            string
		request          request.GetRunMetrics
		expectedResponse response.GetRunMetrics
	}{
		{
			name:  "GetOneRun",
			runID: s.run.ID,
			request: request.GetRunMetrics{
				{
					Context: map[string]string{},
					Name:    "key1",
				},
				{
					Context: map[string]string{},
					Name:    "key2",
				},
			},
			expectedResponse: response.GetRunMetrics{
				response.RunMetrics{
					Name:    "key1",
					Context: map[string]interface{}{},
					Values:  []float64{124.1, 125.1},
					Iters:   []int64{1, 2},
				},
				response.RunMetrics{
					Name:    "key2",
					Context: map[string]interface{}{},
					Values:  []float64{124.1, 125.1},
					Iters:   []int64{1, 2},
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.GetRunMetrics
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"/runs/%s/metric/get-batch", tt.runID,
				),
			)
			s.ElementsMatch(tt.expectedResponse, resp)
		})
	}
}

func (s *GetRunMetricsTestSuite) Test_Error() {
	tests := []struct {
		name  string
		runID string
		error string
	}{
		{
			name:  "GetNonexistentRun",
			runID: uuid.NewString(),
			error: "Not Found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp response.Error
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/metric/get-batch", tt.runID),
			)
			s.Equal(tt.error, resp.Message)
		})
	}
}
