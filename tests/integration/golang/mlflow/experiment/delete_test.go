//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestDeleteExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteExperimentTestSuite))
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
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

	// check that the experiment lifecycle is active
	assert.Equal(s.T(), models.LifecycleStageActive, experiment.LifecycleStage)

	// 2. make actual API call.
	req := request.DeleteExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	}
	resp := fiber.Map{}
	require.Nil(
		s.T(),
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).WithResponse(
			&resp,
		).DoRequest(
			fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsDeleteRoute),
		),
	)

	// 3. check actual API response.
	exp, err := s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment.ID,
	)
	require.Nil(s.T(), err)
	assert.Equal(s.T(), models.LifecycleStageDeleted, exp.LifecycleStage)
}

func (s *DeleteExperimentTestSuite) Test_Error() {
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
		request *request.DeleteExperimentRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.DeleteExperimentRequest{
				ID: "",
			},
		},
		{
			name: "InvalidIDFormat",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'invalid_id': strconv.ParseInt: parsing "invalid_id": invalid syntax`,
			),
			request: &request.DeleteExperimentRequest{
				ID: "invalid_id",
			},
		},
		{
			name: "ExperimentNotFound",
			error: api.NewResourceDoesNotExistError(
				"unable to find experiment '123': error getting experiment by id: 123: record not found",
			),
			request: &request.DeleteExperimentRequest{
				ID: "123",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsDeleteRoute),
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
