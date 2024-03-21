package namespace

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config/auth"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// FilterNamespacesByUserPermissions filter namespaces by provided user permissions.
func FilterNamespacesByUserPermissions(
	namespaces []models.Namespace,
	permissions auth.UserPermissions,
) []models.Namespace {
	var filteredPermissions []models.Namespace

	return filteredPermissions
}
