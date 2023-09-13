package namespace

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"

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

func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, error) {
	return nil, nil
}
