package experiment

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SetExperimentTagTestSuite struct {
	helpers.BaseTestSuite
}

func TestSetExperimentTagTestSuite(t *testing.T) {
	suite.Run(t, &SetExperimentTagTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *SetExperimentTagTestSuite) Test_Ok() {
	// 1. prepare database with test data.
	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           "Test Experiment2",
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	// Set tag on experiment1
	req := request.SetExperimentTagRequest{
		ID:    fmt.Sprintf("%d", *experiment1.ID),
		Key:   "KeyTag1",
		Value: "ValueTag1",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
	)

	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment1.ID,
	)
	s.Require().Nil(err)
	s.True(helpers.CheckTagExists(
		experiment1.Tags, "KeyTag1", "ValueTag1"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'",
	)

	// Update tag on experiment1
	req = request.SetExperimentTagRequest{
		ID:    fmt.Sprintf("%d", *experiment1.ID),
		Key:   "KeyTag1",
		Value: "ValueTag2",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
	)

	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment1.ID,
	)
	s.Require().Nil(err)
	s.True(
		helpers.CheckTagExists(experiment1.Tags, "KeyTag1", "ValueTag2"),
		"Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'",
	)

	// test that setting a tag on 1 experiment1 does not impact another experiment1.
	experiment2, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment2.ID,
	)
	s.Require().Nil(err)
	s.Equal(len(experiment2.Tags), 0)

	// test that setting a tag on different experiments maintain different values across experiments
	req = request.SetExperimentTagRequest{
		ID:    fmt.Sprintf("%d", *experiment2.ID),
		Key:   "KeyTag1",
		Value: "ValueTag3",
	}
	s.Require().Nil(
		s.MlflowClient().WithMethod(
			http.MethodPost,
		).WithRequest(
			req,
		).DoRequest(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
	)
	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment1.ID,
	)
	s.Require().Nil(err)
	s.True(helpers.CheckTagExists(experiment1.Tags, "KeyTag1", "ValueTag2"),
		"Expected 'experiment1.tags' to contain 'KeyTag1' with value 'ValueTag2'",
	)

	experiment2, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), s.DefaultNamespace.ID, *experiment2.ID,
	)
	s.Require().Nil(err)
	s.True(
		helpers.CheckTagExists(experiment2.Tags, "KeyTag1", "ValueTag3"),
		"Expected 'experiment1.tags' to contain 'KeyTag1' with value 'ValueTag3'",
	)
}

func (s *SetExperimentTagTestSuite) Test_Error() {
	testData := []struct {
		name    string
		error   *api.ErrorResponse
		request *request.SetExperimentTagRequest
	}{
		{
			name:  "EmptyIDProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'experiment_id'"),
			request: &request.SetExperimentTagRequest{
				ID: "",
			},
		},
		{
			name:  "EmptyKeyProperty",
			error: api.NewInvalidParameterValueError("Missing value for required parameter 'key'"),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "",
			},
		},
		{
			name: "IncorrectExperimentID",
			error: api.NewBadRequestError(
				`Unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: ` +
					`parsing "incorrect_experiment_id": invalid syntax`,
			),
			request: &request.SetExperimentTagRequest{
				ID:  "incorrect_experiment_id",
				Key: "test_key",
			},
		},
		{
			name: "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(
				`unable to find experiment '1': error getting experiment by id: 1: record not found`,
			),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "test_key",
			},
		},
	}

	for _, tt := range testData {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
		})
	}
}
