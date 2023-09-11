package namespace

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/api/admin/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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

func (s Service) ListNamespaces(ctx context.Context) ([]*models.Namespace, error) {
	return s.namespaceRepository.List(ctx)
}

func (s Service) GetNamespace(ctx context.Context, id uint) (*models.Namespace, error) {
	return s.namespaceRepository.GetByID(ctx, id)
}

func (s Service) CreateNamespace(ctx context.Context, code, description string) (models.Namespace, error) {
	namespace := &models.Namespace{
		Code:        code,
		Description: description,
	}
	err :=  s.namespaceRepository.Create(ctx, namespace)
	return *namespace, err
}

func (s Service) UpdateNamespace(ctx context.Context, id uint, code, description string) (models.Namespace, error) {
	namespace := &models.Namespace{
		ID: id,
		Code:        code,
		Description: description,
	}
	err := s.namespaceRepository.Update(ctx, namespace)
	return *namespace, err
}

func (s Service) DeleteNamespace(ctx context.Context, id uint) error {
	return s.namespaceRepository.Delete(ctx, id)
}
