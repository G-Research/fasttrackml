package request

import (
	"github.com/google/uuid"
)

// CreateTagRequest is a request object for `POST /aim/tags` endpoint.
type CreateTagRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	IsArchived  bool      `json:"archived"`
}

// GetTagRequest is a request object for `GET /aim/tags/:id` endpoint.
type GetTagRequest struct {
	ID uuid.UUID `params:"id"`
}

// UpdateTagRequest is a request object for `PUT /aim/tags/:id` endpoint.
type UpdateTagRequest = CreateTagRequest

// DeleteTagRequest is a request object for `DELETE /aim/tags/:id` endpoint.
type DeleteTagRequest = GetTagRequest

// GetRunsTaggedRequest is a request object for `GET /aim/tags/:id/runs` endpoint.
type GetRunsTaggedRequest = GetTagRequest
