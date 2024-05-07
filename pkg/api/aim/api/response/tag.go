package response

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// TagResponse represents a run tag.
type TagResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	RunCount    int       `json:"run_count"`
	Archived    bool      `json:"archived"`
}

// GetTagsResponse represents a list of run tags.
type GetTagsResponse []TagResponse

// NewGetTagsResponse will convert the []model.Tag to GetTagsResponse
// TODO this is not really implemented
func NewGetTagsResponse(tags []models.Tag) GetTagsResponse {
	tagResponses := make(GetTagsResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = TagResponse{
			ID:   uuid.New(),
			Name: tag.Key,
		}
	}
	return tagResponses
}
