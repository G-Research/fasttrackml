package namespace

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// FilterNamespacesByUserRoles filter namespaces by provided user roles.
func FilterNamespacesByUserRoles(
	roles map[string]struct{},
	namespaces []models.Namespace,
) []models.Namespace {
	var filteredPermissions []models.Namespace
	for _, namespace := range namespaces {
		if _, ok := roles[namespace.Code]; ok {
			filteredPermissions = append(filteredPermissions, namespace)
		}
	}
	return filteredPermissions
}
