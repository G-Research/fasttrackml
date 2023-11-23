//go:build integration

package run

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"gorm.io/datatypes"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunMetricsTestSuite struct {
	helpers.BaseTestSuite
	namespaceID uint
}

func TestGetRunMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunMetricsTestSuite))
}

func (s *GetRunMetricsTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)
	s.namespaceID = namespace.ID
}

func (s *GetRunMetricsTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test data
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    s.namespaceID,
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
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext, err := s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key1": "key1", "value1": "value1"}`),
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     123.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      1,
		ContextID: common.GetPointer(metricContext.ID),
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext, err = s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key2": "key2", "value2": "value2"}`),
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key1",
		Value:     123.2,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      2,
		ContextID: common.GetPointer(metricContext.ID),
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext, err = s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key3": "key3", "value3": "value3"}`),
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key2",
		Value:     124.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      3,
		ContextID: common.GetPointer(metricContext.ID),
	})
	require.Nil(s.T(), err)

	// create context and attach it to own metric.
	metricContext, err = s.ContextFixtures.CreateContext(context.Background(), &models.Context{
		Json: datatypes.JSON(`{"key4": "key4", "value4": "value4"}`),
	})
	require.Nil(s.T(), err)

	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "key2",
		Value:     124.2,
		Timestamp: 123456789,
		Step:      2,
		IsNan:     false,
		RunID:     run.ID,
		Iter:      4,
		ContextID: common.GetPointer(metricContext.ID),
	})
	require.Nil(s.T(), err)

	// runs tests over test data.
	tests := []struct {
		name             string
		runID            string
		request          request.GetRunMetrics
		expectedResponse response.GetRunMetrics
	}{
		{
			name:  "GetOneRun",
			runID: run.ID,
			request: request.GetRunMetrics{
				{
					Name:    "key1",
					Context: map[string]string{},
				},
				{
					Name:    "key2",
					Context: map[string]string{},
				},
			},
			expectedResponse: response.GetRunMetrics{
				response.RunMetrics{
					Name:    "key1",
					Iters:   []int64{1},
					Values:  []float64{123.1},
					Context: []byte(`{"key1":"key1","value1":"value1"}`),
				},
				response.RunMetrics{
					Name:    "key1",
					Iters:   []int64{2},
					Values:  []float64{123.2},
					Context: []byte(`{"key2":"key2","value2":"value2"}`),
				},
				response.RunMetrics{
					Name:    "key2",
					Iters:   []int64{3},
					Values:  []float64{124.1},
					Context: []byte(`{"key3":"key3","value3":"value3"}`),
				},
				response.RunMetrics{
					Name:    "key2",
					Iters:   []int64{4},
					Values:  []float64{124.2},
					Context: []byte(`{"key4":"key4","value4":"value4"}`),
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
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()
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
