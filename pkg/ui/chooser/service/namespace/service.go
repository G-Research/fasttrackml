package namespace

import (
	"context"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/ui/chooser/middleware"
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
func (s Service) ListNamespaces(ctx context.Context) ([]models.Namespace, bool, error) {
	namespaces, err := s.namespaceRepository.List(ctx)
	if err != nil {
		return nil, false, eris.Wrap(err, "error listing namespaces")
	}

	// filter namespaces based on current user permissions.
	switch {
	case s.config.Auth.IsAuthTypeUser():
		authToken, err := middleware.GetAuthTokenFromContext(ctx)
		if err != nil {
			return nil, false, err
		}

		// if user is not an admin user, then filter namespaces for current user,
		// otherwise just show the namespaces for current user.
		if !s.config.Auth.AuthParsedUserPermissions.HasAdminAccess(authToken) {
			roles, ok := s.config.Auth.AuthParsedUserPermissions.GetRolesByAuthToken(authToken)
			if !ok {
				return nil, false, eris.New("error validating user auth token")
			}
			return FilterNamespacesByUserRoles(roles, namespaces),
				s.config.Auth.AuthParsedUserPermissions.HasAdminAccess(authToken),
				nil
		}
	}

	return namespaces, true, nil
}
