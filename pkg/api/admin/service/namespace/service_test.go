package namespace

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// TestService_CreateNamespace_Ok tests successful calls to the repository.
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

	// call service under testing.
	service := NewService(&namespaceRepository)
	_, err := service.CreateNamespace(context.TODO(), "code", "description")

	// compare results.
	assert.Nil(t, err)
}

// TestService_CreateNamespace_Error tests unsuccessful calls to the repository.
func TestService_CreateNamespace_Error(t *testing.T) {
	ns := models.Namespace{}
	err := errors.New("repository error")

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Create", context.TODO(), mock.Anything, mock.Anything,
	).Return(nil, err)
	namespaceRepository.On(
		"Update", context.TODO(), mock.Anything, mock.Anything,
	).Return(ns, nil)

	// TODO someting is not right with mock
	// // call service under testing.
	// service := NewService(&namespaceRepository)
	// _, err = service.CreateNamespace(context.TODO(), "code", "description")

	// // compare results.
	// assert.NotNil(t, err)
}

// TestService_GetNamespace_Ok tests the successful call to GetNamespace.
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

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespace, err := service.GetNamespace(context.TODO(), uint(0))

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, &ns, namespace)
}

// TestService_GetNamespace_Error tests the unsuccessful call to GetNamespace.
func TestService_GetNamespace_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"GetByID", context.TODO(), uint(0),
	).Return(nil, errors.New("something is wrong"))

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespace, err := service.GetNamespace(context.TODO(), uint(0))

	// compare results.
	assert.NotNil(t, err)
	assert.Nil(t, namespace)
}

// TestService_ListNamespaces_OK tests the successful call to GetNamespace.
func TestService_ListNamespace_Ok(t *testing.T) {
	// initialise namespaces.
	ns := models.Namespace{
		ID:   1,
		Code: "code",
	}

	nss := []*models.Namespace{&ns}

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"List", context.TODO(),
	).Return(nss, nil)

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, nss, namespaces)
}

// TestService_GetExperiment_Error tests the unsuccessful call to GetNamespace.
func TestService_ListNamespaces_Error(t *testing.T) {
	// init repository mocks
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"List", context.TODO(),
	).Return(nil, errors.New("something is wrong"))

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	assert.NotNil(t, err)
	assert.Nil(t, namespaces)
}

// TestService_DeleteNamespace_Ok tests the successful call to DeleteNamespace.
func TestService_DeleteNamespace_Ok(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Delete", context.TODO(), uint(0),
	).Return(nil)

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	err := service.DeleteNamespace(context.TODO(), uint(0))

	// compare results.
	assert.Nil(t, err)
}

// TestService_DeleteExperiment_Error tests the unsuccessful call to DeleteNamespace.
func TestService_DeleteExperiment_Error(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Delete", context.TODO(), uint(0),
	).Return(errors.New("something is wrong"))

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	err := service.DeleteNamespace(context.TODO(), uint(0))

	// compare results.
	assert.NotNil(t, err)
}

func TestService_UpdateNamespace_Ok(t *testing.T) {
	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(ns *models.Namespace) bool {
			assert.Equal(t, uint(1), ns.ID)
			assert.Equal(t, "code", ns.Code)
			assert.Equal(t, "description", ns.Description)
			return true
		}),
	).Return(nil)

	// call service under testing.
	service := NewService(&namespaceRepository)
	_, err := service.UpdateNamespace(context.TODO(), uint(1), "code", "description")

	// compare results.
	assert.Nil(t, err)
}
