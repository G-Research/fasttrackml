package run

import (
	"context"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogMetricTestSuite struct {
	helpers.BaseTestSuite
}

func TestLogMetricTestSuite(t *testing.T) {
	suite.Run(t, new(LogMetricTestSuite))
}

func (s *LogMetricTestSuite) Test_Ok() {
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *s.DefaultExperiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	})
	s.Require().Nil(err)

	tests := []struct {
		name           string
		request        *request.LogMetricRequest
		expectedMetric *models.LatestMetric
	}{
		{
			name: "LogMetricWithNormalValue",
			request: &request.LogMetricRequest{
				RunID:     run.ID,
				Key:       "key1",
				Value:     1.1,
				Timestamp: 1234567890,
				Step:      1,
			},
			expectedMetric: &models.LatestMetric{
				Key:       "key1",
				Value:     1.1,
				Timestamp: 1234567890,
				Step:      1,
				IsNan:     false,
				RunID:     run.ID,
				LastIter:  1,
			},
		},
		{
			name: "LogMetricWithNaNValue",
			request: &request.LogMetricRequest{
				RunID:     run.ID,
				Key:       "key1",
				Value:     "NaN",
				Timestamp: 1234567890,
				Step:      1,
			},
			expectedMetric: &models.LatestMetric{
				Key:       "key1",
				Value:     0,
				Timestamp: 1234567890,
				Step:      1,
				IsNan:     true,
				RunID:     run.ID,
				LastIter:  2,
			},
		},
		{
			name: "LogMetricPositiveInfinityValue",
			request: &request.LogMetricRequest{
				RunID:     run.ID,
				Key:       "key1",
				Value:     "Infinity",
				Timestamp: 1234567890,
				Step:      1,
			},
			expectedMetric: &models.LatestMetric{
				Key:       "key1",
				Value:     math.MaxFloat64,
				Timestamp: 1234567890,
				Step:      1,
				RunID:     run.ID,
				LastIter:  3,
			},
		},
		{
			name: "LogMetricNegativeInfinityValue",
			request: &request.LogMetricRequest{
				RunID:     run.ID,
				Key:       "key1",
				Value:     "-Infinity",
				Timestamp: 1234567890,
				Step:      1,
			},
			expectedMetric: &models.LatestMetric{
				Key:       "key1",
				Value:     -math.MaxFloat64,
				Timestamp: 1234567890,
				Step:      1,
				RunID:     run.ID,
				LastIter:  4,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := fiber.Map{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute,
				),
			)
			s.Empty(resp)

			// makes user that records has been created correctly in database.
			metric, err := s.MetricFixtures.GetLatestMetricByRunID(context.Background(), run.ID)
			s.Require().Nil(err)
			s.Equal(tt.expectedMetric, metric)
		})
	}
}

func (s *LogMetricTestSuite) Test_Error() {
	tests := []struct {
		name          string
		error         *api.ErrorResponse
		request       request.LogMetricRequest
		setupDatabase func() string
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.LogMetricRequest{},
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
		},
		{
			name: "EmptyOrIncorrectKey",
			request: request.LogMetricRequest{
				RunID: "id",
			},
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
		},
		{
			name: "NotFoundRun",
			request: request.LogMetricRequest{
				Key:       "key1",
				RunID:     "id",
				Timestamp: 123456789,
			},
			error: api.NewResourceDoesNotExistError("unable to find run 'id'"),
		},
		{
			name: "InvalidMetricValue",
			request: request.LogMetricRequest{
				Key:       "key1",
				Value:     "incorrect_value",
				Timestamp: 123456789,
			},
			error: api.NewInvalidParameterValueError(`invalid metric value 'incorrect_value'`),
			setupDatabase: func() string {
				run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
					ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
					ExperimentID:   *s.DefaultExperiment.ID,
					SourceType:     "JOB",
					LifecycleStage: models.LifecycleStageActive,
					Status:         models.StatusRunning,
				})
				s.Require().Nil(err)
				return run.ID
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// if setupDatabase has been provided then configure database with test data.
			if tt.setupDatabase != nil {
				if runID := tt.setupDatabase(); runID != "" {
					tt.request.RunID = runID
				}
			}

			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
