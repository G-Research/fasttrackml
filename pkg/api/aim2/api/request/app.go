package request

import "github.com/google/uuid"

// AppState represents key/value state data
type AppState map[string]any

// CreateAppRequest is a request object for `POST /aim/apps` endpoint.
type CreateAppRequest struct {
	Type  string   `json:"type"`
	State AppState `json:"state"`
}

// GetAppRequest is a request object for `GET /aim/apps/:id` endpoint.
type GetAppRequest struct {
	ID uuid.UUID `params:"id"`
}

// DeleteAppRequest is a request object for `DELETE /aim/apps/:id` endpoint.
type DeleteAppRequest struct {
	ID uuid.UUID `params:"id"`
}

// UpdateAppRequest is a request object for `PUT /aim/apps/:id` endpoint.
type UpdateAppRequest struct {
	ID    uuid.UUID `params:"id"`
	Type  string    `json:"type"`
	State AppState  `json:"state"`
}
