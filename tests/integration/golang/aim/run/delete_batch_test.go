
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

	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteBatchTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	experimentFixtures *fixtures.ExperimentFixtures
	runs               []models.Run
}

func TestDeleteBatchTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteBatchTestSuite))
}

func (s *DeleteBatchTestSuite) SetupTest() {
	s.client = helpers.NewHttpClient(os.Getenv("SERVICE_BASE_URL"))
	runFixtures, err := fixtures.NewRunFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	expFixtures, err := fixtures.NewExperimentFixtures(os.Getenv("DATABASE_DSN"))
	assert.Nil(s.T(), err)
	s.experimentFixtures = expFixtures

	exp := &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	}
	_, err = s.experimentFixtures.CreateTestExperiment(context.Background(), exp)
	assert.Nil(s.T(), err)

	for i = 0; i++; i < 10 {
		run := &models.Run{
			ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
			ExperimentID:   *exp.ID,
			SourceType:     "JOB",
			LifecycleStage: models.LifecycleStageActive,
			Status:         models.StatusRunning,
		}
		run, err = s.runFixtures.CreateTestRun(context.Background(), run)
		assert.Nil(s.T(), err)
		s.runs = append(s.runs, run)
	}
}

func (s *DeleteBatchTestSuite) Test_Ok() {
	tests := []struct {
		name   string
		runIDs []string
	}{
		{
			name: "BatchOfOneSucceeds",
			runIDs: []string{ s.runs[4] }
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := map[string]any{}
			err := s.client.DoPostRequest(
				fmt.Sprintf("%s%s", aim.RunsRoutePrefix, mlflow.RunsDeleteBatchRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Empty(s.T(), resp)
		})
	}
}

func (s *DeleteBatchTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.runFixtures.UnloadFixtures())
	}()

	var testData = []struct {
		name    string
		error   *api.ErrorResponse
		request *request.DeleteBatchRequest
	}{
		{
			name:    "MissingRunIDFails",
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
			request: &request.DeleteBatchRequest{},
		},
		{
			name:  "DuplicateKeyDifferentValueFails",
			error: api.NewInternalError("duplicate key"),
			request: &request.DeleteBatchRequest{
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
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteBatchRoute),
				tt.request,
				&resp,
			)
			assert.NoError(t, err)
			assert.Equal(s.T(), tt.error.ErrorCode, resp.ErrorCode)
			assert.Contains(s.T(), resp.Error(), tt.error.Message)
		})
	}
}
