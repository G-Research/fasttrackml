package namespace

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
)

// Service provides service layer to work with `namespace` business logic.
type Service struct {
	config              *config.ServiceConfig
	namespaceRepository repositories.NamespaceRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	config *config.ServiceConfig,
	namespaceRepository repositories.NamespaceRepositoryProvider,
) *Service {
	return &Service{
		config:              config,
		namespaceRepository: namespaceRepository,
	}
}

// ListNamespaces returns all namespaces.
func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, error) {
	namespaces, err := s.namespaceRepository.List(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "error listing namespaces")
	}

	// filter namespaces based on current user permissions.
	switch {
	case s.config.Auth.IsAuthTypeUser():
	}

	return namespaces, nil
}
