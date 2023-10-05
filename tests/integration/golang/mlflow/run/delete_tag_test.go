//go:build integration

package run

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hetiansu5/urlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteRunTagTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestDeleteRunTagTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteRunTagTestSuite))
}

func (s *DeleteRunTagTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *DeleteRunTagTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	// create test experiment.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	// create few tags,.
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag1",
		Value: "value1",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)
	_, err = s.TagFixtures.CreateTag(context.Background(), &models.Tag{
		Key:   "tag2",
		Value: "value2",
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	// make actual call to API.
	query, err := urlquery.Marshal(request.GetRunRequest{
		RunID: run.ID,
	})
	assert.Nil(s.T(), err)

	resp := fiber.Map{}
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf("%s%s?%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteTagRoute, query),
		&request.DeleteRunTagRequest{
			RunID: run.ID,
			Key:   "tag1",
		},
		&resp,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), fiber.Map{}, resp)

	// make sure that we still have one tag connected to Run.
	tags, err := s.TagFixtures.GetByRunID(context.Background(), run.ID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(tags))
	assert.Equal(s.T(), []models.Tag{
		{
			Key:   "tag2",
			RunID: run.ID,
			Value: "value2",
		},
	}, tags)
}

func (s *DeleteRunTagTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	// create test experiment.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    namespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	tests := []struct {
		name    string
		error   *api.ErrorResponse
		request request.DeleteRunTagRequest
	}{
		{
			name:    "EmptyOrIncorrectRunID",
			request: request.DeleteRunTagRequest{},
			error:   api.NewInvalidParameterValueError("Missing value for required parameter 'run_id'"),
		},
		{
			name: "NotFoundRun",
			request: request.DeleteRunTagRequest{
				RunID: "id",
			},
			error: api.NewResourceDoesNotExistError("Unable to find active run 'id'"),
		},
		{
			name: "NotFoundTag",
			request: request.DeleteRunTagRequest{
				Key:   "not_found_tag",
				RunID: run.ID,
			},
			error: api.NewResourceDoesNotExistError(
				`Unable to find tag 'not_found_tag' for run '%s': error getting tag by run id: %s and `+
					`tag key: not_found_tag: record not found`,
				run.ID, run.ID,
			),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			err := s.MlflowClient.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.RunsRoutePrefix, mlflow.RunsDeleteTagRoute),
				tt.request,
				&resp,
			)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
