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
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type LogMetricTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestLogMetricTestSuite(t *testing.T) {
	suite.Run(t, new(LogMetricTestSuite))
}

func (s *LogMetricTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *LogMetricTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment := &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
	assert.Nil(s.T(), err)

	run := &models.Run{
		ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
		ExperimentID:   *experiment.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.RunFixtures.CreateRun(context.Background(), run)
	assert.Nil(s.T(), err)

	resp := fiber.Map{}
	err = s.MlflowClient.DoPostRequest(
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
	metric, err := s.MetricFixtures.GetLatestMetricByRunID(context.Background(), run.ID)
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
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			setupDatabase: func() string {
				_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					ID:                  1,
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				assert.Nil(s.T(), err)
				return ""
			},
			cleanDatabase: func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			},
		},
		{
			name: "EmptyOrIncorrectKey",
			request: request.LogMetricRequest{
				RunID: "id",
			},
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			setupDatabase: func() string {
				_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					ID:                  1,
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				assert.Nil(s.T(), err)
				return ""
			},
			cleanDatabase: func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			},
		},
		{
			name: "NotFoundRun",
			request: request.LogMetricRequest{
				Key:       "key1",
				RunID:     "id",
				Timestamp: 123456789,
			},
			error: api.NewResourceDoesNotExistError("unable to find run 'id'"),
			setupDatabase: func() string {
				_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					ID:                  1,
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				assert.Nil(s.T(), err)
				return ""
			},
			cleanDatabase: func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			},
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
				namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					ID:                  1,
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				assert.Nil(s.T(), err)

				experiment := &models.Experiment{
					Name:           uuid.New().String(),
					NamespaceID:    namespace.ID,
					LifecycleStage: models.LifecycleStageActive,
				}
				_, err = s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
				assert.Nil(s.T(), err)

				run := &models.Run{
					ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
					ExperimentID:   *experiment.ID,
					SourceType:     "JOB",
					LifecycleStage: models.LifecycleStageActive,
					Status:         models.StatusRunning,
				}
				run, err = s.RunFixtures.CreateRun(context.Background(), run)
				assert.Nil(s.T(), err)
				return run.ID
			},
			cleanDatabase: func() {
				assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			// if setupDatabase has been provided then configure database with test data.
			if tt.setupDatabase != nil {
				if runID := tt.setupDatabase(); runID != "" {
					tt.request.RunID = runID
				}
			}

			resp := api.ErrorResponse{}
			err := s.MlflowClient.DoPostRequest(
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
