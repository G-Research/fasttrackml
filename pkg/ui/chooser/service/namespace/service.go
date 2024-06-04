package namespace

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/config"
	"github.com/G-Research/fasttrackml/pkg/common/middleware"
)

// Service provides service layer to work with `namespace` business logic.
type Service struct {
	config              *config.Config
	namespaceRepository repositories.NamespaceRepositoryProvider
}

// NewService creates new Service instance.
func NewService(
	config *config.Config,
	namespaceRepository repositories.NamespaceRepositoryProvider,
) *Service {
	return &Service{
		config:              config,
		namespaceRepository: namespaceRepository,
	}
}

// ListNamespaces returns all namespaces.
func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, bool, error) {
	namespaces, err := s.namespaceRepository.List(ctx)
	if err != nil {
		return nil, false, eris.Wrap(err, "error getting namespaces")
	}

	switch {
	case s.config.Auth.IsAuthTypeUser():
		authToken, err := middleware.GetBasicAuthTokenFromContext(ctx)
		if err != nil {
			return nil, false, err
		}
		// if auth token is not admin auth token, then filter namespaces and show
		// only those which belong to the current user, otherwise just show everything.
		if !authToken.HasAdminAccess() {
			return FilterNamespacesByAuthTokenUserRoles(authToken.GetRoles(), namespaces), false, nil
		}
	case s.config.Auth.IsAuthTypeOIDC():
		user, err := middleware.GetOIDCUserFromContext(ctx)
		if err != nil {
			return nil, false, err
		}
		// if auth token is not admin auth token, then filter namespaces and show
		// only those which belong to the current user, otherwise just show everything.
		if !user.IsAdmin() {
			namespaces, err = s.namespaceRepository.GetByRoles(ctx, user.GetRoles())
			if err != nil {
				return nil, false, eris.Wrap(err, "error getting namespaces")
			}
			return namespaces, false, nil
		}
	}

	return namespaces, true, nil
}
