package events

import "github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"

// NamespaceEventAction represents Event action.
type NamespaceEventAction string

// Supported event actions.
const (
	NamespaceEventActionFetched = "fetched"
	NamespaceEventActionCreated = "created"
	NamespaceEventActionDeleted = "deleted"
	NamespaceEventActionUpdated = "updated"
)

// NamespaceEvent represents database event.
type NamespaceEvent struct {
	Action    NamespaceEventAction `json:"action"`
	Namespace models.Namespace     `json:"namespace"`
}
