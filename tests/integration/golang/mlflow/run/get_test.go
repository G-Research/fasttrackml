//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetRunTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetRunTestSuite(t *testing.T) {
	suite.Run(t, new(GetRunTestSuite))
}

func (s *GetRunTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// create test run for the experiment
	run, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
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
	s.Require().Nil(err)

	// create tags, metrics, params.
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	s.Require().Nil(err)

	_, err = s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "metric1",
		Value:     1.1,
		Timestamp: 1234567890,
		RunID:     run.ID,
		Step:      1,
		IsNan:     false,
	})
	s.Require().Nil(err)

	_, err = s.ParamFixtures.CreateParam(context.Background(), &models.Param{
		Key:   "param1",
		Value: "value1",
		RunID: run.ID,
	})
	s.Require().Nil(err)

	query := request.GetRunRequest{
		RunID: run.ID,
	}

	resp := response.GetRunResponse{}
	s.Require().Nil(
		s.MlflowClient().WithQuery(
			query,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute,
		),
	)

	s.NotEmpty(resp.Run.Info.ID)
	s.NotEmpty(resp.Run.Info.UUID)
	s.Equal("TestRun", resp.Run.Info.Name)
	s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.Run.Info.ExperimentID)
	s.Equal(int64(1234567890), resp.Run.Info.StartTime)
	s.Equal(int64(1234567899), resp.Run.Info.EndTime)
	s.Equal(string(models.StatusRunning), resp.Run.Info.Status)
	s.Equal("artifact_uri", resp.Run.Info.ArtifactURI)
	s.Equal(string(models.LifecycleStageActive), resp.Run.Info.LifecycleStage)
	s.Equal([]response.RunTagPartialResponse{
		{
			Key:   "tag1",
			Value: "value1",
		},
	}, resp.Run.Data.Tags)
	s.Equal([]response.RunMetricPartialResponse{
		{
			Key:       "metric1",
			Step:      1,
			Value:     1.1,
			Timestamp: 1234567890,
		},
	}, resp.Run.Data.Metrics)
	s.Equal([]response.RunParamPartialResponse{
		{
			Key:   "param1",
			Value: "value1",
		},
	}, resp.Run.Data.Params)
}

func (s *GetRunTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
			error: api.NewResourceDoesNotExistError("unable to find run 'id'"),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.RunsRoutePrefix, mlflow.RunsGetRoute,
				),
			)
			s.Require().Nil(err)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
