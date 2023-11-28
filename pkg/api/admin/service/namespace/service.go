package namespace

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
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
		Name:           "Default",
		LifecycleStage: models.LifecycleStageActive,
	}
	namespace := &models.Namespace{
		Code:                code,
		Description:         description,
		Experiments:         []models.Experiment{*exp},
		DefaultExperimentID: common.GetPointer(int32(0)),
	}
	if err := s.namespaceRepository.Create(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error creating namespace")
	}
	// update Namespace with correct DefaultExperimentID now that it is known
	namespace.DefaultExperimentID = namespace.Experiments[0].ID
	err := s.namespaceRepository.Update(ctx, namespace)
	if err != nil {
		return nil, eris.Wrap(err, "error setting namespace default experiment id during create")
	}
	return namespace, nil
}

// UpdateNamespace updates the code and description fields.
func (s Service) UpdateNamespace(ctx context.Context, id uint, code, description string) (*models.Namespace, error) {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return nil, eris.Wrapf(err, "error finding namespace by id: %d", id)
	}
	if namespace == nil {
		return nil, eris.Errorf("namespace not found by id: %d", id)
	}
	if err := ValidateNamespace(code); err != nil {
		return nil, eris.Wrap(err, "error validating namespace code")
	}
	namespace.Code = code
	namespace.Description = description

	if err := s.namespaceRepository.Update(ctx, namespace); err != nil {
		return nil, eris.Wrap(err, "error updating namespace")
	}
	return namespace, nil
}

// DeleteNamespace deletes the namespace.
func (s Service) DeleteNamespace(ctx context.Context, id uint) error {
	namespace, err := s.namespaceRepository.GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "error finding namespace by id: %d", id)
	}
	if namespace == nil {
		return eris.Errorf("namespace not found by id: %d", id)
	}
	if err := s.namespaceRepository.Delete(ctx, namespace); err != nil {
		return eris.Wrap(err, "error deleting namespace")
	}
	return nil
}
