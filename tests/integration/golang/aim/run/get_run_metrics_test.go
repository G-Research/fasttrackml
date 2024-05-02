package run

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/dao/types"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunMetricsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetRunMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunMetricsTestSuite))
}

func (s *GetRunMetricsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()
}

func (s *GetRunMetricsTestSuite) Test_Ok() {
	// create test data
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             "id",
		Name:           "TestRun",
		Status:         models.StatusScheduled,
		StartTime:      sql.NullInt64{Int64: 123456789, Valid: true},
		EndTime:        sql.NullInt64{Int64: 123456789, Valid: true},
		SourceType:     "JOB",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     123.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      1,
		Context: models.Context{
			Json: types.JSONB(`{"key1":"key1","value1":"value1"}`),
		},
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     123.2,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      2,
		Context: models.Context{
			Json: types.JSONB(`{"key2":"key2","value2":"value2"}`),
		},
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key2",
		Value:     124.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      3,
		Context: models.Context{
			Json: types.JSONB(`{"key3":"key3","value3":"value3"}`),
		},
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key2",
		Value:     124.2,
		Timestamp: 123456789,
		Step:      2,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      4,
		Context: models.Context{
			Json: types.JSONB(`{"key4":"key4","value4":"value4"}`),
		},
	})
	s.Require().Nil(err)

	// runs tests over test data.
	tests := []struct {
		name             string
		runID            string
		request          request.GetRunMetricsRequest
		expectedResponse []response.GetRunMetricsResponse
	}{
		{
			name:  "GetOneRun",
			runID: run.ID,
			request: request.GetRunMetricsRequest{
				{
					Name: "key1",
				},
				{
					Name: "key2",
				},
			},
			expectedResponse: []response.GetRunMetricsResponse{
				{
					Name:    "key1",
					Iters:   []int{1},
					Values:  []*float64{common.GetPointer(123.1)},
					Context: json.RawMessage(`{"key1":"key1","value1":"value1"}`),
				},
				{
					Name:    "key1",
					Iters:   []int{2},
					Values:  []*float64{common.GetPointer(123.2)},
					Context: json.RawMessage(`{"key2":"key2","value2":"value2"}`),
				},
				{
					Name:    "key2",
					Iters:   []int{3},
					Values:  []*float64{common.GetPointer(124.1)},
					Context: json.RawMessage(`{"key3":"key3","value3":"value3"}`),
				},
				{
					Name:    "key2",
					Iters:   []int{4},
					Values:  []*float64{common.GetPointer(124.2)},
					Context: json.RawMessage(`{"key4":"key4","value4":"value4"}`),
				},
			},
		},
		{
			name:  "GetOneRunWithContextCase1",
			runID: run.ID,
			request: request.GetRunMetricsRequest{
				{
					Name: "key1",
					Context: map[string]string{
						"key1":   "key1",
						"value1": "value1",
					},
				},
				{
					Name: "key1",
					Context: map[string]string{
						"key2":   "key2",
						"value2": "value2",
					},
				},
				{
					Name: "key2",
					Context: map[string]string{
						"key3":   "key3",
						"value3": "value3",
					},
				},
				{
					Name: "key2",
					Context: map[string]string{
						"key4":   "key4",
						"value4": "value4",
					},
				},
			},
			expectedResponse: []response.GetRunMetricsResponse{
				{
					Name:    "key1",
					Iters:   []int{1},
					Values:  []*float64{common.GetPointer(123.1)},
					Context: json.RawMessage(`{"key1":"key1","value1":"value1"}`),
				},
				{
					Name:    "key1",
					Iters:   []int{2},
					Values:  []*float64{common.GetPointer(123.2)},
					Context: json.RawMessage(`{"key2":"key2","value2":"value2"}`),
				},
				{
					Name:    "key2",
					Iters:   []int{3},
					Values:  []*float64{common.GetPointer(124.1)},
					Context: json.RawMessage(`{"key3":"key3","value3":"value3"}`),
				},
				{
					Name:    "key2",
					Iters:   []int{4},
					Values:  []*float64{common.GetPointer(124.2)},
					Context: json.RawMessage(`{"key4":"key4","value4":"value4"}`),
				},
			},
		},
		{
			name:  "GetOneRunWithContextCase2",
			runID: run.ID,
			request: request.GetRunMetricsRequest{
				{
					Name: "key1",
					Context: map[string]string{
						"key1":   "key1",
						"value1": "value1",
					},
				},
				{
					Name: "key2",
					Context: map[string]string{
						"key3":   "key3",
						"value3": "value3",
					},
				},
			},
			expectedResponse: []response.GetRunMetricsResponse{
				{
					Name:    "key1",
					Iters:   []int{1},
					Values:  []*float64{common.GetPointer(123.1)},
					Context: json.RawMessage(`{"key1":"key1","value1":"value1"}`),
				},
				{
					Name:    "key2",
					Iters:   []int{3},
					Values:  []*float64{common.GetPointer(124.1)},
					Context: json.RawMessage(`{"key3":"key3","value3":"value3"}`),
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp []response.GetRunMetricsResponse
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
		error *api.ErrorResponse
	}{
		{
			name:  "GetNonexistentRun",
			runID: "9facdfb7-d502-4172-9325-8df6f4dbcbc0",
			error: &api.ErrorResponse{
				Message:    "Not Found",
				StatusCode: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(
				s.AIMClient().WithResponse(&resp).DoRequest("/runs/%s/metric/get-batch", tt.runID),
			)
			s.Equal(tt.error.Message, resp.Message)
			s.Equal(tt.error.StatusCode, resp.StatusCode)
		})
	}
}
