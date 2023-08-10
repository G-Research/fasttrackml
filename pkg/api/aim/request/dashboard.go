package request

import (
	"github.com/google/uuid"
)

// CreateDashboard represents the data to create a Dashboard
type CreateDashboard struct {
	AppID       uuid.UUID `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// UpdateDashboard represents the data to update a Dashboard
type UpdateDashboard struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
