package namespace

import (
	"context"
	"fmt"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// Service provides service layer to work with `namespace` business logic.
type Service struct {
	namespaceRepository  repositories.NamespaceRepository
}

// NewService creates new Service instance.
func NewService(
	namespaceRepository repositories.NamespaceRepository,
) *Service {
	return &Service{
		namespaceRepository:  namespaceRepository,
	}
}

// ListNamespaces returns all namespaces.
func (s Service) ListNamespaces(ctx context.Context) ([]*models.Namespace, error) {
	return s.namespaceRepository.List(ctx)
}

// GetNamespace returns one namespace by ID.
func (s Service) GetNamespace(ctx context.Context, id uint) (*models.Namespace, error) {
	return s.namespaceRepository.GetByID(ctx, id)
}

// CreateNamespace creates a new namespace with default experiment.
func (s Service) CreateNamespace(ctx context.Context, code, description string) (models.Namespace, error) {
	exp := &models.Experiment{
		Name: fmt.Sprintf("%s-exp", code),
		LifecycleStage: models.LifecycleStageActive,
	}
	// placeholder DefaultExperimentID
	initialDefaultExpID := int32(0)
	namespace := &models.Namespace{
		Code:                code,
		Description:         description,
		Experiments:         []models.Experiment{*exp},
		DefaultExperimentID: &initialDefaultExpID,
	}
	if err := s.namespaceRepository.Create(ctx, namespace); err != nil {
		return *namespace, err
	}
	// update with true DefaultExperimentID
	namespace.DefaultExperimentID = namespace.Experiments[0].ID
	err := s.namespaceRepository.Update(ctx, namespace)
	return *namespace, err
}

// UpdateNamespace updates the code and description fields.
func (s Service) UpdateNamespace(ctx context.Context, id uint, code, description string) (models.Namespace, error) {
	namespace := &models.Namespace{
		ID:          id,
		Code:        code,
		Description: description,
	}
	err := s.namespaceRepository.Update(ctx, namespace)
	return *namespace, err
}

// DeleteNamespace will delete the namespace.
func (s Service) DeleteNamespace(ctx context.Context, id uint) error {
	return s.namespaceRepository.Delete(ctx, id)
}
