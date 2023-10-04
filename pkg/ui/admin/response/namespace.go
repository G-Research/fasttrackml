package response

import (
	"time"
)

// Namespace represents the data for viewing/editing a Namespace.
type Namespace struct {
	ID          uint       `json:"id"`
	Code        string     `json:"code"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}
