//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(GetExperimentTestSuite))
}

func (s *GetExperimentTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	// 1. prepare database with test data.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

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
	s.Require().Nil(err)

	// 2. make actual API call.
	request := request.GetExperimentRequest{
		ID: fmt.Sprintf("%d", *experiment.ID),
	}

	resp := response.GetExperimentResponse{}
	s.Require().Nil(
		s.MlflowClient().WithQuery(
			request,
		).WithResponse(
			&resp,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute,
		),
	)
	// 3. check actual API response.
	s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.Experiment.ID)
	s.Equal(experiment.Name, resp.Experiment.Name)
	s.Equal(string(experiment.LifecycleStage), resp.Experiment.LifecycleStage)
	s.Equal(experiment.ArtifactLocation, resp.Experiment.ArtifactLocation)
	s.Equal([]models.ExperimentTag{
		{
			Key:          "key1",
			Value:        "value1",
			ExperimentID: *experiment.ID,
		},
	}, experiment.Tags)
}

func (s *GetExperimentTestSuite) Test_Error() {
	defer func() {
		s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	s.Require().Nil(err)

	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request request.GetExperimentRequest
	}{
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: ` +
					`parsing "incorrect_experiment_id": invalid syntax`,
			),
			request: request.GetExperimentRequest{
				ID: "incorrect_experiment_id",
			},
		},
		{
			name: "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(
				`unable to find experiment '1': error getting experiment by id: 1: record not found`,
			),
			request: request.GetExperimentRequest{
				ID: "1",
			},
		},
	}

	for _, tt := range testData {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithQuery(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
