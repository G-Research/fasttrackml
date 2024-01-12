package namespace

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

func TestService_CreateNamespace_Ok(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Create",
		context.TODO(),
		mock.MatchedBy(func(ns *models.Namespace) bool {
			assert.Equal(t, "code", ns.Code)
			assert.Equal(t, "description", ns.Description)
			return true
		}),
	).Return(nil)
	namespaceRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(ns *models.Namespace) bool {
			assert.Equal(t, "code", ns.Code)
			assert.Equal(t, "description", ns.Description)
			return true
		}),
	).Return(nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"Create",
		context.TODO(),
		mock.MatchedBy(func(experiment *models.Experiment) bool {
			assert.Equal(t, models.DefaultExperimentName, experiment.Name)
			assert.Equal(t, models.LifecycleStageActive, experiment.LifecycleStage)
			assert.NotNil(t, experiment.CreationTime)
			assert.NotNil(t, experiment.LastUpdateTime)
			experiment.ID = common.GetPointer(int32(1))
			return true
		}),
	).Return(nil)
	experimentRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(experiment *models.Experiment) bool {
			assert.Equal(
				t,
				fmt.Sprintf("default_artifact_root/%d", *experiment.ID),
				experiment.ArtifactLocation,
			)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(&config.ServiceConfig{
		DefaultArtifactRoot: "default_artifact_root",
	}, &namespaceRepository, &experimentRepository)
	_, err := service.CreateNamespace(context.TODO(), "code", "description")

	// compare results.
	require.Nil(t, err)
}

func TestService_CreateNamespace_Error(t *testing.T) {
	err := errors.New("repository error")

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Create", context.TODO(), mock.Anything, mock.Anything,
	).Return(err)
	namespaceRepository.On(
		"Update", context.TODO(), mock.Anything, mock.Anything,
	).Return(nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}
	experimentRepository.On(
		"Update",
		context.TODO(),
		mock.Anything,
	).Return(nil)

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	_, err = service.CreateNamespace(context.TODO(), "code", "description")

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "error creating namespace: repository error", err.Error())
}

func TestService_GetNamespace_Ok(t *testing.T) {
	// initialise namespace.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(0),
	).Return(&ns, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	namespace, err := service.GetNamespace(context.TODO(), uint(0))

	// compare results.
	require.Nil(t, err)
	assert.Equal(t, &ns, namespace)
}

func TestService_GetNamespace_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(0),
	).Return(nil, errors.New("something is wrong"))

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	namespace, err := service.GetNamespace(context.TODO(), uint(0))

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "error getting namespace by id: something is wrong", err.Error())
	assert.Nil(t, namespace)
}

func TestService_ListNamespace_Ok(t *testing.T) {
	// initialise namespaces.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}
	testNamespaces := []models.Namespace{ns}

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"List", context.TODO(),
	).Return(testNamespaces, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	require.Nil(t, err)
	assert.Equal(t, testNamespaces, namespaces)
}

func TestService_ListNamespaces_Error(t *testing.T) {
	// init repository mocks
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"List", context.TODO(),
	).Return(nil, errors.New("error listing namespaces"))

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "error listing namespaces: error listing namespaces", err.Error())
	assert.Nil(t, namespaces)
}

func TestService_DeleteNamespace_Ok(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}
	namespaceRepository.On(
		"Delete", context.TODO(), &ns,
	).Return(nil).On(
		"GetByID", context.TODO(), uint(0),
	).Return(&ns, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	err := service.DeleteNamespace(context.TODO(), uint(0))

	// compare results.
	require.Nil(t, err)
}

func TestService_DeleteNamespace_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(0),
	).Return(nil, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	err := service.DeleteNamespace(context.TODO(), uint(0))

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "namespace not found by id: 0", err.Error())
}

func TestService_DeleteDefaultNamespace_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(0),
	).Return(&models.Namespace{
		ID:   1,
		Code: models.DefaultNamespaceCode,
	}, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	err := service.DeleteNamespace(context.TODO(), uint(0))

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "unable to delete default namespace", err.Error())
}

func TestService_UpdateNamespace_Ok(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	ns := models.Namespace{
		ID: 1,
	}
	namespaceRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(ns *models.Namespace) bool {
			assert.Equal(t, uint(1), ns.ID)
			assert.Equal(t, "code", ns.Code)
			assert.Equal(t, "description", ns.Description)
			return true
		}),
	).Return(nil).On(
		"GetByID", context.TODO(), uint(1),
	).Return(&ns, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	_, err := service.UpdateNamespace(context.TODO(), uint(1), "code", "description")

	// compare results.
	require.Nil(t, err)
}

func TestService_UpdateNamespace_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(1),
	).Return(nil, nil)

	experimentRepository := repositories.MockExperimentRepositoryProvider{}

	// call service under testing.
	service := NewService(&config.ServiceConfig{}, &namespaceRepository, &experimentRepository)
	_, err := service.UpdateNamespace(context.TODO(), uint(1), "code", "description")

	// compare results.
	assert.NotNil(t, err)
	assert.Equal(t, "namespace not found by id: 1", err.Error())
}
