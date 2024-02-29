package response

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
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

// NewGetTagsResponse will convert the []model.TagData to GetTagsResponse
func NewGetTagsResponse(tagDatas []models.TagData) GetTagsResponse {
	tagResponses := make(GetTagsResponse, len(tagDatas))
	idx := 0
	for _, tagData := range tagDatas {
		tagResponses[idx] = TagResponse{
			ID:          tagData.ID,
			Name:        tagData.Key,
			Color:       tagData.Color,
			Description: tagData.Description,
			Archived:    tagData.IsArchived,
			RunCount:    len(tagData.Runs),
		}
		idx++
	}
	return tagResponses
}
