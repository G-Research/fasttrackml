//go:build integration

package run

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogMetricTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestLogMetricTestSuite(t *testing.T) {
	suite.Run(t, new(LogMetricTestSuite))
}

func (s *LogMetricTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures
}

func (s *LogMetricTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	experiment := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err := s.experimentFixtures.CreateExperiment(context.Background(), experiment)
	assert.Nil(s.T(), err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.runFixtures.CreateRun(context.Background(), run)
	assert.Nil(s.T(), err)

	resp := fiber.Map{}
	err = s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute),
		request.LogMetricRequest{
			RunID:     run.ID,
			Key:       "key1",
			Value:     1.1,
			Timestamp: 1234567890,
			Step:      1,
		},
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), resp)

	// makes user that records has been created correctly in database.
	metric, err := s.metricFixtures.GetLatestMetricByRunID(context.Background(), run.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), &models.LatestMetric{
		Key:       "key1",
		Value:     1.1,
		Timestamp: 1234567890,
		Step:      1,
		IsNan:     false,
		RunID:     run.ID,
		LastIter:  1,
	}, metric)
}

func (s *LogMetricTestSuite) Test_Error() {
	tests := []struct {
		name          string
		error         *api.ErrorResponse
		request       request.LogMetricRequest
		setupDatabase func() string
		cleanDatabase func()
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.LogMetricRequest{},
			error: api.NewInvalidParameterValueError(
				"Missing value for required parameter 'run_id'",
			),
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
			error: api.NewResourceDoesNotExistError(
				"unable to find run 'id': error getting 'run' entity by id: id: record not found",
			),
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
				experiment := &models.Experiment{
					Name:           uuid.New().String(),
					LifecycleStage: models.LifecycleStageActive,
				}
				_, err := s.experimentFixtures.CreateExperiment(context.Background(), experiment)
				assert.Nil(s.T(), err)

				run := &models.Run{
					ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
					ExperimentID:   *experiment.ID,
					SourceType:     "JOB",
					LifecycleStage: models.LifecycleStageActive,
					Status:         models.StatusRunning,
				}
				run, err = s.runFixtures.CreateRun(context.Background(), run)
				assert.Nil(s.T(), err)
				return run.ID
			},
			cleanDatabase: func() {
				assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			// if setupDatabase has been provided then configure database with test data.
			if tt.setupDatabase != nil {
				tt.request.RunID = tt.setupDatabase()
			}

			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogMetricRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())

			// if cleanDatabase has been provided then clean database after the test.
			if tt.cleanDatabase != nil {
				tt.cleanDatabase()
			}
		})
	}
}
