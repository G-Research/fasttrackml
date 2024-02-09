package helpers

import (
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/ui/admin/request"
)

// CheckNamespaces checks if the requested namespaces are present in the expected namespaces.
func CheckNamespaces(expectedNamespaces []models.Namespace, requestedNamespaces []request.Namespace) bool {
	for _, testNamespace := range requestedNamespaces {
		found := false
		for _, namespace := range expectedNamespaces {
			if namespace.Code == testNamespace.Code && namespace.Description == testNamespace.Description {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
