//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunTestSuite struct {
	suite.Suite
	client             *helpers.HttpClient
	runFixtures        *fixtures.RunFixtures
	tagFixtures        *fixtures.TagFixtures
	paramFixtures      *fixtures.ParamFixtures
	metricFixtures     *fixtures.MetricFixtures
	experimentFixtures *fixtures.ExperimentFixtures
}

func TestGetRunTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunTestSuite))
}

func (s *GetRunTestSuite) SetupTest() {
	s.client = helpers.NewMlflowApiClient(helpers.GetServiceUri())
	runFixtures, err := fixtures.NewRunFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.runFixtures = runFixtures
	experimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.experimentFixtures = experimentFixtures
	tagFixtures, err := fixtures.NewTagFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.tagFixtures = tagFixtures
	metricFixtures, err := fixtures.NewMetricFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.metricFixtures = metricFixtures
	paramFixtures, err := fixtures.NewParamFixtures(helpers.GetDatabaseUri())
	assert.Nil(s.T(), err)
	s.paramFixtures = paramFixtures
}

func (s *GetRunTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	// create test experiment.
	experiment, err := s.experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// create test run for the experiment
	run, err := s.runFixtures.CreateRun(context.Background(), &models.Run{
		ID:     strings.ReplaceAll(uuid.New().String(), "-", ""),
		Name:   "TestRun",
		Status: models.StatusRunning,
		StartTime: sql.NullInt64{
			Int64: 1234567890,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 1234567899,
			Valid: true,
		},
		SourceType:     "JOB",
		ArtifactURI:    "artifact_uri",
		ExperimentID:   *experiment.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	// create tags, metrics, params.
	_, err = s.tagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	_, err = s.metricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "metric1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
	})
	assert.Nil(s.T(), err)

	_, err = s.paramFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	query, err := urlquery.Marshal(request.GetRunRequest{
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	resp := response.GetRunResponse{}
	err = s.client.DoGetRequest(
		fmt.Sprintf("%s%s?%s", mlflow.RunsRoutePrefix, mlflow.RunsGetTagRoute, query),
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.NotEmpty(s.T(), resp.Run.Info.ID)
	assert.NotEmpty(s.T(), resp.Run.Info.UUID)
	assert.Equal(s.T(), "TestRun", resp.Run.Info.Name)
	assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.Run.Info.ExperimentID)
	assert.Equal(s.T(), int64(1234567890), resp.Run.Info.StartTime)
	assert.Equal(s.T(), int64(1234567899), resp.Run.Info.EndTime)
	assert.Equal(s.T(), string(models.StatusRunning), resp.Run.Info.Status)
	assert.Equal(s.T(), "artifact_uri", resp.Run.Info.ArtifactURI)
	assert.Equal(s.T(), string(models.LifecycleStageActive), resp.Run.Info.LifecycleStage)
	assert.Equal(s.T(), []response.RunTagPartialResponse{
		{
			Key:   "tag1",
			Value: "value1",
		},
	}, resp.Run.Data.Tags)
	assert.Equal(s.T(), []response.RunMetricPartialResponse{
		{
			Key:       "metric1",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, resp.Run.Data.Metrics)
	assert.Equal(s.T(), []response.RunParamPartialResponse{
		{
			Key:   "param1",
			Value: "value1",
		},
	}, resp.Run.Data.Params)
}

func (s *GetRunTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.experimentFixtures.UnloadFixtures())
	}()

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetRunRequest
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.GetRunRequest{},
			error: api.NewInvalidParameterValueError(
				"Missing value for required parameter 'run_id'",
			),
		},
		{
			name: "NotFoundRun",
			request: request.GetRunRequest{
				RunID: "id",
			},
			error: api.NewResourceDoesNotExistError(
				"unable to find run 'id': error getting `run` entity by id: id: record not found",
			),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			query, err := urlquery.Marshal(tt.request)
			assert.Nil(s.T(), err)

			resp := api.ErrorResponse{}
			err = s.client.DoGetRequest(
				fmt.Sprintf("%s%s?%s", mlflow.RunsRoutePrefix, mlflow.RunsGetTagRoute, query),
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
