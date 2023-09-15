package namespace

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// TestService_CreateNamespace_Ok tests successful calls to the repository.
func TestService_CreateNamespace_Ok(t *testing.T) {

	defaultExpID := int32(0)
	expID := int32(21)
	experiment := models.Experiment{
		ID:             &expID,
		Name:           fmt.Sprintf("%s-exp", "code"),
		LifecycleStage: models.LifecycleStageActive,
	}

	ns := models.Namespace{
		Code:                "code",
		Description:         "description",
		DefaultExperimentID: &defaultExpID,
		Experiments:         []models.Experiment{experiment},
	}
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
	).Return(&ns, nil).On(
		"Update",
		context.TODO(),
		mock.MatchedBy(func(ns *models.Namespace) bool {
			assert.Equal(t, "code", ns.Code)
			assert.Equal(t, "description", ns.Description)
			return true
		}),
	).Return(&ns, nil)

	// call service under testing.
	service := NewService(&namespaceRepository)
	namespace, err := service.CreateNamespace(context.TODO(), "code", "description")

	// compare results.
	assert.Nil(t, err)
	assert.Equal(t, ns, namespace)
}

// TestService_CreateNamespace_Error tests unsuccessful calls to the repository.
func TestService_CreateNamespace_Error(t *testing.T) {

	ns := models.Namespace{}
	err := errors.New("repository error")

	// init repository mocks.
	namespaceRepository := repositories.MockNamespaceRepositoryProvider{}
	namespaceRepository.On(
		"Create", context.TODO(), mock.Anything, mock.Anything,
	).Return(nil, err).On(
		"Update", context.TODO(), mock.Anything, mock.Anything,
	).Return(ns, nil)

	// call service under testing.
	service := NewService(&namespaceRepository)
	_, err = service.CreateNamespace(context.TODO(), "code", "description")

	// compare results.
	assert.NotNil(t, err)
}
