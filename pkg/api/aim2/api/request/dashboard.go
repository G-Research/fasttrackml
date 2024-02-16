package request

import (
	"github.com/google/uuid"
)

// CreateDashboardRequest is a request object for `POST /aim/dashboards` endpoint.
type CreateDashboardRequest struct {
	AppID       uuid.UUID `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// GetDashboardRequest is a request object for `GET /aim/dashboards/:id` endpoint.
type GetDashboardRequest struct {
	ID uuid.UUID `params:"id"`
}

// UpdateDashboardRequest is a request object for `PUT /aim/dashboards` endpoint.
type UpdateDashboardRequest struct {
	ID          uuid.UUID `params:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// DeleteDashboardRequest is a request object for `DELETE /aim/dashboards/:id` endpoint.
type DeleteDashboardRequest = GetDashboardRequest
