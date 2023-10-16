//go:build integration

package experiment

import (
	"context"
	"database/sql"
	"net/http"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type DeleteExperimentTestSuite struct {
	suite.Suite
	helpers.BaseTestSuite
}

func TestDeleteExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteExperimentTestSuite))
}

func (s *DeleteExperimentTestSuite) SetupTest() {
	s.BaseTestSuite.SetupTest(s.T())
}

func (s *DeleteExperimentTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
		Tags: []models.ExperimentTag{
			{
				Key:   "key1",
				Value: "value1",
			},
		},
		CreationTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		NamespaceID: namespace.ID,
		LastUpdateTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage:   models.LifecycleStageActive,
		ArtifactLocation: "/artifact/location",
	})
	assert.Nil(s.T(), err)

	experiments, err := s.ExperimentFixtures.GetTestExperiments(context.Background())
	assert.Nil(s.T(), err)
	length := len(experiments)

	var resp response.DeleteExperiment
	assert.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodDelete,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d", *experiment.ID,
		),
	assert.Nil(
		s.T(),
		s.AIMClient.WithMethod(
			http.MethodDelete,
		).WithResponse(
			&resp,
		).DoRequest(
			"/experiments/%d", *experiment.ID,
		),
	)

	remainingExperiments, err := s.ExperimentFixtures.GetTestExperiments(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), length-1, len(remainingExperiments))
}

func (s *DeleteExperimentTestSuite) Test_Error() {
	defer func() {
		assert.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
	}()

	_, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
		ID:                  1,
		Code:                "default",
		DefaultExperimentID: common.GetPointer(int32(0)),
	})
	assert.Nil(s.T(), err)

	tests := []struct {
		name  string
		ID    string
		error string
	}{
		{
			ID:    "123",
			name:  "DeleteWithUnknownIDFails",
			error: "Not Found",
		},
		{
			name: "DeleteIncorrectExperimentID",
			error: `: unable to parse experiment id "incorrect_experiment_id": strconv.ParseInt:` +
				` parsing "incorrect_experiment_id": invalid syntax`,
			ID: "incorrect_experiment_id",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			var resp api.ErrorResponse
			assert.Nil(
				s.T(),
				s.AIMClient.WithMethod(
					http.MethodDelete,
				).WithResponse(
					&resp,
				).DoRequest(
					"/experiments/%s", tt.ID,
				),
			)
			assert.Contains(s.T(), resp.Error(), "Not Found")
			assert.NoError(s.T(), err)
		})
	}
}
