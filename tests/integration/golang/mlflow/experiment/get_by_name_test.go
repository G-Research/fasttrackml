//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentByNameTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentByNameTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentByNameTestSuite))
}

func (s *GetExperimentByNameTestSuite) Test_Ok() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	// 1. prepare database with test data.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
		Tags: []models.ExperimentTag{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		NamespaceID: namespace.ID,
		CreationTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	require.Nil(s.T(), err)

	// 2. make actual API call.
	request := request.GetExperimentRequest{
		Name: experiment.Name,
	}

	resp := response.GetExperimentResponse{}
	require.Nil(
		s.T(),
		s.MlflowClient().WithQuery(
			request,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetByNameRoute,
		),
	)

	// 3. check actual API response.
	assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.Experiment.ID)
	assert.Equal(s.T(), experiment.Name, resp.Experiment.Name)
	assert.Equal(s.T(), string(experiment.LifecycleStage), resp.Experiment.LifecycleStage)
	assert.Equal(s.T(), experiment.ArtifactLocation, resp.Experiment.ArtifactLocation)
	assert.Equal(s.T(), []models.ExperimentTag{
		{
			Key:          "key1",
			Value:        "value1",
			ExperimentID: *experiment.ID,
		},
	}, experiment.Tags)
}

func (s *GetExperimentByNameTestSuite) Test_Error() {
	defer func() {
		require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	require.Nil(s.T(), err)

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetExperimentRequest
	}{
		{
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(`unable to find experiment 'incorrect_experiment_name'`),
			request: request.GetExperimentRequest{
				Name: "incorrect_experiment_name",
			},
		},
		{
			name:  "EmptyExperimentName",
			error: api.NewInvalidParameterValueError(`Missing value for required parameter 'experiment_name'`),
			request: request.GetExperimentRequest{
				Name: "",
			},
		},
	}

	for _, tt := range testData {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetByNameRoute,
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
