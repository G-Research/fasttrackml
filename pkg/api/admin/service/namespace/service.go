package namespace

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// Service provides service layer to work with `namespace` business logic.
type Service struct {
	namespaceRepository repositories.NamespaceRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	namespaceRepository repositories.NamespaceRepositoryProvider,
) *Service {
	return &Service{
		namespaceRepository: namespaceRepository,
	}
}

// ListNamespaces returns all namespaces.
func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, error) {
	namespaces, err := s.namespaceRepository.List(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error listing namespaces")
	}
	return namespaces, nil
}

// GetNamespace returns one namespace by ID.
func (s Service) GetNamespace(ctx context.Context, id uint) (*models.Namespace, error) {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return nil, eris.Wrap(err, "error getting namespace by id")
	}
	return namespace, nil
}

// CreateNamespace creates a new namespace and default experiment.
func (s Service) CreateNamespace(ctx context.Context, code, description string) (*models.Namespace, error) {
	if err := ValidateNamespace(code); err != nil {
		return nil, eris.Wrap(err, "error validating namespace")
	}
	exp := &models.Experiment{
		Name:           fmt.Sprintf("%s-exp", code),
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
		return nil, eris.Wrap(err, "error creating namespace")
	}
	// update with true DefaultExperimentID
	namespace.DefaultExperimentID = namespace.Experiments[0].ID
	err := s.namespaceRepository.Update(ctx, namespace)
	if err != nil {
		return nil, eris.Wrap(err, "error setting namespace default experiment id during create")
	}
	return namespace, nil
}

// UpdateNamespace updates the code and description fields.
func (s Service) UpdateNamespace(ctx context.Context, id uint, code, description string) (*models.Namespace, error) {
	if err := ValidateNamespace(code); err != nil {
		return nil, eris.Wrap(err, "error validating namespace")
	}
	namespace := &models.Namespace{
		ID:          id,
		Code:        code,
		Description: description,
	}
	err := s.namespaceRepository.Update(ctx, namespace)
	if err != nil {
		return nil, eris.Wrap(err, "error updating namespace")
	}
	return namespace, nil
}

// DeleteNamespace deletes the namespace.
func (s Service) DeleteNamespace(ctx context.Context, id uint) error {
	err := s.namespaceRepository.Delete(ctx, id)
	if err != nil {
		return eris.Wrap(err, "error deleting namespace")
	}
	return nil
}
