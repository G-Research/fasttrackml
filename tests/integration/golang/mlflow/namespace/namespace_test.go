//go:build integration

package namespace

import (
	"context"
	"fmt"
	"testing"

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

type NamespaceTestSuite struct {
	helpers.BaseTestSuite
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}

func (s *NamespaceTestSuite) TearDownTest() {
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}

func (s *NamespaceTestSuite) Test_Ok() {
	tests := []struct {
		name      string
		setup     func() *models.Experiment
		namespace string
	}{
		{
			name: "TestCustomNamespaces",
			setup: func() *models.Experiment {
				namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					Code:                "newly-created-namespace",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				require.Nil(s.T(), err)
				experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
					Name:             "Test Experiment",
					NamespaceID:      namespace.ID,
					LifecycleStage:   models.LifecycleStageActive,
					ArtifactLocation: "/artifact/location",
				})
				require.Nil(s.T(), err)
				return experiment
			},
			namespace: "newly-created-namespace",
		},
		{
			name: "TestExplicitDefaultAndCustomNamespaces",
			setup: func() *models.Experiment {
				namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				require.Nil(s.T(), err)
				experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
					Name:             "Test Experiment",
					NamespaceID:      namespace.ID,
					LifecycleStage:   models.LifecycleStageActive,
					ArtifactLocation: "/artifact/location",
				})
				require.Nil(s.T(), err)
				return experiment
			},
			namespace: "default",
		},
		{
			name: "TestImplicitDefaultAndCustomNamespaces",
			setup: func() *models.Experiment {
				namespace, err := s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
					Code:                "default",
					DefaultExperimentID: common.GetPointer(int32(0)),
				})
				require.Nil(s.T(), err)
				experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
					Name:             "Test Experiment",
					NamespaceID:      namespace.ID,
					LifecycleStage:   models.LifecycleStageActive,
					ArtifactLocation: "/artifact/location",
				})
				require.Nil(s.T(), err)
				return experiment
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			defer require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
			experiment := tt.setup()
			resp := response.GetExperimentResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithNamespace(
					tt.namespace,
				).WithQuery(
					request.GetExperimentRequest{
						ID: fmt.Sprintf("%d", *experiment.ID),
					},
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute),
				),
			)
			assert.Equal(s.T(), fmt.Sprintf("%d", *experiment.ID), resp.Experiment.ID)
			assert.Equal(s.T(), experiment.Name, resp.Experiment.Name)
			assert.Equal(s.T(), experiment.ArtifactLocation, resp.Experiment.ArtifactLocation)
			assert.Equal(s.T(), string(models.LifecycleStageActive), resp.Experiment.LifecycleStage)
		})
	}
}

func (s *NamespaceTestSuite) Test_Error() {
	tests := []struct {
		name      string
		error     *api.ErrorResponse
		namespace string
	}{
		{
			name:      "RequestNotExistingNamespace",
			error:     api.NewResourceDoesNotExistError("unable to find namespace with code: not-existing-namespace"),
			namespace: "not-existing-namespace",
		},
		{
			name:      "RequestNotExistingDefaultNamespaceExplicitly",
			error:     api.NewResourceDoesNotExistError("unable to find namespace with code: default"),
			namespace: "default",
		},
		{
			name:  "RequestNotExistingDefaultNamespaceImplicitly",
			error: api.NewResourceDoesNotExistError("unable to find namespace with code: default"),
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			resp := api.ErrorResponse{}
			require.Nil(
				s.T(),
				s.MlflowClient().WithNamespace(
					tt.namespace,
				).WithQuery(
					request.GetExperimentRequest{},
				).WithResponse(
					&resp,
				).DoRequest(
					fmt.Sprintf("%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute),
				),
			)
			assert.Equal(s.T(), tt.error.Error(), resp.Error())
			assert.Equal(s.T(), api.ErrorCodeResourceDoesNotExist, string(resp.ErrorCode))
		})
	}
}
