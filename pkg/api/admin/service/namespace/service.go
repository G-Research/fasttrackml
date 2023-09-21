package namespace

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/admin/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
		return nil, err
	}
	return namespaces, nil
}
