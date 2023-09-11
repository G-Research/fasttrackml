//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SetExperimentTagTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestSetExperimentTagTestSuite(t *testing.T) {
	suite.Run(t, new(SetExperimentTagTestSuite))
}

func (s *SetExperimentTagTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *SetExperimentTagTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()
	// 1. prepare database with test data.
	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  0,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment1, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:        "Test Experiment",
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
	assert.Nil(s.T(), err)
	experiment2, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:        "Test Experiment2",
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
	assert.Nil(s.T(), err)

	// Set tag on experiment1
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
		request.SetExperimentTagRequest{
			ID:    fmt.Sprintf("%d", *experiment1.ID),
			Key:   "KeyTag1",
			Value: "ValueTag1",
		},
		&struct{}{},
	)
	assert.Nil(s.T(), err)
	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment1.ID,
	)
	assert.Nil(s.T(), err)
	assert.True(s.T(), helpers.CheckTagExists(
		experiment1.Tags, "KeyTag1", "ValueTag1"), "Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'",
	)

	// Update tag on experiment1
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
		request.SetExperimentTagRequest{
			ID:    fmt.Sprintf("%d", *experiment1.ID),
			Key:   "KeyTag1",
			Value: "ValueTag2",
		},
		&struct{}{},
	)
	assert.Nil(s.T(), err)
	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment1.ID,
	)
	assert.Nil(s.T(), err)
	assert.True(
		s.T(),
		helpers.CheckTagExists(experiment1.Tags, "KeyTag1", "ValueTag2"),
		"Expected 'experiment.tags' to contain 'KeyTag1' with value 'ValueTag1'",
	)

	// test that setting a tag on 1 experiment1 does not impact another experiment1.
	experiment2, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment2.ID,
	)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), len(experiment2.Tags), 0)

	// test that setting a tag on different experiments maintain different values across experiments
	err = s.MlflowClient.DoPostRequest(
		fmt.Sprintf(
			"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag,
		),
		request.SetExperimentTagRequest{
			ID:    fmt.Sprintf("%d", *experiment2.ID),
			Key:   "KeyTag1",
			Value: "ValueTag3",
		},
		&struct{}{},
	)
	assert.Nil(s.T(), err)
	experiment1, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment1.ID,
	)
	assert.Nil(s.T(), err)
	assert.True(
		s.T(), helpers.CheckTagExists(experiment1.Tags, "KeyTag1", "ValueTag2"),
		"Expected 'experiment1.tags' to contain 'KeyTag1' with value 'ValueTag2'",
	)

	experiment2, err = s.ExperimentFixtures.GetByNamespaceIDAndExperimentID(
		context.Background(), namespace.ID, *experiment2.ID,
	)
	assert.Nil(s.T(), err)
	assert.True(
		s.T(),
		helpers.CheckTagExists(experiment2.Tags, "KeyTag1", "ValueTag3"),
		"Expected 'experiment1.tags' to contain 'KeyTag1' with value 'ValueTag3'",
	)
}

func (s *SetExperimentTagTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  0,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

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
				`Unable to parse experiment id 'incorrect_experiment_id': strconv.ParseInt: parsing "incorrect_experiment_id": invalid syntax`,
			),
			request: &request.SetExperimentTagRequest{
				ID:  "incorrect_experiment_id",
				Key: "test_key",
			},
		},
		{
			name:  "NotFoundExperiment",
			error: api.NewResourceDoesNotExistError(`unable to find experiment '1': error getting experiment by id: 1: record not found`),
			request: &request.SetExperimentTagRequest{
				ID:  "1",
				Key: "test_key",
			},
		},
	}

	for _, tt := range testData {
		s.T().Run(tt.name, func(t *testing.T) {
			resp := api.ErrorResponse{}
			err := s.MlflowClient.DoPostRequest(
				fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsSetExperimentTag),
				tt.request,
				&resp,
			)
			assert.Nil(t, err)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
		})
	}
}
