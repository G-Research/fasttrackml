package response

import "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

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

	for i, namespace := range namespaces {
		response[i] = Namespace{
			ID:          namespace.ID,
			Code:        namespace.Code,
			Description: namespace.Description,
		}
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
