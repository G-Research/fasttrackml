package namespace

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

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

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, testNamespaces, namespaces)
}

// TestService_GetExperiment_Error tests the unsuccessful call to GetNamespace.
func TestService_ListNamespaces_Error(t *testing.T) {
	// init repository mocks
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"List", context.TODO(),
	).Return(nil, errors.New("error listing namespaces"))

	// call service under testing.
	service := NewService(
		&namespaceRepository,
	)
	namespaces, err := service.ListNamespaces(context.TODO())

	// compare results.
	assert.NotNil(t, err)
	assert.Nil(t, namespaces)
}
