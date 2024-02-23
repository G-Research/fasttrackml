package response

import (
	"github.com/google/uuid"
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
func NewGetTagsResponse(tagsRunCount map[string]int) GetTagsResponse {
	tagResponses := make(GetTagsResponse, len(tagsRunCount))
	idx := 0
	for tag, runCount := range tagsRunCount {
		tagResponses[idx] = TagResponse{
			ID:          uuid.New(),
			Name:        tag,
			Color:       "#18AB6D",
			Description: "",
			Archived:    false,
			RunCount:    runCount,
		}
		idx++
	}
	return tagResponses
}
