//go:build integration

package run

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

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

type LogParamTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	run                *models.Run
}

func TestLogParamTestSuite(t *testing.T) {
	suite.Run(t, new(LogParamTestSuite))
}

func (s *LogParamTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	expName := uuid.New().String()
	exp := &models.Experiment{
		Name:           expName,
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	runID := uuid.New().String()
	run := &models.Run{
		ID:             strings.ReplaceAll(runID, "-", ""),
		ExperimentID:   *exp.ID,
		SourceType:     "JOB",
		LifecycleStage: models.LifecycleStageActive,
		Status:         models.StatusRunning,
	}
	run, err = s.runFixtures.CreateTestRun(context.Background(), run)
	assert.Nil(s.T(), err)
	s.run = run
}

func (s *LogParamTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	req := request.LogParamRequest{
		RunID: s.run.ID,
		Key:   "key1",
		Value: "value1",
	}
	resp := map[string]any{}
	err := s.client.DoPostRequest(
		fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogParameterRoute),
		req,
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Empty(s.T(), resp)
}

func (s *LogParamTestSuite) TestBatch_Ok() {
	tests := []struct {
		name    string
		request *request.LogBatchRequest
	}{
		{
			name: "Batch of one should succeed",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
			},
		},
		{
			name: "Duplicat ekeys with same value should succeed",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key2",
						Value: "value2",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
		},
	}
	var err error
	resp := map[string]any{}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			err = s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)
		})
	}
}

func (s *LogParamTestSuite) TestBatch_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	var testData = []struct {
		name    string
		error   string
		request *request.LogBatchRequest
	}{
		{
			name:  "Missing required field",
			error: "Missing value for required parameter 'run_id'",
			request: &request.LogBatchRequest{
				RunID: "",
				Params: []request.ParamPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
			},
		},
		{
			name:  "Duplicate keys with different values should error",
			error: "error creating params in batch: UNIQUE constraint failed",
			request: &request.LogBatchRequest{
				RunID: s.run.ID,
				Params: []request.ParamPartialRequest{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key1",
						Value: "value2",
					},
				},
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsLogBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Contains(s.T(), resp.Error(), tt.error)
		})
	}
}
    
