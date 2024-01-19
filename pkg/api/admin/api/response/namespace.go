package response

import "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

// Namespace is the response struct for the GetCurrentNamespace endpoint.
type Namespace struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// ListNamespaces is the response struct for the ListNamespaces endpoint (slice of Namespace).
type ListNamespaces []Namespace

// NewListNamespacesResponse creates new instance of ListNamespaces.
func NewListNamespacesResponse(
	namespaces []models.Namespace,
) *ListNamespaces {
	response := ListNamespaces(make([]Namespace, len(namespaces)))

	for i := range namespaces {
		response[i] = *NewGetCurrentNamespaceResponse(&namespaces[i])
	}

	return &response
}

// NewGetCurrentNamespaceResponse creates new instance of Namespace.
func NewGetCurrentNamespaceResponse(
	namespace *models.Namespace,
) *Namespace {
	return &Namespace{
		ID:          namespace.ID,
		Code:        namespace.Code,
		Description: namespace.Description,
	}
}
